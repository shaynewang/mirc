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

type client struct {
	ip     net.Addr
	socket *mirc.Connection
}

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
	return &client{
		ip:     conn.LocalAddr(),
		socket: &con,
	}

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

func (c *client) requestToConnect() {
	fmt.Print("Input your nickname:")
	reader := bufio.NewReader(os.Stdin)
	nick, _ := reader.ReadString('\n')
	nick = strings.Replace(nick, "\n", "", -1)
	c.SendMsg(mirc.CLIENT_REQUEST_CONNECTION, nick, nick)
}

func changeNick(c *mirc.Connection) {
	fmt.Print("Input your nickname:")
	reader := bufio.NewReader(os.Stdin)
	nick, _ := reader.ReadString('\n')
	nick = strings.Replace(nick, "\n", "", -1)
	c.SendMsg(mirc.CLIENT_CHANGE_NICK, nick, nick)
}

//Run will run client
func main() {
	client := mirc.Client{}
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
