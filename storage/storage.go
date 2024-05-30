package storage

import (
	"fmt"
	"strings"
)

// type StorageOptions

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

type Storage struct {
	Root string
}
