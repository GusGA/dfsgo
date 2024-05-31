package transport

import (
	"errors"
	"net"
	"sync"

	"go.uber.org/zap"
)

// TCPPeer represent the remote node over TCP established connection.
type TCPPeer struct {
	net.Conn
	// if we dial and retrieve a conn => outbound == true
	// if we accept and retrieve a conn => outbound == false
	outbound bool
	wg       *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

func (p *TCPPeer) CloseStream() {
	p.wg.Done()
}

func (p *TCPPeer) SendData(data []byte) error {
	_, err := p.Conn.Write(data)
	return err
}

type TCPTransportOpts struct {
	ListenAddr    string
	Decoder       Decoder
	Logger        *zap.Logger
	HandshakeFunc HandshakeFunc
}

type TCPTransport struct {
	TCPTransportOpts
	OnPeer     func(Peer) error
	listener   net.Listener
	rpcch      chan RPC
	closedPeer chan string
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	transport := &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC, 1024),
		closedPeer:       make(chan string),
	}

	if opts.Decoder == nil {
		transport.Decoder = TCPDecoder{}
	}

	return transport
}

// Consume implements the Tranport interface, which will return read-only channel
// for reading the incoming messages received from another peer in the network.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

func (t *TCPTransport) ClosedPeer() <-chan string {
	return t.closedPeer
}

// Close implements the Transport interface.
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Addr implements the Transport interface return the address
// the transport is accepting connections.
func (t *TCPTransport) Addr() string {
	return t.ListenAddr
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	t.Logger.Info("TCP transport listening", zap.String("port", t.ListenAddr))

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			t.Logger.Error("TCP accept error", zap.Error(err))
		}

		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error
	defer func() {
		t.Logger.Error(
			"dropping peer connection",
			zap.Error(err),
			zap.String("peer_address", conn.RemoteAddr().String()),
		)
		conn.Close()
		t.closedPeer <- conn.RemoteAddr().String()
	}()

	peer := NewTCPPeer(conn, outbound)

	err = t.HandshakeFunc(peer)
	if err != nil {
		return
	}

	if t.OnPeer != nil {
		t.OnPeer(peer)
		if err != nil {
			return
		}
	}

	// Read Loop
	for {
		rcp := RPC{}
		err = t.Decoder.Decode(conn, &rcp)
		if err != nil {
			return
		}

		rcp.From = conn.RemoteAddr().String()

		if rcp.Stream {
			peer.wg.Add(1)
			t.Logger.Info("incoming stream, waiting...", zap.String("peer", conn.RemoteAddr().String()))

			peer.wg.Wait()
			t.Logger.Info("istream closed, resuming read loop.", zap.String("peer", conn.RemoteAddr().String()))
			continue
		}
	}
}
