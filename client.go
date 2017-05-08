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

func main() {
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
	opcode, msg := con.getMsg()
	for opcode == CONNECTION_FAILURE {
		fmt.Printf("Cannot connect: %s\n", msg)
		con.changeNick()
		opcode, msg = con.getMsg()
	}

	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		//fmt.Fprintf(conn, text+"\n")
		con.sendMsg(CLIENT_SEND_PUB_MESSAGE, text+"\n")
		//message, _ := bufio.NewReader(conn).ReadString('\n')
		_, message := con.getMsg()
		fmt.Print("Message from server: " + message)
	}
}
