package tcp

import (
	"crypto/tls"
	"log"
	"net"
	"time"

	"github.com/millken/tcpwder/balance"
	"github.com/millken/tcpwder/config"
	"github.com/millken/tcpwder/core"
	"github.com/millken/tcpwder/server/scheduler"
	"github.com/millken/tcpwder/server/upstream"
	"github.com/millken/tcpwder/tls/sni"
	"github.com/millken/tcpwder/utils"
)

type Server struct {

	/* Server friendly name */
	name string

	/* Listener */
	listener net.Listener

	/* Configuration */
	cfg config.Server

	/*scheduler deals with upstream */
	scheduler scheduler.Scheduler

	/* Current clients connection */
	clients map[string]net.Conn

	/* Channel for new connections */
	connect chan (*core.TcpContext)

	/* Channel for dropping connections or connectons to drop */
	disconnect chan (net.Conn)

	/* Stop channel */
	stop chan bool
}

/**
 * Creates new server instance
 */
func New(name string, cfg config.Server) (*Server, error) {

	// Create server
	server := &Server{
		name:       name,
		cfg:        cfg,
		stop:       make(chan bool),
		disconnect: make(chan net.Conn),
		connect:    make(chan *core.TcpContext),
		clients:    make(map[string]net.Conn),
		scheduler: scheduler.Scheduler{
			Balancer: balance.New(cfg.Sni, cfg.Balance),
			Upstream: upstream.New(cfg.Upstream),
		},
	}

	log.Printf("Creating '%s': %s %s", name, cfg.Bind, cfg.Balance)

	return server, nil
}

/**
 * Returns current server configuration
 */
func (this *Server) Cfg() config.Server {
	return this.cfg
}

/**
 * Start server
 */
func (this *Server) Start() error {

	go func() {

		for {
			select {
			case client := <-this.disconnect:
				this.HandleClientDisconnect(client)

			case ctx := <-this.connect:
				this.HandleClientConnect(ctx)

			case <-this.stop:
				if this.listener != nil {
					this.listener.Close()
					for _, conn := range this.clients {
						conn.Close()
					}
				}
				this.clients = make(map[string]net.Conn)
				return
			}
		}
	}()

	// Start scheduler
	this.scheduler.Start()

	// Start listening
	if err := this.Listen(); err != nil {
		this.Stop()
		return err
	}

	return nil
}

/**
 * Handle client disconnection
 */
func (this *Server) HandleClientDisconnect(client net.Conn) {
	client.Close()
	delete(this.clients, client.RemoteAddr().String())
}

/**
 * Handle new client connection
 */
func (this *Server) HandleClientConnect(ctx *core.TcpContext) {
	client := ctx.Conn

	if *this.cfg.MaxConnections != 0 && len(this.clients) >= *this.cfg.MaxConnections {
		log.Printf("[WARN] Too many connections to %s", this.cfg.Bind)
		client.Close()
		return
	}

	this.clients[client.RemoteAddr().String()] = client
	go func() {
		this.handle(ctx)
		this.disconnect <- client
	}()
}

/**
 * Stop, dropping all connections
 */
func (this *Server) Stop() {

	log.Printf("Stopping %s", this.name)

	this.stop <- true
}

/**
 * Listen on specified port for a connections
 */
func (this *Server) Listen() (err error) {

	// create tcp listener
	this.listener, err = net.Listen("tcp", this.cfg.Bind)

	var tlsConfig *tls.Config
	sniEnabled := this.cfg.Sni != nil

	if this.cfg.Protocol == "tls" {

		// Create tls listener
		var crt tls.Certificate
		if crt, err = tls.LoadX509KeyPair(this.cfg.Tls.CertPath, this.cfg.Tls.KeyPath); err != nil {
			log.Printf("[ERROR] %s", err)
			return err
		}

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{crt},
		}

	}

	if err != nil {
		log.Printf("[ERROR] Error starting %s server: %s", this.cfg.Protocol, err)
		return err
	}

	go func() {
		for {
			conn, err := this.listener.Accept()

			if err != nil {
				log.Printf("[ERROR] %s", err)
				return
			}

			go this.wrap(conn, sniEnabled, tlsConfig)
		}
	}()

	return nil
}

func (this *Server) wrap(conn net.Conn, sniEnabled bool, tlsConfig *tls.Config) {

	var hostname string
	var err error

	if sniEnabled {
		var sniConn net.Conn
		sniConn, hostname, err = sni.Sniff(conn, utils.ParseDurationOrDefault(this.cfg.Sni.ReadTimeout, time.Second*2))

		if err != nil {
			log.Printf("[ERROR] Failed to get / parse ClientHello for sni: %s", err)
			conn.Close()
			return
		}

		conn = sniConn
	}

	if tlsConfig != nil {
		conn = tls.Server(conn, tlsConfig)
	}

	this.connect <- &core.TcpContext{
		hostname,
		conn,
	}

}

/**
 * Handle incoming connection and prox it to backend
 */
func (this *Server) handle(ctx *core.TcpContext) {
	clientConn := ctx.Conn

	log.Printf("[DEBUG] Accepted %s -> %s", clientConn.RemoteAddr(), this.listener.Addr())

	/* Find out backend for proxying */
	var err error
	backend, err := this.scheduler.TakeBackend(ctx)
	if err != nil {
		log.Printf("[ERROR] %s, Closing connection ", err, clientConn.RemoteAddr())
		return
	}
	log.Printf("[DEBUG] backend %+v", backend)

	/* Connect to backend */
	var backendConn net.Conn

	backendConn, err = net.DialTimeout("tcp", backend.Address(), utils.ParseDurationOrDefault(*this.cfg.BackendConnectionTimeout, 0))

	if err != nil {
		this.scheduler.IncrementRefused(*backend)
		log.Printf("[ERROR] %s", err)
		return
	}
	this.scheduler.IncrementConnection(*backend)
	defer this.scheduler.DecrementConnection(*backend)

	/* Stat proxying */
	log.Printf("[DEBUG] Begin %s%s%s%s%s", clientConn.RemoteAddr(), " -> ", this.listener.Addr(), " -> ", backendConn.RemoteAddr())
	cs := proxy(clientConn, backendConn, utils.ParseDurationOrDefault(*this.cfg.BackendIdleTimeout, 0))
	bs := proxy(backendConn, clientConn, utils.ParseDurationOrDefault(*this.cfg.ClientIdleTimeout, 0))

	isTx, isRx := true, true
	for isTx || isRx {
		select {
		case s, ok := <-cs:
			isRx = ok
			this.scheduler.IncrementRx(*backend, s.CountWrite)
		case s, ok := <-bs:
			isTx = ok
			this.scheduler.IncrementTx(*backend, s.CountWrite)
		}
	}
	log.Printf("[DEBUG] End %s -> %s", clientConn.RemoteAddr(), this.listener.Addr())
}
