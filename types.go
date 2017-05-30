package mirc

import (
	"net"
	"time"
)

// A list of opCodes
const (
	CONNECTION_SUCCESS        = 1
	CONNECTION_FAILURE        = 2
	CONNECTION_PING           = 3
	CONNECTION_ACK            = 4
	CONNECTION_CLOSED         = 5
	CLIENT_REQUEST_CONNECTION = 100
	CLIENT_CREATE_ROOM        = 101
	CLIENT_JOIN_ROOM          = 102
	CLIENT_LEAVE_ROOM         = 103
	CLIENT_LIST_ROOM          = 104
	CLIENT_LIST_MEMBER        = 105
	CLIENT_SEND_MESSAGE       = 106
	CLIENT_SEND_PUB_MESSAGE   = 107
	CLIENT_CHANGE_NICK        = 108
	CLIENT_IN_ROOM            = 109
	SERVER_RPL_LIST_ROOM      = 204
	SERVER_RPL_LIST_MEMBER    = 205
	SERVER_TELL_MESSAGE       = 206
	SERVER_BROADCAST_MESSAGE  = 207
	SERVER_RPL_CLIENT_IN_ROOM = 208
	ERROR                     = 1000
)

// Connection type contains the connection object
type Connection struct {
	net.Conn
}

// Client type contains information of clients in server
type Client struct {
	IP      net.Addr
	Nick    string
	Timeout time.Time
	Socket  *Connection
}

// Room type contains the room name and the list of memebers
type Room struct {
	Name    string
	Members []string
}

// MsgHeader contains header information of messages
type MsgHeader struct {
	OpCode   int16
	Sender   string
	Receiver string
	MsgLen   int
	Timeout  int
}

// Message contain the header object as well as the body of a message
type Message struct {
	Header MsgHeader
	Body   string
}
