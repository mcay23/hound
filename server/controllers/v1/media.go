package v1

import (
	"github.com/gin-gonic/gin"
	"hound/helpers"
	"hound/model/sources"
	"hound/view"
)

func GeneralSearchHandler(c *gin.Context) {
	queryString := c.Query("q")
	// search tmdb
	tvResults, _ := SearchTVShowCore(queryString)
	movieResults, _ := SearchMoviesCore(queryString)
	// search igdb
	gameResults, _ := sources.SearchGameIGDB(queryString)

	helpers.SuccessResponse(c, view.GeneralSearchResponse{
		TVShowSearchResults: tvResults,
		MovieSearchResults:  movieResults,
		GameSearchResults:   &gameResults,
	}, 200)
}