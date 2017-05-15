package mirc

import (
	"net"
	"time"
)

const (
	CONNECTION_SUCCESS        = 1
	CONNECTION_FAILURE        = 2
	CONNECTION_PING           = 3
	CONNECTION_ACK            = 4
	CLIENT_REQUEST_CONNECTION = 100
	CLIENT_CREATE_ROOM        = 101
	CLIENT_JOIN_ROOM          = 102
	CLIENT_LEAVE_ROOM         = 103
	CLIENT_LIST_ROOM          = 104
	CLIENT_LIST_MEMBER        = 105
	CLIENT_SEND_MESSAGE       = 106
	CLIENT_SEND_PUB_MESSAGE   = 107
	CLIENT_CHANGE_NICK        = 108
	SERVER_RPL_LIST_ROOM      = 204
	SERVER_RPL_LIST_MEMBER    = 205
	SERVER_TELL_MESSAGE       = 206
	SERVER_BROADCAST_MESSAGE  = 207
	ERROR                     = 1000
)

type Client struct {
	Ip      net.Addr
	Nick    string
	Timeout time.Time
}

type MsgHeader struct {
	OpCode  int16
	RecNick string
	MsgLen  int
}

type Message struct {
	Header MsgHeader
	Body   string
}
