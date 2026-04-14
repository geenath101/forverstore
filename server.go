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

type MessageGetFile struct {
	Key string
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(key) {
		return s.store.Read(key)
	}

	fmt.Printf("dont have file (%s) locally, fetching from network\n", key)

	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	for _, peer := range s.peers {
		fmt.Println("receiving stream from peer:", peer.RemoteAddr())
		fileBuffer := new(bytes.Buffer)
		n, err := io.Copy(fileBuffer, peer)
		if err != nil {
			return nil, err
		}
		fmt.Println("received bytes over the network", n)
	}

	select {}

	return nil, nil
}

func (s *FileServer) Store(key string, r io.Reader) error {
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
	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return err
	}

	time.Sleep(time.Second * 3)

	//payload := []byte("This is a large file")
	//TODO: use a multiwriter here
	for _, peer := range s.peers {
		// if err := peer.Send(payload); err != nil {
		// 	return err
		// }
		peer.Send([]byte{p2p.IncomingStream})
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
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	}

	return nil

}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	if !s.store.Has(msg.Key) {
		return fmt.Errorf("file (%s) does not exist on the disk", msg.Key)
	}
	fmt.Println("got file and serving it over the wire")
	r, err := s.store.Read(msg.Key)

	if err != nil {
		return err
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not in map ", from)
	}

	n, err := io.Copy(peer, r)

	if err != nil {
		return err
	}

	fmt.Printf("written %d bytes over the network to %s\n", n, from)

	return nil
}

func (s *FileServer) handleMessagesStoreFile(from string, msg MessageStoreFile) error {
	//fmt.Printf("%+v\n", msg)
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list ", from)
	}
	if _, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size)); err != nil {
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
			fmt.Printf("connected to server at %v\n", addr)
		}(addr)
	}

	return nil
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	s.bootstrapNetwork()
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
func (s *FileServer) stream(p Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(p)
}

func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	gob.Register(MessageStoreFile{})
}
