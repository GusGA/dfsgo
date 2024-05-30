package kvstore

import "sync"

type InMemoryKVStore struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewInMemoryKVStore() *InMemoryKVStore {
	return &InMemoryKVStore{
		data: make(map[string]string),
	}
}

func (kv *InMemoryKVStore) Get(key string) (string, bool) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	val, ok := kv.data[key]

	return val, ok
}

func (kv *InMemoryKVStore) Set(key, value string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	kv.data[key] = value
}
