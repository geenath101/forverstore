package p2p

import "net"

// Message represent any arbitary data that been sent over the each
// transport between two nodes in the network.
type RPC struct {
	From    net.Addr
	Payload []byte
}
