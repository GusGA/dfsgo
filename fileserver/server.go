package fileserver

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/gusga/dfsgo/discovery"
	kvstore "github.com/gusga/dfsgo/kv_store"
	"github.com/gusga/dfsgo/storage"
	"github.com/gusga/dfsgo/transport"
	"go.uber.org/zap"
)

type FileServerOpts struct {
	ID                string
	EncKey            []byte
	StorageRoot       string
	PathTransformFunc storage.PathTransformFunc
	Transport         transport.Transport
	Logger            *zap.Logger
	DiscoverySrv      discovery.DiscoveryService
	Storage           *storage.Storage
}

type FileServer struct {
	FileServerOpts
	peerStore kvstore.KVStore[string, transport.Peer]
	peers     map[string]transport.Peer
	quitch    chan struct{}
}

func NewServer(opts FileServerOpts) *FileServer {
	srv := &FileServer{
		FileServerOpts: opts,
		peerStore:      kvstore.NewInMemoryKVStore[string, transport.Peer](),
	}

	return srv
}

func (s *FileServer) Start() error {
	s.Logger.Info("starting fileserver...", zap.String("server_addr", s.Transport.Addr()))

	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.addSelfNode()

	s.bootstrapNetwork()

	s.loop()

	return nil
}

func (s *FileServer) OnPeer(p transport.Peer) error {
	remoteAddr := p.RemoteAddr().String()

	s.peerStore.Set(remoteAddr, p)

	s.Logger.Info("connecting to remote fileserver...", zap.String("remote_addr", remoteAddr))

	return nil
}

func (s *FileServer) Close() {
	s.Transport.Close()
	s.DiscoverySrv.Close()
}

func (s *FileServer) loop() {
	defer func() {
		s.Logger.Warn("file server stopped due to error or user quit action")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("decoding error: ", err)
				s.Logger.Error("decoding error", zap.Error(err))
			}
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				s.Logger.Error("handle message error", zap.Error(err))
			}
		case addr := <-s.Transport.ClosedPeer():
			s.Logger.Info("removing peer from list", zap.String("remote_peer_addr", addr))
			s.peerStore.Delete(addr)
		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		peer.SendData([]byte{transport.IncomingMessage})
		if err := peer.SendData(buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	}

	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	if !s.Storage.HasFile(msg.ID, msg.Key) {
		return fmt.Errorf("[%s] need to serve file (%s) but it does not exist on disk", s.Transport.Addr(), msg.Key)
	}

	fmt.Printf("[%s] serving file (%s) over the network\n", s.Transport.Addr(), msg.Key)

	fileSize, r, err := s.Storage.Read(msg.ID, msg.Key)
	if err != nil {
		return err
	}

	if rc, ok := r.(io.ReadCloser); ok {
		fmt.Println("closing readCloser")
		defer rc.Close()
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not in map", from)
	}

	// First send the "incomingStream" byte to the peer and then we can send
	// the file size as an int64.
	peer.SendData([]byte{transport.IncomingStream})
	binary.Write(peer, binary.LittleEndian, fileSize)
	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] written (%d) bytes over the network to %s\n", s.Transport.Addr(), n, from)

	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}

	n, err := s.Storage.Write(msg.ID, msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}

	fmt.Printf("[%s] written %d bytes to disk\n", s.Transport.Addr(), n)

	peer.CloseStream()

	return nil
}

func (s *FileServer) bootstrapNetwork() error {

	nodes, err := s.DiscoverySrv.GetNodes()
	if err != nil {
		s.Logger.Error("error retriving file server nodes in the network", zap.Error(err))
		return err
	}

	if len(nodes) == 0 {
		return nil
	}

	for _, node := range nodes {

		if node.Address == s.Transport.Addr() {
			continue
		}

		go func(addr string) {
			s.Logger.Info("attemping to connect with remote", zap.String("remote_addr", addr), zap.String("local_addr", s.Transport.Addr()))
			err := s.Transport.Dial(addr)
			if err != nil {
				s.Logger.Info("dial error", zap.Error(err), zap.String("remote_addr", addr))
				s.DiscoverySrv.RemoveDeadNode(discovery.Node{Address: addr})
			}
		}(node.Address)
	}
	return nil
}

func (s *FileServer) addSelfNode() {

	hostname, _ := os.Hostname()

	node := discovery.Node{
		CreatedAt: time.Now(),
		Address:   s.Transport.Addr(),
		Hostmane:  hostname,
	}

	err := s.DiscoverySrv.AddNode(node)
	if err != nil {
		s.Logger.Error("error adding node to discovery service", zap.Error(err))
	}
}
