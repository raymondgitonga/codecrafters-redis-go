package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/utils/ptr"
	"slices"
	"strconv"
	"sync"
	"time"
)

var (
	ErrExpired     = errors.New("entry is expired")
	ErrKeyNotFound = errors.New("key not found")
)

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
	mu         sync.Mutex
	dict       map[string]Data
	expTracker map[string]int64
	waiters    map[string][]chan string // FIFO per key
}

func NewStore() *DataStore {
	return &DataStore{
		dict:       make(map[string]Data),
		expTracker: make(map[string]int64),
		mu:         sync.Mutex{},
		waiters:    make(map[string][]chan string),
	}
}

func (d *DataStore) Set(key string, args []string) error {
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

func (d *DataStore) RPush(key string, data []string) (*int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, v := range data {
		if q := d.waiters[key]; len(q) > 0 {
			//Get the first item in queue
			ch := q[0]
			// dequeue
			d.waiters[key] = q[1:]
			// add channel to its channel
			ch <- v
			continue
		}
	}

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

func (d *DataStore) LPush(key string, data []string) (*int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	baseData := data
	var expiry string
	if slices.Contains(data, "PX") || slices.Contains(data, "px") {
		expiry = data[len(data)-1]
		baseData = data[:len(data)-2]
	}

	cur, ok := d.dict[key]
	if ok && cur.Type != ListDataType {
		return nil, fmt.Errorf("type mismatch: existing value is %v, not a list", cur.Type)
	}

	slices.Reverse(baseData)
	combined := baseData
	if ok {
		combined = append(combined, cur.List...)
	}

	d.dict[key] = Data{
		Type: ListDataType,
		List: combined,
	}

	size := ptr.ToPointer(len(d.dict[key].List))

	if expiry == "" {
		delete(d.expTracker, key)
		return size, nil
	}

	ttlMs, err := strconv.ParseInt(expiry, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid TTL: %w", err)
	}
	expireAtMs := time.Now().Add(time.Duration(ttlMs) * time.Millisecond).UnixMilli()
	d.expTracker[key] = expireAtMs

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

		size = ptr.ToPointer(len(d.dict[key].List))
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
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.get(key)
}

func (d *DataStore) get(key string) (string, error) {
	if expAt, ok := d.expTracker[key]; ok && expAt > 0 {
		if time.Now().UnixMilli() >= expAt {
			delete(d.dict, key)
			delete(d.expTracker, key)
			return "nil", ErrExpired
		}
	}

	v, ok := d.dict[key]
	if !ok {
		return "", ErrKeyNotFound
	}

	if v.Type != StringDataType {
		return "", fmt.Errorf("invalid command for key specified %s", key)
	}

	return v.Simple, nil
}

func (d *DataStore) Type(key string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.get(key)
	if errors.Is(err, ErrKeyNotFound) {
		return "none", nil
	}
	return "string", nil
}

func (d *DataStore) LRange(key string, args []string) ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if expAt, ok := d.expTracker[key]; ok && expAt > 0 {
		if time.Now().UnixMilli() >= expAt {
			delete(d.dict, key)
			delete(d.expTracker, key)
			return nil, ErrExpired
		}
	}

	v, ok := d.dict[key]
	if !ok {
		return []string{}, nil
	}

	if v.Type != ListDataType {
		return nil, fmt.Errorf("invalid command for key specified %s", key)
	}

	if len(args) > 1 {
		start, err := strconv.Atoi(args[0])
		if err != nil {
			return nil, err
		}

		if start < 0 {
			start = len(v.List) + start
		}

		end, err := strconv.Atoi(args[1])
		if err != nil {
			return nil, err
		}

		if end < 0 {
			end = len(v.List) + end
		}

		if start < 0 {
			start = 0
		}

		if end >= len(v.List) {
			end = len(v.List) - 1
		}

		if start > end || start >= len(v.List) {
			return []string{}, nil
		}

		return v.List[start : end+1], nil
	}

	start, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, err
	}

	return v.List[start:], nil
}

func (d *DataStore) LLen(key string) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	v, ok := d.dict[key]
	if !ok {
		return 0, nil
	}

	if v.Type != ListDataType {
		return 0, fmt.Errorf("invalid command for key specified %s", key)
	}

	return len(v.List), nil

}

func (d *DataStore) LPop(key string, cnt int) ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.lpop(key, cnt)
}

func (d *DataStore) BLPop(ctx context.Context, key string, _ int) ([]string, error) {
	respChan := make(chan string, 1)

	d.mu.Lock()
	resp, err := d.lpop(key, 1)
	if err != nil {
		d.mu.Unlock()
		return nil, err
	}

	if len(resp) > 0 {
		d.mu.Unlock()
		return append([]string{key}, resp...), nil
	}

	// Enqueue this waiter (FIFO)
	d.waiters[key] = append(d.waiters[key], respChan)
	d.mu.Unlock()

	select {
	case <-ctx.Done():
		// Remove this waiter if still queued
		d.mu.Lock()
		q := d.waiters[key]
		for i, w := range q {
			if w == respChan {
				d.waiters[key] = append(q[:i], q[i+1:]...)
				break
			}
		}
		d.mu.Unlock()
		return nil, ctx.Err()
	case val := <-respChan:
		return []string{key, val}, nil
	}
}

func (d *DataStore) lpop(key string, cnt int) ([]string, error) {
	v, ok := d.dict[key]
	if !ok {
		return []string{}, nil
	}

	if v.Type != ListDataType {
		return nil, fmt.Errorf("invalid command for key specified %s", key)
	}

	if cnt <= 0 {
		cnt = 1
	}

	if cnt > len(v.List) {
		cnt = len(v.List)
	}

	popped := v.List[:cnt]

	rest := v.List[cnt:]

	d.dict[key] = Data{
		Type: ListDataType,
		List: rest,
	}

	return popped, nil
}

func (d *DataStore) Del(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.dict, key)
	delete(d.expTracker, key)
}
