package main

import (
	"io"

	"github.com/geenath101/forverstore/p2p"
)

type FileServerOpts struct {
	ListenAddr        string
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.TcpTransport
	TCPTransportOpts  p2p.TcpTransportOpts
}

type FileServer struct {
	FileServerOpts
	store *Store
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
	}
}

func (s *FileServerOpts) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	return nil
}

func (s *FileServer) Store(key string, r io.Reader) error {
	return s.Store.Write(key, r)
}
