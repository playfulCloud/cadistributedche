package store

import (
	"sync"
	"time"
)

type Store interface {
	Put(key string, value string) (string, bool, error)
	Get(key string) (string, bool, error)
	Delete(key string) (bool, error)
}

type ExpiringStore interface {
	CleanUpExpired()
}

type KeyValueStore struct {
	storage map[string]KeyValueEntry
	mutex   sync.RWMutex
	clock   Clock
	ttl     time.Duration
}

type KeyValueEntry struct {
	key       string
	value     string
	createdAt time.Time
}

func NewKeyValueStore(clock Clock) *KeyValueStore {
	return &KeyValueStore{
		storage: make(map[string]KeyValueEntry),
		clock:   clock,
	}
}

func (k *KeyValueStore) Put(key string, value string) (string, bool, error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	previousEntry, exists := k.storage[key]
	k.storage[key] = KeyValueEntry{
		key:       key,
		value:     value,
		createdAt: k.clock.Now(),
	}
	return previousEntry.value, exists, nil

}

func (k *KeyValueStore) Get(key string) (string, bool, error) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	entry, exists := k.storage[key]
	return entry.value, exists, nil

}

func (k *KeyValueStore) Delete(key string) (bool, error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	_, exists := k.storage[key]
	if !exists {
		return false, nil
	}
	delete(k.storage, key)
	return true, nil
}

func (k *KeyValueStore) CleanUpExpired() {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	t := k.clock.Now()
	for key, value := range k.storage {
		diff := t.Sub(value.createdAt)
		if diff >= k.ttl {
			delete(k.storage, key)
		}
	}
}
