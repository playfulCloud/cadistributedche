package store

import (
	"errors"
	"testing"
	"time"

	"github.com/playfulCloud/cadistributedche/internal/metrics"
)

func TestMetricsStorePutDoesNotCollectMetricsOnError(t *testing.T) {
	expectedErr := errors.New("put failed")
	collector := &metrics.FakeMetricsCollector{}
	store := NewMetricsStore(&FakeStore{
		PutFunc: func(key string, value string, ttl time.Duration) (KeyValueEntry, bool, error) {
			return KeyValueEntry{}, false, expectedErr
		},
		SizeFunc: func() int {
			return 10
		},
	}, collector)

	_, _, err := store.Put("key", "value", 0)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected put error %v, got %v", expectedErr, err)
	}
	if collector.Writes != 0 {
		t.Fatalf("expected writes to remain 0, got %d", collector.Writes)
	}
	if collector.Keys != 0 {
		t.Fatalf("expected keys to remain 0, got %d", collector.Keys)
	}
}

func TestMetricsStorePutCollectsMetricsOnSuccess(t *testing.T) {
	collector := &metrics.FakeMetricsCollector{}
	store := NewMetricsStore(&FakeStore{
		SizeFunc: func() int {
			return 1
		},
	}, collector)

	_, _, err := store.Put("key", "value", 0)
	if err != nil {
		t.Fatalf("expected put to succeed, got %v", err)
	}
	if collector.Writes != 1 {
		t.Fatalf("expected writes to be 1, got %d", collector.Writes)
	}
	if collector.Keys != 1 {
		t.Fatalf("expected keys to be 1, got %d", collector.Keys)
	}
}

func TestMetricsStoreGetDoesNotCollectMetricsOnError(t *testing.T) {
	expectedErr := errors.New("get failed")
	collector := &metrics.FakeMetricsCollector{}
	store := NewMetricsStore(&FakeStore{
		GetFunc: func(key string) (KeyValueEntry, bool, error) {
			return KeyValueEntry{}, false, expectedErr
		},
	}, collector)

	_, _, err := store.Get("key")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected get error %v, got %v", expectedErr, err)
	}
	if collector.Hits != 0 {
		t.Fatalf("expected hits to remain 0, got %d", collector.Hits)
	}
	if collector.Misses != 0 {
		t.Fatalf("expected misses to remain 0, got %d", collector.Misses)
	}
}

func TestMetricsStoreGetCollectsHitAndMiss(t *testing.T) {
	collector := &metrics.FakeMetricsCollector{}
	store := NewMetricsStore(&FakeStore{}, collector)

	_, found, err := store.Get("exists")
	if err != nil {
		t.Fatalf("expected get to succeed, got %v", err)
	}
	if !found {
		t.Fatal("expected key to be found")
	}

	_, found, err = store.Get("empty")
	if err != nil {
		t.Fatalf("expected get to succeed, got %v", err)
	}
	if found {
		t.Fatal("expected key to be missing")
	}

	if collector.Hits != 1 {
		t.Fatalf("expected hits to be 1, got %d", collector.Hits)
	}
	if collector.Misses != 1 {
		t.Fatalf("expected misses to be 1, got %d", collector.Misses)
	}
}

func TestMetricsStoreDeleteDoesNotCollectMetricsOnError(t *testing.T) {
	expectedErr := errors.New("delete failed")
	collector := &metrics.FakeMetricsCollector{}
	store := NewMetricsStore(&FakeStore{
		DeleteFunc: func(key string) (bool, error) {
			return false, expectedErr
		},
		SizeFunc: func() int {
			return 10
		},
	}, collector)

	_, err := store.Delete("key")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected delete error %v, got %v", expectedErr, err)
	}
	if collector.Deletes != 0 {
		t.Fatalf("expected deletes to remain 0, got %d", collector.Deletes)
	}
	if collector.Misses != 0 {
		t.Fatalf("expected misses to remain 0, got %d", collector.Misses)
	}
	if collector.Keys != 0 {
		t.Fatalf("expected keys to remain 0, got %d", collector.Keys)
	}
}

func TestMetricsStoreDeleteCollectsMetricsFromDeleteResult(t *testing.T) {
	collector := &metrics.FakeMetricsCollector{}
	size := 1
	store := NewMetricsStore(&FakeStore{
		DeleteFunc: func(key string) (bool, error) {
			if key == "exists" {
				size = 0
				return true, nil
			}
			return false, nil
		},
		SizeFunc: func() int {
			return size
		},
	}, collector)

	deleted, err := store.Delete("exists")
	if err != nil {
		t.Fatalf("expected delete to succeed, got %v", err)
	}
	if !deleted {
		t.Fatal("expected key to be deleted")
	}

	deleted, err = store.Delete("missing")
	if err != nil {
		t.Fatalf("expected delete to succeed, got %v", err)
	}
	if deleted {
		t.Fatal("expected key to be missing")
	}

	if collector.Deletes != 1 {
		t.Fatalf("expected deletes to be 1, got %d", collector.Deletes)
	}
	if collector.Misses != 1 {
		t.Fatalf("expected misses to be 1, got %d", collector.Misses)
	}
	if collector.Keys != 0 {
		t.Fatalf("expected keys to be 0, got %d", collector.Keys)
	}
}

func TestMetricsStoreCleanupExpiredCollectsRemovedAmountAndKeys(t *testing.T) {
	collector := &metrics.FakeMetricsCollector{}
	store := NewMetricsStore(&FakeStore{
		CleanupExpiredFunc: func() int {
			return 3
		},
		SizeFunc: func() int {
			return 2
		},
	}, collector)

	removed := store.CleanupExpired()
	if removed != 3 {
		t.Fatalf("expected removed entries to be 3, got %d", removed)
	}
	if collector.Expired != 3 {
		t.Fatalf("expected expired to be 3, got %d", collector.Expired)
	}
	if collector.Keys != 2 {
		t.Fatalf("expected keys to be 2, got %d", collector.Keys)
	}
}
