package p2p

// Peer is an anything that represent the remote node
type Peer interface {
}

// Transport that is anything handles the communication between the nodes in the network
// this can be TCP,UDP,websockets etc
type Transport interface {
	ListenAndAccept() error
}
