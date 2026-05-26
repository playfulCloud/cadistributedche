package store

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/playfulCloud/cadistributedche/internal/metrics"
)

var (
	olderDate = time.Date(
		2024,
		time.January,
		10,
		12, 0, 0, 0,
		time.UTC,
	)

	newerDate = time.Date(
		2026,
		time.May,
		21,
		15, 30, 0, 0,
		time.UTC,
	)
	ttl = 50 * time.Second
)

func TestConcurrentReadsWrites(t *testing.T) {
	store := NewBlockingStore(NewKeyValueStore(provideFakeClock(), ttl))

	store.Put("cloud", "playful", 0)
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(2)
		i := i
		go func(i int) {
			defer wg.Done()
			store.Put("cloud", strconv.Itoa(i), 0)
		}(i)

		go func() {
			defer wg.Done()
			store.Get("cloud")
		}()
	}
	wg.Wait()
}

func TestConcurrentWrites(t *testing.T) {
	store := NewBlockingStore(NewKeyValueStore(provideFakeClock(), ttl))

	store.Put("cloud", "playful", 0)
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		i := i
		go func(i int) {
			defer wg.Done()
			store.Put("cloud", strconv.Itoa(i), 0)
		}(i)

	}
	wg.Wait()
}

func TestConcurrentDeletes(t *testing.T) {
	store := NewBlockingStore(NewKeyValueStore(provideFakeClock(), ttl))
	populateStore(store)
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		i := i
		go func(i int) {
			defer wg.Done()
			store.Delete(strconv.Itoa(i))
		}(i)

	}
	wg.Wait()
}

func TestConcurrentWritesDeletes(t *testing.T) {
	store := NewBlockingStore(NewKeyValueStore(provideFakeClock(), ttl))
	populateStore(store)
	var wg sync.WaitGroup

	for i := 100; i < 200; i++ {
		wg.Add(2)
		i := i
		di := i - 100
		go func(i int) {
			defer wg.Done()
			store.Put(strconv.Itoa(i), strconv.Itoa(i), 0)
		}(i)

		go func(di int) {
			defer wg.Done()
			store.Delete(strconv.Itoa(di))
		}(di)

	}
	wg.Wait()
}

func TestConcurrentWritesDeletesFinalState(t *testing.T) {
	store := NewBlockingStore(NewKeyValueStore(provideFakeClock(), ttl))
	populateStore(store)

	var wg sync.WaitGroup

	for i := 100; i < 200; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			store.Put(strconv.Itoa(i), strconv.Itoa(i), 0)
		}()

		di := i - 100
		go func() {
			defer wg.Done()
			store.Delete(strconv.Itoa(di))
		}()
	}

	wg.Wait()

	for i := range 100 {
		_, exists, _ := store.Get(strconv.Itoa(i))
		if exists {
			t.Fatalf("expected key %d to be deleted", i)
		}
	}

	for i := 100; i < 200; i++ {
		value, exists, _ := store.Get(strconv.Itoa(i))
		if !exists {
			t.Fatalf("expected key %d to exist", i)
		}
		if value.Value() != strconv.Itoa(i) {
			t.Fatalf("expected value %d, got %s", i, value.Value())
		}
	}
}

func TestPutReturnsPreviousValue(t *testing.T) {
	metrics := &metrics.FakeMetricsCollector{}
	store := NewMetricsStore(NewKeyValueStore(provideFakeClock(), ttl), metrics)

	prev, existed, _ := store.Put("key", "first", 0)
	if prev.Value() != "" {
		t.Fatalf("expected empty previous value, got %s", prev.Value())
	}
	if existed {
		t.Fatal("expected key to be new")
	}

	prev, existed, _ = store.Put("key", "second", 0)
	if prev.Value() != "first" {
		t.Fatalf("expected previous value first, got %s", prev.Value())
	}
	if !existed {
		t.Fatal("expected key to exist")
	}

	writes := metrics.Writes
	totalKeys := metrics.Keys

	if writes != 2 {
		t.Fatalf("expected cache writes to be 2, got %d", writes)
	}
	if totalKeys != 1 {
		t.Fatalf("expected total keys to be 1, got %d", totalKeys)
	}
}

func TestPutDetectsExistingEmptyValue(t *testing.T) {
	metrics := &metrics.FakeMetricsCollector{}
	store := NewMetricsStore(NewKeyValueStore(provideFakeClock(), ttl), metrics)

	_, existed, _ := store.Put("key", "", 0)
	if existed {
		t.Fatal("expected key to be new")
	}

	prev, existed, _ := store.Put("key", "second", 0)
	if prev.Value() != "" {
		t.Fatalf("expected empty previous value, got %s", prev.Value())
	}
	if !existed {
		t.Fatal("expected key to exist")
	}
	writes := metrics.Writes
	totalKeys := metrics.Keys

	if writes != 2 {
		t.Fatalf("expected cache writes to be 2, got %d", writes)
	}
	if totalKeys != 1 {
		t.Fatalf("expected total keys to be 1, got %d", totalKeys)
	}
}

func TestDeleteReturnsFound(t *testing.T) {
	metrics := &metrics.FakeMetricsCollector{}
	store := NewMetricsStore(NewKeyValueStore(provideFakeClock(), ttl), metrics)

	store.Put("key", "value", 0)

	deleted, _ := store.Delete("key")
	if !deleted {
		t.Fatal("expected key to be deleted")
	}

	_, exists, _ := store.Get("key")
	if exists {
		t.Fatal("expected key to be deleted")
	}
	deletes := metrics.Deletes
	totalKeys := metrics.Keys

	if deletes != 1 {
		t.Fatalf("expected cache deletes to be 1, got %d", deletes)
	}
	if totalKeys != 0 {
		t.Fatalf("expected total keys to be 0, got %d", totalKeys)
	}
}

func TestGetMissingKey(t *testing.T) {
	metrics := &metrics.FakeMetricsCollector{}
	store := NewMetricsStore(NewKeyValueStore(provideFakeClock(), ttl), metrics)

	value, exists, _ := store.Get("missing")
	if value.Value() != "" {
		t.Fatalf("expected empty value, got %s", value.Value())
	}
	if exists {
		t.Fatal("expected key to be missing")
	}

	misses := metrics.Misses
	if misses != 1 {
		t.Fatalf("expected cache misses to be 1, got %d", misses)
	}
}

func TestGetExistingEmptyValue(t *testing.T) {
	metrics := &metrics.FakeMetricsCollector{}
	store := NewMetricsStore(NewKeyValueStore(provideFakeClock(), ttl), metrics)

	store.Put("key", "", 0)

	value, exists, _ := store.Get("key")
	if value.Value() != "" {
		t.Fatalf("expected empty value, got %s", value.Value())
	}
	if !exists {
		t.Fatal("expected key to exist")
	}

	hits := metrics.Hits
	if hits != 1 {
		t.Fatalf("expected cache hits to be 1, got %d", hits)
	}
}

func TestGetExpiredKeyShouldReturnNothing(t *testing.T) {
	metrics := &metrics.FakeMetricsCollector{}
	store := &KeyValueStore{
		storage: storageWithExpiredEntries(),
		clock:   provideFakeClock(),
		ttl:     time.Duration(30) * time.Second,
	}
	storeWithMetrics := NewMetricsStore(store, metrics)

	value, exists, _ := storeWithMetrics.Get("user:1")
	if value.Value() != "" {
		t.Fatalf("expected empty value, got %s", value.Value())
	}
	if exists {
		t.Fatal("expected key to not exist")
	}

	misses := metrics.Misses
	if misses != 1 {
		t.Fatalf("expected cache misses to be 1, got %d", misses)
	}
}

func TestPutWithCustomTTLOverridesDefaultTTL(t *testing.T) {
	clock := &FakeClock{
		FixedTime: newerDate,
	}
	metrics := &metrics.FakeMetricsCollector{}
	store := NewMetricsStore(NewKeyValueStore(clock, time.Hour), metrics)

	store.Put("key", "value", time.Second)

	clock.FixedTime = newerDate.Add(2 * time.Second)

	value, exists, _ := store.Get("key")
	if value.Value() != "" {
		t.Fatalf("expected empty value, got %s", value.Value())
	}
	if exists {
		t.Fatal("expected key to be expired")
	}

	writes := metrics.Writes
	totalKeys := metrics.Keys

	if writes != 1 {
		t.Fatalf("expected cache writes to be 1, got %d", writes)
	}
	if totalKeys != 1 {
		t.Fatalf("expected total keys to be 1, got %d", totalKeys)
	}
}

func TestPutExpiredKeyShouldReturnNothing(t *testing.T) {
	metrics := &metrics.FakeMetricsCollector{}
	store := &KeyValueStore{
		storage: storageWithExpiredEntries(),
		clock:   provideFakeClock(),
		ttl:     time.Duration(30) * time.Second,
	}
	storeWithMetrics := NewMetricsStore(store, metrics)

	value, exists, _ := storeWithMetrics.Put("user:1", "cloud", 0)

	if value.Value() != "" {
		t.Fatalf("expected empty value, got %s", value.Value())
	}
	if exists {
		t.Fatal("expected key to not exist")
	}

	writes := metrics.Writes
	totalKeys := metrics.Keys

	if writes != 1 {
		t.Fatalf("expected cache writes to be 1, got %d", writes)
	}
	if totalKeys != 3 {
		t.Fatalf("expected total keys to be 3, got %d", totalKeys)
	}
}

func TestDeleteExpiredKeyShouldReturnFalse(t *testing.T) {
	store := &KeyValueStore{
		storage: storageWithExpiredEntries(),
		clock:   provideFakeClock(),
		ttl:     time.Duration(30) * time.Second,
	}

	exists, _ := store.Delete("user:1")

	if exists {
		t.Fatal("expected key to not exist")
	}
}

func TestCleanupExpired(t *testing.T) {
	metrics := &metrics.FakeMetricsCollector{}
	store := &KeyValueStore{
		storage: storageWithExpiredEntries(),
		clock:   provideFakeClock(),
		ttl:     time.Duration(30) * time.Second,
	}
	storeWithMetrics := NewMetricsStore(store, metrics)

	storeWithMetrics.CleanupExpired()
	storageSize := len(store.storage)
	if storageSize != 0 {
		t.Fatalf("expected storage to be empty after clean but got %d elements", storageSize)
	}

	expired := metrics.Expired
	if expired != 3 {
		t.Fatalf("expected expired keys to be 3, but got %d", expired)
	}
}

func BenchmarkGet(b *testing.B) {
	store := &KeyValueStore{
		storage: storageWithExpiredEntries(),
		clock:   provideFakeClock(),
		ttl:     time.Duration(30) * time.Second,
	}

	store.Put("playful", "cloud", 0)

	for b.Loop() {
		store.Get("playful")
	}

}

func BenchmarkPutOverrides(b *testing.B) {
	store := &KeyValueStore{
		storage: storageWithExpiredEntries(),
		clock:   provideFakeClock(),
		ttl:     time.Duration(30) * time.Second,
	}
	b.ResetTimer()

	for b.Loop() {
		store.Put("playful", "cloud", 0)
	}
}

func BenchmarkPut(b *testing.B) {
	store := &KeyValueStore{
		storage: storageWithExpiredEntries(),
		clock:   provideFakeClock(),
		ttl:     time.Duration(30) * time.Second,
	}

	for i := 0; b.Loop(); i++ {
		store.Put(strconv.Itoa(i), "cloud", 0)
	}
}

func populateStore(store Store) {
	for i := range 100 {
		store.Put(strconv.Itoa(i), strconv.Itoa(i), 0)
	}
}

func storageWithExpiredEntries() map[string]KeyValueEntry {
	return map[string]KeyValueEntry{
		"user:1": {
			key:       "user:1",
			value:     "playful",
			createdAt: olderDate,
			ttl:       30 * time.Second,
		},
		"user:2": {
			key:       "user:2",
			value:     "cloud",
			createdAt: olderDate,
			ttl:       30 * time.Second,
		},
		"user:3": {
			key:       "user:3",
			value:     "unemployed",
			createdAt: olderDate,
			ttl:       30 * time.Second,
		},
	}
}

type FakeClock struct {
	FixedTime time.Time
}

func (f FakeClock) Now() time.Time {
	return f.FixedTime
}

func provideFakeClock() FakeClock {
	return FakeClock{
		FixedTime: newerDate,
	}
}
