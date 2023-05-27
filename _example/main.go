package main

import (
	"github.com/pkg6/go-cache"
)

func main() {
	c := cache.New()
	c.Extend(cache.FileCacheName, cache.NewFileCache())
	c.Set("cache", "test", 1)
	c.Get("cache")
	c.GetMulti([]string{"cache"})
	c.Delete("cache")
	c.Has("cache")
	c.Increment("cache_inc", 1)
	c.Decrement("cache_dec", 1)
	c.Clear()
}
