package firewall

import ttlmap "github.com/leprosus/golang-ttl-map"

var ttlMap ttlmap.Heap

func init() {
	ttlMap = ttlmap.New("firewall.tsv")
}

func SetAllow(ip string, ttl int64) {
	Set(ip, "allow", ttl)
}

func DelAllow(ip string) {
	Del(ip, "allow")
}

func SetDeny(ip string, ttl int64) {
	Set(ip, "deny", ttl)
}

func DelDeny(ip string) {
	Del(ip, "deny")
}

func Set(ip, value string, ttl int64) {
	ttlMap.Set(ip, value, ttl)
}

func Del(ip, match string) {
	value := ttlMap.Get(ip)
	if value == match {
		ttlMap.Del(ip)
	}
}

func Allows(ip string) bool {
	value := ttlMap.Get(ip)
	if value == "deny" {
		return false
	}
	return true
}
