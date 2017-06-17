package mirc

import "testing"

func TestNewMsg(t *testing.T) {
	var tests = []struct {
		opCode   int16
		receiver string
		body     string
		msg      Message
	}{{SERVER_BROADCAST_MESSAGE,
		"fakeNick",
		"fakeMsg",
		Message{Header: MsgHeader{OpCode: SERVER_BROADCAST_MESSAGE,
			Sender:   "",
			Receiver: "fakeNick",
			MsgLen:   len("fakeMsg"),
		},
			Body: "fakeMsg"}}}

	for _, test := range tests {
		testMsg := NewMsg(test.opCode, test.receiver, test.body)
		if *testMsg != test.msg {
			t.Error("New message doesn't match.")
		}
	}

}
