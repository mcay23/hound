package model

import (
	"github.com/OrlovEvgeny/go-mcache"
	"time"
)

// In-memory cache layer
var cacheObject *mcache.CacheDriver

func InitializeCache() {
	cacheObject = mcache.New()
}

func SetCache(key string, value interface{}, ttl time.Duration) error {
	err := cacheObject.Set(key, value, ttl)
	if err != nil {
		return err
	}
	return nil
}

func GetCache(key string) (interface{}, bool) {
	return cacheObject.Get(key)
}

func UpdateOrSetCache(key string, value interface{}, ttl time.Duration) error {
	cacheObject.Remove(key)
	return SetCache(key, value, ttl)
}

/*

 */