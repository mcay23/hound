package model

import (
	"log/slog"
	"os"
	"strconv"
)

var (
	MaxConcurrentDownloads        int    = 3
	MaxConcurrentIngests          int    = 3
	ExternalLibraryWorkersEnabled bool   = true
	ExternalLibraryMoviesPath     string = "/app/External Library/Movies"
	ExternalLibraryTVPath         string = "/app/External Library/TV Shows"
	ExternalLibraryScanInterval   int    = 360
	MaxExternalLibraryWorkers     int    = 2
)

func InitializeConfig() {
	MaxConcurrentDownloads = getEnvInt("MAX_DOWNLOAD_WORKERS", MaxConcurrentDownloads)
	MaxConcurrentIngests = getEnvInt("MAX_INGEST_WORKERS", MaxConcurrentIngests)
	MaxExternalLibraryWorkers = getEnvInt("MAX_EXTERNAL_LIBRARY_WORKERS", MaxExternalLibraryWorkers)
	ExternalLibraryWorkersEnabled = getEnvBool("ENABLE_EXTERNAL_LIBRARY", ExternalLibraryWorkersEnabled)
	ExternalLibraryMoviesPath = getEnvString("EXTERNAL_LIBRARY_MOVIES_PATH", ExternalLibraryMoviesPath)
	ExternalLibraryTVPath = getEnvString("EXTERNAL_LIBRARY_SHOWS_PATH", ExternalLibraryTVPath)
	ExternalLibraryScanInterval = getEnvInt("EXTERNAL_LIBRARY_SCAN_INTERVAL_MINUTES", ExternalLibraryScanInterval)
	slog.Info("Config Initialized",
		"MaxConcurrentDownloads", MaxConcurrentDownloads,
		"MaxConcurrentIngests", MaxConcurrentIngests,
		"MaxExternalLibraryWorkers", MaxExternalLibraryWorkers,
		"ExternalLibraryEnabled", ExternalLibraryWorkersEnabled,
		"ExternalLibraryMovies", ExternalLibraryMoviesPath,
		"ExternalLibraryTV", ExternalLibraryTVPath,
		"ExternalScanIntervalMinutes", ExternalLibraryScanInterval,
	)
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return val
}

func getEnvString(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	val, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return val
}
