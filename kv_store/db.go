package kvstore

type KVStore interface {
	Set(key, value string)
	Get(key string) (string, bool)
}
