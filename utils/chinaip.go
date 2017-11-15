package utils

import (
	"errors"
	"net"

	cnip "github.com/millken/go-ipdb"
)

var cnipDB *cnip.DB

// LoadCNIpDB load cnip db path
func LoadCNIpDB(path string) (err error) {
	cnipDB, err = cnip.Load(path)
	return err
}

// FindCN
func FindCN(ip net.IP) (result *cnip.Result, err error) {
	return cnipDB.Find(ip.String())
}

func PrivateIP(ip string) (bool, error) {
	var err error
	private := false
	IP := net.ParseIP(ip)
	if IP == nil {
		err = errors.New("Invalid IP")
	} else {
		_, private24BitBlock, _ := net.ParseCIDR("10.0.0.0/8")
		_, private20BitBlock, _ := net.ParseCIDR("172.16.0.0/12")
		_, private16BitBlock, _ := net.ParseCIDR("192.168.0.0/16")
		private = private24BitBlock.Contains(IP) || private20BitBlock.Contains(IP) || private16BitBlock.Contains(IP)
	}
	return private, err
}
