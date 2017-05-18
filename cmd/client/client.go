package main

import (
	"bufio"
	"encoding/gob"
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

type client mirc.Client

// Initialize new client connection
func newClient(server string) *client {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Println(err)
		// TODO: add error handleing, maybe ask for new server IP
		os.Exit(-1)
	}
	con := mirc.Connection{
		Conn: conn,
		Enc:  *gob.NewEncoder(conn),
		Dec:  *gob.NewDecoder(conn),
	}
	new := client{
		IP:     conn.LocalAddr(),
		Socket: &con,
	}
	new.setNick()
	return &new
}

func newClientConnection(server string) *mirc.Connection {
	var err error
	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Println(err)
		// TODO: add error handleing, maybe ask for new server IP
		os.Exit(-1)
	}
	con := mirc.Connection{
		Conn: conn,
		Enc:  *gob.NewEncoder(conn),
		Dec:  *gob.NewDecoder(conn),
	}
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
		c.Socket.SendMsg(mirc.CLIENT_CHANGE_NICK, c.Nick, c.Nick)
		opCode, msg = c.Socket.GetMsg()
		fmt.Printf("Error: %s\n", msg.Body)
	}
}

// send request to connect to the server
func (c *client) requestToConnect() error {
	var err error
	for i := 0; i < retries; i++ {
		fmt.Printf("Retry connecting... (%d/%d)\n", i, retries)
		err = c.Socket.SendMsg(mirc.CLIENT_REQUEST_CONNECTION, c.Nick, c.Nick)
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	}
	return err
}

// send a request to create a room to server
func (c *client) createRoom(room string) error {
	return c.Socket.SendMsg(mirc.CLIENT_CREATE_ROOM, c.Nick, room)
}

// send a request to join a room to server
func (c *client) joinRoom(room string) error {
	return c.Socket.SendMsg(mirc.CLIENT_JOIN_ROOM, c.Nick, room)
}

// command loop
func (c *client) commandLoop() {
	for {
		fmt.Print(">>> ")
		reader := bufio.NewReader(os.Stdin)
		cmdLine, _ := reader.ReadString('\n')
		cmdLine = strings.Replace(cmdLine, "\n", "", -1)
		args := strings.Fields(cmdLine)
	}
}

// Handles a message
func msgHandler(c *mirc.Connection) int {
	opCode, msg := c.GetMsg()
	daytime := time.Now().String()

	if opCode == mirc.ERROR {
		return -1
	} else if opCode == mirc.SERVER_BROADCAST_MESSAGE {
		fmt.Printf("%s []", mirc.GetTime())
	}
	return 0
}

//Run will run client
func main() {
	client := newClient(SERVER)
	// Initialize Connection
	fmt.Printf("Current server: %s\n", SERVER)
	client.requestToConnect()
	// request new nickname if exisit in server
	opCode, msg := con.GetMsg()
	for opCode == mirc.CONNECTION_FAILURE {
		fmt.Printf("Cannot connect: %s\n", msg.Body)
		changeNick(con)
		opCode, msg = con.GetMsg()
	}
	fmt.Printf("Connected\n")

	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		recever := "shayne"
		//fmt.Fprintf(conn, text+"\n")
		//go con.SendMsg(mirc.CLIENT_SEND_PUB_MESSAGE, text)
		con.SendMsg(mirc.CLIENT_SEND_MESSAGE, nick, recever, text)
		opCode, msg = con.GetMsg()
		if opCode == mirc.SERVER_TELL_MESSAGE {
			fmt.Printf("Message from %s: %s", msg.Header.Sender, msg.Body)
		}

		//message, _ := bufio.NewReader(conn).ReadString('\n')
		//_, message := con.getMsg()
	}
}
