package cache

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type MemoryCache struct {
	sync.RWMutex
	items    map[string]*CacheItem
	Interval time.Duration
}

// NewMemoryCache returns a new MemoryCache.
func NewMemoryCache(interval time.Duration) Cache {
	res := &MemoryCache{
		Interval: interval,
		items:    make(map[string]*CacheItem),
	}
	go res.ClearExpiredKeys()
	return res
}
func (m *MemoryCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	m.Lock()
	defer m.Unlock()
	m.items[key] = &CacheItem{
		Data:       value,
		Lastaccess: time.Now(),
		Expired:    time.Now().Add(ttl),
	}
	return nil
}

func (m *MemoryCache) Has(ctx context.Context, key string) (bool, error) {
	if _, err := m.Get(ctx, key); err == nil {
		return false, err
	}
	return true, nil
}

func (m *MemoryCache) GetMulti(ctx context.Context, keys []string) ([]any, error) {
	rc := make([]interface{}, len(keys))
	keysErr := make([]string, 0)
	for i, ki := range keys {
		val, err := m.Get(context.Background(), ki)
		if err != nil {
			keysErr = append(keysErr, fmt.Sprintf("key [%s] error: %s", ki, err.Error()))
			continue
		}
		rc[i] = val
	}
	if len(keysErr) == 0 {
		return rc, nil
	}
	return rc, errors.New(strings.Join(keysErr, "; "))
}

func (m *MemoryCache) Get(ctx context.Context, key string) (any, error) {
	m.RLock()
	defer m.RUnlock()
	if item, ok := m.items[key]; ok {
		if item.Expired.Before(time.Now()) {
			delete(m.items, key)
			return nil, ErrKeyExpired
		}
		return item.Data, nil
	}
	return nil, ErrKeyNotExist
}

func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	m.Lock()
	defer m.Unlock()
	delete(m.items, key)
	return nil
}

func (m *MemoryCache) Increment(ctx context.Context, key string, step int) error {
	m.Lock()
	defer m.Unlock()
	itm, ok := m.items[key]
	if !ok {
		return ErrKeyNotExist
	}
	val, err := Increment(itm.Data, step)
	if err != nil {
		return err
	}
	itm.Data = val
	return nil
}

func (m *MemoryCache) Decrement(ctx context.Context, key string, step int) error {
	m.Lock()
	defer m.Unlock()
	itm, ok := m.items[key]
	if !ok {
		return ErrKeyNotExist
	}
	val, err := Decrement(itm.Data, step)
	if err != nil {
		return err
	}
	itm.Data = val
	return nil
}

func (m *MemoryCache) Clear(ctx context.Context) {
	m.Lock()
	defer m.Unlock()
	m.items = make(map[string]*CacheItem)
}

func (m *MemoryCache) ClearExpiredKeys() {
	m.RLock()
	m.RUnlock()
	for {
		<-time.After(m.Interval)
		m.RLock()
		if m.items == nil {
			m.RUnlock()
			return
		}
		m.RUnlock()
		for key, item := range m.items {
			if item.Expired.Before(time.Now()) {
				delete(m.items, key)
			}
		}
	}
}
