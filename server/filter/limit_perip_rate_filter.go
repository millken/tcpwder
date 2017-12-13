package filter

import (
	"log"
	"net"
	"time"

	"github.com/millken/tcpwder/config"
	"github.com/millken/tcpwder/core"
	"github.com/millken/tcpwder/firewall"
	"github.com/millken/tcpwder/utils"
)

type LimitPeripRateFilter struct {
	readBytes  uint
	writeBytes uint
	interval   time.Duration
	clients    map[string]*core.ReadWriteCount
	stop       chan bool
}

func (this *LimitPeripRateFilter) Init(cfg config.Server) bool {
	if cfg.LimitPeripRate != nil {
		this.readBytes = cfg.LimitPeripRate.ReadBytes
		this.writeBytes = cfg.LimitPeripRate.WriteBytes
		this.interval = utils.ParseDurationOrDefault(cfg.LimitPeripRate.Interval, time.Second*2)
		this.clients = make(map[string]*core.ReadWriteCount)

		ticker := time.NewTicker(this.interval)
		go func() {
			for {
				select {
				case <-ticker.C:
					this.clients = make(map[string]*core.ReadWriteCount)
				case <-this.stop:
					ticker.Stop()
					return
				}
			}
		}()
		return true
	}
	return false
}

func (this *LimitPeripRateFilter) Connect(client net.Conn) error {
	return nil
}

func (this *LimitPeripRateFilter) Disconnect(client net.Conn) {
}

func (this *LimitPeripRateFilter) Read(client net.Conn, rwc core.ReadWriteCount) {
	host, _, _ := net.SplitHostPort(client.RemoteAddr().String())
	if _, ok := this.clients[host]; !ok {
		this.clients[host] = &core.ReadWriteCount{}
	}
	this.clients[host].CountRead += rwc.CountRead
	if this.readBytes != 0 && this.clients[host].CountRead > this.readBytes {
		log.Printf("[WARN] LimitPeripRateFilter host %s reach read limit %d", host, this.readBytes)
		firewall.SetDeny(host, 3600)
	}
}

func (this *LimitPeripRateFilter) Write(client net.Conn, rwc core.ReadWriteCount) {
	host, _, _ := net.SplitHostPort(client.RemoteAddr().String())
	if _, ok := this.clients[host]; !ok {
		this.clients[host] = &core.ReadWriteCount{}
	}
	this.clients[host].CountWrite += rwc.CountWrite
	if this.writeBytes != 0 && this.clients[host].CountWrite > this.writeBytes {
		log.Printf("[WARN] LimitPeripRateFilter host %s reach write limit %d", host, this.writeBytes)
		firewall.SetDeny(host, 3600)
	}
}

func (this *LimitPeripRateFilter) Request(buf []byte) error {
	return nil
}

func (this *LimitPeripRateFilter) Stop() {
	close(this.stop)
}

func init() {
	RegisterFilter("limit_perip_rate", func() interface{} {
		return new(LimitPeripRateFilter)
	})
}
