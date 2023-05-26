package cache

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	FileCachePath        = os.TempDir() // cache directory
	FileCacheFileSuffix  = ".bin"       // cache file suffix
	FileCacheEmbedExpiry time.Duration  // cache expire time, default is no expire forever.
)

type FileCache struct {
	Path        string
	Suffix      string
	EmbedExpiry int
}
type FileCacheOptions func(c *FileCache)

// FileCacheWithCachePath configures cachePath for FileCache
func FileCacheWithCachePath(cachePath string) FileCacheOptions {
	return func(c *FileCache) {
		c.Path = cachePath
	}
}

// FileCacheWithFileSuffix configures fileSuffix for FileCache
func FileCacheWithFileSuffix(fileSuffix string) FileCacheOptions {
	return func(c *FileCache) {
		c.Suffix = fileSuffix
	}
}

// FileCacheWithEmbedExpiry configures fileCacheEmbedExpiry for FileCache
func FileCacheWithEmbedExpiry(fileCacheEmbedExpiry int) FileCacheOptions {
	return func(c *FileCache) {
		c.EmbedExpiry = fileCacheEmbedExpiry
	}
}

// NewFileCache creates a new file cache with no config.
// The level and expiry need to be set in the method StartAndGC as config string.
func NewFileCache(opts ...FileCacheOptions) (Cache, error) {
	fileCache := &FileCache{
		Path:   FileCachePath,
		Suffix: FileCacheFileSuffix,
	}
	fileCache.EmbedExpiry, _ = strconv.Atoi(
		strconv.FormatInt(int64(FileCacheEmbedExpiry.Seconds()), 10))
	for _, opt := range opts {
		opt(fileCache)
	}
	if err := pathExistOrMkdir(fileCache.Path); err != nil {
		return fileCache, err
	}
	return fileCache, nil
}

// Set value into file cache.
// timeout: how long this file should be kept in ms
// if timeout equals fc.EmbedExpiry(default is 0), cache this item forever.
func (f *FileCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	item := CacheItem{Data: value}
	if ttl == time.Duration(f.EmbedExpiry) {
		item.Expired = time.Now().Add((86400 * 365 * 10) * time.Second) // ten years
	} else {
		item.Expired = time.Now().Add(ttl)
	}
	item.Lastaccess = time.Now()
	data, err := GobEncode(item)
	if err != nil {
		return err
	}
	filename, err := f.getCacheKey(key)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, os.ModePerm)
}

func (f *FileCache) Has(ctx context.Context, key string) (bool, error) {
	_, err := f.Get(ctx, key)
	if err == nil {
		return true, err
	}
	return false, err
}

// GetMulti gets values from file cache.
// if nonexistent or expired return an empty string.
func (f *FileCache) GetMulti(ctx context.Context, keys []string) ([]any, error) {
	rc := make([]any, len(keys))
	keysErr := make([]string, 0)
	for i, ki := range keys {
		val, err := f.Get(ctx, ki)
		if err != nil {
			keysErr = append(keysErr, fmt.Sprintf("key [%s] error: %s", ki, err.Error()))
			continue
		}
		rc[i] = val
	}
	if len(keysErr) == 0 {
		return rc, nil
	}
	return rc, errors.New(strings.Join(keysErr, "; "))
}

// Get value from file cache.
// if nonexistent or expired return an empty string.
func (f *FileCache) Get(ctx context.Context, key string) (any, error) {
	filename, err := f.getCacheKey(key)
	if err != nil {
		return nil, err
	}
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var to CacheItem
	err = GobDecode(fileData, &to)
	if err != nil {
		return nil, err
	}
	if to.Expired.Before(time.Now()) {
		_ = f.Delete(ctx, key)
		return nil, ErrKeyExpired
	}
	return to.Data, nil
}

// Delete file cache value.
func (f *FileCache) Delete(ctx context.Context, key string) error {
	filename, err := f.getCacheKey(key)
	if ok, _ := fileExist(filename); ok {
		err = os.Remove(filename)
		if err != nil {
			return unwrapF("can not delete this file cache key-value, key is %s and file name is %s", key, filename)
		}
	}
	return nil
}

// Increment increases cached int value.
// fc value is saved forever unless deleted.
func (f *FileCache) Increment(ctx context.Context, key string, step int) error {
	data, err := f.Get(ctx, key)
	if err != nil {
		return err
	}
	val, err := Increment(data, step)
	if err != nil {
		return err
	}
	return f.Set(ctx, key, val, time.Duration(f.EmbedExpiry))
}

// Decrement decreases cached int value.
func (f *FileCache) Decrement(ctx context.Context, key string, step int) error {
	data, err := f.Get(ctx, key)
	if err != nil {
		return err
	}
	val, err := Decrement(data, step)
	if err != nil {
		return err
	}
	return f.Set(ctx, key, val, time.Duration(f.EmbedExpiry))
}

// Clear cleans cached files (not implemented)
func (f *FileCache) Clear(ctx context.Context) {
	_ = os.RemoveAll(f.Path)
}

func (f *FileCache) getCacheKey(key string) (string, error) {
	m := md5.New()
	_, _ = io.WriteString(m, key)
	keyHash := fmt.Sprintf("%x", m.Sum(nil))
	paths := []string{f.Path}
	if f.Path == os.TempDir() {
		paths = append(paths, "gocache")
	}
	paths = append(paths, keyHash[0:2])
	path := filepath.Join(paths...)
	if err := pathExistOrMkdir(path); err != nil {
		return "", err
	}
	return filepath.Join(path, fmt.Sprintf("%s%s", keyHash, f.Suffix)), nil
}

// Determine if the file exists, and if it does not exist, create it
func pathExistOrMkdir(path string) error {
	ok, err := fileExist(path)
	if err != nil {
		return err
	}
	if !ok {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return unwrapF("could not create the directory: %s", path)
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
	return false, wrapF("file cache path is invalid: %s", path)
}