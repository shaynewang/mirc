package main

import (
	"net"
	"time"
)

const (
	CONNECTION_SUCCESS        = 1
	CLIENT_REQUEST_CONNECTION = 100
)

type Client struct {
	ip      net.Addr
	nick    string
	timeout time.Time
}

type Msg_header struct {
	Op_code int8
	Msg_len int
}

type Message struct {
	Header Msg_header
	Body   string
}
