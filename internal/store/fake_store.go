package store

import (
	"errors"
	"time"
)

type FakeStore struct {
	PutFunc    func(key string, value string, ttl time.Duration) (string, bool, error)
	GetFunc    func(key string) (string, bool, error)
	DeleteFunc func(key string) (bool, error)
}

func (f *FakeStore) Get(key string) (string, bool, error) {
	if f.GetFunc != nil {
		return f.GetFunc(key)
	}
	if key == "empty" {
		return "", false, nil
	}
	if key == "empty-value" {
		return "", true, nil
	}
	return "value", true, nil
}

func (f *FakeStore) Put(key string, value string, ttl time.Duration) (string, bool, error) {
	if f.PutFunc != nil {
		return f.PutFunc(key, value, ttl)
	}
	if key == "exists" {
		return "previousValue", true, nil
	}
	if key == "empty-value" {
		return "", true, nil
	}
	if key == "error" {
		return "", false, errors.New("some error related with store")
	}
	return "", false, nil
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
