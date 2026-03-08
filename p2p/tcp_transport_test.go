package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	tcpOptions := TcpTransportOpts{
		ListenAddr:    ":3000",
		HandShakeFunc: NOPHandShakeFunc,
		Decoder:       DefaultDecoder{},
	}
	tr := NewTCPTransport(tcpOptions)

	assert.Equal(t, tr.ListenAddr, ":3000 ")

	//Server
	//tr.Accept( )
	assert.Nil(t, tr.ListenAndAccept())
}
