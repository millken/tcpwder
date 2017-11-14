package filter

import (
	"fmt"
	"log"
	"net"

	"github.com/millken/tcpwder/config"
	"github.com/millken/tcpwder/core"
	"github.com/millken/tcpwder/utils"
)

type LimitChinaAccessFilter struct {
	lca []config.LimitChinaAccess
}

func (this *LimitChinaAccessFilter) Init(cfg config.Server) bool {
	if len(cfg.LimitChinaAccess) != 0 {
		this.lca = cfg.LimitChinaAccess
		return true
	}
	return false
}

func (this *LimitChinaAccessFilter) Connect(client net.Conn) error {
	host, _, _ := net.SplitHostPort(client.RemoteAddr().String())
	ip := net.ParseIP(host)
	cn, err := utils.FindCN(ip)
	for _, r := range this.lca {
		if err != nil {
			if r.Area == "" && r.Region == "" && r.Isp == "" && r.Access == "deny" {
				return fmt.Errorf("deny outsite of china")
			}
			log.Printf("%+v, cn=%+v", r, cn)
		}
	}
	return nil
}

func (this *LimitChinaAccessFilter) Disconnect(client net.Conn) {
}

func (this *LimitChinaAccessFilter) Read(client net.Conn, rwc core.ReadWriteCount) {
}

func (this *LimitChinaAccessFilter) Write(client net.Conn, rwc core.ReadWriteCount) {
}

func (this *LimitChinaAccessFilter) Stop() {
}

func init() {
	RegisterFilter("limit_china_access", func() interface{} {
		return new(LimitChinaAccessFilter)
	})
}
