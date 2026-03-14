package main

import (
	"hound/controllers"
	"hound/database"
	"hound/loggers"
	"hound/model"
	"hound/services"
	"hound/sources"
	"hound/workers"
	"log/slog"
	"os"
	"time"
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
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
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
