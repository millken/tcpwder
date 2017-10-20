package tcp

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/millken/tcpwder/balance"
	"github.com/millken/tcpwder/config"
	"github.com/millken/tcpwder/core"
	"github.com/millken/tcpwder/firewall"
	"github.com/millken/tcpwder/server/filter"
	"github.com/millken/tcpwder/server/scheduler"
	"github.com/millken/tcpwder/server/upstream"
	"github.com/millken/tcpwder/stats"
	"github.com/millken/tcpwder/tls/sni"
	"github.com/millken/tcpwder/utils"
	tlsutil "github.com/millken/tcpwder/utils/tls"
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

	/* Stats handler */
	statsHandler *stats.Handler

	/* Channel for new connections */
	connect chan (*core.TcpContext)

	/* Channel for dropping connections or connectons to drop */
	disconnect chan (net.Conn)

	/* Stop channel */
	stop chan bool

	/* Tls config used to connect to backends */
	backendsTlsConfg *tls.Config

	/* filter */
	filter *filter.Filter
}

/**
 * Creates new server instance
 */
func New(name string, cfg config.Server) (*Server, error) {

	var err error = nil

	statsHandler := stats.NewHandler(name)

	// Create server
	server := &Server{
		name:         name,
		cfg:          cfg,
		stop:         make(chan bool),
		disconnect:   make(chan net.Conn),
		connect:      make(chan *core.TcpContext),
		clients:      make(map[string]net.Conn),
		statsHandler: statsHandler,
		scheduler: scheduler.Scheduler{
			Balancer:     balance.New(cfg.Sni, cfg.Balance),
			Upstream:     upstream.New(cfg.Upstream),
			StatsHandler: statsHandler,
		},
		filter: filter.New(cfg),
	}

	/* Add backend tls config if needed */
	if cfg.BackendsTls != nil {
		server.backendsTlsConfg, err = prepareBackendsTlsConfig(cfg)
		if err != nil {
			return nil, err
		}
	}

	log.Printf("[INFO] Creating '%s': %s %s", name, cfg.Bind, cfg.Balance)

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
				this.scheduler.Stop()
				this.statsHandler.Stop()
				this.filter.Stop()
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

	// Start stats handler
	this.statsHandler.Start()

	// Start scheduler
	this.scheduler.Start()

	// Start filter
	this.filter.Start()

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
	this.filter.HandleClientDisconnect(client)
	client.Close()
	delete(this.clients, client.RemoteAddr().String())
	this.statsHandler.Connections <- uint(len(this.clients))
}

/**
 * Handle new client connection
 */
func (this *Server) HandleClientConnect(ctx *core.TcpContext) {
	client := ctx.Conn
	host, _, _ := net.SplitHostPort(client.RemoteAddr().String())

	if !firewall.Allows(host) {
		log.Printf("[WARN] firewall deny: %s", host)
		client.Close()
		return
	}
	if err := this.filter.HandleClientConnect(client); err != nil {
		log.Printf("[WARN] handle client connect: %s", host)
		client.Close()
		return
	}
	/*
		if *this.cfg.MaxConnections != 0 && len(this.clients) >= *this.cfg.MaxConnections {
			log.Printf("[WARN] Too many connections to %s", this.cfg.Bind)
			client.Close()
			return
		}

		this.clients[client.RemoteAddr().String()] = client
	*/
	this.statsHandler.Connections <- uint(len(this.clients))
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

	if this.cfg.BackendsTls != nil {
		backendConn, err = tls.DialWithDialer(&net.Dialer{
			Timeout: utils.ParseDurationOrDefault(*this.cfg.BackendConnectionTimeout, 0),
		}, "tcp", backend.Address(), this.backendsTlsConfg)

	} else {
		backendConn, err = net.DialTimeout("tcp", backend.Address(), utils.ParseDurationOrDefault(*this.cfg.BackendConnectionTimeout, 0))
	}
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
	ticker := time.NewTicker(1 * time.Second)
	for isTx || isRx {
		select {
		case <-ticker.C:
			if !firewall.IsAllowClient(clientConn) {
				clientConn.Close()
				backendConn.Close()
				ticker.Stop()
				break
			}
		case s, ok := <-cs:
			isRx = ok
			this.scheduler.IncrementRx(*backend, s.CountWrite)
			this.filter.HandleClientRead(clientConn, s)
		case s, ok := <-bs:
			isTx = ok
			this.scheduler.IncrementTx(*backend, s.CountWrite)
			this.filter.HandleClientWrite(clientConn, s)
		}
	}
	log.Printf("[DEBUG] End %s -> %s", clientConn.RemoteAddr(), this.listener.Addr())
}

func prepareBackendsTlsConfig(cfg config.Server) (*tls.Config, error) {

	var err error

	result := &tls.Config{
		InsecureSkipVerify:       cfg.BackendsTls.IgnoreVerify,
		CipherSuites:             tlsutil.MapCiphers(cfg.BackendsTls.Ciphers),
		PreferServerCipherSuites: cfg.BackendsTls.PreferServerCiphers,
		MinVersion:               tlsutil.MapVersion(cfg.BackendsTls.MinVersion),
		MaxVersion:               tlsutil.MapVersion(cfg.BackendsTls.MaxVersion),
		SessionTicketsDisabled:   !cfg.BackendsTls.SessionTickets,
	}

	if cfg.BackendsTls.CertPath != nil && cfg.BackendsTls.KeyPath != nil {

		var crt tls.Certificate

		if crt, err = tls.LoadX509KeyPair(*cfg.BackendsTls.CertPath, *cfg.BackendsTls.KeyPath); err != nil {
			log.Printf("[ERROR] prepareBackendsTls : %s", err)
			return nil, err
		}

		result.Certificates = []tls.Certificate{crt}
	}

	if cfg.BackendsTls.RootCaCertPath != nil {

		var caCertPem []byte

		if caCertPem, err = ioutil.ReadFile(*cfg.BackendsTls.RootCaCertPath); err != nil {
			log.Printf("[ERROR] %s", err)
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(caCertPem); !ok {
			log.Printf("[ERROR] Unable to load root pem")
		}

		result.RootCAs = caCertPool

	}

	return result, nil

}
