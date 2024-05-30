package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
		want string
	}{
		{
			name: "Hash file name",
			key:  "my_awesome_file.txt",
			want: "1ee780784d987bef8f5248d1deee53d5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HashKey(tt.key)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestNewEncryptionKey(t *testing.T) {
	key := NewEncryptionKey()

	require.NotEmpty(t, len(key))
}

func TestEncryptionDecryption(t *testing.T) {
	content := "some important text taht should be encripted"

	unencryptedSrc := bytes.NewReader([]byte(content))
	encryptedDst := new(bytes.Buffer)
	encryptionKey := NewEncryptionKey()

	n, err := EncryptContent(encryptionKey, unencryptedSrc, encryptedDst)
	require.NoError(t, err)
	require.NotEmpty(t, n)

	assert.Equal(t, 16+len(content), n)

	newDest := new(bytes.Buffer)
	n, err = DecryptContent(encryptionKey, encryptedDst, newDest)
	require.NoError(t, err)
	require.NotEmpty(t, n)

	assert.Equal(t, newDest.String(), content)

}
