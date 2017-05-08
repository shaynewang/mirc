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

func main() {
	conn, _ := net.Dial("tcp", "127.0.0.1:6667")
	con_msg := new(Message)
	con_msg.Header = Msg_header{Op_code: CLIENT_REQUEST_CONNECTION, Msg_len: 0}
	fmt.Print("Input your nickname:")
	reader := bufio.NewReader(os.Stdin)
	nick, _ := reader.ReadString('\n')
	nick = strings.Replace(nick, "\n", "", -1)
	con_msg.Body = nick
	//fmt.Printf(con_msg)
	//dec := gob.NewDecoder(conn)
	enc := gob.NewEncoder(conn)
	err := enc.Encode(&con_msg)
	if err != nil {
		log.Fatal("encode error:", err)
	}
	//fmt.Fprintf(conn, con_msg)

	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		fmt.Fprintf(conn, text+"\n")
		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print("Message from server: " + message)
	}
}
