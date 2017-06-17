package mirc

import "time"

// GetTime returns the current time as a string
func GetTime() string {
	t := time.Now()
	return t.Format("03:04 PM")
}

// CalDeadline calculates deadline time from current time + timout time
func CalDeadline(timeout int) time.Time {
	return time.Now().Add(time.Duration(timeout) * time.Second)
}

// NewMsg creates a message object from input parameters
func NewMsg(opCode int16, receiver string, body string) *Message {
	msg := Message{
		Header: MsgHeader{
			OpCode:   opCode,
			Receiver: receiver,
			MsgLen:   len(body),
		},
		Body: body,
	}
	return &msg
}
