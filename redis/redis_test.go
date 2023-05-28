package redis

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/pkg6/go-cache"

	"github.com/gomodule/redigo/redis"
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

	dialFunc := func() (c redis.Conn, err error) {
		c, err = redis.Dial("tcp", s.dsn)
		if err != nil {
			return nil, fmt.Errorf("could not dial to remote %s server: %s ", s.driver, s.dsn)
		}
		_, selecterr := c.Do("SELECT", 0)
		if selecterr != nil {
			_ = c.Close()
			return nil, selecterr
		}
		return
	}

	// initialize a new pool
	pool := &redis.Pool{
		Dial:        dialFunc,
		MaxIdle:     3,
		IdleTimeout: 3 * time.Second,
	}
	c := pool.Get()
	defer func() {
		_ = c.Close()
	}()

	// test connection
	err := c.Err()
	for err != nil && maxTryCnt > 0 {
		log.Printf("redis connection exception...")
		c := pool.Get()
		err = c.Err()
		maxTryCnt--
		pool.Stats()
	}
	if err != nil {
		t.Fatal(err)
	}

	bm := NewRedisCache(CacheWithRedisPool(pool))
	if err != nil {
		t.Fatal(err)
	}
	s.cache = bm
}

type RedisCompositionTestSuite struct {
	Suite
}

func (s *RedisCompositionTestSuite) TestRedisCacheGet() {
	testCases := []struct {
		name            string
		key             string
		value           string
		timeoutDuration time.Duration
		wantErr         error
	}{
		//{
		//	name: "get return err",
		//	key:  "key0",
		//	wantErr: func() error {
		//		err := errors.New("the key not exist")
		//		return berror.Wrapf(err, cache.RedisCacheCurdFailed,
		//			"could not execute this command: %s", "GET")
		//	}(),
		//	timeoutDuration: 1 * time.Second,
		//},
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
			vs, _ := redis.String(val, err)
			assert.Equal(t, tc.value, vs)
		})
	}
}

func (s *RedisCompositionTestSuite) TestRedisCacheHas() {
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

func (s *RedisCompositionTestSuite) TestRedisCacheDelete() {
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
		s.T().Run(tc.name, func(t *testing.T) {
			err := s.cache.Set(tc.key, tc.value, tc.timeoutDuration)
			assert.Nil(t, err)

			err = s.cache.Delete(tc.key)
			assert.Nil(t, err)
		})
	}
}

func (s *RedisCompositionTestSuite) TestRedisCacheGetMulti() {
	testCases := []struct {
		name            string
		keys            []string
		values          []string
		timeoutDuration time.Duration
		wantErr         error
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
				vs, _ := redis.String(v, err)
				values = append(values, vs)
			}
			assert.Equal(t, tc.values, values)
		})
	}
}

func (s *RedisCompositionTestSuite) TestRedisCacheIncrAndDecr() {
	testCases := []struct {
		name            string
		key             string
		value           int
		timeoutDuration time.Duration
		wantErr         error
	}{
		{
			name:            "incr and decr",
			key:             "key",
			value:           1,
			timeoutDuration: 5 * time.Second,
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			err := s.cache.Set(tc.key, tc.value, tc.timeoutDuration)
			assert.Nil(t, err)

			val, err := s.cache.Get(tc.key)
			assert.Nil(t, err)
			v1, _ := redis.Int(val, err)
			assert.Equal(t, tc.value, v1)

			assert.Nil(t, s.cache.Increment(tc.key, 1))

			val, err = s.cache.Get(tc.key)
			assert.Nil(t, err)
			v2, _ := redis.Int(val, err)
			assert.Equal(t, v1+1, v2)

			assert.Nil(t, s.cache.Decrement(tc.key, 1))

			val, err = s.cache.Get(tc.key)
			assert.Nil(t, err)
			v3, _ := redis.Int(val, err)
			assert.Equal(t, tc.value, v3)
		})
	}
}

func (s *RedisCompositionTestSuite) TestCacheScan() {
	t := s.T()
	timeoutDuration := 10 * time.Second

	// insert all
	for i := 0; i < 100; i++ {
		assert.Nil(t, s.cache.Set(fmt.Sprintf("astaxie%d", i), fmt.Sprintf("author%d", i), timeoutDuration))
	}
	time.Sleep(time.Second)
	// scan all for the first time
	keys, err := s.cache.(*Cache).Scan(DefaultKey + ":*")
	assert.Nil(t, err)

	assert.Equal(t, 100, len(keys), "scan all error")

	// clear all
	assert.Nil(t, s.cache.Clear())

	// scan all for the second time
	keys, err = s.cache.(*Cache).Scan(DefaultKey + ":*")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(keys))
}

func TestRedisComposition(t *testing.T) {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}
	suite.Run(t, &RedisCompositionTestSuite{
		Suite{
			driver: "redis",
			dsn:    redisAddr,
		},
	})
}
