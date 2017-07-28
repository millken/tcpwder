package core

import "github.com/millken/tcpwder/config"

/**
 * Balancer interface
 */
type Balancer interface {

	/**
	 * Elect backend based on Balancer implementation
	 */
	Elect(Context, []*Backend) (*Backend, error)
}

/**
 * Server interface
 */
type Server interface {

	/**
	 * Start server
	 */
	Start() error

	/**
	 * Stop server and wait until it stop
	 */
	Stop()

	/**
	 * Get server configuration
	 */
	Cfg() config.Server
}
