package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

const LISTEN_PORT = ":6667"

type client struct {
	ip      net.Addr
	nick    string
	timeout time.Time
}

// Handles the connection
func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		buf := make([]byte, 1024)
		conn.Read(buf)
		client_port := conn.RemoteAddr()
		fmt.Printf("%s\n", client_port)
		fmt.Printf("Client says %s\n", buf)
		daytime := time.Now().String()
		conn.Write([]byte(daytime))
		conn.Write([]byte(buf))
	}
}

func main() {
	ln, err := net.Listen("tcp", LISTEN_PORT)

	if err != nil {
		// handle error
		fmt.Print("Server start failed...\n")
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(-1)
	}
	fmt.Print("Server has started.\n")
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleConnection(conn)
	}
}
