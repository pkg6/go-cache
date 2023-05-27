package redis

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg6/go-cache"
	"time"
)

// DefaultKey defines the collection name of redis for the cache adapter.
var DefaultKey = "gocache"

type RedisCache struct {
	Redis *redis.Pool // redis connection pool
	Key   string
}
type CacheOptions func(c *RedisCache)

// CacheWithKey configures key for redis
func CacheWithKey(key string) CacheOptions {
	return func(c *RedisCache) {
		c.Key = key
	}
}

// NewRedisCache creates a new redis cache with default collection name.
func NewRedisCache(pool *redis.Pool, opts ...CacheOptions) cache.Cache {
	c := &RedisCache{
		Redis: pool,
		Key:   DefaultKey,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
func (c *RedisCache) Name() string {
	return cache.RedisCacheName
}

// Set puts cache into redis.
func (c *RedisCache) Set(key string, value any, ttl time.Duration) error {
	_, err := c.do("SETEX", key, int64(ttl/time.Second), value)
	return err
}

func (c *RedisCache) Has(key string) (bool, error) {
	v, err := redis.Bool(c.do("EXISTS", key))
	if err != nil {
		return false, err
	}
	return v, nil
}

// GetMulti gets cache from redis.
func (c *RedisCache) GetMulti(keys []string) ([]any, error) {
	conn := c.Redis.Get()
	defer func() {
		_ = conn.Close()
	}()
	var args []interface{}
	for _, key := range keys {
		args = append(args, c.cacheKey(key))
	}
	return redis.Values(conn.Do("MGET", args...))
}

// Get cache from redis.
func (c *RedisCache) Get(key string) (any, error) {
	if v, err := c.do("GET", key); err == nil {
		return v, nil
	} else {
		return nil, err
	}
}

// Delete deletes a key's cache in redis.
func (c *RedisCache) Delete(key string) error {
	_, err := c.do("DEL", key)
	return err
}

// Increment increases a key's counter in redis.
func (c *RedisCache) Increment(key string, step int) error {
	_, err := redis.Bool(c.do("INCRBY", key, step))
	return err
}

// Decrement decreases a key's counter in redis.
func (c *RedisCache) Decrement(key string, step int) error {
	_, err := redis.Bool(c.do("INCRBY", key, -step))
	return err
}

// Clear deletes all cache in the redis collection
// Be careful about this method, because it scans all keys and the delete them one by one
func (c *RedisCache) Clear() error {
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

// Execute the redis commands. args[0] must be the key name
func (c *RedisCache) do(commandName string, args ...any) (any, error) {
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
		return nil, cache.WrapF("could not execute this command: %s", commandName)
	}
	return reply, nil
}

// cacheKey with config key.
func (c *RedisCache) cacheKey(originKey any) string {
	return fmt.Sprintf("%s:%s", c.Key, originKey)
}

// Scan scans all keys matching a given pattern.
func (c *RedisCache) Scan(pattern string) (keys []string, err error) {
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
