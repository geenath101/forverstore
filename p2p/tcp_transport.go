package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCP Perr represents the remote node over established connection
type TCPPeer struct {

	//conn is the underlying connection of the peer
	conn net.Conn

	//if dial and retrieve a conn => outbound=true
	//if accept and retieve a conn => outbound = false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TcpTransportOpts struct {
	ListenAddr    string
	HandShakeFunc HandShakeFunc
	Decoder       Decoder
}

type TcpTransport struct {
	TcpTransportOpts
	listner net.Listener
	rpcch   chan RPC
	mu      sync.RWMutex
	peers   map[net.Addr]Peer
}

func NewTCPTransport(opts TcpTransportOpts) *TcpTransport {
	return &TcpTransport{
		TcpTransportOpts: opts,
		rpcch:            make(chan RPC),
	}
}

// Consume implements the transport interface, which will return read only channel.
func (t *TcpTransport) Consume() <-chan RPC {
	return t.rpcch
}

func (t *TcpTransport) ListenAndAccept() error {
	var err error
	t.listner, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()
	return nil
}

func (t *TcpTransport) startAcceptLoop() {
	for {
		conn, err := t.listner.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		go t.handleConn(conn)
	}
}

func (t *TcpTransport) handleConn(con net.Conn) {
	peer := NewTCPPeer(con, true)

	if err := t.HandShakeFunc(peer); err != nil {
		con.Close()
		fmt.Printf("TCP handshake error %s\n", err)
		return
	}
	rpc := &RPC{}
	for {
		if err := t.Decoder.Decode(con, rpc); err != nil {
			fmt.Printf("TCP error: %s\n", rpc)
			continue
		}

		fmt.Printf("message: %+v\n", rpc)
	}

}
