package filter

import (
	"fmt"
	"net"

	"github.com/millken/tcpwder/config"
)

type LimitConnectionFilter struct {
	cfg     *config.FilterLimitConnectionConfig
	clients map[string]net.Conn
}

func (this *LimitConnectionFilter) Init(cfg config.Server) bool {
	if cfg.LimitConnection != nil {
		this.cfg = cfg.LimitConnection
		this.clients = make(map[string]net.Conn)
		return true
	}
	return false
}

func (this *LimitConnectionFilter) Connect(client net.Conn) error {
	this.clients[client.RemoteAddr().String()] = client
	if this.cfg.MaxConnections != 0 && len(this.clients) > this.cfg.MaxConnections {
		return fmt.Errorf("Too many connections")
	}
	return nil
}

func init() {
	RegisterFilter("limit_connection", func() interface{} {
		return new(LimitConnectionFilter)
	})
}
