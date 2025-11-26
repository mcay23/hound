package database

import (
	"encoding/json"
	"errors"
	"hound/helpers"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v4"
)

var db *badger.DB

func InitializeCache() {
	opts := badger.DefaultOptions("cache_data")
	opts.Logger = nil

	var err error
	db, err = badger.Open(opts)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Error initializing cache")
	}
	// 10 minute GC cleanup
	gcIntervalSec := 600
	if val := os.Getenv("CACHE_GC_INTERVAL"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			gcIntervalSec = parsed
		}
	}

	go func() {
		ticker := time.NewTicker(time.Duration(gcIntervalSec) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			for {
				err := db.RunValueLogGC(0.5)
				if err != nil {
					break
				}
			}
		}
	}()
	slog.Info("Cache Initialized")
}

func ClearCache() {
	db.RunValueLogGC(0.5)
	db.DropAll()
}

// Stores a key-value pair with TTL, update if key exists
// returns whether the key exists in bool
func SetCache(key string, value interface{}, ttl time.Duration) (bool, error) {
	if db == nil {
		return false, errors.New("cache not initialized")
	}
	data, err := json.Marshal(value)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "set cache: failed to marshal json for key: "+key)
		return false, err
	}
	err = db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), data)
		if ttl > 0 {
			e.WithTTL(ttl)
		}
		return txn.SetEntry(e)
	})
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "set cache: failed to set/update cache for key: "+key)
		return false, err
	}
	slog.Info("Cache set", "key", key)
	return true, nil
}

// Retrieves a key and unmarshals JSON into the provided interface
// returns whether the key exists in bool
// handles the error logging, since we don't usually want failed cache to end in error response
func GetCache(key string, out interface{}) (bool, error) {
	if db == nil {
		panic("Error: GetCache() called while cache not initialized")
	}
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, out)
		})
	})
	if err == badger.ErrKeyNotFound {
		slog.Info("Cache not found", "key", key)
		// don't treat as an actual error
		return false, nil
	}
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Error getting cache key:"+key)
		return false, err
	}
	// slog.Info("Cache found", "key", key)
	return true, nil
}
