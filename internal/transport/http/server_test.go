package http

import (
	"context"
	"testing"
	"time"

	"github.com/playfulCloud/cadistributedche/internal/config"
)

func TestServerShouldShutDown(t *testing.T) {
	storageHandler := NewStorageHandler(&FakeStore{})
	server := NewServer(storageHandler, testServerConfig())

	errCh := make(chan error, 1)

	go func() {
		errCh <- server.Run()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		t.Fatalf("Server should shut down without error, but got %v", err)
	}

	select {
	case e := <-errCh:
		if e != nil {
			t.Fatalf("Error channel should be empty during successful shutdown, but got %v", e)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("Server should shut down within the given time range, but it did not")

	}
}

func testServerConfig() config.ServerConfig {
	return config.ServerConfig{
		Port: 8080,
		Timeouts: config.TimeoutConfig{
			Read:     5 * time.Second,
			Write:    10 * time.Second,
			Idle:     120 * time.Second,
			Shutdown: 15 * time.Second,
		},
	}
}
