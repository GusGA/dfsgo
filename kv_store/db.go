package kvstore

type KVStore[K any, V any] interface {
	Set(K, V)
	Get(K) (V, bool)
	Delete(K)
}
