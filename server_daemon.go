package main

import (
	"fmt"
	"net"
)

const LISTEN_PORT = ":6667"

// Handles the connection
//func (conn net.Conn) handleConnection() {
//}

func main() {
	ln, err := net.Listen("tcp", LISTEN_PORT)

	if err != nil {
		// handle error
		fmt.Print("Problem connecting...")
	}
	for {
		_, err := ln.Accept()
		if err != nil {
			// handle error
		}
		fmt.Print("Client connected!")
		//go handleConnection(conn)
	}
}
