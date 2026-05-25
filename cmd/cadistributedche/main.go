package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/playfulCloud/cadistributedche/internal/config"
	"github.com/playfulCloud/cadistributedche/internal/jobs"
	"github.com/playfulCloud/cadistributedche/internal/metrics"
	"github.com/playfulCloud/cadistributedche/internal/store"
	"github.com/playfulCloud/cadistributedche/internal/transport/http"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	appCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

	cfg, err := config.ReadConfig("configs/cadistributedche.yaml")
	if err != nil {
		slog.Error("failed to load configuration", "path", "configs/cadistributedche.yaml", "error", err)
		os.Exit(1)
	}

	metricsCollector := &metrics.MetricsCollector{}
	storage := store.NewKeyValueStore(&store.ClockProvider{}, metricsCollector, cfg.Store.Ttl)
	storageHandler := http.NewStorageHandler(storage)
	metricsHandler := http.NewMetricsHandler(metricsCollector)
	server := http.NewServer(storageHandler, metricsHandler, cfg.Server)
	errCh := make(chan error, 1)

	go func() {
		errCh <- server.Run()
	}()

	go jobs.RunTTL(appCtx, storage.CleanupExpired, 5*time.Second)

	select {
	case <-appCtx.Done():
		slog.Info("shutdown signal received")
	case serverErr := <-errCh:
		if serverErr != nil {
			slog.Error("http server stopped unexpectedly", "error", serverErr)
			os.Exit(1)
		}
		slog.Info("http server stopped")

	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.Timeouts.Shutdown)
	defer cancel()

	err = server.Shutdown(shutdownCtx)
	if err != nil {
		slog.Error("http server shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("application shutdown complete")
}
