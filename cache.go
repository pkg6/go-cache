package cache

import (
	"bytes"
	"encoding/gob"
	"errors"
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
	Data any
	//expired ttl
	TTL time.Duration
	// now data
	JoinTime time.Time
	// expired data
	ExpirationTime time.Time
}

func (c CacheItem) Encode(v any) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c CacheItem) Decode(data []byte, v *CacheItem) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(&v)
}
