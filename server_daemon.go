package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"time"
)

// Default parameters
const LISTEN_PORT = ":6667"
const TIMEOUT = 10

// Mapping of client's nickname to client object
var activeClients = map[string]Client{}

func addClient(cnick string, cip net.Addr, clientMap map[string]Client) int {
	if _, ok := clientMap[cnick]; ok {
		//  Cannot add duplicated nickname
		return -1
	}
	newClient := Client{
		ip:      cip,
		nick:    cnick,
		timeout: time.Now().Add(time.Second * time.Duration(TIMEOUT)),
	}
	clientMap[cnick] = newClient
	return 0
}

// Handles the connection
func handleConnection(conn net.Conn) {
	defer conn.Close()
	dec := gob.NewDecoder(conn)
	buf := new(Message)
	err := dec.Decode(&buf)
	if err != nil {
		conn.Close()
		return
	}
	ocode := buf.Header.Op_code
	nick := buf.Body
	client_ip := conn.RemoteAddr()
	if ocode != 100 {
		// Drop the invalid connection
		conn.Close()
		return
	}

	for addClient(nick, client_ip, activeClients) < 0 {
		// If nickname exists then client will be asked
		// to change
		msg := []byte("nickname " + nick + " is taken, please choose another name")
		conn.Write(msg)
		dec := gob.NewDecoder(conn)
		buf := new(Message)
		err := dec.Decode(&buf)
		if err != nil {
			conn.Close()
			return
		}
		ocode := buf.Header.Op_code
		if ocode == CLIENT_CHANGE_NICK {
			nick = buf.Body
		}
	}
	fmt.Printf("%s has connected\n", nick)
	fmt.Printf("ip: %s\n", activeClients[nick].ip)
	for {
		buf := make([]byte, 1024)
		conn.Read(buf)
		fmt.Printf("%s\n", client_ip)
		fmt.Printf("%s says %s\n", nick, buf)
		//daytime := time.Now().String()
		conn.Write([]byte(buf))
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
