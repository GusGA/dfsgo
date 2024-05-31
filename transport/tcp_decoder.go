package transport

import "io"

const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)

type TCPDecoder struct{}

func (dec TCPDecoder) Decode(r io.Reader, msg *RPC) error {
	peekBuff := make([]byte, 1)
	_, err := r.Read(peekBuff)
	if err != nil {
		return err
	}

	// In case of a stream we are not decoding what is being sent over the network.
	// We are just setting Stream true so we can handle that in our logic.
	stream := peekBuff[0] == IncomingStream
	if stream {
		msg.Stream = stream
		return nil
	}

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}

	msg.Payload = buf[:n]

	return nil
}
