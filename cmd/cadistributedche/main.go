package main

import (
	"github.com/playfulCloud/cadistributedche/internal/store"
	"github.com/playfulCloud/cadistributedche/internal/transport/http"
)

func main() {

	storage := store.NewKeyValueStore()
	storageHandler := http.NewStorageHandler(storage)
	server := http.NewServer(storageHandler)
	server.Run()
}
