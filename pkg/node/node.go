package node

import (
	"tealfs/pkg/cmds"
	"tealfs/pkg/raw_net"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type NodeId struct {
	value uuid.UUID
}

func NewNodeId() NodeId {
	uuid, err := uuid.NewUUID()
	if err != nil {
		fmt.Println("Error generating UUID:", err)
		os.Exit(1)
	}

	return NodeId{
		value: uuid,
	}
}

type Node struct {
	nodeId      NodeId
	userCmds    chan cmds.User
	listener    net.Listener
	connections *Connections
}

func NewNode(userCmds chan cmds.User) Node {
	return Node{
		nodeId:      NewNodeId(),
		userCmds:    userCmds,
		connections: NewConnections(),
	}
}

func (node Node) listen() {
	if node.listener == nil {
		var keepTryingToListen = true

		for keepTryingToListen {
			var listenErr error
			node.listener, listenErr = net.Listen("tcp", ":0")
			if listenErr != nil {
				fmt.Println("Error listening:", listenErr.Error())
				time.Sleep(2 * time.Second)
			}
		}
	}
}

func (node Node) Start() {
	node.listen()
	defer node.listener.Close()

	go node.handleUiCommands()
	go node.keepConnectionsAlive()

	for {
		conn, err := node.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			return
		}
		fmt.Println("Accepted connection from", conn.RemoteAddr())
		go node.handleConnection(conn)
	}
}

func (node Node) handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		intFromConn, _ := raw_net.IntFrom(conn)
		fmt.Println("Received:", intFromConn)
	}
}

func (node Node) keepConnectionsAlive() {
	for {
		time.Sleep(2 * time.Second)
		node.connections.ConnectAll()
	}
}

func (node Node) handleUiCommands() {
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

func (node Node) addConnection(cmd cmds.User) {
	conn := NodeConnection{
		Address: cmd.Argument,
		Conn:    nil,
	}

	id := node.connections.AddConnection(conn)
	fmt.Println("Received command: add-connnection, address:" + cmd.Argument + ", added connection id:" + strconv.Itoa(int(id.Value)))
}

func (node Node) addStorage(cmd cmds.User) {
	fmt.Println("Received command: add-storage, location:" + cmd.Argument)
}
