package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCacheGet(t *testing.T) {
	testCases := []struct {
		name    string
		key     string
		value   string
		cache   Cache
		wantErr error
	}{
		{
			name:    "key not exist",
			key:     "key0",
			wantErr: ErrKeyNotExist,
			cache: func() Cache {
				bm := NewMemoryCache(1 * time.Second)
				return bm
			}(),
		},
		{
			name: "key expire",
			key:  "key1",
			cache: func() Cache {
				bm := NewMemoryCache(20 * time.Second)
				err := bm.Set("key1", "value1", 1*time.Second)
				time.Sleep(2 * time.Second)
				assert.Nil(t, err)
				return bm
			}(),
			wantErr: ErrKeyExpired,
		},
		{
			name:  "get val",
			key:   "key2",
			value: "author",
			cache: func() Cache {
				bm := NewMemoryCache(1 * time.Second)
				err := bm.Set("key2", "author", 5*time.Second)
				assert.Nil(t, err)
				return bm
			}(),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := tc.cache.Get(tc.key)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.value, val)
		})
	}
}

func TestMemoryCacheConcurrencyIncr(t *testing.T) {
	bm := NewMemoryCache(20)
	err := bm.Set("cacheIncr", 0, time.Second*20)
	assert.Nil(t, err)
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			_ = bm.Increment("cacheIncr", 1)
		}()
	}
	wg.Wait()
	val, _ := bm.Get("cacheIncr")
	if val.(int) != 10 {
		t.Error("Incr err")
	}
}
