package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/logutils"

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
	filter_writer := os.Stdout
	if cfg.Logging.Output != "" {

		switch cfg.Logging.Output {
		case "stdout":
			filter_writer = os.Stdout
		case "stderr":
			filter_writer = os.Stderr

		default:
			filter_writer, err = os.Create(cfg.Logging.Output)
			if err != nil {
				log.Printf("[ERROR] %s", err)
			}

		}

	}
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"INFO", "DEBUG", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(strings.ToUpper(cfg.Logging.Level)),
		Writer:   filter_writer,
	}

	log.SetOutput(filter)
	//log.SetFlags(log.LstdFlags | log.Lshortfile)

	manager.Initialize(cfg)
	<-(chan string)(nil)

}
