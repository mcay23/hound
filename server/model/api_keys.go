package model

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/mcay23/hound/database"
)

func ValidateAPIKey(apiKey string) (*database.APIKey, error) {
	key, err := database.GetAPIKey(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get api key record: %w", err)
	}
	return key, nil
}

func CreateAPIKey(userID int64, name string, expiresAt *time.Time) (*database.APIKey, error) {
	if expiresAt == nil {
		defaultDuration := time.Duration(3650 * 24 * time.Hour)
		temp := time.Now().Add(defaultDuration)
		expiresAt = &temp
	}
	// generate random 32 bytes
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	apiKey := base64.URLEncoding.EncodeToString(b)
	record := &database.APIKey{
		APIKey:    apiKey,
		Name:      name,
		UserID:    userID,
		ExpiresAt: *expiresAt,
	}
	err = database.InsertAPIKey(record)
	if err != nil {
		return nil, fmt.Errorf("failed to insert api key record: %w", err)
	}
	return record, nil
}
