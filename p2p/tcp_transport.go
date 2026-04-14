package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// TCP Perr represents the remote node over established connection
type TCPPeer struct {

	//underlying connection of the peer, in this case a tcp connection
	net.Conn

	//if dial and retrieve a conn => outbound=true
	//if accept and retieve a conn => outbound = false
	outbound bool

	Wg *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		Wg:       &sync.WaitGroup{},
	}
}

type TcpTransportOpts struct {
	ListenAddr    string
	HandShakeFunc HandShakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TcpTransport struct {
	TcpTransportOpts
	listner net.Listener
	rpcch   chan RPC

	// maintaining peers is not a responsiblity of the TCP transport,should be hand over to server
	//mu      sync.RWMutex
	//peers   map[net.Addr]Peer
}

func NewTCPTransport(opts TcpTransportOpts) *TcpTransport {
	return &TcpTransport{
		TcpTransportOpts: opts,
		rpcch:            make(chan RPC, 1024),
	}
}

// Consume implements the transport interface, which will return read only channel.
func (t *TcpTransport) Consume() <-chan RPC {
	return t.rpcch
}

// close implements the transport interface
func (t *TcpTransport) Close() error {
	return t.listner.Close()
}

// Close implements the peer interface
func (p *TCPPeer) Close() error {
	return p.Conn.Close()
}

// RemoteAddr implements the peer interface and will return the remote
// address of its undelying connection of the peer.
func (p *TCPPeer) RemoteAddr() net.Addr {
	return p.Conn.RemoteAddr()
}

func (t *TcpTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	//since this connection is initiating from the server, boolian is true
	go t.handleConn(conn, true)

	return nil
}

func (t *TcpTransport) ListenAndAccept() error {
	var err error
	t.listner, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	log.Printf("TCP transport listening on port :%s\n", t.ListenAddr)
	return nil
}

func (t *TcpTransport) startAcceptLoop() {
	for {
		conn, err := t.listner.Accept()
		// if connection is closed return
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		// this is accepting outside connection, hense the value is false
		go t.handleConn(conn, false)
	}
}

func (t *TcpTransport) handleConn(con net.Conn, isOutBound bool) {
	var err error

	defer func() {
		fmt.Printf("dropping peer connection: %s", err)
		con.Close()
	}()

	peer := NewTCPPeer(con, true)

	if err := t.HandShakeFunc(peer); err != nil {
		//con.Close()
		//fmt.Printf("TCP handshake error %s\n", err)
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			//Instead of droping the peer for all the errors, do this only for specific errors
			return
		}
	}

	rpc := RPC{}
	for {
		if err := t.Decoder.Decode(con, &rpc); err != nil {
			fmt.Printf("TCP error: %s\n", rpc)
			continue
		}
		//fmt.Printf("message: %+v\n", rpc)
		rpc.From = con.RemoteAddr().String()

		if rpc.Stream {
			peer.Wg.Add(1)
			fmt.Printf("[%s] incoming stream, waiting ", con.RemoteAddr())
			peer.Wg.Wait()
			fmt.Printf("[%s] stream closed,resuming the read loop", con.RemoteAddr())
			continue
		}
		t.rpcch <- rpc
	}

}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}
