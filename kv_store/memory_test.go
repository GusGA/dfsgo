package kvstore

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInMemoryKVStore_Get(t *testing.T) {
	t.Parallel()

	kv := NewInMemoryKVStore[string, string]()

	kv.data = map[string]string{
		"127.0.0.1:3000": "alive",
		"127.0.0.1:4000": "alive",
		"127.0.0.1:5000": "dead",
		"127.0.0.1:6000": "dead",
	}

	for i := 3; i <= 6; i++ {
		key := fmt.Sprintf("127.0.0.1:%d000", i)

		go func() {
			val, ok := kv.Get(key)
			require.True(t, ok)
			require.NotEmpty(t, val)
		}()

	}

}

func TestInMemoryKVStore_Set(t *testing.T) {
	t.Parallel()

	kv := NewInMemoryKVStore[string, string]()

	values := []string{"alive", "dead"}

	for i := 1; i <= 20; i++ {
		key := fmt.Sprintf("127.0.0.1:%d000", i)

		val := values[i%2]
		go func() {
			kv.Set(key, val)
		}()

	}
}

func TestInMemoryKVStore_Delete(t *testing.T) {
	t.Parallel()
	wg := &sync.WaitGroup{}
	kv := NewInMemoryKVStore[string, string]()

	kv.data = map[string]string{
		"127.0.0.1:3000": "alive",
		"127.0.0.1:4000": "alive",
		"127.0.0.1:5000": "dead",
		"127.0.0.1:6000": "dead",
	}
	wg.Add(4)
	for i := 3; i <= 6; i++ {
		key := fmt.Sprintf("127.0.0.1:%d000", i)
		go func(key string) {
			kv.Delete(key)
			wg.Done()
		}(key)
	}

	wg.Wait()

	for i := 3; i <= 6; i++ {
		key := fmt.Sprintf("127.0.0.1:%d000", i)

		_, ok := kv.Get(key)
		require.False(t, ok)
	}

}
