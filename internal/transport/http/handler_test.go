package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStorageHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		url        string
		status     int
		httpMethod string
		response   string
		allow      string
		store      FakeStore
	}{
		{
			name:       "PUT Element in store exists",
			body:       `{"key":"exists","value":"value"}`,
			url:        "/storage",
			status:     http.StatusOK,
			httpMethod: http.MethodPut,
			response:   `{"status":"updated"}` + "\n",
		},
		{
			name:       "PUT Element does not exist in store",
			body:       `{"key":"new","value":"value"}`,
			url:        "/storage",
			status:     http.StatusCreated,
			httpMethod: http.MethodPut,
			response:   `{"status":"created"}` + "\n",
		},
		{
			name:       "PUT Element exists with empty previous value",
			body:       `{"key":"empty-value","value":"value"}`,
			url:        "/storage",
			status:     http.StatusOK,
			httpMethod: http.MethodPut,
			response:   `{"status":"updated"}` + "\n",
		},
		{
			name:       "PUT invalid body",
			body:       `{`,
			url:        "/storage",
			status:     http.StatusBadRequest,
			httpMethod: http.MethodPut,
			response:   "Invalid request body\n",
		},
		{
			name:       "PUT missing key",
			body:       `{"value":"value"}`,
			url:        "/storage",
			status:     http.StatusBadRequest,
			httpMethod: http.MethodPut,
			response:   "Invalid request body\n",
		},
		{
			name:       "PUT store error",
			body:       `{"key":"error","value":"value"}`,
			url:        "/storage",
			status:     http.StatusInternalServerError,
			httpMethod: http.MethodPut,
			response:   "Internal server error\n",
			store: FakeStore{
				put: func(key string, value string) (string, bool, error) {
					return "", false, errors.New("some error related with store")
				},
			},
		},
		{
			name:       "GET Element in store exists",
			url:        "/storage?key=exists",
			status:     http.StatusOK,
			httpMethod: http.MethodGet,
			response:   `{"key":"exists","value":"value"}` + "\n",
		},
		{
			name:       "GET Element exists with empty value",
			url:        "/storage?key=empty-value",
			status:     http.StatusOK,
			httpMethod: http.MethodGet,
			response:   `{"key":"empty-value","value":""}` + "\n",
		},
		{
			name:       "GET Element does not exist in store",
			url:        "/storage?key=empty",
			status:     http.StatusNotFound,
			httpMethod: http.MethodGet,
			response:   "Not found\n",
		},
		{
			name:       "GET missing key",
			url:        "/storage",
			status:     http.StatusBadRequest,
			httpMethod: http.MethodGet,
			response:   "Missing query param: key\n",
		},
		{
			name:       "GET store error",
			url:        "/storage?key=exists",
			status:     http.StatusInternalServerError,
			httpMethod: http.MethodGet,
			response:   "Internal server error\n",
			store: FakeStore{
				get: func(key string) (string, bool, error) {
					return "", false, errors.New("some error related with store")
				},
			},
		},
		{
			name:       "DELETE Element in store exists",
			url:        "/storage?key=exists",
			status:     http.StatusNoContent,
			httpMethod: http.MethodDelete,
		},
		{
			name:       "DELETE Element does not exist in store",
			url:        "/storage?key=empty",
			status:     http.StatusNotFound,
			httpMethod: http.MethodDelete,
			response:   "Not found\n",
		},
		{
			name:       "DELETE missing key",
			url:        "/storage",
			status:     http.StatusBadRequest,
			httpMethod: http.MethodDelete,
			response:   "Missing query param: key\n",
		},
		{
			name:       "DELETE store error",
			url:        "/storage?key=exists",
			status:     http.StatusInternalServerError,
			httpMethod: http.MethodDelete,
			response:   "Internal server error\n",
			store: FakeStore{
				delete: func(key string) (bool, error) {
					return false, errors.New("some error related with store")
				},
			},
		},
		{
			name:       "POST method not allowed",
			body:       `{"key":"exists","value":"value"}`,
			url:        "/storage",
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
			rr := httptest.NewRecorder()

			h.handleStorage(rr, req)

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

type FakeStore struct {
	put    func(key string, value string) (string, bool, error)
	get    func(key string) (string, bool, error)
	delete func(key string) (bool, error)
}

func (f *FakeStore) Get(key string) (string, bool, error) {
	if f.get != nil {
		return f.get(key)
	}
	if key == "empty" {
		return "", false, nil
	}
	if key == "empty-value" {
		return "", true, nil
	}
	return "value", true, nil
}

func (f *FakeStore) Put(key string, value string) (string, bool, error) {
	if f.put != nil {
		return f.put(key, value)
	}
	if key == "exists" {
		return "previousValue", true, nil
	}
	if key == "empty-value" {
		return "", true, nil
	}
	if key == "error" {
		return "", false, errors.New("Some error related with store")
	}
	return "", false, nil
}

func (f *FakeStore) Delete(key string) (bool, error) {
	if f.delete != nil {
		return f.delete(key)
	}
	if key == "exists" {
		return true, nil
	}
	return false, nil
}
