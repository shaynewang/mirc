package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"time"
)

const LISTEN_PORT = ":6667"
const TIMEOUT = 5

var clientMap = map[string]Client{}

func addClient(cnick string, cip net.Addr, clientMap map[string]Client) int {
	newClient := Client{
		ip:      cip,
		nick:    cnick,
		timeout: time.Now().Add(time.Second * time.Duration(TIMEOUT)),
	}
	clientMap[cnick] = newClient
	return 1
}

// Handles the connection
func handleConnection(conn net.Conn) {
	defer conn.Close()
	dec := gob.NewDecoder(conn)
	buf := new(Message)
	err := dec.Decode(&buf)
	if err != nil {
		conn.Close()
	}
	ocode := buf.Header.Op_code
	nick := buf.Body
	client_port := conn.RemoteAddr()
	if ocode != 100 {
		// Drop the invalid connection
		conn.Close()
	} else {
		addClient(nick, client_port, clientMap)
		fmt.Printf("%s has connected\n", nick)
		fmt.Printf("ip: %s\n", clientMap[nick].ip)
		for {
			buf := make([]byte, 1024)
			conn.Read(buf)
			fmt.Printf("%s\n", client_port)
			fmt.Printf("%s says %s\n", nick, buf)
			//daytime := time.Now().String()
			conn.Write([]byte(buf))
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", LISTEN_PORT)

	if err != nil {
		// handle error
		fmt.Print("Server failed to start...\n")
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
