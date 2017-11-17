package udp

import (
	"log"
	"net"

	"github.com/millken/tcpwder/balance"
	"github.com/millken/tcpwder/config"
	"github.com/millken/tcpwder/core"
	"github.com/millken/tcpwder/server/scheduler"
	"github.com/millken/tcpwder/server/upstream"
	"github.com/millken/tcpwder/stats"
	"github.com/millken/tcpwder/utils"
)

const UDP_PACKET_SIZE = 65507

/**
 * UDP server implementation
 */
type Server struct {

	/* Server name */
	name string

	/* Server configuration */
	cfg config.Server

	/* Scheduler */
	scheduler scheduler.Scheduler

	/* Stats handler */
	statsHandler *stats.Handler

	/* Server connection */
	serverConn *net.UDPConn

	/* Flag indicating that server is stopped */
	stopped bool

	/* ----- channels ----- */
	getOrCreate chan *sessionRequest
	remove      chan net.UDPAddr
	stop        chan bool
}

/**
 * Request to get session for clientAddr
 */
type sessionRequest struct {
	clientAddr net.UDPAddr
	response   chan sessionResponse
}

/**
 * Sessnion request response
 */
type sessionResponse struct {
	session *session
	err     error
}

/**
 * Creates new UDP server
 */
func New(name string, cfg config.Server) (*Server, error) {

	statsHandler := stats.NewHandler(name)
	server := &Server{
		name: name,
		cfg:  cfg,
		scheduler: scheduler.Scheduler{
			Balancer:     balance.New(nil, cfg.Balance),
			Upstream:     upstream.New(cfg.Upstream),
			StatsHandler: statsHandler,
		},
		statsHandler: statsHandler,
		getOrCreate:  make(chan *sessionRequest),
		remove:       make(chan net.UDPAddr),
		stop:         make(chan bool),
	}

	log.Printf("[INFO] Creating UDP server '%s': %s %s", name, cfg.Bind, cfg.Balance)
	return server, nil
}

/**
 * Returns current server configuration
 */
func (this *Server) Cfg() config.Server {
	return this.cfg
}

/**
 * Starts server
 */
func (this *Server) Start() error {
	this.scheduler.Start()
	this.statsHandler.Start()
	// Start listening
	if err := this.listen(); err != nil {
		this.Stop()
		log.Printf("[ERROR] starting UDP Listen %s", err)
		return err
	}

	go func() {
		sessions := make(map[string]*session)
		for {
			select {

			/* handle get session request */
			case sessionRequest := <-this.getOrCreate:
				session, ok := sessions[sessionRequest.clientAddr.String()]

				if ok {
					sessionRequest.response <- sessionResponse{
						session: session,
						err:     nil,
					}
					break
				}

				session, err := this.makeSession(sessionRequest.clientAddr)
				if err == nil {
					sessions[sessionRequest.clientAddr.String()] = session
				}

				sessionRequest.response <- sessionResponse{
					session: session,
					err:     err,
				}

			/* handle session remove */
			case clientAddr := <-this.remove:
				session, ok := sessions[clientAddr.String()]
				if !ok {
					break
				}
				session.stop()
				delete(sessions, clientAddr.String())

			/* handle server stop */
			case <-this.stop:
				for _, session := range sessions {
					session.stop()
				}
				return
			}
		}
	}()

	return nil
}

/**
 * Start accepting connections
 */
func (this *Server) listen() error {

	listenAddr, err := net.ResolveUDPAddr("udp", this.cfg.Bind)
	if err != nil {
		log.Printf("[ERROR] resolving server bind addr %s", err)
		return err
	}

	this.serverConn, err = net.ListenUDP("udp", listenAddr)

	if err != nil {
		log.Printf("[ERROR] starting UDP server: %s", err)
		return err
	}

	// Main proxy loop goroutine
	go func() {
		for {
			buf := make([]byte, UDP_PACKET_SIZE)
			n, clientAddr, err := this.serverConn.ReadFromUDP(buf)

			if err != nil {
				if this.stopped {
					return
				}
				log.Printf("[ERROR] ReadFromUDP: %s", err)
				continue
			}

			go func(buf []byte) {
				responseChan := make(chan sessionResponse, 1)

				this.getOrCreate <- &sessionRequest{
					clientAddr: *clientAddr,
					response:   responseChan,
				}

				response := <-responseChan

				if response.err != nil {
					log.Printf("[ERROR] creating session %s", response.err)
					return
				}

				err := response.session.send(buf)

				if err != nil {
					log.Printf("[ERROR] sending data to backend %s", err)
				}

			}(buf[0:n])
		}
	}()

	return nil
}

/**
 * Makes new session
 */
func (this *Server) makeSession(clientAddr net.UDPAddr) (*session, error) {

	log.Printf("[DEBUG] Accepted %s%s%s", clientAddr.String(), " -> ", this.serverConn.LocalAddr())

	var maxRequests uint64
	var maxResponses uint64

	if this.cfg.Udp != nil {
		maxRequests = this.cfg.Udp.MaxRequests
		maxResponses = this.cfg.Udp.MaxResponses
	}

	backend, err := this.scheduler.TakeBackend(&core.UdpContext{
		RemoteAddr: clientAddr,
	})

	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] backend %+v", backend)

	session := &session{
		clientIdleTimeout:  utils.ParseDurationOrDefault(*this.cfg.ClientIdleTimeout, 0),
		backendIdleTimeout: utils.ParseDurationOrDefault(*this.cfg.BackendIdleTimeout, 0),
		maxRequests:        maxRequests,
		maxResponses:       maxResponses,
		scheduler:          this.scheduler,
		notifyClosed: func() {
			this.remove <- clientAddr
		},
		serverConn: this.serverConn,
		clientAddr: clientAddr,
		backend:    backend,
	}

	err = session.start()
	if err != nil {
		session.stop()
		return nil, err
	}

	return session, nil
}

/**
 * Stop, dropping all connections
 */
func (this *Server) Stop() {
	log.Printf("[INFO] Stopping %s", this.name)

	this.stopped = true
	this.serverConn.Close()

	this.scheduler.Stop()
	this.scheduler.Stop()
	this.stop <- true
}
