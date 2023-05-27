package cache

import (
	"fmt"
	"time"
)

type GoCache struct {
	name  string
	Maps  map[string]Cache
	Names []string
}

func New() CacheManager {
	return &GoCache{Maps: make(map[string]Cache)}
}

func NewCache(caches ...Cache) CacheManager {
	c := New()
	for _, cache := range caches {
		c.Extend(cache.Name(), cache)
	}
	return c
}

// Extend 扩展
func (f *GoCache) Extend(name string, cache Cache) CacheManager {
	f.Maps[name] = cache
	f.Names = append(f.Names, name)
	return f
}
func (f GoCache) Name() string {
	return CacheName
}

func (f *GoCache) Disk(name string) CacheManager {
	return &GoCache{
		name:  name,
		Maps:  f.Maps,
		Names: f.Names,
	}
}

// FindAdapter Find Adapter
func (f *GoCache) FindAdapter() Cache {
	var name string
	if f.name != "" {
		name = f.name
	} else if len(f.Names) > 0 {
		name = f.Names[0]
	}
	if adapter, ok := f.Maps[name]; ok {
		return adapter
	}
	panic(fmt.Sprintf("Unable to find %s cache", name))
}

func (f *GoCache) Set(key string, value any, ttl time.Duration) error {
	return f.FindAdapter().Set(key, value, ttl)
}

func (f *GoCache) Has(key string) (bool, error) {
	return f.FindAdapter().Has(key)
}

func (f *GoCache) GetMulti(keys []string) ([]any, error) {
	return f.FindAdapter().GetMulti(keys)
}

func (f *GoCache) Get(key string) (any, error) {
	return f.FindAdapter().Get(key)
}

func (f *GoCache) Delete(key string) error {
	return f.FindAdapter().Delete(key)
}

func (f *GoCache) Increment(key string, step int) error {
	return f.FindAdapter().Increment(key, step)
}

func (f *GoCache) Decrement(key string, step int) error {
	return f.FindAdapter().Decrement(key, step)
}

func (f *GoCache) Clear() error {
	return f.FindAdapter().Clear()
}
