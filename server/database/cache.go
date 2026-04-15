package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/mcay23/hound/internal"

	"github.com/dgraph-io/badger/v4"
)

var db *badger.DB

func InitializeCache() {
	opts := badger.DefaultOptions(filepath.Join("Hound Data", "cache_data"))
	opts.Logger = nil

	var err error
	db, err = badger.Open(opts)
	if err != nil {
		_ = internal.LogErrorWithMessage(err, "Error initializing cache")
		panic(err)
	}
	// 10 minute GC cleanup
	go func() {
		ticker := time.NewTicker(time.Duration(600) * time.Second)
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
		return false, fmt.Errorf("cache not initialized: %w", internal.InternalServerError)
	}
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("set cache for key %s: %w", key, err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), data)
		if ttl > 0 {
			e.WithTTL(ttl)
		}
		return txn.SetEntry(e)
	})
	if err != nil {
		return false, fmt.Errorf("set cache for key %s: %w", key, err)
	}
	slog.Debug("cache set", "key", key)
	return true, nil
}

// Retrieves a key and unmarshals JSON into the provided interface.
// Returns whether the key exists in bool
// Handles the error logging, since we don't usually want failed cache to end in error response
func GetCache(key string, out interface{}) (bool, error) {
	if db == nil {
		panic("Error: GetCache() called while cache not initialized")
	}
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return fmt.Errorf("get cache for key %s: %w", key, err)
		}
		return item.Value(func(val []byte) error {
			err := json.Unmarshal(val, out)
			if err != nil {
				return fmt.Errorf("get cache for key %s: %w", key, err)
			}
			return nil
		})
	})
	if errors.Is(err, badger.ErrKeyNotFound) {
		// don't treat as an actual error
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("get cache for key %s: %w", key, err)
	}
	slog.Debug("cache found", "key", key)
	return true, nil
}

// Returns all keys starting with the given prefix
func GetKeysWithPrefix(prefix string) ([]string, error) {
	if db == nil {
		return nil, fmt.Errorf("cache not initialized")
	}
	var keys []string
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		prefixBytes := []byte(prefix)
		for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
			item := it.Item()
			key := item.Key()
			keys = append(keys, string(key))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("get cache for keys with prefix %s: %w", prefix, err)
	}
	return keys, nil
}

// Deletes a key from the cache
func DeleteCache(key string) error {
	if db == nil {
		return errors.New("cache not initialized")
	}
	err := db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	if err != nil {
		return fmt.Errorf("delete cache for key %s: %w", key, err)
	}
	slog.Debug("cache deleted", "key", key)
	return nil
}
