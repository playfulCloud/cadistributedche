package store

import (
	"context"
	"log/slog"
	"time"
)

type Store interface {
	Put(key string, value string, ttl time.Duration) (KeyValueEntry, bool, error)
	Get(key string) (KeyValueEntry, bool, error)
	Delete(key string) (bool, error)
	Size() int
}

type ExpiringStore interface {
	Store
	CleanupExpired() int
}

type KeyValueStore struct {
	storage map[string]KeyValueEntry
	clock   Clock
	ttl     time.Duration
}

func NewKeyValueStore(clock Clock, ttl time.Duration) *KeyValueStore {
	return &KeyValueStore{
		storage: make(map[string]KeyValueEntry),
		clock:   clock,
		ttl:     ttl,
	}
}

func (k *KeyValueStore) Put(key string, value string, ttl time.Duration) (KeyValueEntry, bool, error) {
	previousEntry, exists := k.storage[key]
	entryTtl := ttl
	if ttl == 0 {
		entryTtl = k.ttl
	}

	entry := KeyValueEntry{
		key:       key,
		value:     value,
		createdAt: k.clock.Now(),
		ttl:       entryTtl,
	}

	k.storage[key] = entry

	if !exists || previousEntry.isExpired(k.clock.Now()) {
		return KeyValueEntry{}, false, nil
	}

	return previousEntry, true, nil

}

func (k *KeyValueStore) Get(key string) (KeyValueEntry, bool, error) {
	entry, exists := k.storage[key]

	if !exists || entry.isExpired(k.clock.Now()) {
		return KeyValueEntry{}, false, nil
	}

	return entry, true, nil

}

func (k *KeyValueStore) Delete(key string) (bool, error) {
	entry, exists := k.storage[key]
	if !exists || entry.isExpired(k.clock.Now()) {
		return false, nil
	}

	delete(k.storage, key)
	return true, nil
}

func (k *KeyValueStore) Size() int {
	return len(k.storage)
}

func (k *KeyValueStore) CleanupExpired() int {
	startedAt := time.Now()
	removed := 0

	for key, value := range k.storage {
		if value.isExpired(k.clock.Now()) {
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
	return removed
}
