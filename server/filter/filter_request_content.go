package filter

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	"github.com/millken/tcpwder/config"
	"github.com/millken/tcpwder/core"
)

type FilterRequestContent struct {
	frc           []config.FilterRequestContent
	defaultAccess bool
	client        net.Conn
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
	this.client = client
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
	hit := false
	for _, r := range this.frc {
		switch r.Mode {
		case "hex":
			src := hex.EncodeToString(buf)
			if strings.Contains(src, r.Content) {
				hit = true
			}
		default:
			if bytes.Contains(buf, []byte(r.Content)) {
				hit = true
			}
		}
		if hit {
			switch r.Access {
			case "allow":
				return nil
			case "deny":
				return fmt.Errorf("deny by FilterRequestContent %s:%s", this.client.RemoteAddr().String(), r.Content)
			}
		}
	}
	if this.defaultAccess {
		return fmt.Errorf("deny content default")
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
