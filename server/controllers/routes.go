package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mcay23/hound/config"
	v1 "github.com/mcay23/hound/controllers/v1"

	_ "github.com/mcay23/hound/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes() {
	if !config.DebugLogging {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	v1.SetupRoutes(r)
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Serve static files from the build directory in prod
	if config.AppEnvironment == "production" {
		r.Static("/static", "./build/static")

		r.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api") {
				c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
				return
			}
			path := filepath.Join("./build", c.Request.URL.Path)
			if _, err := os.Stat(path); err == nil {
				c.File(path)
				return
			}
			c.File(filepath.Join("./build", "index.html"))
		})
	}

	err := r.Run(":" + os.Getenv("SERVER_PORT"))
	if err != nil {
		panic("Error parsing SERVER_PORT .env variable")
	}
}
