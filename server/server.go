package main

import (
	"fmt"
	"net"
	"os"
	"strings"
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
	r := rooms["public"]
	r.addMember(cnick)
	rooms["public"] = r
	fmt.Printf("%s added to %s\n", cnick, r)
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
	fmt.Printf("room %s created\n", roomName)
	return nil

}

// add a client to a room
func (c *client) joinRoom(roomName string) error {
	if _, ok := rooms[roomName]; !ok {
		return errors.New("room doesn't exist")
	}
	r := rooms[roomName]
	r.addMember(c.Nick)
	rooms[roomName] = r
	fmt.Printf("%s added to %s\n", c.Nick, r)
	return nil
}

// add member to a room
func (r *room) addMember(nick string) error {
	if contain(r.Members, nick) >= 0 {
		//  Cannot add duplicated nickname
		fmt.Printf("%s is already in %s\n", nick, r.Name)
		return errors.New("nickname exists")
	}

	r.Members = append(r.Members, nick)
	return nil

}

// remove member from a room
func (r *room) removeMember(nick string) error {
	i := contain(r.Members, nick)
	if i >= 0 {
		r.Members = append(r.Members[:i], r.Members[i+1:]...)
		return nil
	}
	//  Cannot delete non member
	fmt.Printf("DEBUG: member %s removed from %s: %s", nick, r.Name, r.Members)
	return errors.New("cannot remove memeber")

}

// check if element in list
func contain(list []string, el string) int {
	for i, v := range list {
		if v == el {
			// Found!
			return i
		}
	}
	return -1
}

// remove client from the client list
func removeClient(nick string, clientMap map[string]client) int {
	if _, ok := clientMap[nick]; ok {
		delete(clientMap, nick)
	}
	for roomName := range rooms {
		r := rooms[roomName]
		r.removeMember(nick)
		rooms[roomName] = r
	}
	return 0
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

// listRoomHandler
func (c *client) listRoomHandler() {
	var roomList []string
	for name := range rooms {
		roomList = append(roomList, name)
	}
	msgBody := strings.Join(roomList, " ,")
	c.Socket.SendMsg(newMsg(mirc.SERVER_RPL_LIST_ROOM, c.Nick, msgBody))
	return
}

func (c *client) listMemberHandler(room string) {
	if _, ok := rooms[room]; !ok {
		c.Socket.SetWriteDeadline(mirc.CalDeadline(timeout))
		msgBody := "room " + room + " doesn't exist.\n"
		c.Socket.SendMsg(newMsg(mirc.SERVER_TELL_MESSAGE, c.Nick, msgBody))
		return
	}
	msgBody := strings.Join(rooms[room].Members, " ,")
	c.Socket.SendMsg(newMsg(mirc.SERVER_RPL_LIST_MEMBER, c.Nick, msgBody))
	return
}

// handles request if client is a member of a room
// replies with "true" if it's a member and "false" if not a member
func (c *client) inRoomHandler(room string) {
	if _, ok := rooms[room]; !ok {
		c.Socket.SetWriteDeadline(mirc.CalDeadline(timeout))
		msgBody := "room " + room + " doesn't exist.\n"
		c.Socket.SendMsg(newMsg(mirc.SERVER_TELL_MESSAGE, c.Nick, msgBody))
		return
	}
	if contain(rooms[room].Members, c.Nick) < 0 {
		c.Socket.SendMsg(newMsg(mirc.SERVER_TELL_MESSAGE, c.Nick, "not a member of the room"))
		return
	}
	fmt.Print("DEBUG: in room")
	c.Socket.SendMsg(newMsg(mirc.SERVER_RPL_CLIENT_IN_ROOM, c.Nick, room))
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

// broadCastMsg sends passes message to all members in a room
func broadCastMsg(m *mirc.Message) {
	if _, ok := rooms[m.Header.Receiver]; !ok {
		msgBody := "Room " + m.Header.Receiver + " doesn't exist.\n"
		activeClients[m.Header.Sender].Socket.SendMsg(newMsg(mirc.SERVER_TELL_MESSAGE, m.Header.Sender, msgBody))
		return
	}
	receiverList := rooms[m.Header.Receiver].Members
	m.Header.OpCode = mirc.SERVER_BROADCAST_MESSAGE
	fmt.Printf("%s list\n", receiverList)
	for i := 0; i < len(receiverList); i++ {
		c := receiverList[i]
		if c != "server" {
			fmt.Printf("client %s\n", c)
			activeClients[c].Socket.SendMsg(m)
		}
	}
	return

}

func (c *client) requestHandler() {
	for {
		c.Socket.Conn.SetReadDeadline(mirc.CalDeadline(inactiveTimeout))
		opCode, msg := c.Socket.GetMsg()
		if opCode == mirc.ERROR {
			fmt.Printf("Server DEBUG: %d\n", opCode)
			c.Socket.Conn.Close()
			removeClient(c.Nick, activeClients)
			c.Socket.SendMsg(newMsg(mirc.CONNECTION_CLOSED, c.Nick, "server has closed your connection"))
			fmt.Printf("%s has disconnected\n", c.Nick)
			return
		}
		if opCode == mirc.CLIENT_SEND_PUB_MESSAGE {
			fmt.Printf("%s says %s\n", c.Nick, msg.Body)
			broadCastMsg(msg)
		} else if opCode == mirc.CLIENT_SEND_MESSAGE {
			rallyMsg(msg)
		} else if opCode == mirc.CONNECTION_PING {
			c.Socket.SendMsg(newMsg(mirc.CONNECTION_PING, c.Nick, "ping"))
			continue
		} else if opCode == mirc.CONNECTION_CLOSED {
			removeClient(c.Nick, activeClients)
		} else if opCode == mirc.CLIENT_CREATE_ROOM {
			c.addRoomHandler(msg)
		} else if opCode == mirc.CLIENT_JOIN_ROOM {
			c.joinRoomHandler(msg)
		} else if opCode == mirc.CLIENT_LIST_ROOM {
			c.listRoomHandler()
		} else if opCode == mirc.CLIENT_IN_ROOM {
			c.inRoomHandler(msg.Body)
		} else if opCode == mirc.CLIENT_LEAVE_ROOM {
			r := rooms[msg.Body]
			err := r.removeMember(c.Nick)
			if err != nil {
				c.Socket.SendMsg(newMsg(mirc.ERROR, c.Nick, "cannot remove member"))
			} else {
				rooms[msg.Body] = r
				m := "you left room " + msg.Body
				fmt.Printf("%s says %s\n", c.Nick, msg.Body)
				c.Socket.SendMsg(newMsg(mirc.SERVER_TELL_MESSAGE, c.Nick, m))
			}
		} else if opCode == mirc.CLIENT_LIST_MEMBER {
			c.listMemberHandler(msg.Body)
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
	addRoom("public", "server")
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleConnection(conn)
	}
}
