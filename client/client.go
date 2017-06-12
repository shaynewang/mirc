package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"

	"sync"

	"github.com/jroimartin/gocui"
	"github.com/shaynewang/mirc"
)

// Default parameters
const retries = 3
const timeout = 20
const dialTimeout = 5
const ping = 10

type client mirc.Client
type conf struct {
	Server string
}

// Initialize new client connection
func newClient(server string) *client {
	//conn, err := net.Dial("tcp", server)
	conn, err := net.DialTimeout("tcp", server, dialTimeout*time.Second)
	if err != nil {
		log.Printf("ERROR: server %s is not available\n", server)
		// TODO: add error handleing, maybe ask for new server IP
		os.Exit(-1)
	}
	con := mirc.Connection{conn}
	new := client{
		IP:     conn.LocalAddr(),
		Room:   "public",
		Socket: &con,
	}
	new.Nick = setNick()
	return &new
}

// Generate new message object from opcode, receiver nick and message body
func (c *client) newMsg(opCode int16, receiver string, body string) *mirc.Message {
	msg := mirc.NewMsg(opCode, receiver, body)
	msg.Header.Sender = c.Nick
	return msg
}

// Generate new message to server from opcode and message body to server
func (c *client) newServMsg(opCode int16, body string) *mirc.Message {
	return c.newMsg(opCode, "server", body)
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
			if opCode == mirc.CONNECTION_FAILURE {
				fmt.Printf("Cannot connect: %s\n", msg.Body)
				c.changeNick(setNick())
			}
			fmt.Printf("Connected\n")
			break
		}
	}
	return err
}

// closeConnection sends a close connection request to the server
func (c *client) closeConnection() {
	reqMsg := c.newServMsg(mirc.CONNECTION_CLOSED, "")
	c.Socket.SendMsg(reqMsg)
	return
}

// Sets nickname locally
func setNick() string {
	fmt.Print("Input your nickname:")
	reader := bufio.NewReader(os.Stdin)
	nick, _ := reader.ReadString('\n')
	nick = strings.Replace(nick, "\n", "", -1)
	for len(nick) <= 0 {
		fmt.Print("Input a valid nickname:")
		nick, _ = reader.ReadString('\n')
		nick = strings.Replace(nick, "\n", "", -1)
	}
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
		if opCode == mirc.CONNECTION_SUCCESS {
			c.Nick = nick
		} else {
			fmt.Printf("Error: %s\n", msg.Body)
		}
	}
	return
}

// listRoom lists all rooms in the server
func (c *client) listRoom() error {
	reqMsg := c.newServMsg(mirc.CLIENT_LIST_ROOM, "")
	return c.Socket.SendMsg(reqMsg)
}

// listMember lists all members in a room
func (c *client) listMember(room string) error {
	reqMsg := c.newServMsg(mirc.CLIENT_LIST_MEMBER, room)
	return c.Socket.SendMsg(reqMsg)
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

// changeRoom if input room exists in the server and the client is
// a memeber to that room then change the current room to that
func (c *client) changeRoom(room string) error {
	reqMsg := c.newServMsg(mirc.CLIENT_IN_ROOM, room)
	return c.Socket.SendMsg(reqMsg)
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
	msg := c.newMsg(mirc.CLIENT_SEND_PUB_MESSAGE, c.Room, msgBody)
	return c.Socket.SendMsg(msg)
}

/*********** Helper functions ************/
// Get configuration setup from file
func getConf(config *conf) {
	configFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("Cannot open configuration file %v ", err)
	}
	err = yaml.Unmarshal(configFile, config)
	if err != nil {
		log.Fatalf("Cannot parse configuration file: %v", err)
	}
	return
}

// periodically send ping to the server to notify this client is alive
func (c *client) keepAliveLoop() {
	keepAliveMsg := c.newServMsg(mirc.CONNECTION_PING, "ping")
	for {
		c.Socket.SendMsg(keepAliveMsg)
		time.Sleep(ping * time.Second)
	}
}

// command parser return command word and message
func comParser(cmdLine string) (string, string) {
	cmdLine = strings.TrimSpace(cmdLine)
	args := strings.SplitN(cmdLine, " ", 2)
	cmd := args[0]
	arg := ""
	if len(args) > 1 {
		arg = args[1]
	}
	return cmd, arg
}

/*********** UI functions ************/
func main() {
	config := conf{}
	getConf(&config)
	fmt.Printf("server: %s\n", config.Server)
	currentClient := newClient(config.Server)
	// Initialize Connection
	currentClient.requestToConnect()
	go currentClient.keepAliveLoop()
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()
	g.Cursor = true
	maxX, maxY := g.Size()
	if lv, err := g.SetView("view", 2, 1, maxX-2, maxY-8); err != nil {
		if err != gocui.ErrUnknownView {
			fmt.Printf("Error: %s\n", err)
		}
		lv.Title = currentClient.Room
		lv.Autoscroll = true
		lv.Wrap = true
		displayHelp(g)
	}

	if iv, err := g.SetView("input", 2, maxY-6, maxX-2, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			fmt.Printf("Error: %s\n", err)
		}
		iv.Title = currentClient.Nick
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
	if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, currentClient.inputFunc); err != nil {
		log.Panicln(err)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go g.MainLoop()
	go currentClient.msgHandlerLoop(g)

	wg.Wait()
	os.Exit(0)
}

// display help message
func displayHelp(g *gocui.Gui) {
	g.Execute(func(g *gocui.Gui) error {
		v, err := g.View("view")
		if err != nil {
			return err
		}
		helpMsg := "\nUSAGE EXAMPLE:\n" +
			"create a room:          \\create roomName\n" +
			"join a room:            \\join roomName\n" +
			"list all rooms:         \\listRoom\n" +
			"change current room:    \\changeRoom roomName\n" +
			"list members of a room: \\listMember roomName\n" +
			"leave a room:           \\leave roomName\n" +
			"send private message:   @nick message\n" +
			"display this message:   \\help\n" +
			"exit:                   \\exit\n"

		fmt.Fprintf(v, helpMsg)
		return nil
	})
	return
}

// message handler loop
func (c *client) msgHandlerLoop(g *gocui.Gui) {
	for {
		c.msgHandler(g)
	}
}

// Handles an incoming message
func (c *client) msgHandler(g *gocui.Gui) int {
	c.Socket.SetDeadline(mirc.CalDeadline(timeout))
	opCode, msg := c.Socket.GetMsg()

	if opCode == mirc.ERROR {
		g.Close()
		fmt.Print("Server connection has lost...Client exited\n")
		os.Exit(0)
	} else if opCode == mirc.SERVER_BROADCAST_MESSAGE {
		g.Execute(func(g *gocui.Gui) error {
			v, err := g.View("view")
			if err != nil {
				return err
			}
			fmt.Fprintf(v, "\n%s [%s] %s: %s\n", mirc.GetTime(), msg.Header.Receiver, msg.Header.Sender, msg.Body)
			return nil
		})
	} else if opCode == mirc.SERVER_TELL_MESSAGE {
		g.Execute(func(g *gocui.Gui) error {
			v, err := g.View("view")
			if err != nil {
				return err
			}
			fmt.Fprintf(v, "\n%s [PRIVATE] %s: %s\n", mirc.GetTime(), msg.Header.Sender, msg.Body)
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
			fmt.Fprintf(v, "Members: %s\n", msg.Body)
			return nil
		})
	} else if opCode == mirc.SERVER_RPL_CLIENT_IN_ROOM {
		c.Room = msg.Body
		g.Execute(func(g *gocui.Gui) error {
			v, err := g.View("view")
			if err != nil {
				return err
			}
			fmt.Fprintf(v, "current Room: %s\n", c.Room)
			v.Title = c.Room
			return nil
		})
	}
	return 0
}

// handles commands from user input
func (c *client) inputFunc(g *gocui.Gui, iv *gocui.View) error {
	v, _ := g.View("input")
	cmdLine := iv.ViewBuffer()
	v.Clear()
	v.SetCursor(0, 0)
	cmd, arg := comParser(cmdLine)
	if len(cmd) == 0 { // ignore empty input
		return nil
	}

	if cmd == "\\create" { // create a chat room
		c.createRoom(arg)
	} else if cmd == "\\join" { // join a chat room
		c.joinRoom(arg)
	} else if cmd == "\\listRoom" { // list all char rooms on a server
		c.listRoom()
	} else if cmd == "\\changeRoom" { // change current chat room
		c.changeRoom(arg)
	} else if cmd == "\\listMember" { // list membership of a room
		c.listMember(arg)
	} else if cmd == "\\leave" { // leave room
		c.leaveRoom(arg)
		c.Room = "public"
	} else if cmd[0] == '@' { // Private user message
		g.Execute(func(g *gocui.Gui) error {
			v, err := g.View("view")
			if err != nil {
				return err
			}
			fmt.Fprintf(v, "\n%s [PRIVATE] %s: %s\n", mirc.GetTime(), c.Nick, arg)
			return nil
		})
		c.sendPrivateMsg(cmd[1:], arg)
	} else if cmd == "\\help" { // display help info
		displayHelp(g)
	} else if cmd == "\\exit" { // exits the client
		g.Close()
		c.closeConnection()
		fmt.Printf("Bye!\n")
		os.Exit(0)
	} else { // send public message
		c.sendPubMsg(cmdLine)
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
