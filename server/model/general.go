package model

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
)

type RemoteVersionInfo struct {
	LatestServerVersion string `json:"latest_server_version"`
}

var (
	githubVersionCheckURL = "https://raw.githubusercontent.com/Hound-Media-Server/hound/refs/heads/main/version-check.json"
)

func FetchRemoteVersionInfo() (*RemoteVersionInfo, error) {
	// fetch from github
	// then compare versions
	cacheKey := "remote_version_info"
	var remoteVersionInfo RemoteVersionInfo
	cacheExists, _ := database.GetCache(cacheKey, &remoteVersionInfo)
	if cacheExists {
		return &remoteVersionInfo, nil
	}
	// cache miss
	client := &http.Client{
		Timeout: time.Second * 30,
	}
	resp, err := client.Get(githubVersionCheckURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote version info: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get remote version info status %d: %w", resp.StatusCode, internal.GatewayTimeoutError)
	}
	if err := json.NewDecoder(resp.Body).Decode(&remoteVersionInfo); err != nil {
		return nil, fmt.Errorf("failed to decode remote version info: %w", err)
	}
	_, _ = database.SetCache(cacheKey, remoteVersionInfo, 30*time.Minute)
	return &remoteVersionInfo, nil
}
