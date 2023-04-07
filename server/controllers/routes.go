package controllers

import (
	"github.com/gin-gonic/gin"
	v1 "hound/controllers/v1"
	"os"
)

func SetupRoutes() {
	r := gin.Default()
	v1.SetupRoutes(r)
	err := r.Run(":" + os.Getenv("PORT"))
	if err != nil {
		panic(err)
	}
}
