package storage

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
)

// type StorageOptions

func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blocksize := 5
	sliceLen := len(hashStr) / blocksize
	paths := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blocksize, (i*blocksize)+blocksize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		pathName: strings.Join(paths, "/"),
		fileName: hashStr,
	}
}

type PathTransformFunc func(string) PathKey

type PathKey struct {
	pathName string
	fileName string
}

func (pk *PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", pk.pathName, pk.fileName)
}

func (pk *PathKey) FirstDirectoryFromPath() string {
	paths := strings.Split(pk.pathName, "/")
	return paths[0]
}

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		pathName: key,
		fileName: key,
	}
}

type StorageOpts struct {
	Root              string
	PathTransformFunc PathTransformFunc
	Logger            *zap.Logger
}

type Storage struct {
	StorageOpts
}

func NewStorage(opts StorageOpts) *Storage {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}

	return &Storage{
		StorageOpts: opts,
	}
}

func (s *Storage) HasFile(serverID, key string) bool {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, serverID, pathKey.FullPath())

	_, err := os.Stat(fullPathWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Storage) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Storage) Delete(serverID string, key string) error {
	pathKey := s.PathTransformFunc(key)

	defer func() {
		message := fmt.Sprintf("deleted [%s] from disk", pathKey.fileName)
		s.Logger.Info(message, zap.String("server_id", serverID))
	}()

	firstPathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, serverID, pathKey.FirstDirectoryFromPath())

	return os.RemoveAll(firstPathNameWithRoot)
}

func (s *Storage) openFileForWriting(serverID, key string) (*os.File, error) {
	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, serverID, pathKey.pathName)
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return nil, err
	}

	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, serverID, pathKey.FullPath())

	return os.Create(fullPathWithRoot)
}
