package model

import (
	"bytes"
	"encoding/gob"
	"log"
	"github.com/dgraph-io/badger/v4"
	"time"
)

var db *badger.DB

func InitializeCache() {
	var err error
	db, err = badger.Open(badger.DefaultOptions("/tmp/badger"))
	if err != nil {
		log.Fatal(err)
	}
	go startBadgerGC()
}

func startBadgerGC() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
	again:
		err := db.RunValueLogGC(0.5)
		if err == nil {
			goto again
		}
	}
}

func encode(value interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(value)
	return buf.Bytes(), err
}

func decode(data []byte, out interface{}) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(out)
}

func SetCache(key string, value interface{}, ttl time.Duration) error {
	data, err := encode(value)
	if err != nil {
		return err
	}

	return db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), data).WithTTL(ttl)
		return txn.SetEntry(e)
	})
}

func GetCache(key string) (interface{}, bool) {
	return cacheObject.Get(key)
}

func UpdateOrSetCache(key string, value interface{}, ttl time.Duration) error {
	cacheObject.Remove(key)
	return SetCache(key, value, ttl)
}

//
//// In-memory cache layer
//var cacheObject *mcache.CacheDriver
//
//func InitializeCache() {
//	cacheObject = mcache.New()
//}
//
//func SetCache(key string, value interface{}, ttl time.Duration) error {
//	err := cacheObject.Set(key, value, ttl)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func GetCache(key string) (interface{}, bool) {
//	return cacheObject.Get(key)
//}
//
//func UpdateOrSetCache(key string, value interface{}, ttl time.Duration) error {
//	cacheObject.Remove(key)
//	return SetCache(key, value, ttl)
//}
