package http

import (
	"log"
	"net/http"
)

type Server struct {
	storageHandler *StorageHandler
}

func NewServer(storageHandler *StorageHandler) *Server {
	return &Server{
		storageHandler: storageHandler,
	}
}

func (s *Server) Run() {

	router := http.NewServeMux()
	router.HandleFunc("/storage", s.storageHandler.handleStorageTrafic)

	handler := loggingMiddleware(router)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}
