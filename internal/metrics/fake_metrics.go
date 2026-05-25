package metrics

type FakeMetricsReader struct {
	GetMetricsFunc func() CacheMetrics
}

func (f *FakeMetricsReader) GetMetrics() CacheMetrics {
	if f.GetMetricsFunc != nil {
		return f.GetMetricsFunc()
	}

	return CacheMetrics{
		CacheHits:      1,
		CacheMisses:    2,
		CacheDeletes:   3,
		CacheWrites:    4,
		CacheTotalKeys: 5,
	}
}
