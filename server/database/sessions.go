package database

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/mcay23/hound/internal"
)

/*
	Session tokens are stored in cache, no db representation
	For each session created, two entries are made:
	- auth_session|session_id|<session_id>
	- So we can check valid session ids without knowing the user

	- auth_session|user_id:<user_id>|session_id:<session_id>
	- So we can check a user's sessions for invalidation when necessary
		by doing a prefix search (fast in BadgerDB)
*/

const (
	AuthSessionTTL = 365 * 24 * time.Hour
)

func getUserAuthSessionCacheKey(userID int64, sessionID string) string {
	return fmt.Sprintf("auth_session|user_id:%d|session_id:%s", userID, sessionID)
}

func getAuthSessionCacheKey(sessionID string) string {
	return fmt.Sprintf("auth_session|session_id:%s", sessionID)
}

type AuthSession struct {
	UserID         int64
	DeviceID       string
	ClientID       string
	ClientPlatform string
}

func ValidateAuthSession(sessionID string) (*AuthSession, error) {
	cacheKey := getAuthSessionCacheKey(sessionID)
	var sessionObject AuthSession
	found, err := GetCache(cacheKey, &sessionObject)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("session not found (expired?): %w", internal.UnauthorizedError)
	}
	return &sessionObject, nil
}

func GenerateAuthSession(userID int64, clientID, clientPlatform, deviceID string) (string, error) {
	b := make([]byte, 32) // 256 bits
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	sessionID := base64.RawURLEncoding.EncodeToString(b)
	sessionCacheKey := getAuthSessionCacheKey(sessionID)
	cacheObject := AuthSession{
		UserID:         userID,
		DeviceID:       deviceID,
		ClientID:       clientID,
		ClientPlatform: clientPlatform,
	}
	_, err := SetCache(sessionCacheKey, cacheObject, AuthSessionTTL)
	if err != nil {
		return "", fmt.Errorf("failed to set cache for auth session: %w", err)
	}
	userSessionCacheKey := getUserAuthSessionCacheKey(userID, sessionID)
	_, err = SetCache(userSessionCacheKey, nil, AuthSessionTTL)
	// set user cache failed, also delete session key
	if err != nil {
		DeleteCache(sessionCacheKey)
		return "", fmt.Errorf("failed to set cache for auth session: %w", err)
	}
	return sessionID, nil
}

func DeleteUserAuthSessions(userID int64) error {
	keys, err := GetKeysWithPrefix(getUserAuthSessionCacheKey(userID, ""))
	if err != nil {
		return err
	}
	for _, userSessionKey := range keys {
		temp := strings.Split(userSessionKey, "|")
		if len(temp) != 3 {
			return fmt.Errorf("invalid cache key format %s: %w", userSessionKey, internal.InternalServerError)
		}
		// auth_session|session_id:<session_id>
		sessKey := temp[0] + "|" + temp[2]
		err := DeleteCache(sessKey)
		if err != nil {
			slog.Error("failed to delete cache for auth session", "error", err)
		}
		err = DeleteCache(userSessionKey)
		if err != nil {
			slog.Error("failed to delete cache for auth session", "error", err)
		}
	}
	return nil
}

func DeleteAuthSession(userID int64, sessionID string) error {
	sessionCacheKey := getAuthSessionCacheKey(sessionID)
	userSessionCacheKey := getUserAuthSessionCacheKey(userID, sessionID)
	err := DeleteCache(sessionCacheKey)
	if err != nil {
		slog.Error("failed to delete cache for auth session", "sessionID", sessionID, "error", err)
	}
	err = DeleteCache(userSessionCacheKey)
	if err != nil {
		slog.Error("failed to delete cache for auth session", "userSessionKey", userSessionCacheKey, "error", err)
	}
	return nil
}
