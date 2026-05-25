package metrics

import (
	"sync"
	"testing"
)

func TestConcurrentMetricIncrement(t *testing.T) {
	m := &MetricsCollector{}

	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(2)
		i := i
		go func(i int) {
			defer wg.Done()
			m.IncreaseCacheWrites()
		}(i)

		go func() {
			defer wg.Done()
			m.IncreaseCacheWrites()
		}()
	}
	wg.Wait()

	if m.cacheWrites.Load() != 200 {
		t.Fatalf("Concurrent metric increments should be equal to 200 but got %d", m.cacheWrites.Load())
	}
}
