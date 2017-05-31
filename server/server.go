package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"errors"

	"sync"

	"github.com/shaynewang/mirc"
)

// Default parameters
const listenPort = ":6667"
const timeout = 10
const inactiveTimeout = 30

/******************** types ********************/
type client mirc.Client
type room mirc.Room
type clientList struct {
	mu   sync.Mutex
	list map[string]client
}
type roomList struct {
	mu   sync.Mutex
	list map[string]room
}

/********************* Globals ******************/

// list of all clients on the server
var clients = clientList{
	mu:   sync.Mutex{},
	list: map[string]client{},
}

// list of all rooms on the server
var rooms = roomList{
	mu:   sync.Mutex{},
	list: map[string]room{},
}

/********************** Server funtions *****************/
// newMsg creates a message object from input parameters
func newMsg(opCode int16, receiver string, body string) *mirc.Message {
	msg := mirc.NewMsg(opCode, receiver, body)
	msg.Header.Sender = "server"
	msg.Header.Timeout = timeout
	return msg
}

// add client to the client list
func addClient(cnick string, conn net.Conn, clients *clientList) (*client, error) {
	clients.mu.Lock()
	if _, ok := clients.list[cnick]; ok {
		//  Cannot add duplicated nickname
		clients.mu.Unlock()
		return nil, errors.New("nickname exists")
	}

	newClient := client{
		IP:      conn.RemoteAddr(),
		Nick:    cnick,
		Timeout: time.Now().Add(time.Second * time.Duration(timeout)),
		Socket:  &mirc.Connection{conn},
	}
	clients.list[cnick] = newClient
	clients.mu.Unlock()
	rooms.mu.Lock()
	r := rooms.list["public"]
	r.addMember(cnick)
	rooms.list["public"] = r
	rooms.mu.Unlock()
	fmt.Printf("%s added to %s\n", cnick, r)
	return &newClient, nil
}

// remove client from the client list
func removeClient(nick string, clientMap map[string]client) int {
	clients.mu.Lock()
	if _, ok := clientMap[nick]; ok {
		delete(clientMap, nick)
	}
	clients.mu.Unlock()
	rooms.mu.Lock()
	for roomName := range rooms.list {
		r := rooms.list[roomName]
		r.removeMember(nick)
		if len(r.Members) > 0 {
			rooms.list[roomName] = r
		}
	}
	rooms.mu.Unlock()
	return 0
}

// create a new room
func addRoom(roomName string, nick string) error {
	rooms.mu.Lock()
	if _, ok := rooms.list[roomName]; ok {
		rooms.mu.Unlock()
		return errors.New("room exists")
	}
	newRoom := room{Name: roomName}
	newRoom.addMember(nick)
	rooms.list[roomName] = newRoom
	rooms.mu.Unlock()
	return nil
}

// add a client to a room
func (c *client) joinRoom(roomName string) error {
	rooms.mu.Lock()
	if _, ok := rooms.list[roomName]; !ok {
		rooms.mu.Unlock()
		return errors.New("room doesn't exist")
	}
	r := rooms.list[roomName]
	r.addMember(c.Nick)
	rooms.list[roomName] = r
	rooms.mu.Unlock()
	fmt.Printf("%s is added to %s\n", c.Nick, r)
	return nil
}

// add member to a room assumes lock is held
func (r *room) addMember(nick string) error {
	if contain(r.Members, nick) >= 0 {
		//  Cannot add duplicated nickname
		fmt.Printf("%s is already in %s\n", nick, r.Name)
		return errors.New("nickname exists")
	}

	r.Members = append(r.Members, nick)
	if len(r.Members) > 1 {
		m := nick + " joined"
		broadCastMsg(newMsg(mirc.SERVER_BROADCAST_MESSAGE, r.Name, m))
	}
	return nil
}

// remove member from a room assumes lock is held
func (r *room) removeMember(nick string) error {
	i := contain(r.Members, nick)
	if i >= 0 {
		r.Members = append(r.Members[:i], r.Members[i+1:]...)
		if len(r.Members) <= 0 {
			delete(rooms.list, r.Name)
			fmt.Printf("empty room %s has been removed\n", r.Name)
		}
		return nil
	}
	//  Cannot delete non member
	return errors.New("cannot remove memeber")
}

// handles client's error messages
func (c *client) errorHandler() {
	c.Socket.Conn.Close()
	removeClient(c.Nick, clients.list)
	c.Socket.SendMsg(newMsg(mirc.CONNECTION_CLOSED, c.Nick, "server has closed your connection"))
	fmt.Printf("%s has disconnected\n", c.Nick)
}

// remove client from the client list
func (c *client) removeClientHandler() int {
	return removeClient(c.Nick, clients.list)
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

// joinRoomHandler
func (c *client) joinRoomHandler(m *mirc.Message) {
	err := c.joinRoom(m.Body)
	if err != nil {
		c.Socket.SetWriteDeadline(mirc.CalDeadline(timeout))
		c.Socket.SendMsg(newMsg(mirc.SERVER_TELL_MESSAGE, c.Nick, err.Error()))
	}
	c.Socket.SetWriteDeadline(mirc.CalDeadline(timeout))
	msgBody := "You joined " + m.Body + "!\n"
	c.Socket.SendMsg(newMsg(mirc.SERVER_TELL_MESSAGE, c.Nick, msgBody))
	return
}

// handles clent's leave room request
func (c *client) leaveRoomHandler(m *mirc.Message) {
	r := rooms.list[m.Body]
	err := r.removeMember(c.Nick)
	if err != nil {
		c.Socket.SendMsg(newMsg(mirc.ERROR, c.Nick, "cannot remove member"))
	} else {
		if len(r.Members) > 0 {
			rooms.list[m.Body] = r
			msg := m.Header.Sender + " left the room"
			broadCastMsg(newMsg(mirc.SERVER_BROADCAST_MESSAGE, m.Body, msg))
		}
		m := "you have left the room " + m.Body
		c.Socket.SendMsg(newMsg(mirc.SERVER_TELL_MESSAGE, c.Nick, m))
	}
	return
}

// list all rooms of client's request
func (c *client) listRoomHandler() {
	var roomList []string
	rooms.mu.Lock()
	for name := range rooms.list {
		roomList = append(roomList, name)
	}
	rooms.mu.Unlock()
	msgBody := strings.Join(roomList, " ,")
	c.Socket.SendMsg(newMsg(mirc.SERVER_RPL_LIST_ROOM, c.Nick, msgBody))
	return
}

// list all members of a room that client's requested
func (c *client) listMemberHandler(room string) {
	rooms.mu.Lock()
	if _, ok := rooms.list[room]; !ok {
		rooms.mu.Unlock()
		c.Socket.SetWriteDeadline(mirc.CalDeadline(timeout))
		msgBody := "room " + room + " doesn't exist.\n"
		c.Socket.SendMsg(newMsg(mirc.SERVER_TELL_MESSAGE, c.Nick, msgBody))
		return
	}
	msgBody := strings.Join(rooms.list[room].Members, " ,")
	rooms.mu.Unlock()
	c.Socket.SendMsg(newMsg(mirc.SERVER_RPL_LIST_MEMBER, c.Nick, msgBody))
	return
}

// handles request if client is a member of a room
// replies with "true" if it's a member and "false" if not a member
func (c *client) inRoomHandler(room string) {
	rooms.mu.Lock()
	if _, ok := rooms.list[room]; !ok {
		rooms.mu.Unlock()
		c.Socket.SetWriteDeadline(mirc.CalDeadline(timeout))
		msgBody := "room " + room + " doesn't exist.\n"
		c.Socket.SendMsg(newMsg(mirc.SERVER_TELL_MESSAGE, c.Nick, msgBody))
		return
	}
	if contain(rooms.list[room].Members, c.Nick) < 0 {
		rooms.mu.Unlock()
		c.Socket.SendMsg(newMsg(mirc.SERVER_TELL_MESSAGE, c.Nick, "not a member of the room"))
		return
	}
	rooms.mu.Unlock()
	c.Socket.SendMsg(newMsg(mirc.SERVER_RPL_CLIENT_IN_ROOM, c.Nick, room))
	return
}

// server passes rallied message to the receiver
func rallyMsg(m *mirc.Message) {
	if _, ok := clients.list[m.Header.Receiver]; !ok {
		msgBody := "Receiver " + m.Header.Receiver + " doesn't exist.\n"
		c := clients.list[m.Header.Sender]
		c.Socket.SendMsg(newMsg(mirc.SERVER_TELL_MESSAGE, m.Header.Sender, msgBody))
		return
	}
	m.Header.OpCode = mirc.SERVER_TELL_MESSAGE
	c := clients.list[m.Header.Receiver]
	c.Socket.SendMsg(m)
	return
}

// broadCastMsg sends passes message to all members in a room
func broadCastMsg(m *mirc.Message) {
	if _, ok := rooms.list[m.Header.Receiver]; !ok {
		msgBody := "Room " + m.Header.Receiver + " doesn't exist.\n"
		c := clients.list[m.Header.Sender]
		c.Socket.SendMsg(newMsg(mirc.SERVER_TELL_MESSAGE, m.Header.Sender, msgBody))
		return
	}
	receiverList := rooms.list[m.Header.Receiver].Members
	m.Header.OpCode = mirc.SERVER_BROADCAST_MESSAGE
	for i := 0; i < len(receiverList); i++ {
		cNick := receiverList[i]
		if cNick != "server" {
			c := clients.list[cNick]
			c.Socket.SendMsg(m)
		}
	}
	return
}

// handles requests from clients
func (c *client) requestHandler() {
	for {
		c.Socket.Conn.SetReadDeadline(mirc.CalDeadline(inactiveTimeout))
		opCode, msg := c.Socket.GetMsg()
		if opCode == mirc.ERROR {
			c.errorHandler()
			return
		}
		if opCode == mirc.CLIENT_SEND_PUB_MESSAGE {
			broadCastMsg(msg)
		} else if opCode == mirc.CLIENT_SEND_MESSAGE {
			rallyMsg(msg)
		} else if opCode == mirc.CONNECTION_PING {
			c.Socket.SendMsg(newMsg(mirc.CONNECTION_ACK, c.Nick, "pong"))
		} else if opCode == mirc.CONNECTION_CLOSED {
			removeClient(c.Nick, clients.list)
		} else if opCode == mirc.CLIENT_CREATE_ROOM {
			c.addRoomHandler(msg)
		} else if opCode == mirc.CLIENT_JOIN_ROOM {
			c.joinRoomHandler(msg)
		} else if opCode == mirc.CLIENT_LIST_ROOM {
			c.listRoomHandler()
		} else if opCode == mirc.CLIENT_IN_ROOM {
			c.inRoomHandler(msg.Body)
		} else if opCode == mirc.CLIENT_LEAVE_ROOM {
			c.leaveRoomHandler(msg)
		} else if opCode == mirc.CLIENT_LIST_MEMBER {
			c.listMemberHandler(msg.Body)
		}
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
	client, err := addClient(nick, conn, &clients)
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
		client, err = addClient(nick, conn, &clients)
	}
	con.Conn.SetWriteDeadline(mirc.CalDeadline(timeout))
	con.SendMsg(newMsg(mirc.CONNECTION_SUCCESS, nick, "Connection established"))

	fmt.Printf("%s has connected\n", nick)
	fmt.Printf("ip: %s\n", clients.list[nick].IP)
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
	addRoom("public", "server")
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleConnection(conn)
	}
}

// check if a string is in a list of strings
// returns the index of the element if it's in the list
// returns -1 if the element is not in the list
func contain(list []string, el string) int {
	for i, v := range list {
		if v == el {
			// found it
			return i
		}
	}
	return -1
}
