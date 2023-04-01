package main

import (
	"hound/config"
	"hound/controllers"
	"hound/model/database"
	"hound/model/sources"
)

func main() {
	config.InitializeConfig()
	database.InstantiateDB()
	sources.InitializeSources()
	controllers.SetupRoutes()
}
