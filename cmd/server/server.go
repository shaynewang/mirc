package main

import (
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
const inactiveTimeout = 30

type client mirc.Client
type room mirc.Room

// Mapping of client's nickname to client object
var activeClients = map[string]client{}
var rooms = map[string]room{}

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
		Socket:  &mirc.Connection{conn},
	}
	clientMap[cnick] = newClient
	return &newClient, nil
}

// create a new room
func addRoom(roomName string, nick string) error {
	if _, ok := rooms[roomName]; ok {
		return errors.New("room exists")
	}

	newRoom := room{Name: roomName}
	newRoom.addMember(nick)
	rooms[roomName] = newRoom
	return nil

}

// add member to a room
func (r *room) addMember(nick string) error {
	if _, ok := rooms[nick]; ok {
		//  Cannot add duplicated nickname
		return errors.New("nickname exists")
	}

	r.Members = append(r.Members, nick)
	return nil

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
func (c *client) addRoomHandler(m *mirc.Message) {
	err := addRoom(m.Body, m.Header.Sender)
	if err != nil {
		c.Socket.SetWriteDeadline(mirc.CalDeadline(timeout))
		c.Socket.SendMsg(newMsg(mirc.SERVER_TELL_MESSAGE, c.Nick, err.Error()))
	}
	c.Socket.SetWriteDeadline(mirc.CalDeadline(timeout))
	msgBody := "Room " + m.Body + " created!\n"
	c.Socket.SendMsg(newMsg(mirc.SERVER_TELL_MESSAGE, c.Nick, msgBody))
	fmt.Printf("room %s created\n", m.Body)
	return
}

// server passes rallied message to the receiver
func rallyMsg(m *mirc.Message) {
	if _, ok := activeClients[m.Header.Receiver]; !ok {
		msgBody := "Receiver " + m.Header.Receiver + " doesn't exist.\n"
		activeClients[m.Header.Sender].Socket.SendMsg(newMsg(mirc.SERVER_TELL_MESSAGE, m.Header.Sender, msgBody))
		return
	}
	m.Header.OpCode = mirc.SERVER_TELL_MESSAGE
	activeClients[m.Header.Receiver].Socket.SendMsg(m)
	//messageQ = append(messageQ, *m)I
	return
}

func (c *client) requestHandler() {
	for {
		c.Socket.Conn.SetReadDeadline(mirc.CalDeadline(inactiveTimeout))
		opCode, msg := c.Socket.GetMsg()
		if opCode == mirc.ERROR {
			fmt.Printf("Server DEBUG: %s", opCode)
			c.Socket.Conn.Close()
			removeClient(c.Nick, activeClients)
			c.Socket.SendMsg(newMsg(mirc.CONNECTION_CLOSED, c.Nick, "server has closed your connection"))
			fmt.Printf("%s has disconnected\n", c.Nick)
			return
		}
		if opCode == mirc.CLIENT_SEND_PUB_MESSAGE {
			fmt.Printf("%s says %s\n", c.Nick, msg.Body)
		} else if opCode == mirc.CLIENT_SEND_MESSAGE {
			fmt.Printf("DEBUG Rally message\n")
			rallyMsg(msg)
		} else if opCode == mirc.CONNECTION_PING {
			continue
		} else if opCode == mirc.CLIENT_CREATE_ROOM {
			c.addRoomHandler(msg)
		}
		//daytime := time.Now().String()
	}
}

// when a new client connects adds it to the list, then initialize a message queue
// for that client.
func handleConnection(conn net.Conn) {
	defer conn.Close()
	// boostrap client connection
	con := mirc.Connection{conn}
	con.Conn.SetReadDeadline(mirc.CalDeadline(timeout))
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
		con.Conn.SetWriteDeadline(mirc.CalDeadline(timeout))
		con.SendMsg(newMsg(mirc.CONNECTION_FAILURE, nick, "nickname exists"))
		con.Conn.SetReadDeadline(mirc.CalDeadline(timeout))
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
	con.Conn.SetWriteDeadline(mirc.CalDeadline(timeout))
	con.SendMsg(newMsg(mirc.CONNECTION_SUCCESS, nick, "Connection established"))

	fmt.Printf("%s has connected\n", nick)
	fmt.Printf("ip: %s\n", activeClients[nick].IP)
	client.requestHandler()
	fmt.Printf("Client %s has left\n", nick)
	return
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
