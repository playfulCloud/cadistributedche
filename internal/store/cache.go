package store

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/playfulCloud/cadistributedche/internal/metrics"
)

type Store interface {
	Put(key string, value string, ttl time.Duration) (string, bool, error)
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
	metrics metrics.CacheMetricsCollector
	ttl     time.Duration
}

type KeyValueEntry struct {
	key       string
	value     string
	createdAt time.Time
	ttl       time.Duration
}

func NewKeyValueStore(clock Clock, metrics metrics.CacheMetricsCollector, ttl time.Duration) *KeyValueStore {
	return &KeyValueStore{
		storage: make(map[string]KeyValueEntry),
		clock:   clock,
		ttl:     ttl,
		metrics: metrics,
	}
}

func (k *KeyValueStore) Put(key string, value string, ttl time.Duration) (string, bool, error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	previousEntry, exists := k.storage[key]
	entryTtl := ttl
	if ttl == 0 {
		entryTtl = k.ttl
	}

	k.storage[key] = KeyValueEntry{
		key:       key,
		value:     value,
		createdAt: k.clock.Now(),
		ttl:       entryTtl,
	}
	k.metrics.IncreaseCacheWrites()
	k.metrics.SetCacheTotalKeys(uint64(len(k.storage)))

	if !exists || k.isExpired(previousEntry) {
		return "", false, nil
	}

	return previousEntry.value, true, nil

}

func (k *KeyValueStore) Get(key string) (string, bool, error) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	entry, exists := k.storage[key]
	if !exists || k.isExpired(entry) {
		k.metrics.IncreaseCacheMisses()
		return "", false, nil
	}
	k.metrics.IncreaseCacheHits()

	return entry.value, exists, nil

}

func (k *KeyValueStore) Delete(key string) (bool, error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	entry, exists := k.storage[key]
	if !exists || k.isExpired(entry) {
		k.metrics.IncreaseCacheMisses()
		return false, nil
	}

	delete(k.storage, key)

	k.metrics.IncreaseCacheDeletes()
	k.metrics.SetCacheTotalKeys(uint64(len(k.storage)))

	return true, nil
}

func (k *KeyValueStore) CleanupExpired() {
	startedAt := time.Now()
	removed := 0

	k.mutex.Lock()
	defer k.mutex.Unlock()

	for key, value := range k.storage {
		if k.isExpired(value) {
			delete(k.storage, key)
			removed++
		}
	}

	level := slog.LevelDebug
	if removed > 0 {
		level = slog.LevelInfo
	}

	k.metrics.SetCacheTotalKeys(uint64(len(k.storage)))
	slog.Log(
		context.Background(),
		level,
		"expired entries cleanup completed",
		"removed", removed,
		"remaining", len(k.storage),
		"duration_ms", time.Since(startedAt).Milliseconds(),
	)
}

func (k *KeyValueStore) isExpired(entry KeyValueEntry) bool {
	if entry.ttl <= 0 {
		return false
	}

	t := k.clock.Now()
	diff := t.Sub(entry.createdAt)
	return diff >= entry.ttl
}
