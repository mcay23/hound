package controllers

import (
	"fmt"
	"github.com/Open-pi/gol"
	"github.com/gin-gonic/gin"
	v1 "hound/controllers/v1"
	"os"
)

func SetupRoutes() {
	r := gin.Default()
	v1.SetupRoutes(r)
	err := r.Run(os.Getenv("PORT"))
	if err != nil {
		panic(err)
	}
	// Construct the SearchUrl
	url := gol.SearchUrl().All("the selfish gene").Author("Richard Dawkins").Construct()
	// search
	search, err := gol.Search(url)
	fmt.Println(search.)
}
