package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/millken/tcpwder/codec"
	"github.com/millken/tcpwder/config"
	"github.com/millken/tcpwder/manager"
)

const version = "1.0.0"

var (
	flagConfigFile = flag.String("c", "./config.json", "Path to config.")
)

func main() {
	log.Printf("tcpwder v%s // by millken\n", version)
	flag.Parse()

	var cfg config.Config

	data, err := ioutil.ReadFile(*flagConfigFile)
	if err != nil {
		log.Fatal(err)
	}
	if err = codec.Decode(string(data), &cfg, "toml"); err != nil {
		log.Fatal(err)
	}
	manager.Initialize(cfg)
	<-(chan string)(nil)

}
