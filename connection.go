package mirc

// SendMsg accepts OpCode, sender nick,receiver nick and message body as arguments then return
// an error if send is not successful
func (c *Connection) SendMsg_old(opCode int16, sender string, msgBody string) error {
	//fmt.Printf("DEBUG: sendMsg opcode %d\n", opcode)
	//fmt.Printf("DEBUG: sendMsg msg_body %s\n", msg_body)
	msg := new(Message)
	msg.Header = MsgHeader{OpCode: opCode, Sender: sender, MsgLen: len(msgBody)}
	msg.Body = msgBody
	err := c.Enc.Encode(&msg)
	return err
}

// SendMsg passes a message to the server
func (c *Connection) SendMsg(opCode int16, msg *Message) error {
	//fmt.Printf("DEBUG: sendMsg opcode %d\n", opcode)
	//fmt.Printf("DEBUG: sendMsg msg_body %s\n", msg_body)
	msg := new(Message)
	msg.Header = MsgHeader{OpCode: opCode, Sender: sender, MsgLen: len(msgBody)}
	msg.Body = msgBody
	err := c.Enc.Encode(&msg)
	return err
}

// GetMsg returns opCode, message if a message is in queue
func (c *Connection) GetMsg() (int16, *Message) {
	recvMsg := new(Message)
	err := c.Dec.Decode(&recvMsg)
	if err != nil {
		return ERROR, nil
	}
	return recvMsg.Header.OpCode, recvMsg
}
