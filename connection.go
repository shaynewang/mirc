package mirc

import "encoding/gob"

// SendMsg passes a message object to a reciever client
func (c *Connection) SendMsg(msg *Message) error {
	c.SetWriteDeadline(CalDeadline(10))
	encoder := gob.NewEncoder(c)
	err := encoder.Encode(&msg)
	return err
}

// GetMsg returns opCode, message if a message is in queue
func (c *Connection) GetMsg() (int16, *Message) {
	recvMsg := new(Message)
	decoder := gob.NewDecoder(c)
	err := decoder.Decode(&recvMsg)
	if err != nil {
		return ERROR, nil
	}
	return recvMsg.Header.OpCode, recvMsg
}
