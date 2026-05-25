package metrics

import "sync/atomic"

type MetricsCollector struct {
	cacheHits      atomic.Uint64
	cacheMisses    atomic.Uint64
	cacheDeletes   atomic.Uint64
	cacheWrites    atomic.Uint64
	cacheTotalKeys atomic.Uint64
}

type CacheMetrics struct {
	CacheHits      uint64 `json:"cacheHits"`
	CacheMisses    uint64 `json:"cacheMisses"`
	CacheDeletes   uint64 `json:"cacheDeletes"`
	CacheWrites    uint64 `json:"cacheWrites"`
	CacheTotalKeys uint64 `json:"cacheTotalKeys"`
}

type CacheMetricsCollector interface {
	IncreaseCacheHits()
	IncreaseCacheMisses()
	IncreaseCacheDeletes()
	IncreaseCacheWrites()
	SetCacheTotalKeys(totalKeys uint64)
}

type CacheMetricsReader interface {
	GetMetrics() CacheMetrics
}

func (m *MetricsCollector) IncreaseCacheHits() {
	m.cacheHits.Add(1)
}

func (m *MetricsCollector) IncreaseCacheMisses() {
	m.cacheMisses.Add(1)
}

func (m *MetricsCollector) IncreaseCacheDeletes() {
	m.cacheDeletes.Add(1)
}

func (m *MetricsCollector) IncreaseCacheWrites() {
	m.cacheWrites.Add(1)
}

func (m *MetricsCollector) SetCacheTotalKeys(totalKeys uint64) {
	m.cacheTotalKeys.Store(totalKeys)
}

func (m *MetricsCollector) GetMetrics() CacheMetrics {
	return CacheMetrics{
		CacheHits:      m.cacheHits.Load(),
		CacheMisses:    m.cacheMisses.Load(),
		CacheDeletes:   m.cacheDeletes.Load(),
		CacheWrites:    m.cacheWrites.Load(),
		CacheTotalKeys: m.cacheTotalKeys.Load(),
	}
}
