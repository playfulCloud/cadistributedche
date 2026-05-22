package store

import (
	"strconv"
	"sync"
	"testing"
	"time"
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
	store := NewKeyValueStore(provideFakeClock(), ttl)

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
	store := NewKeyValueStore(provideFakeClock(), ttl)

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
	store := NewKeyValueStore(provideFakeClock(), ttl)
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
	store := NewKeyValueStore(provideFakeClock(), ttl)
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
	store := NewKeyValueStore(provideFakeClock(), ttl)
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
		if value != strconv.Itoa(i) {
			t.Fatalf("expected value %d, got %s", i, value)
		}
	}
}

func TestPutReturnsPreviousValue(t *testing.T) {
	store := NewKeyValueStore(provideFakeClock(), ttl)

	prev, existed, _ := store.Put("key", "first", 0)
	if prev != "" {
		t.Fatalf("expected empty previous value, got %s", prev)
	}
	if existed {
		t.Fatal("expected key to be new")
	}

	prev, existed, _ = store.Put("key", "second", 0)
	if prev != "first" {
		t.Fatalf("expected previous value first, got %s", prev)
	}
	if !existed {
		t.Fatal("expected key to exist")
	}
}

func TestPutDetectsExistingEmptyValue(t *testing.T) {
	store := NewKeyValueStore(provideFakeClock(), ttl)

	_, existed, _ := store.Put("key", "", 0)
	if existed {
		t.Fatal("expected key to be new")
	}

	prev, existed, _ := store.Put("key", "second", 0)
	if prev != "" {
		t.Fatalf("expected empty previous value, got %s", prev)
	}
	if !existed {
		t.Fatal("expected key to exist")
	}
}

func TestDeleteReturnsFound(t *testing.T) {
	store := NewKeyValueStore(provideFakeClock(), ttl)

	store.Put("key", "value", 0)

	deleted, _ := store.Delete("key")
	if !deleted {
		t.Fatal("expected key to be deleted")
	}

	_, exists, _ := store.Get("key")
	if exists {
		t.Fatal("expected key to be deleted")
	}
}

func TestGetMissingKey(t *testing.T) {
	store := NewKeyValueStore(provideFakeClock(), ttl)

	value, exists, _ := store.Get("missing")
	if value != "" {
		t.Fatalf("expected empty value, got %s", value)
	}
	if exists {
		t.Fatal("expected key to be missing")
	}
}

func TestGetExistingEmptyValue(t *testing.T) {
	store := NewKeyValueStore(provideFakeClock(), ttl)

	store.Put("key", "", 0)

	value, exists, _ := store.Get("key")
	if value != "" {
		t.Fatalf("expected empty value, got %s", value)
	}
	if !exists {
		t.Fatal("expected key to exist")
	}
}

func TestGetExpiredKeyShouldReturnNothing(t *testing.T) {
	store := &KeyValueStore{
		storage: storageWithExpiredEntries(),
		clock:   provideFakeClock(),
		ttl:     time.Duration(30) * time.Second,
	}

	value, exists, _ := store.Get("user:1")
	if value != "" {
		t.Fatalf("expected empty value, got %s", value)
	}
	if exists {
		t.Fatal("expected key to not exist")
	}
}

func TestPutWithCustomTTLOverridesDefaultTTL(t *testing.T) {
	clock := &FakeClock{
		FixedTime: newerDate,
	}
	store := NewKeyValueStore(clock, time.Hour)

	store.Put("key", "value", time.Second)

	clock.FixedTime = newerDate.Add(2 * time.Second)

	value, exists, _ := store.Get("key")
	if value != "" {
		t.Fatalf("expected empty value, got %s", value)
	}
	if exists {
		t.Fatal("expected key to be expired")
	}
}

func TestPutExpiredKeyShouldReturnNothing(t *testing.T) {
	store := &KeyValueStore{
		storage: storageWithExpiredEntries(),
		clock:   provideFakeClock(),
		ttl:     time.Duration(30) * time.Second,
	}

	value, exists, _ := store.Put("user:1", "cloud", 0)

	if value != "" {
		t.Fatalf("expected empty value, got %s", value)
	}
	if exists {
		t.Fatal("expected key to not exist")
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
	store := &KeyValueStore{
		storage: storageWithExpiredEntries(),
		clock:   provideFakeClock(),
		ttl:     time.Duration(30) * time.Second,
	}

	store.CleanupExpired()
	storageSize := len(store.storage)
	if storageSize != 0 {
		t.Fatalf("Expected storage to be empty after clean but got %d elements", storageSize)
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
