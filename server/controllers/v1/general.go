package v1

import (
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/model"
	"hound/sources"
	"hound/view"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/gin-gonic/gin"
)

var (
	backdropCacheKey = "server-backdrop-cache"
)

// @Router /api/v1/search [get]
// @Summary General Media Search
// @Tags Search
// @Accept json
// @Produce json
// @Param q query string true "Search Query"
// @Success 200 {object} V1SuccessResponse{data=view.GeneralSearchResponse}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GeneralSearchHandler(c *gin.Context) {
	queryString := c.Query("q")
	// search tmdb
	tvResults, _ := model.SearchTVShows(queryString)
	movieResults, _ := model.SearchMovies(queryString)
	// search igdb
	//gameResults, _ := sources.SearchGameIGDB(queryString)

	helpers.SuccessResponse(c, view.GeneralSearchResponse{
		TVShowSearchResults: tvResults,
		MovieSearchResults:  movieResults,
		GameSearchResults:   nil,
	}, 200)
}

// @Router /api/v1/backdrops [get]
// @Summary Get Media Backdrops
// @Tags Media
// @Accept json
// @Produce json
// @Success 200 {object} V1SuccessResponse{data=string} "URL to backdrop"
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetMediaBackdrops(c *gin.Context) {
	// refresh backdrop every 24 hours, store data in cache
	var backdropCache string
	cacheExists, _ := database.GetCache(backdropCacheKey, &backdropCache)
	if cacheExists {
		helpers.SuccessResponse(c, backdropCache, 200)
		return
	}
	shows, err := sources.GetTrendingTVShowsTMDB("1")
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to get trending tv shows: %w", err))
		return
	}
	movies, err := sources.GetTrendingMoviesTMDB("1")
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to get trending movies: %w", err))
		return
	}
	candidateURL := ""
	if shows != nil && movies != nil {
		concat := append(shows.Results, movies.Results...)
		var popularity float32 = 0
		for _, item := range concat {
			if item.Popularity > popularity {
				popularity = item.Popularity
				candidateURL = helpers.GetTMDBImageURL(item.BackdropPath, tmdb.Original)
			}
		}
	}
	if candidateURL == "" {
		helpers.ErrorResponse(c, fmt.Errorf("failed to get backdrop: %w", helpers.InternalServerError))
		return
	}
	_, _ = database.SetCache(backdropCacheKey, candidateURL, time.Hour*24)
	helpers.SuccessResponse(c, candidateURL, 200)
}
