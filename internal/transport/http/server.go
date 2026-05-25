package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/playfulCloud/cadistributedche/internal/config"
)

type Server struct {
	storageHandler *StorageHandler
	statsHandler   *StatsHandler
	server         *http.Server
}

func NewServer(storageHandler *StorageHandler, statsHandler *StatsHandler, cfg config.ServerConfig) *Server {
	router := http.NewServeMux()
	router.HandleFunc("/cache/{key}", storageHandler.handleCache)
	router.HandleFunc("/stats", statsHandler.handleStats)

	handler := loggingMiddleware(router)
	return &Server{
		storageHandler: storageHandler,
		statsHandler:   statsHandler,
		server: &http.Server{
			Addr:         ":" + strconv.Itoa(cfg.Port),
			Handler:      handler,
			ReadTimeout:  cfg.Timeouts.Read,
			WriteTimeout: cfg.Timeouts.Write,
			IdleTimeout:  cfg.Timeouts.Idle,
		},
	}
}

func (s *Server) Run() error {
	slog.Info("http server starting", "addr", s.server.Addr)
	err := s.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("http server shutdown started")
	return s.server.Shutdown(ctx)
}
