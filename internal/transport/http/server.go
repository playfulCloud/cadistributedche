package http

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/playfulCloud/cadistributedche/internal/config"
)

type Server struct {
	storageHandler *StorageHandler
	server         *http.Server
}

func NewServer(storageHandler *StorageHandler, cfg config.ServerConfig) *Server {
	router := http.NewServeMux()
	router.HandleFunc("/storage", storageHandler.handleStorage)

	handler := loggingMiddleware(router)
	return &Server{
		storageHandler: storageHandler,
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
	err := s.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Print("Shutting down server")
	return s.server.Shutdown(ctx)
}
