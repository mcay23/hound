package v1

import (
	"github.com/gin-gonic/gin"
	"hound/helpers"
	"hound/view"
)

func GeneralSearchHandler(c *gin.Context) {
	queryString := c.Query("q")
	// search
	tvResults, _ := SearchTVShowCore(queryString)
	movieResults, _ := SearchMoviesCore(queryString)
	helpers.SuccessResponse(c, view.GeneralSearchResponse{
		TVShowSearchResults: tvResults,
		MovieSearchResults:  movieResults,
	}, 200)
}