package filter

import (
	"fmt"
	"net"
	"strings"

	"github.com/millken/tcpwder/config"
	"github.com/millken/tcpwder/core"
	"github.com/millken/tcpwder/utils"
)

type LimitChinaAccessFilter struct {
	lca  []config.LimitChinaAccess
	lcad bool
}

func (this *LimitChinaAccessFilter) Init(cfg config.Server) bool {
	if len(cfg.LimitChinaAccess) != 0 {
		this.lca = cfg.LimitChinaAccess
		this.lcad = cfg.LimitChinaAccessDefault == "deny"
		return true
	}
	return false
}

func (this *LimitChinaAccessFilter) Connect(client net.Conn) error {
	host, _, _ := net.SplitHostPort(client.RemoteAddr().String())
	isPrivate, _ := utils.PrivateIP(host)
	if isPrivate {
		return nil
	}
	ip := net.ParseIP(host)
	cn, err := utils.FindCN(ip)

	hitSplit := 0
	allow := true
	for _, r := range this.lca {
		n := 0
		countSplit := strings.Count(fmt.Sprintf("%s%s%s", r.Area, r.Region, r.Isp), "-")
		if err != nil {
			if countSplit == 0 && r.Access == "deny" {
				return fmt.Errorf("deny outsite of china")
			}
		} else {
			if countSplit < hitSplit {
				continue
			}
			if r.Area != "" && r.Area == cn.Area {
				n = n + 1
			}
			if r.Region != "" && r.Region == cn.Region {
				n = n + 1
			}
			if r.Isp != "" && r.Isp == cn.Isp {
				n = n + 1
			}
			if n != 0 && n == countSplit {
				hitSplit = n
				if r.Access == "deny" {
					allow = false
				} else {
					allow = true
				}
			}
		}
	}
	if allow {
		return nil
	}
	if this.lcad {
		return fmt.Errorf("deny default")
	}
	return nil
}

func (this *LimitChinaAccessFilter) Disconnect(client net.Conn) {
}

func (this *LimitChinaAccessFilter) Read(client net.Conn, rwc core.ReadWriteCount) {
}

func (this *LimitChinaAccessFilter) Write(client net.Conn, rwc core.ReadWriteCount) {
}

func (this *LimitChinaAccessFilter) Request(buf []byte) error {
	return nil
}

func (this *LimitChinaAccessFilter) Stop() {
}

func init() {
	RegisterFilter("limit_china_access", func() interface{} {
		return new(LimitChinaAccessFilter)
	})
}
