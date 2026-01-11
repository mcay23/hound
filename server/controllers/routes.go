package controllers

import (
	v1 "hound/controllers/v1"
	"os"

	"github.com/gin-gonic/gin"
)

func SetupRoutes() {
	r := gin.Default()
	v1.SetupRoutes(r)

	// Serve static files from the build directory in prod
	if os.Getenv("APP_ENV") == "production" {
		r.Static("/static", "./build/static")
		r.StaticFile("/manifest.json", "./build/manifest.json")
		r.StaticFile("/favicon.ico", "./build/favicon.ico")
		r.StaticFile("/logo192.png", "./build/logo192.png")
		r.StaticFile("/logo512.png", "./build/logo512.png")
		r.StaticFile("/asset-manifest.json", "./build/asset-manifest.json")
		r.StaticFile("/robots.txt", "./build/robots.txt")
		r.NoRoute(func(c *gin.Context) {
			c.File("./build/index.html")
		})
	}

	err := r.Run(":" + os.Getenv("SERVER_PORT"))
	if err != nil {
		panic("Error parsing SERVER_PORT .env variable")
	}
}
