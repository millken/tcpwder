package main

import (
	"fmt"
	"flag"
	"sync"
	"time"
	"net"
)

var Hostname string
var Port string
var numberOfDevices int
var wg sync.WaitGroup

func init() {
	flag.IntVar(&numberOfDevices, "devices", 25000, "specify how many devices to simulate")
	flag.StringVar(&Hostname, "hostname", "localhost", "dns entry for clearblade")
	flag.StringVar(&Port, "port", "1883", "mqtt port for clearblade")
	flag.Parse()
}

func main() {
	var arrayClients []net.Conn
	isErr := false
	i := 0
	fmt.Printf("Connected clients are %d\n", len(arrayClients))
	wg.Add(numberOfDevices)
	for i < numberOfDevices {
		go func(v int) {
			start := time.Now()
			fmt.Printf("Connecting client # %d\n", v)
			conn, err := net.Dial("tcp", Hostname + ":" + Port)
			if err != nil {
				fmt.Printf("Got error connecting client: %s\n", err.Error())
				isErr = true
				return
			}
			arrayClients = append(arrayClients, conn)
			fmt.Printf("Array elemets are %d\n", len(arrayClients))
			fmt.Printf("Connected client # %d\n", v)
			end := time.Now()
			required := end.Sub(start)
			wg.Done()
			fmt.Printf("Time is %s\n", required.String())
			wg.Wait()
		}(i)
		if isErr {
			break
		}
		time.Sleep(time.Millisecond * 20)
		i = i + 1
	}
	for {
		fmt.Printf("Connected clients are %d\n", len(arrayClients))
		time.Sleep(time.Second * 5)
	}
}
