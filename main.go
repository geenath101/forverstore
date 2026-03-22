package main

import (
	"fmt"
	"log"

	"github.com/geenath101/forverstore/p2p"
)

func OnPeer(p2p.Peer) error {
	fmt.Printf("doing some logic with the peer outside of TCPTransport")
	return nil
}

func main() {
	tcpOpts := p2p.TcpTransportOpts{
		ListenAddr:    ":3000",
		HandShakeFunc: p2p.NOPHandShakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tr := p2p.NewTCPTransport(tcpOpts)

	fileServerOpts := FileServerOpts{
		StorageRoot:       "3000_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         *tr,
	}

	s := NewFileServer(fileServerOpts)

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}

	select {}
}
