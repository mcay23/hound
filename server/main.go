package main

import (
	"hound/config"
	"hound/controllers"
	"hound/model"
	"hound/model/database"
	"hound/model/sources"
)

func main() {
	config.InitializeConfig()
	database.InstantiateDB()
	model.InitializeCache()
	sources.InitializeSources()
	controllers.SetupRoutes()
}
