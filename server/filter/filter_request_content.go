package filter

import (
	"bytes"
	"fmt"
	"log"
	"net"

	"github.com/millken/tcpwder/config"
	"github.com/millken/tcpwder/core"
)

type FilterRequestContent struct {
	frc           []config.FilterRequestContent
	defaultAccess bool
}

func (this *FilterRequestContent) Init(cfg config.Server) bool {
	if len(cfg.FilterRequestContent) != 0 {
		this.frc = cfg.FilterRequestContent
		this.defaultAccess = cfg.FilterRequestContentDefault == "deny"

		return true
	}
	return false
}

func (this *FilterRequestContent) Connect(client net.Conn) error {
	return nil
}

func (this *FilterRequestContent) Disconnect(client net.Conn) {
}

func (this *FilterRequestContent) Read(client net.Conn, rwc core.ReadWriteCount) {
}

func (this *FilterRequestContent) Write(client net.Conn, rwc core.ReadWriteCount) {
}

func (this *FilterRequestContent) Request(buf []byte) error {
	if len(bytes.TrimSpace(buf)) == 0 {
		return nil
	}
	for _, r := range this.frc {
		log.Printf("b=%v, r.access=%s", bytes.Contains(buf, []byte(r.Content)), r.Access)
		if bytes.Contains(buf, []byte(r.Content)) {
			switch r.Access {
			case "allow":
				return nil
			case "deny":
				return fmt.Errorf("deny content")
			}
		}
	}
	if this.defaultAccess {
		return fmt.Errorf("deny content default [%s]", buf)
	}
	return nil
}

func (this *FilterRequestContent) Stop() {
}

func init() {
	RegisterFilter("filter_request_content", func() interface{} {
		return new(FilterRequestContent)
	})
}
