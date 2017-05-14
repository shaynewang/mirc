package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"strings"
)

func (c *connection) requestToConnect() {
	fmt.Print("Input your nickname:")
	reader := bufio.NewReader(os.Stdin)
	nick, _ := reader.ReadString('\n')
	nick = strings.Replace(nick, "\n", "", -1)
	c.sendMsg(CLIENT_REQUEST_CONNECTION, nick)
}

func (c *connection) changeNick() {
	fmt.Print("Input your nickname:")
	reader := bufio.NewReader(os.Stdin)
	nick, _ := reader.ReadString('\n')
	nick = strings.Replace(nick, "\n", "", -1)
	c.sendMsg(CLIENT_CHANGE_NICK, nick)
}

//Run will run client
func Run() {
	// Initialize connection
	fmt.Printf("Current server: %s\n", SERVER)
	conn, _ := net.Dial("tcp", SERVER)
	con := connection{
		conn: conn,
		enc:  *gob.NewEncoder(conn),
		dec:  *gob.NewDecoder(conn),
	}
	con.requestToConnect()
	// request new nickname if exisit in server
	opCode, msg := con.getMsg()
	for opCode == CONNECTION_FAILURE {
		fmt.Printf("Cannot connect: %s\n", msg)
		con.changeNick()
		opCode, msg = con.getMsg()
	}
	fmt.Printf("Connected\n")

	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		//fmt.Fprintf(conn, text+"\n")
		con.sendMsg(CLIENT_SEND_PUB_MESSAGE, text)
		//message, _ := bufio.NewReader(conn).ReadString('\n')
		//_, message := con.getMsg()
		//fmt.Print("Message from server: " + message)
	}
}
