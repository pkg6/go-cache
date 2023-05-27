package memcache

import (
	"errors"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/pkg6/go-cache"
	"strings"
	"time"
)

type MemcacheCache struct {
	Memcache *memcache.Client
}
type CacheOptions func(c *MemcacheCache)

// NewMemCache creates new memcache adapter.
func NewMemCache(memcache *memcache.Client, opts ...CacheOptions) cache.Cache {
	c := &MemcacheCache{
		Memcache: memcache,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (m MemcacheCache) Set(key string, value any, ttl time.Duration) error {
	item := memcache.Item{Key: key, Expiration: int32(ttl / time.Second)}
	if v, ok := value.([]byte); ok {
		item.Value = v
	} else if str, ok := value.(string); ok {
		item.Value = []byte(str)
	} else {
		return fmt.Errorf("the value must be string or byte[]. key: %s, value:%v", key, value)
	}
	return m.Memcache.Set(&item)
}

func (m MemcacheCache) Has(key string) (bool, error) {
	_, err := m.Get(key)
	return err == nil, err
}

func (m MemcacheCache) GetMulti(keys []string) ([]any, error) {
	rv := make([]interface{}, len(keys))
	mv, err := m.Memcache.GetMulti(keys)
	if err != nil {
		return rv, cache.WrapF("could not read multiple key-values from memcache, please check your keys, network and connection. Root cause: %s", err.Error())
	}
	keysErr := make([]string, 0)
	for i, ki := range keys {
		if _, ok := mv[ki]; !ok {
			keysErr = append(keysErr, fmt.Sprintf("key [%s] error: %s", ki, "key not exist"))
			continue
		}
		rv[i] = mv[ki].Value
	}
	if len(keysErr) == 0 {
		return rv, nil
	}
	return rv, errors.New(strings.Join(keysErr, "; "))
}

func (m MemcacheCache) Get(key string) (any, error) {
	if item, err := m.Memcache.Get(key); err == nil {
		return item.Value, nil
	} else {
		return nil, cache.WrapF("could not read data from memcache, please check your key, network and connection. Root cause: %s", err.Error())
	}
}

func (m MemcacheCache) Delete(key string) error {
	return m.Memcache.Delete(key)
}

func (m MemcacheCache) Increment(key string, step int) error {
	_, err := m.Memcache.Increment(key, uint64(step))
	return err
}

func (m MemcacheCache) Decrement(key string, step int) error {
	_, err := m.Memcache.Decrement(key, uint64(step))
	return err
}

func (m MemcacheCache) Clear() error {
	return m.Memcache.FlushAll()
}
