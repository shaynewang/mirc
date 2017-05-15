package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/shaynewang/mirc"
)

// Default parameters
const LISTEN_PORT = ":6667"
const TIMEOUT = 10

// Mapping of client's nickname to client object
var activeClients = map[string]mirc.Client{}

func newServerConnection(server string) *mirc.Connection {
	var err error
	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Println(err)
		return nil
	}
	con := mirc.Connection{
		Conn: conn,
		Enc:  *gob.NewEncoder(conn),
		Dec:  *gob.NewDecoder(conn),
	}
	return &con
}

func addClient(cnick string, cip net.Addr, clientMap map[string]mirc.Client) int {
	if _, ok := clientMap[cnick]; ok {
		//  Cannot add duplicated nickname
		return -1
	}
	newClient := mirc.Client{
		Ip:      cip,
		Nick:    cnick,
		Timeout: time.Now().Add(time.Second * time.Duration(TIMEOUT)),
	}
	clientMap[cnick] = newClient
	return 0
}

func removeClient(cnick string, clientMap map[string]mirc.Client) int {
	if _, ok := clientMap[cnick]; ok {
		delete(clientMap, cnick)
		return 0
	}
	return -1
}

// Handles the Connection
func handleConnection(conn net.Conn) {
	defer conn.Close()
	con := mirc.Connection{
		Conn: conn,
		Enc:  *gob.NewEncoder(conn),
		Dec:  *gob.NewDecoder(conn),
	}
	opCode, nick := con.GetMsg()
	clientIP := con.Conn.RemoteAddr()
	if opCode != 100 {
		// Drop the invalid Connection
		conn.Close()
		return
	}

	for addClient(nick, clientIP, activeClients) < 0 {
		// If nickname exists then client will be asked
		// to change
		con.SendMsg(mirc.CONNECTION_FAILURE, "nickname exists")
		opCode, msg := con.GetMsg()
		if opCode == mirc.CLIENT_CHANGE_NICK {
			nick = msg
		} else if opCode < 0 {
			// Client quit unexpectly
			con.Conn.Close()
			return
		}
	}
	con.SendMsg(mirc.CONNECTION_SUCCESS, "Connection established")
	fmt.Printf("%s has connected\n", nick)
	fmt.Printf("ip: %s\n", activeClients[nick].Ip)
	for {
		opCode, message := con.GetMsg()
		if opCode == mirc.ERROR {
			conn.Close()
			removeClient(nick, activeClients)
			fmt.Printf("%s has disconnected\n", nick)
			return
		}
		if opCode == mirc.CLIENT_SEND_PUB_MESSAGE {
			fmt.Printf("%s says %s\n", nick, message)
		}
		//daytime := time.Now().String()
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
