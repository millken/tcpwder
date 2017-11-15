package main

import (
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/logutils"

	"github.com/millken/tcpwder/api"
	"github.com/millken/tcpwder/codec"
	"github.com/millken/tcpwder/config"
	"github.com/millken/tcpwder/manager"
	"github.com/millken/tcpwder/utils"
)

const version = "1.0.1"

var (
	flagConfigFile = flag.String("c", "./config.json", "Path to config.")
)

/**
 * Initialize package
 */
func init() {

	// Set GOMAXPROCS if not set
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	// Init random seed
	rand.Seed(time.Now().UnixNano())

	// Save info
	config.Version = version
	config.StartTime = time.Now()

}

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

	if cfg.Defaults.ChinaIpdbPath != "" {
		err = utils.LoadCNIpDB(cfg.Defaults.ChinaIpdbPath)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("[INFO] loading china ip")
	}

	// Start API
	go api.Start(cfg.Api)

	manager.Initialize(cfg)
	<-(chan string)(nil)

}
