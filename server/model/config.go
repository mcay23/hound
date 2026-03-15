package model

import (
	"fmt"
	"hound/helpers"
	"log/slog"

	"github.com/spf13/viper"
)

var configFileName = "config.yaml"

var (
	MaxConcurrentDownloads    int
	MaxConcurrentIngests      int
	ExternalLibraryEnabled    bool
	ExternalLibraryMovies     string
	ExternalLibraryTV         string
	ExternalScanInterval      int
	MaxExternalLibraryWorkers int
)

func InitializeConfig() {
	// read yaml config
	viper.AddConfigPath("./config")
	viper.SetConfigType("yaml")
	viper.SetConfigName(configFileName)
	err := viper.MergeInConfig()
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to read .yaml config")
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	// hot reload functionality
	viper.WatchConfig()
	viper.SetDefault("max_download_workers", 3)
	viper.SetDefault("max_ingest_workers", 3)
	viper.SetDefault("external_library.enabled", false)
	viper.SetDefault("external_library.movies_root_path", "")
	viper.SetDefault("external_library.tv_root_path", "")
	viper.SetDefault("external_library.scan_interval_minutes", 360)
	viper.SetDefault("external_library.max_workers", 2)
	MaxConcurrentDownloads = viper.GetInt("max_download_workers")
	MaxConcurrentIngests = viper.GetInt("max_ingest_workers")
	ExternalLibraryEnabled = viper.GetBool("external_library.enabled")
	ExternalLibraryMovies = viper.GetString("external_library.movies_root_path")
	ExternalLibraryTV = viper.GetString("external_library.tv_root_path")
	ExternalScanInterval = viper.GetInt("external_library.scan_interval_minutes")
	MaxExternalLibraryWorkers = viper.GetInt("external_library.max_workers")
	slog.Info("Config Initialized",
		"MaxConcurrentDownloads", MaxConcurrentDownloads,
		"ExternalLibraryEnabled", ExternalLibraryEnabled,
		"ExternalLibraryMovies", ExternalLibraryMovies,
		"ExternalLibraryTV", ExternalLibraryTV,
		"ExternalScanIntervalMinutes", ExternalScanInterval,
		"ExternalLibraryWorkers", MaxExternalLibraryWorkers,
	)
}
