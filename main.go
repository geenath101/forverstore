package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/geenath101/forverstore/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {

	tcpOpts := p2p.TcpTransportOpts{
		ListenAddr:    listenAddr,
		HandShakeFunc: p2p.NOPHandShakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}

	tcpTransport := p2p.NewTCPTransport(tcpOpts)

	fileServerOpts := FileServerOpts{
		StorageRoot:       listenAddr + "_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         *tcpTransport,
		BootstrapNodes:    []string{":4000"},
	}
	s := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s

}

func OnPeer(p2p.Peer) error {
	fmt.Printf("doing some logic with the peer outside of TCPTransport")
	return nil
}

func main() {

	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")

	go func() {
		log.Print("Starting server s1")
		log.Fatal(s1.Start())
	}()

	time.Sleep(1 * time.Second)
	log.Print("Starting server s2")
	log.Fatal(s2.Start())

	time.Sleep(1 * time.Second)

	data := bytes.NewReader([]byte("my big data file"))
	s2.StoreData("privatedata", data)

}
