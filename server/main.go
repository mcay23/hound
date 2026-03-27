package main

import (
	"github.com/mcay23/hound/controllers"
	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/loggers"
	"github.com/mcay23/hound/model"
	"github.com/mcay23/hound/services"
	"github.com/mcay23/hound/sources"
	"github.com/mcay23/hound/workers"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// @title Hound API Documentation V1
// @version beta-0.1.0
// @description V1 Hound Media Server API Documentation
// @description While in beta, please expect breaking api changes in short/no notice
// @host localhost:2323
// @BasePath /
func main() {
	// initialize logging
	time.Local, _ = time.LoadLocation("UTC")

	// load env file to os for dev
	if os.Getenv("APP_ENV") != "production" {
		_ = godotenv.Load("dev.env")
	}
	logLevel := slog.LevelInfo
	if strings.ToLower(os.Getenv("DEBUG_LOGGING")) == "true" {
		logLevel = slog.LevelDebug
	}
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	slog.SetDefault(slog.New(handler))

	model.InitializeConfig()
	loggers.InitializeLoggers()
	database.InitializeCache()
	database.InstantiateDB()
	sources.InitializeSources()
	model.InitializeP2P()
	model.InitializeMedia()
	services.InitializeFFMPEG()
	model.InitializeOnboarding()
	// workers should run after db, since some row cleanups are done during startup
	workers.InitializeWorkers()
	controllers.SetupRoutes()
}
