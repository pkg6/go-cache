package memcache

import (
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/pkg6/go-cache"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
	driver string
	dsn    string
	cache  cache.Cache
}

func (s *Suite) SetupSuite() {
	t := s.T()
	maxTryCnt := 10
	pool := memcache.New(s.dsn)
	// test connection
	err := pool.Ping()
	for err != nil && maxTryCnt > 0 {
		log.Printf("memcache connection exception...")
		err = pool.Ping()
		maxTryCnt--
	}
	if err != nil {
		t.Fatal(err)
	}

	bm := NewMemCache(CacheWithMemcacheClient(pool))

	s.cache = bm

}

type MemcacheCompositionTestSuite struct {
	Suite
}

func (s *MemcacheCompositionTestSuite) TearDownTest() {
	// test clear all
	assert.Nil(s.T(), s.cache.Clear())
}

func (s *MemcacheCompositionTestSuite) TestMemcacheCacheGet() {
	testCases := []struct {
		name            string
		key             string
		value           string
		timeoutDuration time.Duration
	}{
		{
			name:            "get val",
			key:             "key1",
			value:           "author",
			timeoutDuration: 5 * time.Second,
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			err := s.cache.Set(tc.key, tc.value, tc.timeoutDuration)
			assert.Nil(t, err)
			time.Sleep(2 * time.Second)

			val, err := s.cache.Get(tc.key)
			assert.Nil(t, err)
			vs := val.([]byte)
			assert.Equal(t, tc.value, string(vs))
		})
	}
}

func (s *MemcacheCompositionTestSuite) TestMemcacheCacheIsExist() {
	testCases := []struct {
		name            string
		key             string
		value           string
		timeoutDuration time.Duration
		isExist         bool
	}{
		{
			name:            "not exist",
			key:             "key0",
			value:           "value0",
			timeoutDuration: 1 * time.Second,
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
		s.T().Run(tc.name, func(t *testing.T) {
			err := s.cache.Set(tc.key, tc.value, tc.timeoutDuration)
			assert.Nil(t, err)

			time.Sleep(2 * time.Second)

			res, _ := s.cache.Has(tc.key)
			assert.Equal(t, res, tc.isExist)
		})
	}
}

func (s *MemcacheCompositionTestSuite) TestMemcacheCacheIncrAndDecr() {
	testCases := []struct {
		name            string
		key             string
		value           string
		timeoutDuration time.Duration
		wantErr         error
	}{
		{
			name:            "incr and decr",
			key:             "key",
			value:           "1",
			timeoutDuration: 5 * time.Second,
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			err := s.cache.Set(tc.key, tc.value, tc.timeoutDuration)
			assert.Nil(t, err)
			v, _ := strconv.Atoi(tc.value)

			val, err := s.cache.Get(tc.key)
			assert.Nil(t, err)
			v1, _ := strconv.Atoi(string(val.([]byte)))
			assert.Equal(t, v, v1)

			assert.Nil(t, s.cache.Increment(tc.key, 1))

			val, err = s.cache.Get(tc.key)
			assert.Nil(t, err)
			v2, _ := strconv.Atoi(string(val.([]byte)))
			assert.Equal(t, v1+1, v2)

			assert.Nil(t, s.cache.Decrement(tc.key, 1))

			val, err = s.cache.Get(tc.key)
			assert.Nil(t, err)
			v3, _ := strconv.Atoi(string(val.([]byte)))
			assert.Equal(t, v, v3)
		})
	}
}

func (s *MemcacheCompositionTestSuite) TestMemcacheCacheDelete() {
	testCases := []struct {
		name            string
		key             string
		value           string
		timeoutDuration time.Duration
		wantErr         error
	}{
		{
			name:            "delete val",
			key:             "key1",
			value:           "author",
			timeoutDuration: 5 * time.Second,
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			err := s.cache.Set(tc.key, tc.value, tc.timeoutDuration)
			assert.Nil(t, err)
			err = s.cache.Delete(tc.key)
			assert.Nil(t, err)
		})
	}
}

func (s *MemcacheCompositionTestSuite) TestMemcacheCacheGetMulti() {
	testCases := []struct {
		name            string
		keys            []string
		values          []string
		timeoutDuration time.Duration
	}{
		{
			name:            "get multi val",
			keys:            []string{"key2", "key3"},
			values:          []string{"value2", "value3"},
			timeoutDuration: 5 * time.Second,
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			for idx, key := range tc.keys {
				value := tc.values[idx]
				err := s.cache.Set(key, value, tc.timeoutDuration)
				assert.Nil(t, err)
			}

			time.Sleep(2 * time.Second)

			vals, err := s.cache.GetMulti(tc.keys)
			assert.Nil(t, err)
			values := make([]string, 0, len(tc.values))
			for _, v := range vals {
				vs := v.([]byte)
				values = append(values, string(vs))
			}
			assert.Equal(t, tc.values, values)
		})
	}
}

func TestSsdbComposition(t *testing.T) {
	memCacheAddr := os.Getenv("MEMCACHE_ADDR")
	if memCacheAddr == "" {
		memCacheAddr = "127.0.0.1:11211"
	}
	suite.Run(t, &MemcacheCompositionTestSuite{
		Suite{
			driver: "memcache",
			dsn:    memCacheAddr,
		},
	})
}
