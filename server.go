package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/geenath101/forverstore/p2p"
)

type FileServerOpts struct {
	ListenAddr        string
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.TcpTransport
	TCPTransportOpts  p2p.TcpTransportOpts
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	store  *Store
	quitch chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p
	log.Printf("connected with the remote %s", p.RemoteAddr())
	return nil
}

func (s *FileServer) loop() {

	defer func() {
		log.Println("file server stopped due to user quit action")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Fatal(err)
			}

			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println(err)
				return
			}
			// fmt.Printf("%+v\n", msg.Payload)

			// peer, ok := s.peers[rpc.From]

			// if !ok {
			// 	panic("peer not found peer map")
			// }

			// b := make([]byte, 1000)
			// if _, err := peer.Read(b); err != nil {
			// 	panic(err)
			// }
			// fmt.Printf("%s\n", string(b))

			// //should fix this later
			// peer.(*p2p.TCPPeer).Wg.Done()

			// do something here
			//if err := s.handleMessage(&m); err != nil {
			//	log.Println(err)
			//}
		case <-s.quitch:
			return
		}

	}
}

type Message struct {
	From    string
	Payload any
}

type DataMessage struct {
	Key  string
	Data []byte
}

type MessageStoreFile struct {
	Key  string
	Size int64
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	// 1. store this file to the disk
	// 2. broadcast this file to all the known peers in the network

	// commenting out the simple implementation infavour of streaming implementation.
	// buf := new(bytes.Buffer)
	// tee := io.TeeReader(r, buf)

	// if err := s.store.Write(key, tee); err != nil {
	// 	return err
	// }
	// p := &DataMessage{
	// 	Key:  key,
	// 	Data: buf.Bytes(),
	// }

	// fmt.Println(buf.Bytes())

	// return s.broadCast(&Message{
	// 	From:    "todo",
	// 	Payload: p,
	// })

	fileBuffer := new(bytes.Buffer)
	tee := io.TeeReader(r, fileBuffer)
	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}

	msgBuf := new(bytes.Buffer)
	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}
	if err := gob.NewEncoder(msgBuf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		if err := peer.Send(msgBuf.Bytes()); err != nil {
			return err
		}
	}

	time.Sleep(time.Second * 3)

	//payload := []byte("This is a large file")
	for _, peer := range s.peers {
		// if err := peer.Send(payload); err != nil {
		// 	return err
		// }
		n, err := io.Copy(peer, fileBuffer)
		if err != nil {
			return err
		}
		fmt.Println("received and written bytes to disk", n)
	}
	return nil
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		//fmt.Printf("received data %v\n",v)
		return s.handleMessagesStoreFile(from, v)
	}

	return nil

}

func (s *FileServer) handleMessagesStoreFile(from string, msg MessageStoreFile) error {
	//fmt.Printf("%+v\n", msg)
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list ", from)
	}
	if err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size)); err != nil {
		return err
	}
	peer.(*p2p.TCPPeer).Wg.Done()

	return nil

}

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {

		if len(addr) == 0 {
			continue
		}

		// s.Transport
		go func(addr string) {
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("dial error ", err)
			}
		}(addr)
	}

	return nil
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	s.loop()
	return nil
}

// type Payload struct {
// 	Key  string
// 	Data []byte
// }

// func (s *FileServer) broadCast(p Payload) error {
// 	buf := new(bytes.Buffer)
// 	for _, peer := range s.peers {
// 		if err := gob.NewEncoder(buf).Encode(p); err != nil {
// 			return err
// 		}
// 		peer.Send(buf.Bytes())
// 	}
// 	return nil
// }

// double check the implementation here
func (s *FileServer) broadCast(p Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(p)
}

func init() {
	gob.Register(MessageStoreFile{})
}
