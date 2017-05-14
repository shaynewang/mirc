package main

import (
	"encoding/gob"
	"log"
	"net"
)

// SERVER is the default server ip and port number
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

func (c *connection) sendMsg(opCode int16, msgBody string) error {
	//fmt.Printf("DEBUG: sendMsg opcode %d\n", opcode)
	//fmt.Printf("DEBUG: sendMsg msg_body %s\n", msg_body)
	msg := new(Message)
	msg.Header = MsgHeader{OpCode: opCode, MsgLen: len(msgBody)}
	msg.Body = msgBody
	err := c.enc.Encode(&msg)
	return err
}

func (c *connection) getMsg() (int16, string) {
	recvMsg := new(Message)
	err := c.dec.Decode(&recvMsg)
	if err != nil {
		return ERROR, ""
	}
	return recvMsg.Header.OpCode, recvMsg.Body
}
