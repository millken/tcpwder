package main

import (
	"fmt"
	"net"
)


func main() {
	fmt.Printf("Starting server on port 1883\n")
	lsnr, err := net.Listen("tcp", ":1883")
	if err != nil {
		fmt.Printf("Error in listen: %s\n", err.Error())
		return
	}

	fmt.Printf("Listening on 1883\n")
	i := 0
	for {
		conn, err := lsnr.Accept()
		if err != nil {
			fmt.Printf("Cannot accept connection: %s\n", err.Error())
			continue
		}
		i++
		fmt.Printf("Handling client # %d\n", i)
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	for {
		conn.Read(make([]byte, 1))
	}
	fmt.Printf("Should not come here\n")
}
