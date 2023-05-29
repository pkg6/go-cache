package cache

import (
	"encoding/json"
	"fmt"
	"time"
)

var IndefiniteTime = (86400 * 365 * 20) * time.Second

type ICacheItem interface {
	SetCacheItem(data any, ttl time.Duration) (string, error)
	GetCacheItem(data any) (ICacheItem, error)
	GetExpirationTime() time.Time
	IsNeverExpires() bool
	GetTTL() time.Duration
	GetData() any
	GetJoinTime() time.Time
}

type CacheItem struct {
	// data
	Data any `json:"data"`
	//expired ttl
	TTL time.Duration `json:"ttl"`
	// now data
	JoinTime time.Time `json:"join_time"`
	// expired data
	ExpirationTime time.Time `json:"expiration_time"`
	//Is it indefinite
	NeverExpires bool `json:"never_expires"`
}

func (c *CacheItem) GetTTL() time.Duration {
	return c.TTL
}
func (c *CacheItem) GetData() any {
	return c.Data
}
func (c *CacheItem) GetJoinTime() time.Time {
	return c.JoinTime
}

func (c *CacheItem) GetExpirationTime() time.Time {
	return c.ExpirationTime
}
func (c *CacheItem) IsNeverExpires() bool {
	return c.NeverExpires
}

func (c *CacheItem) SetCacheItem(data any, ttl time.Duration) (string, error) {
	c.Data = data
	c.JoinTime = time.Now()
	c.TTL = ttl
	if c.TTL == time.Duration(0) || c.TTL == IndefiniteTime {
		c.NeverExpires = true
		c.TTL = IndefiniteTime
	}
	c.ExpirationTime = c.JoinTime.Add(c.TTL)
	marshal, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}

func (c *CacheItem) GetCacheItem(data any) (item ICacheItem, err error) {
	item = new(CacheItem)
	if bytes, ok := data.([]byte); ok {
		err := json.Unmarshal(bytes, &item)
		if err != nil {
			return item, err
		}
		if item.GetExpirationTime().Before(time.Now()) && !item.IsNeverExpires() {
			return item, ErrKeyExpired
		}
		return item, nil
	}
	return item, fmt.Errorf("data must be []byte")
}
