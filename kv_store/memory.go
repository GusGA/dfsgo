package kvstore

import "sync"

type InMemoryKVStore[K comparable, V any] struct {
	data map[K]V
	mu   sync.RWMutex
}

func NewInMemoryKVStore[K comparable, V any]() *InMemoryKVStore[K, V] {
	return &InMemoryKVStore[K, V]{
		data: make(map[K]V),
	}
}

func (kv *InMemoryKVStore[K, V]) Get(key K) (V, bool) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	val, ok := kv.data[key]

	return val, ok
}

func (kv *InMemoryKVStore[K, V]) Set(key K, value V) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	kv.data[key] = value
}
