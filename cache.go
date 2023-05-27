package cache

import (
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
)

type Cache interface {
	Set(key string, value any, ttl time.Duration) error
	Has(key string) (bool, error)
	GetMulti(keys []string) ([]any, error)
	Get(key string) (any, error)
	Delete(key string) error
	Increment(key string, step int) error
	Decrement(key string, step int) error
	Clear() error
}

type CacheItem struct {
	// data
	Data any
	// now data
	Lastaccess time.Time
	// expired data
	Expired time.Time
}
