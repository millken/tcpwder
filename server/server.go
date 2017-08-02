package server

import (
	"errors"

	"github.com/millken/tcpwder/config"
	"github.com/millken/tcpwder/core"
	"github.com/millken/tcpwder/server/tcp"
	"github.com/millken/tcpwder/server/udp"
)

func New(name string, cfg config.Server) (core.Server, error) {
	switch cfg.Protocol {
	case "tls", "tcp":
		return tcp.New(name, cfg)
	case "udp":
		return udp.New(name, cfg)
	default:
		return nil, errors.New("Can't create server for protocol " + cfg.Protocol)
	}
}
