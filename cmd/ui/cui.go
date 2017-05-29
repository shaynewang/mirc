package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"sync"

	"github.com/jroimartin/gocui"
	"github.com/shaynewang/mirc"
)

// SERVER is the default server ip and port number
// TODO: use config.yaml file for server ip address
const SERVER = "127.0.0.1:6667"
const retries = 3
const timeout = 20
const ping = 10

type client mirc.Client

var currentClient *client
var currentRoom = "public"

// Initialize new client connection
func newClient(server string) *client {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Printf("ERROR: server %s is not available\n", server)
		//log.Printf("ERROR: %s", err)
		// TODO: add error handleing, maybe ask for new server IP
		os.Exit(-1)
	}
	con := mirc.Connection{conn}
	new := client{
		IP:     conn.LocalAddr(),
		Socket: &con,
	}
	new.Nick = setNick()
	return &new
}

func (c *client) newMsg(opCode int16, receiver string, body string) *mirc.Message {
	msg := mirc.NewMsg(opCode, receiver, body)
	msg.Header.Sender = c.Nick
	return msg
}

func (c *client) newServMsg(opCode int16, body string) *mirc.Message {
	return c.newMsg(opCode, "server", body)
}

func newClientConnection(server string) *mirc.Connection {
	var err error
	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Println(err)
		// TODO: add error handleing, maybe ask for new server IP
		os.Exit(-1)
	}
	con := mirc.Connection{conn}
	return &con
}

// Sets nickname locally
func setNick() string {
	fmt.Print("Input your nickname:")
	reader := bufio.NewReader(os.Stdin)
	nick, _ := reader.ReadString('\n')
	nick = strings.Replace(nick, "\n", "", -1)
	return nick
}

// Prompt user to input to a nick name
// Then update it with the server
func (c *client) changeNick(nick string) {
	var opCode int16
	var msg *mirc.Message
	opCode = mirc.CONNECTION_FAILURE
	for opCode == mirc.CONNECTION_FAILURE {
		msg = c.newServMsg(mirc.CLIENT_CHANGE_NICK, nick)
		c.Socket.Conn.SetDeadline(mirc.CalDeadline(timeout))
		c.Socket.SendMsg(msg)
		opCode, msg = c.Socket.GetMsg()
		fmt.Printf("Error: %s\n", msg.Body)
	}
	c.Nick = nick
}

// send request to connect to the server
func (c *client) requestToConnect() error {
	var err error
	for i := 0; i < retries+1; i++ {
		if i > 0 {
			fmt.Printf("Retry connecting... (%d/%d)\n", i, retries)
		}
		msg := c.newServMsg(mirc.CLIENT_REQUEST_CONNECTION, c.Nick)
		c.Socket.Conn.SetDeadline(mirc.CalDeadline(timeout))
		err = c.Socket.SendMsg(msg)
		if err != nil {
			fmt.Printf("%s\n", err)
		} else {
			// request new nickname if exisit in server
			c.Socket.Conn.SetDeadline(mirc.CalDeadline(timeout))
			opCode, msg := c.Socket.GetMsg()
			for opCode == mirc.CONNECTION_FAILURE {
				fmt.Printf("Cannot connect: %s\n", msg.Body)
				c.changeNick(setNick())
				c.Socket.Conn.SetDeadline(mirc.CalDeadline(timeout))
				opCode, msg = c.Socket.GetMsg()
			}
			fmt.Printf("Connected\n")
			break
		}
	}
	return err
}

// send a request to create a room to server
func (c *client) createRoom(room string) error {
	//fmt.Printf("requesting new room %s\n", room)
	msg := c.newServMsg(mirc.CLIENT_CREATE_ROOM, room)
	return c.Socket.SendMsg(msg)
}

// send a request to join a room to server
func (c *client) joinRoom(room string) error {
	msg := c.newServMsg(mirc.CLIENT_JOIN_ROOM, room)
	return c.Socket.SendMsg(msg)
}

// send a request to leave a room
func (c *client) leaveRoom(room string) error {
	msg := c.newServMsg(mirc.CLIENT_LEAVE_ROOM, room)
	return c.Socket.SendMsg(msg)
}

// send a private message to a client
func (c *client) sendPrivateMsg(receiver string, msgBody string) error {
	msg := c.newMsg(mirc.CLIENT_SEND_MESSAGE, receiver, msgBody)
	return c.Socket.SendMsg(msg)
}

// send a private message to a client
func (c *client) sendPubMsg(msgBody string) error {
	msg := c.newMsg(mirc.CLIENT_SEND_PUB_MESSAGE, currentRoom, msgBody)
	return c.Socket.SendMsg(msg)
}

// command parser return command word and message
func comParser(cmdLine string) (string, string) {
	cmdLine = strings.Replace(cmdLine, "\n", "", -1)
	args := strings.SplitN(cmdLine, " ", 2)
	cmd := args[0]
	arg := ""
	if len(args) > 1 {
		arg = args[1]
	}
	return cmd, arg
}

// changeRoom changes client's current room
func changeRoom(roomName string) {
	currentRoom = roomName
}

// listRoom lists all rooms in the server
func (c *client) listRoom() error {
	reqMsg := c.newServMsg(mirc.CLIENT_LIST_ROOM, "")
	return c.Socket.SendMsg(reqMsg)
}

// message handler loop
func msgHandlerLoop(c *mirc.Connection, g *gocui.Gui) {
	for {
		msgHandler(c, g)
	}
}

// Handles a message
func msgHandler(c *mirc.Connection, g *gocui.Gui) int {
	c.Conn.SetDeadline(mirc.CalDeadline(timeout))
	opCode, msg := c.GetMsg()

	if opCode == mirc.ERROR {
		return -1
	} else if opCode == mirc.SERVER_BROADCAST_MESSAGE || opCode == mirc.SERVER_TELL_MESSAGE {
		g.Execute(func(g *gocui.Gui) error {
			v, err := g.View("view")
			if err != nil {
				return err
			}
			fmt.Fprintf(v, "\n%s [%s] %s: %s\n", mirc.GetTime(), currentRoom, msg.Header.Sender, msg.Body)
			return nil
		})
	} else if opCode == mirc.CONNECTION_CLOSED {
		g.Execute(func(g *gocui.Gui) error {
			v, err := g.View("view")
			if err != nil {
				return err
			}
			fmt.Fprintf(v, "Server connection closed, client exiting...\n")
			return nil
		})
		g.Close()
		os.Exit(0)
	} else if opCode == mirc.SERVER_RPL_LIST_ROOM {
		g.Execute(func(g *gocui.Gui) error {
			v, err := g.View("view")
			if err != nil {
				return err
			}
			fmt.Fprintf(v, "Rooms available: %s\n", msg.Body)
			return nil
		})
	} else if opCode == mirc.SERVER_RPL_LIST_MEMBER {
		g.Execute(func(g *gocui.Gui) error {
			v, err := g.View("view")
			if err != nil {
				return err
			}
			fmt.Fprintf(v, "Members in %s: %s\n", currentRoom, msg.Body)
			return nil
		})
	}
	return 0
}

// periodically send ping to the server to notify this client is alive
func (c *client) keepAliveLoop() {
	keepAliveMsg := c.newServMsg(mirc.CONNECTION_PING, "ping")
	for {
		c.Socket.SendMsg(keepAliveMsg)
		time.Sleep(ping * time.Second)
	}
}

/* ==================================================================================== */
func main() {
	currentClient = newClient(SERVER)
	// Initialize Connection
	currentClient.requestToConnect()
	go currentClient.keepAliveLoop()
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true

	g.SetManagerFunc(layout)
	maxX, maxY := g.Size()
	if lv, err := g.SetView("view", 2, 1, maxX-2, maxY-8); err != nil {
		if err != gocui.ErrUnknownView {
			fmt.Printf("Error: %s\n", err)
		}
		lv.Title = "Messages"
		lv.Autoscroll = true
	}

	if iv, err := g.SetView("input", 2, maxY-6, maxX-2, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			fmt.Printf("Error: %s\n", err)
		}
		iv.Title = "Input"
		iv.Editable = true
		err = iv.SetCursor(0, 0)
		_, err = g.SetCurrentView("input")
		if err != nil {
			log.Println("Cannot set focus to input view:", err)
		}
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, inputFunc); err != nil {
		log.Panicln(err)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go g.MainLoop()
	go msgHandlerLoop(currentClient.Socket, g)

	wg.Wait()

}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if lv, err := g.SetView("view", 2, 1, maxX-2, maxY-8); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		lv.Title = "Messages"
		lv.Autoscroll = true
	}

	if iv, err := g.SetView("input", 2, maxY-6, maxX-2, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		iv.Title = "Input"
		iv.Editable = true
		err = iv.SetCursor(0, 0)
		_, err = g.SetCurrentView("input")
		if err != nil {
			log.Println("Cannot set focus to input view:", err)
		}
	}

	return nil
}

func inputFunc(g *gocui.Gui, iv *gocui.View) error {
	//lv, _ := g.View("view")
	v, _ := g.View("input")
	//n = v.ViewBuffer()
	cmdLine := iv.ViewBuffer()
	//lv.Clear()
	v.Clear()
	v.SetCursor(0, 0)
	cmd, arg := comParser(cmdLine)
	if len(cmd) == 0 {
		return nil
	}
	lv, _ := g.View("view")

	if cmd == "\\create" {
		currentClient.createRoom(arg)
	} else if cmd == "\\nick" {
		currentClient.changeNick(arg)
	} else if cmd == "\\join" {
		currentClient.joinRoom(arg)
	} else if cmd == "\\listRoom" {
		currentClient.listRoom()
	} else if cmd == "\\leave" {
		currentClient.leaveRoom(arg)
	} else if cmd[0] == '@' {
		g.Execute(func(g *gocui.Gui) error {
			v, err := g.View("view")
			if err != nil {
				return err
			}
			fmt.Fprintf(v, "\n%s [%s] %s: %s\n", mirc.GetTime(), currentRoom, currentClient.Nick, arg)
			return nil
		})
		currentClient.sendPrivateMsg(cmd[1:], arg)
	} else if cmd == "\\exit" {
		fmt.Fprintf(lv, "Bye!\n")
		g.Close()
		os.Exit(0)
	} else {
		currentClient.sendPubMsg(cmdLine)
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
