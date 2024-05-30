package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPathKey(t *testing.T) {
	pk := PathKey{
		pathName: "a/b/c",
		fileName: "file.txt",
	}

	require.Equal(t, "a/b/c/file.txt", pk.FullPath())

	require.Equal(t, "a", pk.FirstDirectoryFromPath())

	emptyPK := PathKey{}

	require.Equal(t, "", emptyPK.FirstDirectoryFromPath())

}
