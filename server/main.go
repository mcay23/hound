package main

import (
	"hound/config"
	"hound/controllers"
	"hound/model"
	"hound/model/database"
	"hound/model/sources"
	"log/slog"
	"os"
)

func main() {
	// initialize logging
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))

	config.InitializeConfig()
	database.InstantiateDB()
	model.InitializeCache()
	sources.InitializeSources()
	controllers.SetupRoutes()
}
