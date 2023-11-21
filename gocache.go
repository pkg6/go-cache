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

func New() *GoCache {
	return &GoCache{Maps: make(map[string]Cache)}
}

func NewCache(caches ...Cache) *GoCache {
	c := New()
	for _, cache := range caches {
		c.Extend(cache)
	}
	return c
}

// Extend 扩展
func (f *GoCache) Extend(cache Cache, names ...string) *GoCache {
	name := cache.Name()
	if len(names) > 0 {
		name = names[0]
	}
	f.Maps[name] = cache
	f.Names = append(f.Names, name)
	return f
}

func (f *GoCache) Cache(name string) (Cache, error) {
	f.name = name
	if f.name == "" {
		if len(f.Names) > 0 {
			f.name = f.Names[0]
		}
	}
	if cache, ok := f.Maps[f.name]; ok {
		return cache, nil
	}
	return nil, fmt.Errorf("unable to find %s cache", name)
}

// Pull 读取缓存并删除
func (f *GoCache) Pull(key string) (any, error) {
	adapter, err := f.Cache("")
	if err != nil {
		return nil, err
	}
	if val, err := adapter.Get(key); err != nil {
		return val, err
	} else {
		_ = adapter.Delete(key)
		return val, nil
	}
}

func (f *GoCache) Remember(key string, value any, ttl time.Duration) (any, error) {
	adapter, err := f.Cache("")
	if err != nil {
		return nil, err
	}
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

func (f *GoCache) Set(key string, value any, ttl time.Duration) error {
	adapter, err := f.Cache("")
	if err != nil {
		return err
	}
	return adapter.Set(key, value, ttl)
}

func (f *GoCache) Has(key string) (bool, error) {
	adapter, err := f.Cache("")
	if err != nil {
		return false, err
	}
	return adapter.Has(key)
}

func (f *GoCache) GetMulti(keys []string) ([]any, error) {
	adapter, err := f.Cache("")
	if err != nil {
		return nil, err
	}
	return adapter.GetMulti(keys)
}

func (f *GoCache) Get(key string) (any, error) {
	adapter, err := f.Cache("")
	if err != nil {
		return nil, err
	}
	return adapter.Get(key)
}

func (f *GoCache) Delete(key string) error {
	adapter, err := f.Cache("")
	if err != nil {
		return err
	}
	return adapter.Delete(key)
}

func (f *GoCache) Increment(key string, step int) error {
	adapter, err := f.Cache("")
	if err != nil {
		return err
	}
	return adapter.Increment(key, step)
}

func (f *GoCache) Decrement(key string, step int) error {
	adapter, err := f.Cache("")
	if err != nil {
		return err
	}
	return adapter.Decrement(key, step)
}

func (f *GoCache) Clear() error {
	adapter, err := f.Cache("")
	if err != nil {
		return err
	}
	return adapter.Clear()
}
