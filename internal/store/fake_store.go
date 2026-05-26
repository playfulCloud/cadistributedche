package store

import (
	"errors"
	"time"
)

type FakeStore struct {
	PutFunc            func(key string, value string, ttl time.Duration) (KeyValueEntry, bool, error)
	GetFunc            func(key string) (KeyValueEntry, bool, error)
	DeleteFunc         func(key string) (bool, error)
	SizeFunc           func() int
	CleanupExpiredFunc func() int
}

func (f *FakeStore) Get(key string) (KeyValueEntry, bool, error) {
	if f.GetFunc != nil {
		return f.GetFunc(key)
	}
	if key == "empty" {
		return KeyValueEntry{}, false, nil
	}
	if key == "empty-value" {
		return KeyValueEntry{key: key}, true, nil
	}
	return KeyValueEntry{key: key, value: "value"}, true, nil
}

func (f *FakeStore) Put(key string, value string, ttl time.Duration) (KeyValueEntry, bool, error) {
	if f.PutFunc != nil {
		return f.PutFunc(key, value, ttl)
	}
	if key == "exists" {
		return KeyValueEntry{key: key, value: value, ttl: ttl}, true, nil
	}
	if key == "empty-value" {
		return KeyValueEntry{key: key, value: value, ttl: ttl}, true, nil
	}
	if key == "error" {
		return KeyValueEntry{}, false, errors.New("some error related with store")
	}
	return KeyValueEntry{}, false, nil
}

func (f *FakeStore) Delete(key string) (bool, error) {
	if f.DeleteFunc != nil {
		return f.DeleteFunc(key)
	}
	if key == "exists" {
		return true, nil
	}
	return false, nil
}

func (f *FakeStore) Size() int {
	if f.SizeFunc != nil {
		return f.SizeFunc()
	}
	return 0
}

func (f *FakeStore) CleanupExpired() int {
	if f.CleanupExpiredFunc != nil {
		return f.CleanupExpiredFunc()
	}

	return 0
}
