package metrics

type FakeMetricsReader struct {
	GetStatsFunc func() CacheStats
}

func (f *FakeMetricsReader) GetStats() CacheStats {
	if f.GetStatsFunc != nil {
		return f.GetStatsFunc()
	}

	return CacheStats{
		Keys:    5,
		Hits:    1,
		Misses:  2,
		Expired: 6,
		Deletes: 3,
		Writes:  4,
	}
}

type FakeMetricsCollector struct {
	Hits    int
	Misses  int
	Deletes int
	Writes  int
	Keys    uint64
	Expired int
}

func (f *FakeMetricsCollector) IncreaseHits() {
	f.Hits++
}

func (f *FakeMetricsCollector) IncreaseMisses() {
	f.Misses++
}

func (f *FakeMetricsCollector) IncreaseDeletes() {
	f.Deletes++
}

func (f *FakeMetricsCollector) IncreaseWrites() {
	f.Writes++
}

func (f *FakeMetricsCollector) SetKeys(keys uint64) {
	f.Keys = keys
}

func (f *FakeMetricsCollector) IncreaseExpired() {
	f.Expired++
}
