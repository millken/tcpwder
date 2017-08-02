package udp

import (
	"log"
	"net"

	"github.com/millken/tcpwder/config"
)

/**
 * UDP server implementation
 */
type Server struct {

	/* Server name */
	name string

	/* Server configuration */
	cfg config.Server

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
	err error
}

/**
 * Creates new UDP server
 */
func New(name string, cfg config.Server) (*Server, error) {

	server := &Server{
		name:        name,
		cfg:         cfg,
		getOrCreate: make(chan *sessionRequest),
		remove:      make(chan net.UDPAddr),
		stop:        make(chan bool),
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
	return nil
}

/**
 * Stop, dropping all connections
 */
func (this *Server) Stop() {
}
