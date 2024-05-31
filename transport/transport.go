package transport

import (
	"io"
	"net"
)

type RPC struct {
	From    string
	Payload []byte
	Stream  bool
}

type Transport interface {
	Addr() string
	Dial(address string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	ClosedPeer() <-chan string
	Close() error
}

// Peer is an interface that represents the remote node.
type Peer interface {
	net.Conn
	SendData(data []byte) error
	CloseStream()
}

type Decoder interface {
	Decode(io.Reader, *RPC) error
}

type HandshakeFunc func(Peer) error
