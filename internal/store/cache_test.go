package store

import (
	"strconv"
	"sync"
	"testing"
)

func TestConcurrentReadsWrites(t *testing.T) {
	store := NewKeyValueStore()

	store.Put("cloud", "playful")
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(2)
		i := i
		go func(i int) {
			defer wg.Done()
			store.Put("cloud", strconv.Itoa(i))
		}(i)

		go func() {
			defer wg.Done()
			store.Get("cloud")
		}()
	}
	wg.Wait()
}

func TestConcurrentWrites(t *testing.T) {
	store := NewKeyValueStore()

	store.Put("cloud", "playful")
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		i := i
		go func(i int) {
			defer wg.Done()
			store.Put("cloud", strconv.Itoa(i))
		}(i)

	}
	wg.Wait()
}

func TestConcurrentDeletes(t *testing.T) {
	store := NewKeyValueStore()
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
	store := NewKeyValueStore()
	populateStore(store)
	var wg sync.WaitGroup

	for i := 100; i < 200; i++ {
		wg.Add(2)
		i := i
		di := i - 100
		go func(i int) {
			defer wg.Done()
			store.Put(strconv.Itoa(i), strconv.Itoa(i))
		}(i)

		go func(di int) {
			defer wg.Done()
			store.Delete(strconv.Itoa(di))
		}(di)

	}
	wg.Wait()
}

func TestConcurrentWritesDeletesFinalState(t *testing.T) {
	store := NewKeyValueStore()
	populateStore(store)

	var wg sync.WaitGroup

	for i := 100; i < 200; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			store.Put(strconv.Itoa(i), strconv.Itoa(i))
		}()

		di := i - 100
		go func() {
			defer wg.Done()
			store.Delete(strconv.Itoa(di))
		}()
	}

	wg.Wait()

	for i := range 100 {
		value, _ := store.Get(strconv.Itoa(i))
		if value != "" {
			t.Fatalf("expected key %d to be deleted", i)
		}
	}

	for i := 100; i < 200; i++ {
		value, _ := store.Get(strconv.Itoa(i))
		if value == "" {
			t.Fatalf("expected key %d to exist", i)
		}
		if value != strconv.Itoa(i) {
			t.Fatalf("expected value %d, got %s", i, value)
		}
	}
}

func TestPutReturnsPreviousValue(t *testing.T) {
	store := NewKeyValueStore()

	prev, _ := store.Put("key", "first")
	if prev != "" {
		t.Fatalf("expected empty previous value, got %s", prev)
	}

	prev, _ = store.Put("key", "second")
	if prev != "first" {
		t.Fatalf("expected previous value first, got %s", prev)
	}
}

func TestDeleteReturnsDeletedKey(t *testing.T) {
	store := NewKeyValueStore()

	store.Put("key", "value")

	deleted, _ := store.Delete("key")
	if deleted != "key" {
		t.Fatalf("expected deleted key, got %s", deleted)
	}

	value, _ := store.Get("key")
	if value != "" {
		t.Fatalf("expected key to be deleted, got %s", value)
	}
}

func TestGetMissingKey(t *testing.T) {
	store := NewKeyValueStore()

	value, _ := store.Get("missing")
	if value != "" {
		t.Fatalf("expected empty value, got %s", value)
	}
}

func populateStore(store Store) {
	for i := range 100 {
		store.Put(strconv.Itoa(i), strconv.Itoa(i))
	}
}
