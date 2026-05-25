package metrics

import "sync/atomic"

type MetricsCollector struct {
	hits    atomic.Uint64
	misses  atomic.Uint64
	deletes atomic.Uint64
	writes  atomic.Uint64
	keys    atomic.Uint64
	expired atomic.Uint64
}

type CacheStats struct {
	Keys    uint64 `json:"keys"`
	Hits    uint64 `json:"hits"`
	Misses  uint64 `json:"misses"`
	Expired uint64 `json:"expired"`
	Deletes uint64 `json:"deletes"`
	Writes  uint64 `json:"writes"`
}

type CacheStatsCollector interface {
	IncreaseHits()
	IncreaseMisses()
	IncreaseDeletes()
	IncreaseWrites()
	SetKeys(keys uint64)
	IncreaseExpired()
}

type CacheStatsReader interface {
	GetStats() CacheStats
}

func (m *MetricsCollector) IncreaseHits() {
	m.hits.Add(1)
}

func (m *MetricsCollector) IncreaseMisses() {
	m.misses.Add(1)
}

func (m *MetricsCollector) IncreaseDeletes() {
	m.deletes.Add(1)
}

func (m *MetricsCollector) IncreaseWrites() {
	m.writes.Add(1)
}

func (m *MetricsCollector) SetKeys(keys uint64) {
	m.keys.Store(keys)
}

func (m *MetricsCollector) IncreaseExpired() {
	m.expired.Add(1)
}

func (m *MetricsCollector) GetStats() CacheStats {
	return CacheStats{
		Keys:    m.keys.Load(),
		Hits:    m.hits.Load(),
		Misses:  m.misses.Load(),
		Expired: m.expired.Load(),
		Deletes: m.deletes.Load(),
		Writes:  m.writes.Load(),
	}
}
