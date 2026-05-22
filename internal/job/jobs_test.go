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
		t.Fatalf("Expected ttl to be run withing 5 second but did not")
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
		t.Fatalf("Expected ttl to be run withing 5 second but did not")
	}
}
