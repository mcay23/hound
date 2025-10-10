package controllers

import (
	"github.com/gin-gonic/gin"
	v1 "hound/controllers/v1"
	"os"
)

func SetupRoutes() {
	r := gin.Default()
	v1.SetupRoutes(r)
	err := r.Run(":" + os.Getenv("SERVER_PORT"))
	if err != nil {
		panic("Error parsing SERVER_PORT .env variable")
	}
}
