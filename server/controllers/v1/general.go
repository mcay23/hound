package v1

import (
	"fmt"
	"time"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/model"
	"github.com/mcay23/hound/sources"
	"github.com/mcay23/hound/view"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/gin-gonic/gin"
)

var (
	backdropCacheKey = "server-backdrop-cache"
)

// @Router /api/v1/search [get]
// @Summary General Media Search
// @ID search-media
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

	internal.SuccessResponse(c, view.GeneralSearchResponse{
		TVShowSearchResults: tvResults,
		MovieSearchResults:  movieResults,
	}, 200)
}

// @Router /api/v1/backdrops [get]
// @Summary Get Media Backdrops
// @ID get-media-backdrops
// @Tags General
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
		internal.SuccessResponse(c, backdropCache, 200)
		return
	}
	shows, err := sources.GetTrendingTVShowsTMDB("1")
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get trending tv shows: %w", err))
		return
	}
	movies, err := sources.GetTrendingMoviesTMDB("1")
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get trending movies: %w", err))
		return
	}
	candidateURL := ""
	if shows != nil && movies != nil {
		concat := append(shows.Results, movies.Results...)
		var popularity float32 = 0
		for _, item := range concat {
			if item.Popularity > popularity {
				popularity = item.Popularity
				candidateURL = internal.GetTMDBImageURL(item.BackdropPath, tmdb.Original)
			}
		}
	}
	if candidateURL == "" {
		internal.ErrorResponse(c, fmt.Errorf("failed to get backdrop: %w", internal.InternalServerError))
		return
	}
	_, _ = database.SetCache(backdropCacheKey, candidateURL, time.Hour*24)
	internal.SuccessResponse(c, candidateURL, 200)
}

type ServerInfoResponse struct {
	ServerID string `json:"server_id"`
	internal.BuildInfo
}

// @Router /api/v1/server_info [get]
// @Summary Get Server Info
// @ID get-server-info
// @Tags General
// @Accept json
// @Produce json
// @Success 200 {object} V1SuccessResponse{data=internal.BuildInfo}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetServerInfoHandler(c *gin.Context) {
	serverID, err := database.GetServerID()
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get server ID: %w", err))
		return
	}
	response := ServerInfoResponse{
		ServerID:  serverID,
		BuildInfo: internal.GetBuildInfo(),
	}
	internal.SuccessResponse(c, response, 200)
}
