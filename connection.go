package mirc

import (
	"encoding/gob"
	"net"
)

type Connection struct {
	Conn net.Conn
	Enc  gob.Encoder
	Dec  gob.Decoder
}

func (c *Connection) SendMsg(opCode int16, msgBody string) error {
	//fmt.Printf("DEBUG: sendMsg opcode %d\n", opcode)
	//fmt.Printf("DEBUG: sendMsg msg_body %s\n", msg_body)
	msg := new(Message)
	msg.Header = MsgHeader{OpCode: opCode, MsgLen: len(msgBody)}
	msg.Body = msgBody
	err := c.Enc.Encode(&msg)
	return err
}

func (c *Connection) GetMsg() (int16, string) {
	recvMsg := new(Message)
	err := c.Dec.Decode(&recvMsg)
	if err != nil {
		return ERROR, ""
	}
	return recvMsg.Header.OpCode, recvMsg.Body
}
