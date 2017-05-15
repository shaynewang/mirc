package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/shaynewang/mirc"
)

// SERVER is the default server ip and port number
// TODO: use config.yaml file for server ip address
const SERVER = "127.0.0.1:6667"

func newClientConnection(server string) *mirc.Connection {
	var err error
	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	con := mirc.Connection{
		Conn: conn,
		Enc:  *gob.NewEncoder(conn),
		Dec:  *gob.NewDecoder(conn),
	}
	return &con
}

func requestToConnect(c *mirc.Connection) {
	fmt.Print("Input your nickname:")
	reader := bufio.NewReader(os.Stdin)
	nick, _ := reader.ReadString('\n')
	nick = strings.Replace(nick, "\n", "", -1)
	c.SendMsg(mirc.CLIENT_REQUEST_CONNECTION, nick)
}

func changeNick(c *mirc.Connection) {
	fmt.Print("Input your nickname:")
	reader := bufio.NewReader(os.Stdin)
	nick, _ := reader.ReadString('\n')
	nick = strings.Replace(nick, "\n", "", -1)
	c.SendMsg(mirc.CLIENT_CHANGE_NICK, nick)
}

//Run will run client
func main() {
	// Initialize Connection
	fmt.Printf("Current server: %s\n", SERVER)
	con := newClientConnection(SERVER)
	//conn, _ := net.Dial("tcp", SERVER)
	//con := mirc.Connection{
	//	conn: conn,
	//	enc:  *gob.NewEncoder(conn),
	//	dec:  *gob.NewDecoder(conn),
	//}
	requestToConnect(con)
	// request new nickname if exisit in server
	opCode, msg := con.GetMsg()
	for opCode == mirc.CONNECTION_FAILURE {
		fmt.Printf("Cannot connect: %s\n", msg)
		changeNick(con)
		opCode, msg = con.GetMsg()
	}
	fmt.Printf("Connected\n")

	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		//fmt.Fprintf(conn, text+"\n")
		con.SendMsg(mirc.CLIENT_SEND_PUB_MESSAGE, text)
		//message, _ := bufio.NewReader(conn).ReadString('\n')
		//_, message := con.getMsg()
		//fmt.Print("Message from server: " + message)
	}
}
