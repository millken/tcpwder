//CLIENT
package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var requestCount uint64
var totalPingsPerConnection uint64 = 100
var concurrentConnections uint64 = 1
var totalPings = concurrentConnections * totalPingsPerConnection

func monitor(done chan bool) chan bool {
	out := make(chan bool)
	go func() {
		var last uint64
		start := time.Now()

		for {
			select {
			case <-done:
				elapsed := time.Since(start)
				fmt.Printf("%f ns\n", float64(elapsed)/float64(requestCount))
				fmt.Printf("%d requests\n", requestCount)
				fmt.Printf("%f requests per second\n", float64(time.Second)/(float64(elapsed)/float64(requestCount)))
				fmt.Printf("elapsed: %s\r\n", elapsed)
				out <- true
				return
			case <-time.After(1 * time.Second):
				current := atomic.LoadUint64(&requestCount)
				fmt.Printf("%d combined requests per second (%d)\n", current-last, current)
				last = current

				if current >= uint64(totalPings) {
					return
				}
			}
		}
	}()
	return out
}

func (c *client) readLoop(wg *sync.WaitGroup) {
	defer wg.Done()

	rd := bufio.NewReader(c.conn)
	buf := make([]byte, 4)

	for atomic.LoadUint64(&c.revcd) < totalPingsPerConnection {

		n, err := rd.Read(buf)
		if n > 0 {
			atomic.AddUint64(&c.revcd, 1)
			atomic.AddUint64(&requestCount, 1)
		} else if err == io.EOF {
			return
		}
		if err != nil && err != io.EOF {
			fmt.Println(err)
			return
		}
	}
	// fmt.Printf("total recvd: %d\r\n", c.revcd)
}

func (c *client) writeLoop(wg *sync.WaitGroup) {
	defer wg.Done()

	wr := bufio.NewWriterSize(c.conn, 65536)
	outBuf := []byte("Ping")
	// var buffered int

	for atomic.LoadUint64(&c.sent) < totalPingsPerConnection {

		n, err := wr.Write(outBuf)
		if n > 0 {
		}
		if err != nil && err != io.EOF {
			fmt.Println(err)
			return
		}
		atomic.AddUint64(&c.sent, 1)
	}
	wr.Flush()
	fmt.Printf("total sent: %d\r\n", c.sent)
}

const RingBufferCapacity = 1024 * 1024

type client struct {
	sent  uint64
	revcd uint64
	conn  *net.TCPConn
}

func NewClient(wg *sync.WaitGroup) {
	defer wg.Done()

	tcpAddr, _ := net.ResolveTCPAddr("tcp4", "localhost:1880")
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	var w sync.WaitGroup
	c := client{conn: conn}

	w.Add(2)
	go c.writeLoop(&w)
	go c.readLoop(&w)

	w.Wait()
	conn.Close()
}

func main() {
	runtime.GOMAXPROCS(8)

	var wg sync.WaitGroup
	done := make(chan bool)
	c := monitor(done)

	for i := uint64(0); i < concurrentConnections; i++ {
		wg.Add(1)
		go NewClient(&wg)
	}

	wg.Wait()
	done <- true
	<-c
}
