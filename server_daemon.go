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

func removeClient(cnick string, clientMap map[string]Client) int {
	if _, ok := clientMap[cnick]; ok {
		delete(clientMap, cnick)
		return 0
	}
	return -1
}

// Handles the connection
func handleConnection(conn net.Conn) {
	defer conn.Close()
	con := connection{
		conn: conn,
		enc:  *gob.NewEncoder(conn),
		dec:  *gob.NewDecoder(conn),
	}
	//dec := gob.NewDecoder(conn)
	//buf := new(Message)
	//err := dec.Decode(&buf)
	//if err != nil {
	//	conn.Close()
	//	return
	//}
	//ocode := buf.Header.Op_code
	opcode, nick := con.getMsg()
	//nick := buf.Body
	client_ip := con.conn.RemoteAddr()
	if opcode != 100 {
		// Drop the invalid connection
		conn.Close()
		return
	}

	for addClient(nick, client_ip, activeClients) < 0 {
		// If nickname exists then client will be asked
		// to change
		con.sendMsg(CONNECTION_FAILURE, "nickname exists")
		opcode, msg := con.getMsg()
		if opcode == CLIENT_CHANGE_NICK {
			nick = msg
		} else if opcode < 0 {
			// Client quit unexpectly
			con.conn.Close()
			return
		}
	}
	fmt.Printf("%s has connected\n", nick)
	fmt.Printf("ip: %s\n", activeClients[nick].ip)
	for {
		//buf := make([]byte, 1024)
		//conn.Read(buf)
		err, message := con.getMsg()
		if err == ERROR {
			conn.Close()
			removeClient(nick, activeClients)
			return
		}
		fmt.Printf("%s\n", client_ip)
		fmt.Printf("%s says %s\n", nick, message)
		//daytime := time.Now().String()
		//conn.Write([]byte(buf))
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
