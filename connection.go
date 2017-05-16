package mirc

func (c *Connection) SendMsg(opCode int16, sender string, recever string, msgBody string) error {
	//fmt.Printf("DEBUG: sendMsg opcode %d\n", opcode)
	//fmt.Printf("DEBUG: sendMsg msg_body %s\n", msg_body)
	msg := new(Message)
	msg.Header = MsgHeader{OpCode: opCode, Sender: sender, Recever: recever, MsgLen: len(msgBody)}
	msg.Body = msgBody
	err := c.Enc.Encode(&msg)
	return err
}

func (c *Connection) GetMsg() (int16, *Message) {
	recvMsg := new(Message)
	err := c.Dec.Decode(&recvMsg)
	if err != nil {
		return ERROR, nil
	}
	return recvMsg.Header.OpCode, recvMsg
}
