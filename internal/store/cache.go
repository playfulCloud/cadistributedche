package store

import (
	"sync"
)

type Store interface {
	Put(key string, value string) (string, error)
	Get(key string) (string, error)
	Delete(key string) (string, error)
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

func (k *KeyValueStore) Put(key string, value string) (string, error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	previousValue, exists := k.storage[key]
	k.storage[key] = value
	if !exists {
		return "", nil
	}
	return previousValue, nil

}

func (k *KeyValueStore) Get(key string) (string, error) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	value, exists := k.storage[key]
	if !exists {
		return "", nil
	}
	return value, nil

}

func (k *KeyValueStore) Delete(key string) (string, error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	_, exists := k.storage[key]
	if !exists {
		return "", nil
	}
	delete(k.storage, key)
	return key, nil
}
