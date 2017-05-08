package main

import (
	"encoding/gob"
	"log"
	"net"
)

// Default server ip
// TODO: use config.yaml file for server ip address
const SERVER = "127.0.0.1:6667"

type connection struct {
	conn net.Conn
	enc  gob.Encoder
	dec  gob.Decoder
}

func (c *connection) initializeConnection() {
	var err error
	c.conn, err = net.Dial("tcp", SERVER)
	if err != nil {
		log.Println(err)
		return
	}
}

func (c *connection) sendMsg(opcode int16, msg_body string) error {
	//fmt.Printf("DEBUG: sendMsg opcode %d\n", opcode)
	//fmt.Printf("DEBUG: sendMsg msg_body %s\n", msg_body)
	msg := new(Message)
	msg.Header = Msg_header{Op_code: opcode, Msg_len: len(msg_body)}
	msg.Body = msg_body
	err := c.enc.Encode(&msg)
	return err
}

func (c *connection) getMsg() (int16, string) {
	recv_msg := new(Message)
	err := c.dec.Decode(&recv_msg)
	if err != nil {
		return ERROR, ""
	}
	return recv_msg.Header.Op_code, recv_msg.Body
}
