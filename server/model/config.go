package model

import (
	"fmt"
	"hound/helpers"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var configFileName = "config.yaml"

var (
	MaxConcurrentDownloads int
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
	// load env file to os for dev
	if os.Getenv("APP_ENV") != "production" {
		slog.Info("Loading dev.env")
		_ = godotenv.Load("dev.env")
	}
	// hot reload functionality
	viper.WatchConfig()
	viper.SetDefault("max_concurrent_downloads", 3)
	MaxConcurrentDownloads = viper.GetInt("max_concurrent_downloads")
	slog.Info("Config Initialized", "MaxConcurrentDownloads", MaxConcurrentDownloads)
}
