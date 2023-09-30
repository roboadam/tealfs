package node

import (
	"fmt"
	"net"
	"strconv"
	"tealfs/pkg/cmds"
	"tealfs/pkg/raw_net"
	"time"
)

type Node struct {
	Id          NodeId
	userCmds    chan cmds.User
	listener    net.Listener
	connections *Connections
	hostToBind  string
}

func NewNode(userCmds chan cmds.User) Node {
	node := Node{
		Id:          NewNodeId(),
		userCmds:    userCmds,
		connections: NewConnections(),
		hostToBind:  "",
	}

	go node.handleUiCommands()
	go node.keepConnectionsAlive()
	go node.acceptConnections()

	return node
}

func (node *Node) SetHostToBind(hostToBind string) {
	if node.hostToBind != hostToBind {
		node.hostToBind = hostToBind
		if node.IsListening() {
			node.setListener()
		}
	}
}

func (node *Node) GetAddress() net.Addr {
	if node.listener != nil {
		return node.listener.Addr()
	}
	return nil
}

func (node *Node) Close() {
	if node.listener != nil {
		node.listener.Close()
	}
}

func (node *Node) acceptConnections() {
	for {
		node.acceptAndHandleConnection()
	}
}

func (node *Node) Listen() {
	if node.listener == nil {
		node.setListener()
	}
}

func (node *Node) IsListening() bool {
	return node.listener != nil
}

func (node *Node) setListener() {
	var listenErr error
	for {
		node.listener, listenErr = net.Listen("tcp", node.hostToBind+":0")
		if listenErr == nil {
			return
		}

		fmt.Println("Error listening:", listenErr.Error())
		time.Sleep(2 * time.Second)
	}
}

func (node *Node) acceptAndHandleConnection() {
	if node.listener != nil {
		conn, err := node.listener.Accept()
		if err == nil {
			defer conn.Close()
			fmt.Println("Accepted connection from", conn.RemoteAddr())
			go node.handleConnection(conn)
		} else {
			fmt.Println("Error accepting connection:", err.Error())
		}
	}
}

func (node *Node) handleConnection(conn net.Conn) {
	for {
		intFromConn, _ := raw_net.IntFrom(conn)
		fmt.Println("Received:", intFromConn)
	}
}

func (node *Node) keepConnectionsAlive() {
	for {
		time.Sleep(2 * time.Second)
		node.connections.ConnectAll()
	}
}

func (node *Node) handleUiCommands() {
	for {
		command := <-node.userCmds
		switch command.CmdType {
		case cmds.ConnectTo:
			node.addConnection(command)
		case cmds.AddStorage:
			node.addStorage(command)
		}
	}
}

func (node *Node) addConnection(cmd cmds.User) {
	conn := NodeConnection{
		Address: cmd.Argument,
		Conn:    nil,
	}

	id := node.connections.AddConnection(conn)
	fmt.Println("Received command: add-connnection, address:" + cmd.Argument + ", added connection id:" + strconv.Itoa(int(id.Value)))
}

func (node *Node) addStorage(cmd cmds.User) {
	fmt.Println("Received command: add-storage, location:" + cmd.Argument)
}
