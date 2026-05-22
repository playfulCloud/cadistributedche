package job

import (
	"context"
	"testing"
	"time"
)

func TestRunTTL(t *testing.T) {
	ctx := t.Context()
	sent := make(chan struct{})

	go func() {
		RunTTL(ctx, func() {
			sent <- struct{}{}
		}, 1*time.Millisecond)
	}()

	select {
	case <-sent:
	case <-time.After(5 * time.Second):
		t.Fatalf("Expected TTL to run within 5 seconds, but it did not")
	}
}

func TestRunTTLClosingJob(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	cancel()
	go func() {
		defer close(done)
		RunTTL(ctx, func() {
		}, 1*time.Millisecond)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatalf("Expected TTL to stop within 5 seconds, but it did not")
	}
}
