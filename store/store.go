package store

import (
	"fmt"
	"sync"
)

type DataStore struct {
	mu   sync.RWMutex
	dict map[string][]byte
}

func NewStore() *DataStore {
	return &DataStore{
		mu:   sync.RWMutex{},
		dict: make(map[string][]byte),
	}
}

func (d *DataStore) Set(key string, data []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.dict[key] = data

	return nil
}

func (d *DataStore) Get(key string) ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	v, ok := d.dict[key]
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}

	return v, nil
}
