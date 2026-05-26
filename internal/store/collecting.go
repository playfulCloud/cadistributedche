package store

import (
	"time"

	"github.com/playfulCloud/cadistributedche/internal/metrics"
)

type MetricsStore struct {
	store   ExpiringStore
	metrics metrics.CacheStatsCollector
}

func NewMetricsStore(store ExpiringStore, metrics metrics.CacheStatsCollector) *MetricsStore {
	return &MetricsStore{
		store:   store,
		metrics: metrics,
	}
}

func (m *MetricsStore) Put(key string, value string, ttl time.Duration) (KeyValueEntry, bool, error) {
	entry, exists, err := m.store.Put(key, value, ttl)
	if err != nil {
		return entry, exists, err
	}

	m.metrics.IncreaseWrites()
	m.metrics.SetKeys(uint64(m.Size()))

	return entry, exists, nil
}

func (m *MetricsStore) Get(key string) (KeyValueEntry, bool, error) {
	entry, exists, err := m.store.Get(key)
	if err != nil {
		return entry, exists, err
	}

	if !exists {
		m.metrics.IncreaseMisses()
	} else {
		m.metrics.IncreaseHits()
	}

	return entry, exists, nil
}

func (m *MetricsStore) Delete(key string) (bool, error) {
	deleted, err := m.store.Delete(key)
	if err != nil {
		return deleted, err
	}

	if deleted {
		m.metrics.IncreaseDeletes()
		m.metrics.SetKeys(uint64(m.store.Size()))
	} else {
		m.metrics.IncreaseMisses()
	}

	return deleted, nil
}

func (m *MetricsStore) Size() int {
	return m.store.Size()
}

func (m *MetricsStore) CleanupExpired() int {
	removed := m.store.CleanupExpired()
	m.metrics.IncreaseExpired(uint64(removed))
	m.metrics.SetKeys(uint64(m.store.Size()))
	return removed
}
