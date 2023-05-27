package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFileCacheGet(t *testing.T) {
	testCases := []struct {
		name    string
		key     string
		value   string
		cache   Cache
		wantErr error
	}{
		{
			name:  "get val",
			key:   "key1",
			value: "author",
			cache: func() Cache {
				bm, err := NewFileCache(
					FileCacheWithCachePath("cache"),
					FileCacheWithFileSuffix(".bin"),
					FileCacheWithEmbedExpiry(0))
				assert.Nil(t, err)
				err = bm.Set("key1", "author", 5*time.Second)
				assert.Nil(t, err)
				return bm
			}(),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := tc.cache.Get(tc.key)
			assert.Nil(t, err)
			assert.Equal(t, tc.value, val)
		})
	}
	assert.Nil(t, os.RemoveAll("cache"))
}

func TestFileCacheIsExist(t *testing.T) {
	cache, err := NewFileCache(
		FileCacheWithCachePath("cache"),
		FileCacheWithFileSuffix(".bin"),
		FileCacheWithEmbedExpiry(0))
	assert.Nil(t, err)
	testCases := []struct {
		name            string
		key             string
		value           string
		timeoutDuration time.Duration
		isExist         bool
	}{
		{
			name:            "expired",
			key:             "key0",
			value:           "value0",
			timeoutDuration: 1 * time.Second,
			isExist:         true,
		},
		{
			name:            "exist",
			key:             "key1",
			value:           "author",
			timeoutDuration: 5 * time.Second,
			isExist:         true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := cache.Set(tc.key, tc.value, tc.timeoutDuration)
			assert.Nil(t, err)
			time.Sleep(2 * time.Second)
		})
	}
	assert.Nil(t, os.RemoveAll("cache"))
}

func TestFileCacheDelete(t *testing.T) {
	cache, err := NewFileCache(
		FileCacheWithCachePath("cache"),
		FileCacheWithFileSuffix(".bin"),
		FileCacheWithEmbedExpiry(0))
	assert.Nil(t, err)
	testCases := []struct {
		name            string
		key             string
		value           string
		timeoutDuration time.Duration
	}{
		{
			name:            "delete val",
			key:             "key1",
			value:           "author",
			timeoutDuration: 5 * time.Second,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := cache.Set(tc.key, tc.value, tc.timeoutDuration)
			assert.Nil(t, err)
			err = cache.Delete(tc.key)
			assert.Nil(t, err)
		})
	}
	assert.Nil(t, os.RemoveAll("cache"))
}

func TestFileCacheGetMulti(t *testing.T) {
	cache, err := NewFileCache(
		FileCacheWithCachePath("cache"),
		FileCacheWithFileSuffix(".bin"),
		FileCacheWithEmbedExpiry(0))
	assert.Nil(t, err)
	testCases := []struct {
		name            string
		keys            []string
		values          []any
		timeoutDuration time.Duration
	}{
		{
			name:            "key expired",
			keys:            []string{"key0", "key1"},
			values:          []any{"value0", "value1"},
			timeoutDuration: 1 * time.Second,
		},
		{
			name:            "get multi val",
			keys:            []string{"key2", "key3"},
			values:          []any{"value2", "value3"},
			timeoutDuration: 5 * time.Second,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for idx, key := range tc.keys {
				value := tc.values[idx]
				err := cache.Set(key, value, tc.timeoutDuration)
				assert.Nil(t, err)
			}
			time.Sleep(2 * time.Second)
			vals, err := cache.GetMulti(tc.keys)
			if err != nil {
				assert.ErrorContains(t, err, ErrKeyExpired.Error())
				return
			}
			assert.Equal(t, tc.values, vals)
		})
	}
	assert.Nil(t, os.RemoveAll("cache"))
}

func TestFileGetContents(t *testing.T) {
	_, err := os.ReadFile("/bin/aaa")
	assert.NotNil(t, err)
	fn := filepath.Join(os.TempDir(), "fileCache.txt")
	f, err := os.Create(fn)
	assert.Nil(t, err)
	_, err = f.WriteString("text")
	assert.Nil(t, err)
	data, err := os.ReadFile(fn)
	assert.Nil(t, err)
	assert.Equal(t, "text", string(data))
}

func TestGobEncodeDecode(t *testing.T) {
	_, err := GobEncode(func() {
		fmt.Print("test func")
	})
	assert.NotNil(t, err)
	data, err := GobEncode(&CacheItem{
		Data: "hello",
	})
	assert.Nil(t, err)
	err = GobDecode([]byte("wrong data"), &CacheItem{})
	assert.NotNil(t, err)
	dci := &CacheItem{}
	err = GobDecode(data, dci)
	assert.Nil(t, err)
	assert.Equal(t, "hello", dci.Data)
}
