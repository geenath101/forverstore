package main

import (
	"fmt"
	"log"

	"github.com/geenath101/forverstore/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {

	tcpOpts := p2p.TcpTransportOpts{
		ListenAddr:    listenAddr,
		HandShakeFunc: p2p.NOPHandShakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		// TODO onPeer func
	}

	tcpTransport := p2p.NewTCPTransport(tcpOpts)

	fileServerOpts := FileServerOpts{
		StorageRoot:       listenAddr + "_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         *tcpTransport,
		BootstrapNodes:    []string{":4000"},
	}

	return NewFileServer(fileServerOpts)

}

func OnPeer(p2p.Peer) error {
	fmt.Printf("doing some logic with the peer outside of TCPTransport")
	return nil
}

func main() {

	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")

	go func() {
		log.Fatal(s1.Start())
	}()

	s2.Start()

}
