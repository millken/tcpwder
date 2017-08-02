package manager

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/millken/tcpwder/codec"
	"github.com/millken/tcpwder/config"
	"github.com/millken/tcpwder/core"
	"github.com/millken/tcpwder/server"
)

var servers = struct {
	sync.RWMutex
	m map[string]core.Server
}{
	m: make(map[string]core.Server),
}

/* default configuration for server */
var defaults config.ConnectionOptions

/* original cfg read from the file */
var originalCfg config.Config

/**
 * Initialize manager from the initial/default configuration
 */
func Initialize(cfg config.Config) {

	log.Println("[INFO] Initializing...")

	originalCfg = cfg

	// save defaults for futher reuse
	defaults = cfg.Defaults

	// Go through config and start servers for each server
	for name, serverCfg := range cfg.Servers {
		err := Create(name, serverCfg)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("[INFO] Initialized")
}

/**
 * Dumps current [servers] section to
 * the config file
 */
func DumpConfig(format string) (string, error) {

	originalCfg.Servers = map[string]config.Server{}

	servers.RLock()
	for name, server := range servers.m {
		originalCfg.Servers[name] = server.Cfg()
	}
	servers.RUnlock()

	var out *string = new(string)
	if err := codec.Encode(originalCfg, out, format); err != nil {
		return "", err
	}

	return *out, nil
}

/**
 * Returns map of servers with configurations
 */
func All() map[string]config.Server {
	result := map[string]config.Server{}

	servers.RLock()
	for name, server := range servers.m {
		result[name] = server.Cfg()
	}
	servers.RUnlock()

	return result
}

/**
 * Returns server configuration by name
 */
func Get(name string) interface{} {

	servers.RLock()
	server, ok := servers.m[name]
	servers.RUnlock()

	if !ok {
		return nil
	}

	return server.Cfg()
}

/**
 * Create new server and launch it
 */
func Create(name string, cfg config.Server) error {

	servers.Lock()
	defer servers.Unlock()

	if _, ok := servers.m[name]; ok {
		return errors.New("Server with this name already exists: " + name)
	}

	c, err := prepareConfig(name, cfg, defaults)
	if err != nil {
		return err
	}

	server, err := server.New(name, c)
	if err != nil {
		return err
	}

	if err = server.Start(); err != nil {
		return err
	}

	servers.m[name] = server

	return nil
}

/**
 * Delete server stopping all active connections
 */
func Delete(name string) error {

	servers.Lock()
	defer servers.Unlock()

	server, ok := servers.m[name]
	if !ok {
		return errors.New("Server not found")
	}

	server.Stop()
	delete(servers.m, name)

	return nil
}

/**
 * Returns stats for the server
 */
func Stats(name string) interface{} {

	servers.Lock()
	server := servers.m[name]
	servers.Unlock()

	return server
}

/**
 * Prepare config (merge default configuration, and try to validate)
 * TODO: make validation better
 */
func prepareConfig(name string, server config.Server, defaults config.ConnectionOptions) (config.Server, error) {

	/* ----- Prerequisites ----- */

	if server.Bind == "" {
		return config.Server{}, errors.New("No bind specified")
	}
	if server.Sni != nil {

		if server.Sni.ReadTimeout == "" {
			server.Sni.ReadTimeout = "2s"
		}

		if server.Sni.UnexpectedHostnameStrategy == "" {
			server.Sni.UnexpectedHostnameStrategy = "default"
		}

		switch server.Sni.UnexpectedHostnameStrategy {
		case
			"default",
			"reject",
			"any":
		default:
			return config.Server{}, errors.New("Not supported sni unexprected hostname strategy " + server.Sni.UnexpectedHostnameStrategy)
		}

		if server.Sni.HostnameMatchingStrategy == "" {
			server.Sni.HostnameMatchingStrategy = "exact"
		}

		switch server.Sni.HostnameMatchingStrategy {
		case
			"exact",
			"regexp":
		default:
			return config.Server{}, errors.New("Not supported sni matching " + server.Sni.HostnameMatchingStrategy)
		}

		if _, err := time.ParseDuration(server.Sni.ReadTimeout); err != nil {
			return config.Server{}, errors.New("timeout parsing error")
		}
	}

	/* ----- Connections params and overrides ----- */

	/* Protocol */
	switch server.Protocol {
	case "":
		server.Protocol = "tcp"
	case "tls":
		if server.Tls == nil {
			return config.Server{}, errors.New("Need tls section for tls protocol")
		}
		fallthrough
	case "tcp":
	case "udp":
	default:
		return config.Server{}, errors.New("Not supported protocol " + server.Protocol)
	}
	/* Balance */
	switch server.Balance {
	case
		"weight",
		"leastconn",
		"roundrobin",
		"leastbandwidth",
		"iphash":
	case "":
		server.Balance = "weight"
	default:
		return config.Server{}, errors.New("Not supported balance type " + server.Balance)
	}

	/* TODO: Still need to decide how to get rid of this */

	if defaults.MaxConnections == nil {
		defaults.MaxConnections = new(int)
	}
	if server.MaxConnections == nil {
		server.MaxConnections = new(int)
		*server.MaxConnections = *defaults.MaxConnections
	}

	if defaults.ClientIdleTimeout == nil {
		defaults.ClientIdleTimeout = new(string)
		*defaults.ClientIdleTimeout = "0"
	}
	if server.ClientIdleTimeout == nil {
		server.ClientIdleTimeout = new(string)
		*server.ClientIdleTimeout = *defaults.ClientIdleTimeout
	}

	if defaults.BackendIdleTimeout == nil {
		defaults.BackendIdleTimeout = new(string)
		*defaults.BackendIdleTimeout = "0"
	}
	if server.BackendIdleTimeout == nil {
		server.BackendIdleTimeout = new(string)
		*server.BackendIdleTimeout = *defaults.BackendIdleTimeout
	}

	if defaults.BackendConnectionTimeout == nil {
		defaults.BackendConnectionTimeout = new(string)
		*defaults.BackendConnectionTimeout = "0"
	}
	if server.BackendConnectionTimeout == nil {
		server.BackendConnectionTimeout = new(string)
		*server.BackendConnectionTimeout = *defaults.BackendConnectionTimeout
	}

	return server, nil
}
