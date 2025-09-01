package store

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

var ErrExpired = errors.New("entry is expired")

type ExpiryStruct struct {
	DateAdded int64
	Expiry    int64
}

type DataStore struct {
	mu         sync.RWMutex
	dict       map[string][]byte
	expTracker map[string]int64
}

func NewStore() *DataStore {
	return &DataStore{
		mu:         sync.RWMutex{},
		dict:       make(map[string][]byte),
		expTracker: make(map[string]int64),
	}
}

func (d *DataStore) Set(key string, data []byte, ex string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.dict[key] = data
	if ex == "" {
		delete(d.expTracker, key)
		return nil
	}

	ttlMs, err := strconv.ParseInt(ex, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid TTL: %w", err)
	}
	expireAtMs := time.Now().Add(time.Duration(ttlMs) * time.Millisecond).UnixMilli()
	d.expTracker[key] = expireAtMs

	return nil
}

func (d *DataStore) Get(key string) ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if expAt, ok := d.expTracker[key]; ok && expAt > 0 {
		if time.Now().UnixMilli() >= expAt {
			delete(d.dict, key)
			delete(d.expTracker, key)
			return nil, ErrExpired
		}
	}

	v, ok := d.dict[key]
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}

	return v, nil
}

func (d *DataStore) Del(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.dict, key)
	delete(d.expTracker, key)
}
