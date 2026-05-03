package providers

import (
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/mcay23/hound/database"
)

const (
	serviceFailureCacheKey = "service|%s|fail"
	serviceBlockCacheKey   = "service|%s|block"
	serviceFailThreshold   = 3
	serviceFailWindow      = 5 * time.Minute  // note that each failure resets the window to this value, but a success clears the failures
	serviceBlockDuration   = 10 * time.Minute // how long to block the service for after threshold is met
)

/*
Health check service to determine if external providers are responsive,
especially important for subtitles since in the current implementation,
failing to fetch subtitles blocks playback on Android Exoplayer and MPV
until all subtitles have been fetched or timeout.
*/
func IncrementServiceFailure(rawURL string) error {
	provider, err := getBaseURL(rawURL)
	if err != nil {
		slog.Error("Failed to get base url", "error", err, "url", rawURL)
		return err
	}
	// TODO: fix technically a race condition here, but low severity
	count := 0
	_, err = database.GetCache(fmt.Sprintf(serviceFailureCacheKey, provider), &count)
	if err != nil {
		slog.Error("Failed to get service failure cache", "error", err)
		return err
	}
	count++
	_, err = database.SetCache(fmt.Sprintf(serviceFailureCacheKey, provider), count, serviceFailWindow)
	if err != nil {
		slog.Error("Failed to get set failure cache", "error", err)
		return err
	}
	slog.Info("Incremented service failures", "provider", provider, "count", count)
	// if we have enough failures, block the service
	if count >= serviceFailThreshold {
		_, err = database.SetCache(fmt.Sprintf(serviceBlockCacheKey, provider), "blocked", serviceBlockDuration)
		if err != nil {
			return err
		}
		slog.Info("Blocked provider due to multiple failures", "provider", provider)
	}
	return nil
}

// call whenever there is a successful response
func ClearServiceFailures(rawURL string) error {
	provider, err := getBaseURL(rawURL)
	if err != nil {
		return err
	}
	err = database.DeleteCache(fmt.Sprintf(serviceFailureCacheKey, provider))
	if err != nil {
		slog.Error("Failed to delete service failure cache", "error", err)
	}
	err = database.DeleteCache(fmt.Sprintf(serviceBlockCacheKey, provider))
	if err != nil {
		slog.Error("Failed to delete service block cache", "error", err)
	}
	slog.Info("Cleared service failures", "provider", provider)
	return nil
}

func IsServiceBlocked(rawURL string) (bool, error) {
	provider, err := getBaseURL(rawURL)
	if err != nil {
		return false, err
	}
	var block string
	blocked, err := database.GetCache(fmt.Sprintf(serviceBlockCacheKey, provider), &block)
	if err != nil {
		return false, err
	}
	slog.Info("service access currently blocked due to multiple failures", "provider", provider, "blocked", blocked)
	return blocked, nil
}

func getBaseURL(raw string) (string, error) {
	if !strings.Contains(raw, "://") {
		return "", fmt.Errorf("failed to parse url: %s, no scheme provided", raw)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %s, error: %w", raw, err)
	}
	if u.Host == "" {
		return "", fmt.Errorf("invalid url: %s", raw)
	}
	return u.Scheme + "://" + u.Host, nil
}
