package cache

import (
	"context"
	"errors"
	"time"
)

var (
	ErrKeyExpired = errors.New("the key is expired")
)

const (
	MinUint32 uint32 = 0
	MinUint64 uint64 = 0
)

type Cache interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Has(ctx context.Context, key string) (bool, error)
	GetMulti(ctx context.Context, keys []string) ([]any, error)
	Get(ctx context.Context, key string) (any, error)
	Delete(ctx context.Context, key string) error
	Increment(ctx context.Context, key string, step int) error
	Decrement(ctx context.Context, key string, step int) error
	Clear(ctx context.Context)
}

type CacheItem struct {
	Data       any
	Lastaccess time.Time
	Expired    time.Time
}
