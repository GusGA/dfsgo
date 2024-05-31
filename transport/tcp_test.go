package transport

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTCPTransport(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.Nil(t, err)

	opts := TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: TCPNoHandshakeFunc,
		Logger:        logger,
	}
	tr := NewTCPTransport(opts)
	assert.Equal(t, tr.ListenAddr, ":3000")

	assert.Nil(t, tr.ListenAndAccept())
}
