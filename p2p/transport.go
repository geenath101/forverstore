package p2p

import "net"

// Peer is an anything that represent the remote node
type Peer interface {
	//RemoteAddr() net.Addr
	//Close() error
	net.Conn
	Send([]byte) error
}

// Transport that is anything handles the communication between the nodes in the network
// this can be TCP,UDP,websockets etc
type Transport interface {
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
