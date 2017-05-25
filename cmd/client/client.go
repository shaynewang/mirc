package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/shaynewang/mirc"
)

// SERVER is the default server ip and port number
// TODO: use config.yaml file for server ip address
const SERVER = "127.0.0.1:6667"
const retries = 3
const timeout = 20
const ping = 10

type client mirc.Client

var currentRoom = "public"

// Initialize new client connection
func newClient(server string) *client {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Printf("ERROR: server %s is not available\n", server)
		//log.Printf("ERROR: %s", err)
		// TODO: add error handleing, maybe ask for new server IP
		os.Exit(-1)
	}
	con := mirc.Connection{conn}
	new := client{
		IP:     conn.LocalAddr(),
		Socket: &con,
	}
	new.setNick()
	return &new
}

func (c *client) newMsg(opCode int16, receiver string, body string) *mirc.Message {
	msg := mirc.NewMsg(opCode, receiver, body)
	msg.Header.Sender = c.Nick
	return msg
}

func (c *client) newServMsg(opCode int16, body string) *mirc.Message {
	return c.newMsg(opCode, "server", body)
}

func newClientConnection(server string) *mirc.Connection {
	var err error
	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Println(err)
		// TODO: add error handleing, maybe ask for new server IP
		os.Exit(-1)
	}
	con := mirc.Connection{conn}
	return &con
}

// Sets nickname locally
func (c *client) setNick() {
	fmt.Print("Input your nickname:")
	reader := bufio.NewReader(os.Stdin)
	nick, _ := reader.ReadString('\n')
	nick = strings.Replace(nick, "\n", "", -1)
	c.Nick = nick
}

// Prompt user to input to a nick name
// Then update it with the server
func (c *client) changeNick() {
	var opCode int16
	var msg *mirc.Message
	opCode = mirc.CONNECTION_FAILURE
	for opCode == mirc.CONNECTION_FAILURE {
		c.setNick()
		msg = c.newServMsg(mirc.CLIENT_CHANGE_NICK, c.Nick)
		c.Socket.Conn.SetDeadline(mirc.CalDeadline(timeout))
		c.Socket.SendMsg(msg)
		opCode, msg = c.Socket.GetMsg()
		fmt.Printf("Error: %s\n", msg.Body)
	}
}

// send request to connect to the server
func (c *client) requestToConnect() error {
	var err error
	for i := 0; i < retries+1; i++ {
		if i > 0 {
			fmt.Printf("Retry connecting... (%d/%d)\n", i, retries)
		}
		msg := c.newServMsg(mirc.CLIENT_REQUEST_CONNECTION, c.Nick)
		c.Socket.Conn.SetDeadline(mirc.CalDeadline(timeout))
		err = c.Socket.SendMsg(msg)
		if err != nil {
			fmt.Printf("%s\n", err)
		} else {
			// request new nickname if exisit in server
			c.Socket.Conn.SetDeadline(mirc.CalDeadline(timeout))
			opCode, msg := c.Socket.GetMsg()
			for opCode == mirc.CONNECTION_FAILURE {
				fmt.Printf("Cannot connect: %s\n", msg.Body)
				c.changeNick()
				c.Socket.Conn.SetDeadline(mirc.CalDeadline(timeout))
				opCode, msg = c.Socket.GetMsg()
			}
			fmt.Printf("Connected\n")
			break
		}
	}
	return err
}

// send a request to create a room to server
func (c *client) createRoom(room string) error {
	fmt.Printf("requesting new room %s\n", room)
	msg := c.newServMsg(mirc.CLIENT_CREATE_ROOM, room)
	return c.Socket.SendMsg(msg)
}

// send a request to join a room to server
func (c *client) joinRoom(room string) error {
	msg := c.newServMsg(mirc.CLIENT_JOIN_ROOM, room)
	return c.Socket.SendMsg(msg)
}

// send a private message to a client
func (c *client) sendPrivateMsg(receiver string, msgBody string) error {
	msg := c.newMsg(mirc.CLIENT_SEND_MESSAGE, receiver, msgBody)
	return c.Socket.SendMsg(msg)
}

// command parser return command word and message
func comParser(cmdLine string) (string, string) {
	cmdLine = strings.Replace(cmdLine, "\n", "", -1)
	args := strings.SplitN(cmdLine, " ", 2)
	cmd := args[0]
	arg := ""
	if len(args) > 1 {
		arg = args[1]
	}
	return cmd, arg
}

// command loop
func (c *client) commandLoop() {
	go msgHandlerLoop(c.Socket)
	for {

		fmt.Print(">>> ")
		reader := bufio.NewReader(os.Stdin)
		cmdLine, _ := reader.ReadString('\n')
		cmd, arg := comParser(cmdLine)
		fmt.Printf("DEBUG cmd: %s\n", cmd)
		fmt.Printf("DEBUG arg: %s\n", arg)

		if cmd == "\\create" {
			c.createRoom(arg)
		} else if cmd == "\\join" {
			c.joinRoom(arg)
		} else if cmd[0] == '@' {
			c.sendPrivateMsg(cmd[1:], arg)
		} else if cmd == "\\exit" {
			fmt.Printf("Bye!\n")
			os.Exit(0)
		}

	}
}

// message handler loop
func msgHandlerLoop(c *mirc.Connection) {
	for {
		msgHandler(c)
	}
}

// periodically send ping to the server to notify this client is alive
func (c *client) keepAliveLoop() {
	keepAliveMsg := c.newServMsg(mirc.CONNECTION_PING, "ping")
	for {
		c.Socket.SendMsg(keepAliveMsg)
		time.Sleep(ping * time.Second)
	}
}

// Handles a message
func msgHandler(c *mirc.Connection) int {
	c.Conn.SetDeadline(mirc.CalDeadline(timeout))
	opCode, msg := c.GetMsg()

	if opCode == mirc.ERROR {
		return -1
	} else if opCode == mirc.SERVER_BROADCAST_MESSAGE || opCode == mirc.SERVER_TELL_MESSAGE {
		fmt.Printf("\n%s [%s]: %s\n", mirc.GetTime(), msg.Header.Sender, msg.Body)
	} else if opCode == mirc.CONNECTION_CLOSED {
		fmt.Printf("Server connection closed, client exiting...\n")
		os.Exit(0)
	}
	return 0
}

// Main event loop
func main() {
	client := newClient(SERVER)
	// Initialize Connection
	fmt.Printf("Current server: %s\n", SERVER)
	client.requestToConnect()
	go client.keepAliveLoop()
	client.commandLoop()

}
