package model

import (
	"fmt"
	"hound/helpers"
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var envFileName = ".env"
var yamlFileName = "config.yaml"
var configFilePath = "."

var (
	MaxConcurrentDownloads int
)

func InitializeConfig() {
	// read yaml config
	viper.AddConfigPath(configFilePath)
	viper.SetConfigType("yaml")
	viper.SetConfigName(yamlFileName)
	err := viper.MergeInConfig()
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to read .yaml config")
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	// load env file to os
	err = godotenv.Load(envFileName)
	if err != nil {
		panic(fmt.Errorf("fatal error loading .env config file: %w", err))
	}
	// hot reload functionality
	viper.WatchConfig()
	viper.SetDefault("max_concurrent_downloads", 3)
	MaxConcurrentDownloads = viper.GetInt("max_concurrent_downloads")
	slog.Info("Config Initialized", "MaxConcurrentDownloads", MaxConcurrentDownloads)
}
