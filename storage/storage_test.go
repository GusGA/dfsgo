package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestCASPathTransformFunc(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want PathKey
	}{
		{
			name: "create a pathkey with hashed name",
			args: args{
				"awesome_file.txt",
			},
			want: PathKey{
				fileName: "522c2cd75b27e2a78896b08e0bc690424139d607",
				pathName: "522c2/cd75b/27e2a/78896/b08e0/bc690/42413/9d607",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pathKey := CASPathTransformFunc(tt.args.key)

			assert.Equal(t, pathKey.fileName, tt.want.fileName)
			assert.Equal(t, pathKey.pathName, tt.want.pathName)

		})
	}
}
