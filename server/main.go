package main

import (
	"hound/controllers"
	"hound/database"
	"hound/model"
	"hound/model/sources"
	"hound/services"
	"log/slog"
	"os"
	"time"
)

func main() {
	// initialize logging
	time.Local, _ = time.LoadLocation("UTC")
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))

	model.InitializeConfig()
	database.InstantiateDB()
	database.InitializeCache()
	sources.InitializeSources()
	model.InitializeP2P()
	model.InitializeMedia()
	services.InitializeFFMPEG()
	controllers.SetupRoutes()
}
