package cache

import (
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
	c := &MemoryCache{
		Interval: interval,
		items:    make(map[string]*CacheItem),
	}
	go c.ClearExpiredKeys()
	return c
}

func (m *MemoryCache) Name() string {
	return MemoryCacheName
}

func (m *MemoryCache) Set(key string, value any, ttl time.Duration) error {
	m.Lock()
	defer m.Unlock()
	item := &CacheItem{Data: value, JoinTime: time.Now()}
	item.TTL = ttl
	if item.TTL == time.Duration(0) {
		item.TTL = (86400 * 365 * 20) * time.Second
	}
	item.ExpirationTime = item.JoinTime.Add(item.TTL)
	m.items[key] = item
	return nil
}

func (m *MemoryCache) Has(key string) (bool, error) {
	if _, err := m.Get(key); err == nil {
		return false, err
	}
	return true, nil
}

func (m *MemoryCache) GetMulti(keys []string) ([]any, error) {
	rc := make([]interface{}, len(keys))
	keysErr := make([]string, 0)
	for i, ki := range keys {
		val, err := m.Get(ki)
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

func (m *MemoryCache) Get(key string) (any, error) {
	m.RLock()
	defer m.RUnlock()
	if item, ok := m.items[key]; ok {
		if item.ExpirationTime.Before(time.Now()) {
			return nil, ErrKeyExpired
		}
		return item.Data, nil
	}
	return nil, ErrKeyNotExist
}

func (m *MemoryCache) Delete(key string) error {
	m.Lock()
	defer m.Unlock()
	delete(m.items, key)
	return nil
}

func (m *MemoryCache) Increment(key string, step int) error {
	m.Lock()
	defer m.Unlock()
	itm, ok := m.items[key]
	if !ok {
		return m.Set(key, step, 0)
	}
	val, err := Increment(itm.Data, step)
	if err != nil {
		return err
	}
	itm.Data = val
	return nil
}

func (m *MemoryCache) Decrement(key string, step int) error {
	m.Lock()
	defer m.Unlock()
	itm, ok := m.items[key]
	if !ok {
		return m.Set(key, step, 0)
	}
	val, err := Decrement(itm.Data, step)
	if err != nil {
		return err
	}
	itm.Data = val
	return nil
}

func (m *MemoryCache) Clear() error {
	m.Lock()
	defer m.Unlock()
	m.items = make(map[string]*CacheItem)
	return nil
}

func (m *MemoryCache) ClearExpiredKeys() {
	for {
		<-time.After(m.Interval)
		m.RLock()
		if m.items == nil {
			m.RUnlock()
			return
		}
		m.RUnlock()
		for key, item := range m.items {
			if item.ExpirationTime.Before(time.Now()) {
				delete(m.items, key)
			}
		}
	}
}
