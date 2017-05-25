package mirc

import "encoding/gob"

// SendMsg passes a message object to a reciever client
func (c *Connection) SendMsg(msg *Message) error {
	c.SetWriteDeadline(CalDeadline(10))
	encoder := gob.NewEncoder(c)
	err := encoder.Encode(&msg)
	//err := c.Enc.Encode(&msg)
	//	fmt.Printf("DEBUG: %s\n", err)
	return err
}

// GetMsg returns opCode, message if a message is in queue
func (c *Connection) GetMsg() (int16, *Message) {
	recvMsg := new(Message)
	//err := c.Dec.Decode(&recvMsg)
	decoder := gob.NewDecoder(c)
	err := decoder.Decode(&recvMsg)
	if err != nil {
		//fmt.Printf("DEBUG: %s\n", err)
		return ERROR, nil
	}
	return recvMsg.Header.OpCode, recvMsg
}
