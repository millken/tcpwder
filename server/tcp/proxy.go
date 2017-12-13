package tcp

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/millken/tcpwder/core"
)

const (

	/* Buffer size to handle data from socket */
	BUFFER_SIZE = 16 * 1024

	/* Interval of pushing aggregated read/write stats */
	PROXY_STATS_PUSH_INTERVAL = 1 * time.Second
)

/**
 * Perform copy/proxy data from 'from' to 'to' socket, counting r/w stats and
 * dropping connection if timeout exceeded
 */
func (this *Server) proxy(to net.Conn, from net.Conn, timeout time.Duration) <-chan core.ReadWriteCount {

	stats := make(chan core.ReadWriteCount)
	outStats := make(chan core.ReadWriteCount)

	rwcBuffer := core.ReadWriteCount{}
	ticker := time.NewTicker(PROXY_STATS_PUSH_INTERVAL)
	flushed := false

	// Stats collecting goroutine
	go func() {

		if timeout > 0 {
			from.SetReadDeadline(time.Now().Add(timeout))
		}

		for {
			select {
			case <-ticker.C:
				if !rwcBuffer.IsZero() {
					outStats <- rwcBuffer
				}
				flushed = true
			case rwc, ok := <-stats:

				if !ok {
					ticker.Stop()
					if !flushed && !rwcBuffer.IsZero() {
						outStats <- rwcBuffer
					}
					close(outStats)
					return
				}

				if timeout > 0 && rwc.CountRead > 0 {
					from.SetReadDeadline(time.Now().Add(timeout))
				}

				// Remove non blocking
				if flushed {
					rwcBuffer = rwc
				} else {
					rwcBuffer.CountWrite += rwc.CountWrite
					rwcBuffer.CountRead += rwc.CountRead
				}

				flushed = false
			}
		}
	}()

	// Run proxy copier
	go func() {
		err := this.Copy(to, from, stats)
		// hack to determine normal close. TODO: fix when it will be exposed in golang
		e, ok := err.(*net.OpError)
		if err != nil && (!ok || e.Err.Error() != "use of closed network connection") {
			log.Printf("[WARN] proxy net %s", err)
		}

		to.Close()
		from.Close()

		// Stop stats collecting goroutine
		close(stats)
	}()

	return outStats
}

/**
 * It's build by analogy of io.Copy
 */
func (this *Server) Copy(to io.Writer, from io.Reader, ch chan<- core.ReadWriteCount) error {

	buf := make([]byte, BUFFER_SIZE)
	var err error = nil

	for {
		readN, readErr := from.Read(buf)

		if readN > 0 {
			if err = this.filter.HandleClientRequest(buf); err != nil {
				return err
			}

			writeN, writeErr := to.Write(buf[0:readN])

			if writeN > 0 {
				ch <- core.ReadWriteCount{CountRead: uint(readN), CountWrite: uint(writeN)}
			}

			if writeErr != nil {
				err = writeErr
				break
			}

			if readN != writeN {
				err = io.ErrShortWrite
				break
			}
		}

		if readErr == io.EOF {
			break
		}

		if readErr != nil {
			err = readErr
			break
		}
	}

	return err
}
