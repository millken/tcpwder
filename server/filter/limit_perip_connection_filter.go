package filter

import (
	"fmt"
	"net"

	"github.com/millken/tcpwder/config"
)

type LimitPerIPConnectionFilter struct {
	connections *uint
	clients     map[string]uint
}

func (this *LimitPerIPConnectionFilter) Init(cfg config.Server) bool {
	if cfg.PerIpConnections != nil && *cfg.PerIpConnections > 0 {
		this.connections = cfg.PerIpConnections
		this.clients = make(map[string]uint)
		return true
	}
	return false
}

func (this *LimitPerIPConnectionFilter) Connect(client net.Conn) error {
	host, _, _ := net.SplitHostPort(client.RemoteAddr().String())
	if this.clients[host] >= *this.connections {
		return fmt.Errorf("per ip connections %s, limit %d", host, *this.connections)
	}
	this.clients[host] += 1
	return nil
}

func (this *LimitPerIPConnectionFilter) Disconnect(client net.Conn) {
	host, _, _ := net.SplitHostPort(client.RemoteAddr().String())
	if _, ok := this.clients[host]; ok {
		if this.clients[host] > 1 {
			this.clients[host] -= 1
		} else {
			delete(this.clients, host)
		}
	}
}

func (this *LimitPerIPConnectionFilter) Stop() {
}

func init() {
	RegisterFilter("limit_perip_connection", func() interface{} {
		return new(LimitPerIPConnectionFilter)
	})
}
