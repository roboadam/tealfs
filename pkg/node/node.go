package node

import (
	"errors"
	"fmt"
	"net"
	"tealfs/pkg/cmds"
	"tealfs/pkg/raw_net"
	"time"
)

type Node struct {
	Id          Id
	userCmds    chan cmds.User
	listener    net.Listener
	connections *RemoteNodes
	hostToBind  string
}

func NewNode(userCmds chan cmds.User) Node {
	node := Node{
		Id:          NewNodeId(),
		userCmds:    userCmds,
		connections: NewRemoteNodes(),
		hostToBind:  "",
	}

	go node.handleUiCommands()
	go node.acceptConnections()

	return node
}

func (node *Node) SetHostToBind(hostToBind string) {
	if node.hostToBind != hostToBind {
		node.hostToBind = hostToBind
		node.applyListenerChanges()
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
		_ = node.listener.Close()
	}
}

func (node *Node) applyListenerChanges() {
	if node.IsListening() {
		node.setListener()
	}
}

func (node *Node) acceptConnections() {
	for {
		node.acceptAndHandleConnection()
	}
}

func (node *Node) Listen() {
	if !node.IsListening() {
		node.setListener()
	}
}

func (node *Node) IsListening() bool {
	return node.listener != nil
}

func (node *Node) setListener() {
	listenErr := errors.New("")
	for listenErr != nil {
		node.listener, listenErr = listenOrSleepOnError("tcp", node.hostToBind+":0")
	}
}

func listenOrSleepOnError(network string, address string) (net.Listener, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		time.Sleep(2 * time.Second)
	}
	return listener, err
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
	conn := RemoteNode{
		NodeId:  node.Id,
		Address: cmd.Argument,
		Conn:    nil,
	}

	node.connections.AddConnection(conn)
	fmt.Println("Received command: add-connection, address:" + cmd.Argument + ", added connection id:" + conn.NodeId.value.String())
}

func (node *Node) addStorage(cmd cmds.User) {
	fmt.Println("Received command: add-storage, location:" + cmd.Argument)
}
