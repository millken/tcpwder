package filter

import (
	"fmt"
	"net"

	"github.com/millken/tcpwder/config"
	"github.com/millken/tcpwder/core"
)

type LimitMaxConnectionFilter struct {
	maxConnections *int
	clients        map[string]bool
}

func (this *LimitMaxConnectionFilter) Init(cfg config.Server) bool {
	if cfg.MaxConnections != nil && *cfg.MaxConnections > 0 {
		this.maxConnections = cfg.MaxConnections
		this.clients = make(map[string]bool)
		return true
	}
	return false
}

func (this *LimitMaxConnectionFilter) Connect(client net.Conn) error {
	if len(this.clients) >= *this.maxConnections {
		return fmt.Errorf("Too many connections, more than %d", *this.maxConnections)
	}
	this.clients[client.RemoteAddr().String()] = true
	return nil
}

func (this *LimitMaxConnectionFilter) Disconnect(client net.Conn) {
	delete(this.clients, client.RemoteAddr().String())
}

func (this *LimitMaxConnectionFilter) Read(client net.Conn, rwc core.ReadWriteCount) {
}

func (this *LimitMaxConnectionFilter) Write(client net.Conn, rwc core.ReadWriteCount) {
}

func (this *LimitMaxConnectionFilter) Stop() {
}

func init() {
	RegisterFilter("limit_max_connection", func() interface{} {
		return new(LimitMaxConnectionFilter)
	})
}
