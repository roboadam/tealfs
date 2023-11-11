package node

// import (
// 	"net"
// 	"tealfs/pkg/raw_net"
// 	"tealfs/pkg/tnet"
// )

// type RemoteNode struct {
// 	Base Node
// 	tNet tnet.TNet
// 	conn net.Conn
// }

// type UnknownRemoteNode struct {
// 	address Address
// 	tNet tnet.TNet
// }

// func NewUnknownRemoteNode(address Address, tNet tnet.TNet) *UnknownRemoteNode {
// 	return &UnknownRemoteNode{
// 		address: address,
// 		tNet: tNet,
// 	}
// }

// func (u *UnknownRemoteNode) IntoConnectedKnown() *RemoteNode {
// 	conn := u.tNet.Dial(u.address.value)
// 	conn.Write(HelloToBytes())
// 	address := NewAddress(conn.RemoteAddr().String())
// 	return &RemoteNode{
// 		Base: Node{ Address: address,  }
// 	}
// }

// func NewRemoteNode(base Node, tNet tnet.TNet) *RemoteNode {
// 	return &RemoteNode{Base: base, tNet: tNet}
// }

// func (r *RemoteNode) Connect() {
// 	r.conn = r.tNet.Dial(r.Base.Address.value)
// }

// func (r *RemoteNode) SendHello(id Id) {
// 	_ = raw_net.Int8To(r.conn, 1)
// 	_ = raw_net.UInt32To(r.conn, uint32(len(id.String())))
// 	_ = raw_net.StringTo(r.conn, id.String())
// }

// func (r *RemoteNode) Disconnect() {
// 	r.tNet.Close()
// }
