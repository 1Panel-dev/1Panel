package badger_db

import (
	"github.com/patrickmn/go-cache"
	"strings"
	"time"
)

type Cache struct {
	db *cache.Cache
}

func NewCacheDB() *Cache {
	db := cache.New(5*time.Minute, 10*time.Minute)
	return &Cache{
		db: db,
	}
}

func (c *Cache) Set(key string, value interface{}) {
	c.db.Set(key, value, cache.DefaultExpiration)
}

func (c *Cache) SetWithTTL(key string, value interface{}, d time.Duration) {
	c.db.Set(key, value, d)
}
func (c *Cache) Del(key string) {
	c.db.Delete(key)
}

func (c *Cache) Clean() error {
	return nil
}

func (c *Cache) Get(key string) string {
	obj, exist := c.db.Get(key)
	if !exist {
		return ""
	}
	return obj.(string)
}

func (c *Cache) PrefixScanKey(prefixStr string) []string {
	var res []string
	values := c.db.Items()
	for key := range values {
		if strings.HasPrefix(key, prefixStr) {
			res = append(res, key)
		}
	}
	return res
}
