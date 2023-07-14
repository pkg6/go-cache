package cache

import (
	"fmt"
	"sync"
	"time"
)

type GoCache struct {
	name  string
	Maps  map[string]Cache
	Names []string
	l     sync.RWMutex
}

func New() CacheManager {
	return &GoCache{Maps: make(map[string]Cache)}
}

func NewCache(caches ...Cache) CacheManager {
	c := New()
	for _, cache := range caches {
		c.Extend(cache)
	}
	return c
}

// Extend 扩展
func (f *GoCache) Extend(cache Cache, names ...string) CacheManager {
	name := cache.Name()
	if len(names) > 0 {
		name = names[0]
	}
	f.Maps[name] = cache
	f.Names = append(f.Names, name)
	return f
}
func (f *GoCache) Name() string {
	return CacheName
}

func (f *GoCache) Disk(name string) CacheManager {
	return &GoCache{
		name:  name,
		Maps:  f.Maps,
		Names: f.Names,
	}
}

// Pull 读取缓存并删除
func (f *GoCache) Pull(key string) (any, error) {
	adapter := f.FindAdapter()
	if val, err := adapter.Get(key); err != nil {
		return val, err
	} else {
		_ = adapter.Delete(key)
		return val, nil
	}
}

func (f *GoCache) Remember(key string, value any, ttl time.Duration) (any, error) {
	adapter := f.FindAdapter()
	has, _ := adapter.Has(key)
	if has {
		if val, err := adapter.Get(key); err != nil {
			return val, err
		} else {
			return val, nil
		}
	}
	f.l.Lock()
	defer f.l.Unlock()
	if valFun, ok := value.(func() any); ok {
		value = valFun()
	}
	if err := adapter.Set(key, value, ttl); err != nil {
		return value, err
	}
	return value, nil
}

// FindAdapter Find Adapter
func (f *GoCache) FindAdapter() Cache {
	var name string
	if f.name == "" {
		if len(f.Names) > 0 {
			f.name = f.Names[0]
		}
	}
	if adapter, ok := f.Maps[f.name]; ok {
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
