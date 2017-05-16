package main

import (
	"encoding/gob"
	"fmt"
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

// add client to the client list
func addClient(cnick string, conn net.Conn, clientMap map[string]mirc.Client) int {
	if _, ok := clientMap[cnick]; ok {
		//  Cannot add duplicated nickname
		return -1
	}

	newClient := mirc.Client{
		Ip:      conn.RemoteAddr(),
		Nick:    cnick,
		Timeout: time.Now().Add(time.Second * time.Duration(TIMEOUT)),
		Socket: &mirc.Connection{
			Conn: conn,
			Enc:  *gob.NewEncoder(conn),
			Dec:  *gob.NewDecoder(conn)}}
	clientMap[cnick] = newClient
	return 0
}

// remove client from the client list
func removeClient(cnick string, clientMap map[string]mirc.Client) int {
	if _, ok := clientMap[cnick]; ok {
		delete(clientMap, cnick)
		return 0
	}
	return -1
}

//addRoomHandler
func addRoomHandler(rname string) {
	return
}

func rallyMsg(from string, m *mirc.Message) {
	m.Header.Sender = from
	fmt.Printf("sender %s\n", m.Header.Sender)
	fmt.Printf("recever %s\n", m.Header.Recever)
	activeClients[m.Header.Recever].Socket.SendMsg(mirc.SERVER_TELL_MESSAGE, from, m.Body)
	//messageQ = append(messageQ, *m)I
	return
}

// when a new client connects adds it to the list, then initialize a message queue
// for that client.
func handleConnection(conn net.Conn) {
	defer conn.Close()
	con := mirc.Connection{
		Conn: conn,
		Enc:  *gob.NewEncoder(conn),
		Dec:  *gob.NewDecoder(conn),
	}
	opCode, msg := con.GetMsg()
	nick := msg.Body
	//clientIP := con.Conn.RemoteAddr()
	if opCode != 100 {
		// Silently drop the invalid Connection
		return
	}

	for addClient(nick, conn, activeClients) < 0 {
		// If nickname exists then client will be asked
		// to change
		con.SendMsg(mirc.CONNECTION_FAILURE, nick, "nickname exists")
		opCode, msg = con.GetMsg()
		if opCode == mirc.CLIENT_CHANGE_NICK {
			nick = msg.Body
		} else if opCode < 0 {
			// Client quit unexpectly
			con.Conn.Close()
			return
		}
	}
	con.SendMsg(mirc.CONNECTION_SUCCESS, nick, "Connection established")

	fmt.Printf("%s has connected\n", nick)
	fmt.Printf("ip: %s\n", activeClients[nick].Ip)
	for {
		if len(activeClients[nick].MsgQ) != 0 {
			con.SendMsg(mirc.SERVER_TELL_MESSAGE, nick, activeClients[nick].MsgQ[0].Body)
			//fmt.Printf("%s", con.MsgQ[0].Body)
		}
		opCode, msg = con.GetMsg()
		if opCode == mirc.ERROR {
			con.Conn.Close()
			removeClient(nick, activeClients)
			fmt.Printf("%s has disconnected\n", nick)
			return
		}
		if opCode == mirc.CLIENT_SEND_PUB_MESSAGE {
			fmt.Printf("%s says %s\n", nick, msg.Body)
		} else if opCode == mirc.CLIENT_SEND_MESSAGE {
			rallyMsg(nick, msg)
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
