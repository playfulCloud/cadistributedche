package store

import (
	"sync"
)

type Store interface {
	Put(key string, value string) (string, bool, error)
	Get(key string) (string, bool, error)
	Delete(key string) (bool, error)
}

type KeyValueStore struct {
	storage map[string]string
	mutex   sync.RWMutex
}

func NewKeyValueStore() *KeyValueStore {
	return &KeyValueStore{
		storage: make(map[string]string),
	}
}

func (k *KeyValueStore) Put(key string, value string) (string, bool, error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	previousValue, exists := k.storage[key]
	k.storage[key] = value
	return previousValue, exists, nil

}

func (k *KeyValueStore) Get(key string) (string, bool, error) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	value, exists := k.storage[key]
	return value, exists, nil

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
