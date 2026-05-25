package http

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/playfulCloud/cadistributedche/internal/metrics"
	"github.com/playfulCloud/cadistributedche/internal/store"
)

var errInvalidTTL = errors.New("ttl must be greater than zero")

type StorageHandler struct {
	storage store.Store
}

type StatsHandler struct {
	reader metrics.CacheStatsReader
}

func NewStorageHandler(storage store.Store) *StorageHandler {
	return &StorageHandler{
		storage: storage,
	}
}

func NewStatsHandler(reader metrics.CacheStatsReader) *StatsHandler {
	return &StatsHandler{
		reader: reader,
	}
}

func (h *StorageHandler) handleCache(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetCache(w, r)
	case http.MethodPut:
		h.handlePutCache(w, r)
	case http.MethodDelete:
		h.handleDeleteCache(w, r)
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *StatsHandler) handleStats(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	switch r.Method {
	case http.MethodGet:
		stats := h.reader.GetStats()
		writeJson(w, http.StatusOK, stats)
	default:
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *StorageHandler) handleGetCache(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	defer r.Body.Close()

	if key == "" {
		http.Error(w, "Missing path param: key", http.StatusBadRequest)
		return
	}

	value, found, err := h.storage.Get(key)

	if err != nil {
		slog.Error("storage get failed", "operation", "get", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !found {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	writeJson(w, http.StatusOK, map[string]string{
		"key":   key,
		"value": value,
	})

}

func (h *StorageHandler) handlePutCache(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	defer r.Body.Close()

	if key == "" {
		http.Error(w, "Missing path param: key", http.StatusBadRequest)
		return
	}

	ttl, err := parseTTL(r)
	if err != nil {
		http.Error(w, "Invalid ttl query param", http.StatusBadRequest)
		return
	}

	value, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, existed, err := h.storage.Put(key, string(value), ttl)

	if err != nil {
		slog.Error("storage put failed", "operation", "put", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !existed {
		writeJson(w, http.StatusCreated, map[string]string{
			"status": "created",
		})
		return
	}

	writeJson(w, http.StatusOK, map[string]string{
		"status": "updated",
	})
}

func (h *StorageHandler) handleDeleteCache(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	defer r.Body.Close()

	if key == "" {
		http.Error(w, "Missing path param: key", http.StatusBadRequest)
		return
	}

	found, err := h.storage.Delete(key)

	if err != nil {
		slog.Error("storage delete failed", "operation", "delete", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !found {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseTTL(r *http.Request) (time.Duration, error) {
	rawTTL := r.URL.Query().Get("ttl")
	if rawTTL == "" {
		return 0, nil
	}

	ttl, err := time.ParseDuration(rawTTL)
	if err != nil {
		return 0, err
	}

	if ttl <= 0 {
		return 0, errInvalidTTL
	}

	return ttl, nil
}

func writeJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
