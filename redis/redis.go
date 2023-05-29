package redis

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg6/go-cache"
)

// DefaultKey defines the collection name of redis for the cache adapter.
var DefaultKey = "gocache"

type Cache struct {
	Redis     *redis.Pool // redis connection pool
	Key       string
	CacheItem cache.ICacheItem
}
type CacheOptions func(c *Cache)

// CacheWithKey configures key for redis
func CacheWithKey(key string) CacheOptions {
	return func(c *Cache) {
		c.Key = key
	}
}

// CacheWithRedisPool configures prefix for redis
func CacheWithRedisPool(pool *redis.Pool) CacheOptions {
	return func(c *Cache) {
		c.Redis = pool
	}
}
func defaultRedisPool() *redis.Pool {
	return &redis.Pool{
		Dial: func() (c redis.Conn, err error) {
			c, err = redis.Dial("tcp", "127.0.0.1:6379")
			if err != nil {
				return nil, fmt.Errorf("could not dial to remote redis server: %s ", "127.0.0.1:6379")
			}
			if _, doErr := c.Do("SELECT", 0); doErr != nil {
				_ = c.Close()
				return nil, doErr
			}
			return
		},
		MaxIdle:     3,
		IdleTimeout: 3 * time.Second,
	}
}

// NewRedisCache creates a new redis cache with default collection name.
func NewRedisCache(opts ...CacheOptions) cache.Cache {
	c := &Cache{
		Redis:     defaultRedisPool(),
		Key:       DefaultKey,
		CacheItem: &cache.CacheItem{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
func (c *Cache) Name() string {
	return cache.RedisCacheName
}

// Set puts cache into redis.
func (c *Cache) Set(key string, value any, ttl time.Duration) error {
	valStr, err := c.CacheItem.SetCacheItem(value, ttl)
	if err != nil {
		return err
	}
	commandName := "SETEX"
	args := []any{key}
	if c.CacheItem.IsNeverExpires() {
		commandName = "SET"
		args = append(args, valStr)
	} else {
		args = append(args, int64(c.CacheItem.GetTTL()/time.Second), valStr)
	}
	_, err = c.do(commandName, args...)
	return err
}

func (c *Cache) Has(key string) (bool, error) {
	v, err := redis.Bool(c.do("EXISTS", key))
	if err != nil {
		return false, err
	}
	return v, nil
}

// GetMulti gets cache from redis.
func (c *Cache) GetMulti(keys []string) ([]any, error) {
	conn := c.Redis.Get()
	defer func() {
		_ = conn.Close()
	}()
	var args []interface{}
	for _, key := range keys {
		args = append(args, c.cacheKey(key))
	}
	values, err := redis.Values(conn.Do("MGET", args...))
	if err != nil {
		return nil, err
	}
	newValues := make([]any, len(values))
	keysErr := make([]string, len(values))
	for i, value := range values {
		item, err := c.CacheItem.GetCacheItem(value)
		if err != nil {
			keysErr[i] = err.Error()
			continue
		}
		newValues[i] = item.GetData()
	}
	if len(keysErr) == 0 {
		return newValues, nil
	}
	return newValues, fmt.Errorf(strings.Join(keysErr, "; "))
}

// Get cache from redis.
func (c *Cache) Get(key string) (any, error) {
	item, err := c.getCacheItem(key)
	if err != nil {
		return nil, err
	}
	return item.GetData(), nil
}

// Delete deletes a key's cache in redis.
func (c *Cache) Delete(key string) error {
	_, err := c.do("DEL", key)
	return err
}

// Increment increases a key's counter in redis.
func (c *Cache) Increment(key string, step int) error {
	item, err := c.getCacheItem(key)
	if err != nil {
		return c.Set(key, step, 0)
	}
	data, err := cache.Increment(item.GetData(), step)
	if err != nil {
		return err
	}
	return c.Set(key, data, item.GetTTL())
}

// Decrement decreases a key's counter in redis.
func (c *Cache) Decrement(key string, step int) error {
	item, err := c.getCacheItem(key)
	if err != nil {
		return c.Set(key, step, 0)
	}
	data, err := cache.Decrement(item.GetData(), step)
	if err != nil {
		return err
	}
	return c.Set(key, data, item.GetTTL())
}

// Clear deletes all cache in the redis collection
// Be careful about this method, because it scans all keys and the delete them one by one
func (c *Cache) Clear() error {
	cachedKeys, err := c.Scan(c.Key + ":*")
	if err != nil {
		return err
	}
	conn := c.Redis.Get()
	defer func() {
		_ = conn.Close()
	}()
	for _, str := range cachedKeys {
		if _, err = conn.Do("DEL", str); err != nil {
			return err
		}
	}
	return err
}
func (c *Cache) getCacheItem(key string) (item cache.ICacheItem, err error) {
	if v, err := c.do("GET", key); err == nil {
		item, err = c.CacheItem.GetCacheItem(v)
		if err != nil {
			return nil, err
		}
		return item, nil
	} else {
		return nil, err
	}
}

// cacheKey with config key.
func (c *Cache) cacheKey(originKey any) string {
	return fmt.Sprintf("%s:%s", c.Key, originKey)
}

// Execute the redis commands. args[0] must be the key name
func (c *Cache) do(commandName string, args ...any) (any, error) {
	if len(args) == 0 {
		return nil, errors.New("args is 0")
	}
	args[0] = c.cacheKey(args[0])
	conn := c.Redis.Get()
	defer func() {
		_ = conn.Close()
	}()
	reply, err := conn.Do(commandName, args...)
	if err != nil {
		return nil, fmt.Errorf("could not execute this command: %s", commandName)
	}
	return reply, nil
}

// Scan scans all keys matching a given pattern.
func (c *Cache) Scan(pattern string) (keys []string, err error) {
	conn := c.Redis.Get()
	defer func() {
		_ = conn.Close()
	}()
	var (
		cursor uint64 = 0 // start
		result []interface{}
		list   []string
	)
	for {
		result, err = redis.Values(conn.Do("SCAN", cursor, "MATCH", pattern, "COUNT", 1024))
		if err != nil {
			return
		}
		list, err = redis.Strings(result[1], nil)
		if err != nil {
			return
		}
		keys = append(keys, list...)
		cursor, err = redis.Uint64(result[0], nil)
		if err != nil {
			return
		}
		if cursor == 0 { // over
			return
		}
	}
}
