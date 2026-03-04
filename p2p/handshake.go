package p2p

// Handshake func is
type HandShakeFunc func(Peer) error

func NOPHandShakeFunc(Peer) error { return nil }
