package store

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"sync"
	"time"
)

var ErrExpired = errors.New("entry is expired")

type DataType uint8

const (
	StringDataType DataType = iota
	ListDataType
)

type Data struct {
	Type   DataType
	List   []string
	Simple string
}

type DataStore struct {
	mu         sync.RWMutex
	dict       map[string]Data
	expTracker map[string]int64
}

func NewStore() *DataStore {
	return &DataStore{
		mu:         sync.RWMutex{},
		dict:       make(map[string]Data),
		expTracker: make(map[string]int64),
	}
}

func (d *DataStore) SetString(key string, args []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var data string
	var expiry string
	if len(args) >= 3 {
		expiry = args[2]
		data = args[len(args)-3]
	} else {
		data = args[0]
	}

	_, err := d.set(key, data, expiry)
	return err
}

func (d *DataStore) SetList(key string, data []string) (*int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	baseData := data
	var expiry string
	if slices.Contains(data, "PX") || slices.Contains(data, "px") {
		expiry = data[len(data)-1]
		baseData = data[:len(data)-2]
	}

	size, err := d.set(key, baseData, expiry)
	if err != nil {
		return nil, err
	}

	return size, nil
}

func (d *DataStore) set(key string, data any, ex string) (*int, error) {
	var size *int

	switch v := data.(type) {
	case string:
		d.dict[key] = Data{
			Type:   StringDataType,
			Simple: v,
		}
	case []string:
		cur, ok := d.dict[key]
		if ok && cur.Type != ListDataType {
			return nil, fmt.Errorf("type mismatch: existing value is %v, not a list", cur.Type)
		}

		combined := v
		if ok {
			combined = append(cur.List, combined...)
		}

		d.dict[key] = Data{
			Type: ListDataType,
			List: combined,
		}

		size = Ptr(len(d.dict[key].List))
	default:
		return nil, fmt.Errorf("invalid data type")
	}

	if ex == "" {
		delete(d.expTracker, key)
		return size, nil
	}

	ttlMs, err := strconv.ParseInt(ex, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid TTL: %w", err)
	}
	expireAtMs := time.Now().Add(time.Duration(ttlMs) * time.Millisecond).UnixMilli()
	d.expTracker[key] = expireAtMs

	return size, nil
}

func (d *DataStore) Get(key string) (string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if expAt, ok := d.expTracker[key]; ok && expAt > 0 {
		if time.Now().UnixMilli() >= expAt {
			delete(d.dict, key)
			delete(d.expTracker, key)
			return "", ErrExpired
		}
	}

	v, ok := d.dict[key]
	if !ok {
		return "nil", fmt.Errorf("key %s not found", key)
	}

	if v.Type == StringDataType {
		return v.Simple, nil
	}

	return "", nil
}

func (d *DataStore) Del(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.dict, key)
	delete(d.expTracker, key)
}

// TODO: move this
func Ptr[T any](v T) *T {
	return &v
}
