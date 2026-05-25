package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/playfulCloud/cadistributedche/internal/metrics"
	"github.com/playfulCloud/cadistributedche/internal/store"
)

func TestStorageHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		key        string
		url        string
		status     int
		httpMethod string
		response   string
		allow      string
		store      store.FakeStore
	}{
		{
			name:       "PUT element in store exists",
			body:       "value",
			key:        "exists",
			url:        "/cache/exists",
			status:     http.StatusOK,
			httpMethod: http.MethodPut,
			response:   `{"status":"updated"}` + "\n",
		},
		{
			name:       "PUT element does not exist in store",
			body:       "value",
			key:        "new",
			url:        "/cache/new",
			status:     http.StatusCreated,
			httpMethod: http.MethodPut,
			response:   `{"status":"created"}` + "\n",
		},
		{
			name:       "PUT element exists with empty previous value",
			body:       "value",
			key:        "empty-value",
			url:        "/cache/empty-value",
			status:     http.StatusOK,
			httpMethod: http.MethodPut,
			response:   `{"status":"updated"}` + "\n",
		},
		{
			name:       "PUT with ttl query param",
			body:       "value",
			key:        "new",
			url:        "/cache/new?ttl=10s",
			status:     http.StatusCreated,
			httpMethod: http.MethodPut,
			response:   `{"status":"created"}` + "\n",
			store: store.FakeStore{
				PutFunc: func(key string, value string, ttl time.Duration) (string, bool, error) {
					if ttl != 10*time.Second {
						return "", false, errors.New("unexpected ttl")
					}
					return "", false, nil
				},
			},
		},
		{
			name:       "PUT invalid ttl",
			body:       "value",
			key:        "new",
			url:        "/cache/new?ttl=invalid",
			status:     http.StatusBadRequest,
			httpMethod: http.MethodPut,
			response:   "Invalid ttl query param\n",
		},
		{
			name:       "PUT missing key",
			body:       "value",
			url:        "/cache/",
			status:     http.StatusBadRequest,
			httpMethod: http.MethodPut,
			response:   "Missing path param: key\n",
		},
		{
			name:       "PUT store error",
			body:       "value",
			key:        "error",
			url:        "/cache/error",
			status:     http.StatusInternalServerError,
			httpMethod: http.MethodPut,
			response:   "Internal server error\n",
			store: store.FakeStore{
				PutFunc: func(key string, value string, ttl time.Duration) (string, bool, error) {
					return "", false, errors.New("some error related with store")
				},
			},
		},
		{
			name:       "GET element in store exists",
			key:        "exists",
			url:        "/cache/exists",
			status:     http.StatusOK,
			httpMethod: http.MethodGet,
			response:   `{"key":"exists","value":"value"}` + "\n",
		},
		{
			name:       "GET element exists with empty value",
			key:        "empty-value",
			url:        "/cache/empty-value",
			status:     http.StatusOK,
			httpMethod: http.MethodGet,
			response:   `{"key":"empty-value","value":""}` + "\n",
		},
		{
			name:       "GET element does not exist in store",
			key:        "empty",
			url:        "/cache/empty",
			status:     http.StatusNotFound,
			httpMethod: http.MethodGet,
			response:   "Not found\n",
		},
		{
			name:       "GET missing key",
			url:        "/cache/",
			status:     http.StatusBadRequest,
			httpMethod: http.MethodGet,
			response:   "Missing path param: key\n",
		},
		{
			name:       "GET store error",
			key:        "exists",
			url:        "/cache/exists",
			status:     http.StatusInternalServerError,
			httpMethod: http.MethodGet,
			response:   "Internal server error\n",
			store: store.FakeStore{
				GetFunc: func(key string) (string, bool, error) {
					return "", false, errors.New("some error related with store")
				},
			},
		},
		{
			name:       "DELETE element in store exists",
			key:        "exists",
			url:        "/cache/exists",
			status:     http.StatusNoContent,
			httpMethod: http.MethodDelete,
		},
		{
			name:       "DELETE element does not exist in store",
			key:        "empty",
			url:        "/cache/empty",
			status:     http.StatusNotFound,
			httpMethod: http.MethodDelete,
			response:   "Not found\n",
		},
		{
			name:       "DELETE missing key",
			url:        "/cache/",
			status:     http.StatusBadRequest,
			httpMethod: http.MethodDelete,
			response:   "Missing path param: key\n",
		},
		{
			name:       "DELETE store error",
			key:        "exists",
			url:        "/cache/exists",
			status:     http.StatusInternalServerError,
			httpMethod: http.MethodDelete,
			response:   "Internal server error\n",
			store: store.FakeStore{
				DeleteFunc: func(key string) (bool, error) {
					return false, errors.New("some error related with store")
				},
			},
		},
		{
			name:       "POST method not allowed",
			body:       "value",
			key:        "exists",
			url:        "/cache/exists",
			status:     http.StatusMethodNotAllowed,
			httpMethod: http.MethodPost,
			response:   "method not allowed\n",
			allow:      "GET, PUT, DELETE",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &tt.store
			h := NewStorageHandler(f)

			req := httptest.NewRequest(tt.httpMethod, tt.url, strings.NewReader(tt.body))
			req.SetPathValue("key", tt.key)
			rr := httptest.NewRecorder()

			h.handleCache(rr, req)

			if rr.Code != tt.status {
				t.Fatalf("The status code expected to be %d but got %d", tt.status, rr.Code)
			}

			if rr.Body.String() != tt.response {
				t.Fatalf("The body expected to be %s but got %s", tt.response, rr.Body.String())
			}

			if rr.Header().Get("Allow") != tt.allow {
				t.Fatalf("The Allow header expected to be %s but got %s", tt.allow, rr.Header().Get("Allow"))
			}

		})
	}
}

func TestHandleMetrics(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		status     int
		httpMethod string
		response   string
		allow      string
		metrics    metrics.CacheMetrics
	}{
		{
			name:       "GET metrics",
			url:        "/cache/metrics",
			status:     http.StatusOK,
			httpMethod: http.MethodGet,
			response:   `{"cacheHits":10,"cacheMisses":20,"cacheDeletes":30,"cacheWrites":40,"cacheTotalKeys":50}` + "\n",
			metrics: metrics.CacheMetrics{
				CacheHits:      10,
				CacheMisses:    20,
				CacheDeletes:   30,
				CacheWrites:    40,
				CacheTotalKeys: 50,
			},
		},
		{
			name:       "POST method not allowed",
			url:        "/cache/metrics",
			status:     http.StatusMethodNotAllowed,
			httpMethod: http.MethodPost,
			response:   "method not allowed\n",
			allow:      "GET",
		},
		{
			name:       "PUT method not allowed",
			url:        "/cache/metrics",
			status:     http.StatusMethodNotAllowed,
			httpMethod: http.MethodPut,
			response:   "method not allowed\n",
			allow:      "GET",
		},
		{
			name:       "DELETE method not allowed",
			url:        "/cache/metrics",
			status:     http.StatusMethodNotAllowed,
			httpMethod: http.MethodDelete,
			response:   "method not allowed\n",
			allow:      "GET",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &metrics.FakeMetricsReader{
				GetMetricsFunc: func() metrics.CacheMetrics {
					return tt.metrics
				},
			}
			h := NewMetricsHandler(f)

			req := httptest.NewRequest(tt.httpMethod, tt.url, strings.NewReader(""))
			rr := httptest.NewRecorder()

			h.handleMetrics(rr, req)

			if rr.Code != tt.status {
				t.Fatalf("The status code expected to be %d but got %d", tt.status, rr.Code)
			}

			if rr.Body.String() != tt.response {
				t.Fatalf("The body expected to be %s but got %s", tt.response, rr.Body.String())
			}

			if rr.Header().Get("Allow") != tt.allow {
				t.Fatalf("The Allow header expected to be %s but got %s", tt.allow, rr.Header().Get("Allow"))
			}

		})
	}
}
