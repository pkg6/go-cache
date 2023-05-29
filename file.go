package cache

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	fileCacheSuffix        = ".bin"
	fileCacheTempDirAppend = "gcache"
)

var (
	FileCachePath = os.TempDir()
)

type FileCache struct {
	Path      string
	CacheItem ICacheItem
}
type FileCacheOptions func(c *FileCache)

func FileCacheWithCachePath(cachePath string) FileCacheOptions {
	return func(c *FileCache) {
		c.Path = cachePath
	}
}
func FileCacheWithCacheItem(cacheItem ICacheItem) FileCacheOptions {
	return func(c *FileCache) {
		c.CacheItem = cacheItem
	}
}

func NewFileCache(opts ...FileCacheOptions) Cache {
	c := &FileCache{
		Path:      FileCachePath,
		CacheItem: &CacheItem{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
func (f *FileCache) Name() string {
	return FileCacheName
}
func (f *FileCache) Get(key string) (any, error) {
	item, err := f.getCacheItem(key)
	if err != nil {
		return nil, err
	}
	return item.GetData(), nil
}

func (f *FileCache) Set(key string, val any, ttl time.Duration) error {
	dataStr, err := f.CacheItem.SetCacheItem(val, ttl)
	if err != nil {
		return err
	}
	filename, err := f.getCacheKey(key)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, []byte(dataStr), os.ModePerm)
}

func (f *FileCache) Delete(key string) error {
	filename, err := f.getCacheKey(key)
	if err != nil {
		return err
	}
	if ok, _ := fileExist(filename); ok {
		err = os.Remove(filename)
		if err != nil {
			return fmt.Errorf("can not delete this file cache key-value, key is %s and file name is %s", key, filename)
		}
	}
	return nil
}

func (f *FileCache) Clear() error {
	return os.RemoveAll(f.savePath())
}

func (f *FileCache) GetMulti(keys []string) ([]any, error) {
	values := make([]any, len(keys))
	keysErr := make([]string, 0)
	for i, key := range keys {
		val, err := f.Get(key)
		if err != nil {
			keysErr = append(keysErr, fmt.Sprintf("key [%s] error: %s", key, err.Error()))
			continue
		}
		values[i] = val
	}
	if len(keysErr) == 0 {
		return values, nil
	}
	return values, errors.New(strings.Join(keysErr, "; "))
}

func (f *FileCache) DeleteMultiple(keys []string) error {
	keysErr := make([]string, 0)
	for _, key := range keys {
		err := f.Delete(key)
		if err != nil {
			keysErr = append(keysErr, fmt.Sprintf("key [%s] error: %s", key, err.Error()))
		}
	}
	if len(keysErr) > 0 {
		return errors.New(strings.Join(keysErr, "; "))
	}
	return nil
}

func (f *FileCache) Increment(key string, step int) error {
	item, err := f.getCacheItem(key)
	if err != nil {
		return f.Set(key, step, 0)
	}
	val, err := Increment(item.GetData(), step)
	if err != nil {
		return err
	}
	return f.Set(key, val, item.GetTTL())
}

func (f *FileCache) Decrement(key string, step int) error {
	item, err := f.getCacheItem(key)
	if err != nil {
		return f.Set(key, step, 0)
	}
	val, err := Decrement(item.GetData(), step)
	if err != nil {
		return err
	}
	return f.Set(key, val, item.GetTTL())
}

func (f *FileCache) Has(key string) (bool, error) {
	if _, err := f.Get(key); err != nil {
		return false, err
	}
	return true, nil
}

func (f *FileCache) savePath() string {
	var paths []string
	if f.Path == "" || f.Path == os.TempDir() {
		paths = append(paths, os.TempDir(), fileCacheTempDirAppend)
	} else {
		paths = append(paths, f.Path)
	}
	return filepath.Join(paths...)
}

func (f *FileCache) getCacheKey(key string) (string, error) {
	m := md5.New()
	_, _ = io.WriteString(m, key)
	keyHash := fmt.Sprintf("%x", m.Sum(nil))
	path := filepath.Join(f.savePath(), keyHash[0:2])
	if err := ensureDirectory(path); err != nil {
		return "", err
	}
	return filepath.Join(path, fmt.Sprintf("%s%s", keyHash, fileCacheSuffix)), nil
}

func (f *FileCache) getCacheItem(key string) (item ICacheItem, err error) {
	filename, err := f.getCacheKey(key)
	if err != nil {
		return item, err
	}
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return item, err
	}
	return f.CacheItem.GetCacheItem(fileData)
}

func ensureDirectory(path string) error {
	var err error
	if _, err = os.Stat(path); os.IsNotExist(err) {
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			return fmt.Errorf("create directory %s err=%v", path, err)
		}
	}
	return err
}

//Determine if the file exists
func fileExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("file cache path is invalid: %s", path)
}
