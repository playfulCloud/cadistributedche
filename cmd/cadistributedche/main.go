package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/playfulCloud/cadistributedche/internal/config"
	"github.com/playfulCloud/cadistributedche/internal/job"
	"github.com/playfulCloud/cadistributedche/internal/store"
	"github.com/playfulCloud/cadistributedche/internal/transport/http"
)

func main() {
	appCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

	cfg, err := config.ReadConfig("configs/cadistributedche.yaml")
	if err != nil {
		log.Fatal("Error while reading the config file")
	}

	storage := store.NewKeyValueStore(&store.ClockProvider{}, cfg.Store.Ttl)
	storageHandler := http.NewStorageHandler(storage)
	server := http.NewServer(storageHandler, cfg.Server)
	errCh := make(chan error, 1)

	go func() {
		errCh <- server.Run()
	}()

	go job.RunTTL(appCtx, storage.CleanUpExpired, 5*time.Second)

	select {
	case <-appCtx.Done():
		log.Println("Shut down signal received")
	case serverErr := <-errCh:
		if serverErr != nil {
			log.Fatalf("Server stopped unexpectedly: %v", serverErr)
		}
		log.Print("Server stopped")

	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.Timeouts.Shutdown)
	defer cancel()

	err = server.Shutdown(shutdownCtx)
	if err != nil {
		log.Fatalf("Server shutdown failed %v", err)
	}

}
