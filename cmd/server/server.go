package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"time"

	"errors"

	"github.com/shaynewang/mirc"
)

// Default parameters
const listenPort = ":6667"
const timeout = 10

type client mirc.Client

// Mapping of client's nickname to client object
var activeClients = map[string]client{}

// newMsg creates a message object from input parameters
func newMsg(opCode int16, receiver string, body string) *mirc.Message {
	msg := mirc.NewMsg(opCode, receiver, body)
	msg.Header.Sender = "server"
	msg.Header.Timeout = timeout
	return msg
}

// add client to the client list
func addClient(cnick string, conn net.Conn, clientMap map[string]client) (*client, error) {
	if _, ok := clientMap[cnick]; ok {
		//  Cannot add duplicated nickname
		return nil, errors.New("nickname exists")
	}

	newClient := client{
		IP:      conn.RemoteAddr(),
		Nick:    cnick,
		Timeout: time.Now().Add(time.Second * time.Duration(timeout)),
		Socket: &mirc.Connection{
			Conn: conn,
			Enc:  *gob.NewEncoder(conn),
			Dec:  *gob.NewDecoder(conn)}}
	clientMap[cnick] = newClient
	return &newClient, nil
}

// remove client from the client list
func removeClient(cnick string, clientMap map[string]client) int {
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

func rallyMsg(m *mirc.Message) {
	//m.Header.Sender = from
	fmt.Printf("sender %s\n", m.Header.Sender)
	fmt.Printf("recever %s\n", m.Header.Receiver)
	activeClients[m.Header.Receiver].Socket.SendMsg(m)
	//messageQ = append(messageQ, *m)I
	return
}

func (c *client) requestHandler() {
	for {
		opCode, msg := c.Socket.GetMsg()
		if opCode == mirc.ERROR {
			c.Socket.Conn.Close()
			removeClient(c.Nick, activeClients)
			fmt.Printf("%s has disconnected\n", c.Nick)
			return
		}
		if opCode == mirc.CLIENT_SEND_PUB_MESSAGE {
			fmt.Printf("%s says %s\n", c.Nick, msg.Body)
		} else if opCode == mirc.CLIENT_SEND_MESSAGE {
			rallyMsg(msg)
		}
		//daytime := time.Now().String()
	}
}

// when a new client connects adds it to the list, then initialize a message queue
// for that client.
func handleConnection(conn net.Conn) {
	defer conn.Close()
	// boostrap client connection
	con := mirc.Connection{
		Conn: conn,
		Enc:  *gob.NewEncoder(conn),
		Dec:  *gob.NewDecoder(conn),
	}
	con.Conn.SetDeadline(mirc.CalDeadline(timeout))
	opCode, msg := con.GetMsg()
	nick := msg.Body
	if opCode != 100 {
		// Silently drop the invalid Connection
		return
	}

	// ask client to change their nickname if it's taken
	client, err := addClient(nick, conn, activeClients)
	for err != nil {
		// If nickname exists then client will be asked
		// to change
		con.SendMsg(newMsg(mirc.CONNECTION_FAILURE, nick, "nickname exists"))
		con.Conn.SetDeadline(mirc.CalDeadline(timeout))
		opCode, msg = con.GetMsg()
		if opCode == mirc.CLIENT_CHANGE_NICK {
			nick = msg.Body
		} else if opCode < 0 {
			// Client quit unexpectly
			con.Conn.Close()
			return
		}
		client, err = addClient(nick, conn, activeClients)
	}
	con.Conn.SetDeadline(mirc.CalDeadline(timeout))
	con.SendMsg(newMsg(mirc.CONNECTION_SUCCESS, nick, "Connection established"))

	fmt.Printf("%s has connected\n", nick)
	fmt.Printf("ip: %s\n", activeClients[nick].IP)
	client.requestHandler()
	fmt.Printf("Client %s has left\n", nick)
	/*
		for {
			if len(activeClients[nick].MsgQ) != 0 {
				con.SendMsg(newMsg(mirc.SERVER_TELL_MESSAGE, nick, activeClients[nick].MsgQ[0].Body))
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
	*/
}

func main() {
	ln, err := net.Listen("tcp", listenPort)

	if err != nil {
		// handle error
		fmt.Print("Server failed to start...\n")
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(-1)
	}
	fmt.Print("Server has started. Control + C to exit\n")
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleConnection(conn)
	}
}
