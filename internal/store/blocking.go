package store

import (
	"sync"
	"time"
)

type BlockingStore struct {
	mutex sync.RWMutex
	store ExpiringStore
}

func NewBlockingStore(store ExpiringStore) *BlockingStore {
	return &BlockingStore{
		store: store,
	}
}

func (b *BlockingStore) Put(key string, value string, ttl time.Duration) (KeyValueEntry, bool, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.store.Put(key, value, ttl)
}

func (b *BlockingStore) Get(key string) (KeyValueEntry, bool, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.store.Get(key)
}

func (b *BlockingStore) Delete(key string) (bool, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.store.Delete(key)
}

func (b *BlockingStore) Size() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.store.Size()
}

func (b *BlockingStore) CleanupExpired() int {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.store.CleanupExpired()
}
