package http

import (
	"encoding/json"
	"net/http"

	"github.com/playfulCloud/cadistributedche/internal/model"
	"github.com/playfulCloud/cadistributedche/internal/store"
)

type StorageHandler struct {
	storage store.Store
}

func NewStorageHandler(storage store.Store) *StorageHandler {
	return &StorageHandler{
		storage: storage,
	}
}

func (h *StorageHandler) handleStorageTrafic(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetStorage(w, r)
	case http.MethodPut:
		h.handlePutStorage(w, r)
	case http.MethodDelete:
		h.handleDeleteStorage(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *StorageHandler) handleGetStorage(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	defer r.Body.Close()

	if key == "" {
		http.Error(w, "Missing querry param: key", http.StatusBadRequest)
		return
	}

	value, err := h.storage.Get(key)

	if err != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
	}

	if value == "" {
		http.Error(w, "Not found", http.StatusNotFound)
	}

	writeJson(w, http.StatusOK, map[string]string{
		"key":   key,
		"value": value,
	})

}

func (h *StorageHandler) handlePutStorage(w http.ResponseWriter, r *http.Request) {
	var req model.StorageKeyValueRequest
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.storage.Put(req.Key, req.Value)

	if err != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	if result == "" {
		writeJson(w, http.StatusCreated, map[string]string{
			"status": "created",
		})
		return
	}

	writeJson(w, http.StatusCreated, map[string]string{
		"status": "updated",
	})
}

func (h *StorageHandler) handleDeleteStorage(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	defer r.Body.Close()

	if key == "" {
		http.Error(w, "Missing querry param: key", http.StatusBadRequest)
	}

	result, err := h.storage.Delete(key)

	if err != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
	}

	if result == "" {
		writeJson(w, http.StatusNotFound, result)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
