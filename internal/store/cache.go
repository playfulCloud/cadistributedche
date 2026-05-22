package store

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Store interface {
	Put(key string, value string) (string, bool, error)
	Get(key string) (string, bool, error)
	Delete(key string) (bool, error)
}

type ExpiringStore interface {
	CleanupExpired()
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

func NewKeyValueStore(clock Clock, ttl time.Duration) *KeyValueStore {
	return &KeyValueStore{
		storage: make(map[string]KeyValueEntry),
		clock:   clock,
		ttl:     ttl,
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
	if !exists || k.isExpired(previousEntry.createdAt) {
		return "", false, nil
	}
	return previousEntry.value, true, nil

}

func (k *KeyValueStore) Get(key string) (string, bool, error) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	entry, exists := k.storage[key]
	if !exists || k.isExpired(entry.createdAt) {
		return "", false, nil
	}

	return entry.value, exists, nil

}

func (k *KeyValueStore) Delete(key string) (bool, error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	entry, exists := k.storage[key]
	if !exists || k.isExpired(entry.createdAt) {
		return false, nil
	}
	delete(k.storage, key)
	return true, nil
}

func (k *KeyValueStore) CleanupExpired() {
	startedAt := time.Now()
	removed := 0

	k.mutex.Lock()
	defer k.mutex.Unlock()

	for key, value := range k.storage {
		if k.isExpired(value.createdAt) {
			delete(k.storage, key)
			removed++
		}
	}

	level := slog.LevelDebug
	if removed > 0 {
		level = slog.LevelInfo
	}

	slog.Log(
		context.Background(),
		level,
		"expired entries cleanup completed",
		"removed", removed,
		"remaining", len(k.storage),
		"duration_ms", time.Since(startedAt).Milliseconds(),
	)
}

func (k *KeyValueStore) isExpired(createdAt time.Time) bool {
	t := k.clock.Now()
	diff := t.Sub(createdAt)
	return diff >= k.ttl
}
