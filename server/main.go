package main

import (
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

	model.InitializeConfig()
	database.InstantiateDB()
	model.InitializeCache()
	sources.InitializeSources()
	model.InitializeP2P()
	controllers.SetupRoutes()
}
