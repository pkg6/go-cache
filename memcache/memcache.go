package memcache

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/pkg6/go-cache"
)

type Cache struct {
	Memcache *memcache.Client
}
type CacheOptions func(c *Cache)

func CacheWithMemcacheClient(memcache *memcache.Client) CacheOptions {
	return func(c *Cache) {
		c.Memcache = memcache
	}
}

// NewMemCache creates new memcache adapter.
func NewMemCache(opts ...CacheOptions) cache.Cache {
	c := &Cache{
		Memcache: memcache.New("127.0.0.1:11211"),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
func (m *Cache) Name() string {
	return cache.MemcacheCacheName
}
func (m *Cache) Set(key string, value any, ttl time.Duration) error {
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

func (m *Cache) Has(key string) (bool, error) {
	_, err := m.Get(key)
	return err == nil, err
}

func (m *Cache) GetMulti(keys []string) ([]any, error) {
	rv := make([]interface{}, len(keys))
	mv, err := m.Memcache.GetMulti(keys)
	if err != nil {
		return rv, fmt.Errorf("could not read multiple key-values from memcache, please check your keys, network and connection. Root cause: %s", err)
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

func (m *Cache) Get(key string) (any, error) {
	if item, err := m.Memcache.Get(key); err == nil {
		return item.Value, nil
	} else {
		return nil, fmt.Errorf("could not read data from memcache, please check your key, network and connection. Root cause: %s", err.Error())
	}
}

func (m *Cache) Delete(key string) error {
	return m.Memcache.Delete(key)
}

func (m *Cache) Increment(key string, step int) error {
	_, err := m.Memcache.Increment(key, uint64(step))
	return err
}

func (m *Cache) Decrement(key string, step int) error {
	_, err := m.Memcache.Decrement(key, uint64(step))
	return err
}

func (m *Cache) Clear() error {
	return m.Memcache.FlushAll()
}
