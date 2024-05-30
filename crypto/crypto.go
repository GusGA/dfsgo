package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
)

func HashKey(key string) string {
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}

func NewEncryptionKey() []byte {
	keyBuffer := make([]byte, 32)
	io.ReadFull(rand.Reader, keyBuffer)
	return keyBuffer
}

func EncryptContent(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, nil
	}
	iv := make([]byte, block.BlockSize()) // 16 bytes size

	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return 0, nil
	}

	// prepend the iv to the file
	_, err = dst.Write(iv)
	if err != nil {
		return 0, nil
	}

	stream := cipher.NewCTR(block, iv)

	return copyStream(stream, block.BlockSize(), src, dst)
}

func DecryptContent(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, nil
	}
	iv := make([]byte, block.BlockSize())
	_, err = src.Read(iv)
	if err != nil {
		return 0, nil
	}

	stream := cipher.NewCTR(block, iv)
	return copyStream(stream, block.BlockSize(), src, dst)
}

func copyStream(stream cipher.Stream, blockSize int, src io.Reader, dst io.Writer) (int, error) {
	buff := make([]byte, 1024*32)
	nw := blockSize

	for {
		n, err := src.Read(buff)
		if n > 0 {
			stream.XORKeyStream(buff, buff[:n])
			nn, err := dst.Write(buff[:n])
			if err != nil {
				return 0, nil
			}
			nw += nn
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, nil
		}
	}
	return nw, nil
}
