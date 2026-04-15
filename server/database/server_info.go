package database

import (
	"github.com/google/uuid"
)

var (
	serverIDCacheKey = "hound_server_id"
)

// Server ID should persist through restarts,
// note that hound uses BadgerDB which is a persistent cache
func GetServerID() (string, error) {
	var serverID string
	cacheExists, err := GetCache(serverIDCacheKey, &serverID)
	if err != nil {
		return "", err
	}
	if !cacheExists {
		// create new server ID as v4 uuid, persist in cache
		serverID = uuid.New().String()
		_, err = SetCache(serverIDCacheKey, serverID, -1)
		if err != nil {
			return "", err
		}
	}
	return serverID, nil
}
