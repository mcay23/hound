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
		r.StaticFile("/favicon-16x16.png", "./build/favicon-16x16.png")
		r.StaticFile("/favicon-32x32.png", "./build/favicon-32x32.png")
		r.StaticFile("/apple-touch-icon.png", "./build/apple-touch-icon.png")
		r.StaticFile("/android-chrome-192x192.png", "./build/android-chrome-192x192.png")
		r.StaticFile("/android-chrome-512x512.png", "./build/android-chrome-512x512.png")
		r.StaticFile("/avatar-placeholder.png", "./build/avatar-placeholder.png")
		r.StaticFile("/landscape-placeholder.jpg", "./build/landscape-placeholder.jpg")
		r.StaticFile("/login-bg.jpg", "./build/login-bg.jpg")
		// r.StaticFile("/asset-manifest.json", "./build/asset-manifest.json")
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
