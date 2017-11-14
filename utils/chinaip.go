package utils

import (
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
