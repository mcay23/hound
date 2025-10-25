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

func InitializeConfig() {
	// read env config
	//viper.SetConfigName(envFileName)
	//viper.SetConfigType("env")
	//viper.AddConfigPath(".")
	//err := viper.ReadInConfig()
	//if err != nil {
	//	_ = helpers.LogErrorWithMessage(err, "Failed to read .env config")
	//	panic(fmt.Errorf("fatal error config file: %w", err))
	//}
	// read yaml config
	viper.SetConfigName(yamlFileName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configFilePath)
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
	slog.Info("Config Initialized")
}
