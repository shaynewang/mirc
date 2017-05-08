package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

var enc = gob.NewEncoder(conn)

func requestToConnect(conn net.Conn) {
	fmt.Print("Input your nickname:")
	reader := bufio.NewReader(os.Stdin)
	nick, _ := reader.ReadString('\n')
	nick = strings.Replace(nick, "\n", "", -1)
	msg := new(Message)
	msg.Header = Msg_header{Op_code: CLIENT_REQUEST_CONNECTION, Msg_len: len(nick)}
	msg.Body = nick
	err := enc.Encode(&msg)
	if err != nil {
		log.Fatal("encode error:", err)
	}
}

func changeNick(conn net.Conn) {
	fmt.Print("Input your nickname:")
	reader := bufio.NewReader(os.Stdin)
	nick, _ := reader.ReadString('\n')
	nick = strings.Replace(nick, "\n", "", -1)
	msg := new(Message)
	msg.Header = Msg_header{Op_code: CLIENT_CHANGE_NICK, Msg_len: len(nick)}
	msg.Body = nick
	//enc := gob.NewEncoder(conn)
	err := enc.Encode(&msg)
	if err != nil {
		log.Fatal("encode error:", err)
	}
}

func main() {
	conn, _ := net.Dial("tcp", "127.0.0.1:6667")
	//dec := gob.NewDecoder(conn)
	//fmt.Fprintf(conn, con_msg)
	go requestToConnect(conn)

	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		fmt.Fprintf(conn, text+"\n")
		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print("Message from server: " + message)
	}
}
