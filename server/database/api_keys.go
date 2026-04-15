package database

import (
	"fmt"
	"time"

	"github.com/mcay23/hound/internal"
)

const apiKeysTable = "api_keys"

// Note: keys are stored in plaintext
type APIKey struct {
	KeyID     int64      `xorm:"pk autoincr 'key_id'" json:"key_id"`
	APIKey    string     `xorm:"'api_key'" json:"api_key"`
	Name      string     `xorm:"'name'" json:"name"`
	UserID    int64      `xorm:"'user_id'" json:"user_id"`
	RevokedAt *time.Time `xorm:"timestampz 'revoked_at'" json:"revoked_at"`
	ExpiresAt time.Time  `xorm:"timestampz 'expires_at'" json:"expires_at"`
	CreatedAt time.Time  `xorm:"timestampz created" json:"created_at"`
}

func instantiateAPIKeysTable() error {
	return databaseEngine.Table(apiKeysTable).Sync2(new(APIKey))
}

func InsertAPIKey(apiKey *APIKey) error {
	_, err := databaseEngine.Table(apiKeysTable).Insert(apiKey)
	return err
}

func GetUserAPIKeys(userID int64) ([]APIKey, error) {
	var records []APIKey
	err := databaseEngine.Table(apiKeysTable).Where("user_id = ?", userID).Where("revoked_at IS NULL").Find(&records)
	if err != nil {
		return nil, fmt.Errorf("query api keys: %w", err)
	}
	return records, nil
}

func GetAPIKey(apiKey string) (*APIKey, error) {
	cacheKey := fmt.Sprintf("api_key:%s", apiKey)
	var record APIKey
	found, err := GetCache(cacheKey, &record)
	if err != nil {
		return nil, err
	}
	if found {
		return &record, nil
	}
	has, err := databaseEngine.Table(apiKeysTable).Where("api_key = ?", apiKey).Get(&record)
	if err != nil {
		return nil, fmt.Errorf("query api key: %w", err)
	}
	if !has {
		return nil, fmt.Errorf("query api key, key not found: %w", internal.UnauthorizedError)
	}
	if record.RevokedAt != nil {
		return nil, fmt.Errorf("query api key, key revoked: %w", internal.UnauthorizedError)
	}
	// add 5-sec buffer
	now := time.Now().Add(5 * time.Second)
	if record.ExpiresAt.Before(now) {
		return nil, fmt.Errorf("query api key, key expired: %w", internal.UnauthorizedError)
	}
	SetCache(cacheKey, &record, record.ExpiresAt.Sub(now))
	return &record, nil
}

func RevokeAPIKey(keyID int64) error {
	var record APIKey
	_, err := databaseEngine.Table(apiKeysTable).Where("key_id = ?", keyID).Get(&record)
	if err != nil {
		return fmt.Errorf("query api key: %w", err)
	}
	cacheKey := fmt.Sprintf("api_key:%s", record.APIKey)
	_ = DeleteCache(cacheKey)
	now := time.Now()
	_, err = databaseEngine.Table(apiKeysTable).Where("key_id = ?", keyID).Update(&APIKey{RevokedAt: &now})
	if err != nil {
		return fmt.Errorf("failed to delete api key record: %w", err)
	}
	return nil
}
