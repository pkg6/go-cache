package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrKeyExpired  = errors.New("the key is expired")
	ErrKeyNotExist = errors.New("the key isn't exist")
)

const (
	MinUint32 uint32 = 0
	MinUint64 uint64 = 0

	FileCacheName     = "file"
	MemoryCacheName   = "memory"
	RedisCacheName    = "redis"
	MemcacheCacheName = "memcache"
	CacheName         = "go-cache"
)

type Cache interface {
	Name() string
	Set(key string, value any, ttl time.Duration) error
	Has(key string) (bool, error)
	GetMulti(keys []string) ([]any, error)
	Get(key string) (any, error)
	Delete(key string) error
	Increment(key string, step int) error
	Decrement(key string, step int) error
	Clear() error
}

type CacheManager interface {
	Extend(name string, cache Cache) CacheManager
	Disk(name string) CacheManager
	Cache
}
type CacheItemCrypto interface {
	Encode(v any) ([]byte, error)
	Decode(data []byte, v *CacheItem) error
}

type CacheItem struct {
	// data
	Data any `json:"data"`
	//expired ttl
	TTL time.Duration `json:"ttl"`
	// now data
	JoinTime time.Time `json:"join_time"`
	// expired data
	ExpirationTime time.Time `json:"expiration_time"`
	//Is it indefinite
	IsIndefinite bool `json:"is_indefinite"`
}

func (c *CacheItem) CheckIndefinite() {
	if c.TTL == time.Duration(0) || c.TTL == (86400*365*20)*time.Second {
		c.IsIndefinite = true
	}
}

func GetCacheItem(crypto CacheItemCrypto, data any) (item CacheItem, err error) {
	if bytes, ok := data.([]byte); ok {
		err = crypto.Decode(bytes, &item)
		if err != nil {
			return item, err
		}
		if item.ExpirationTime.Before(time.Now()) && !item.IsIndefinite {
			return item, ErrKeyExpired
		}
		return item, nil
	}
	return item, fmt.Errorf("data must be []byte")
}

type CacheItemEncryption struct{}

func (c CacheItemEncryption) Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (c CacheItemEncryption) Decode(data []byte, v *CacheItem) error {
	return json.Unmarshal(data, v)
}
