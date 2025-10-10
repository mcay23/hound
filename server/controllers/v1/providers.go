package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"hound/helpers"
	"hound/model/database"
	"hound/model/sources"
	"hound/providers"
	"strconv"
	"strings"
)

func SearchProvidersHandler(c *gin.Context) {
	_, sourceID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "request id param invalid" + err.Error()))
		return
	}
	mediaType := ""
	imdbID := ""
	path := c.FullPath() // gives the registered route path like "/api/tv/:id"
	// TODO Cache might be good here
	if strings.HasPrefix(path, "/api/v1/tv") {
		mediaType = database.MediaTypeTVShow
		imdbID, err = sources.GetTVShowIMDBID(sourceID, nil)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Error retrieving TMDB tv" + err.Error()))
			return
		}
		// cannot find IMDB id
		if imdbID == "" {
			res := map[string]interface{}{
				"results": []interface{}{}, // empty array
				"media_type":   mediaType,
				"message": "No results found",
			}
			helpers.SuccessResponse(c, gin.H{"status": "success", "data": res}, 200)
			return
		}
	} else if strings.HasPrefix(path, "/api/v1/movie") {
		mediaType = database.MediaTypeMovie
		movie, err := sources.GetMovieFromIDTMDB(sourceID, nil)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Error retrieving TMDB movie" + err.Error()))
			return
		}
		imdbID = movie.IMDbID
	}
	query := providers.ProviderQueryObject{
		IMDbID:    		imdbID,
		MediaType: 		mediaType,
		MediaSource: 	sources.SourceTMDB,
		SourceID: 		sourceID,
		Query:     		"",
		Season:     	0,
		Episode:    	0,
	}
	if mediaType == database.MediaTypeTVShow {
		season, err := strconv.Atoi(c.Query("season"))
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Invalid season query param" + err.Error()))
		}
		episode, err := strconv.Atoi(c.Query("episode"))
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Invalid episode query param" + err.Error()))
		}
		query.Season = season
		query.Episode = episode
	}
	res, err := providers.SearchProviders(query)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to search providers")
		helpers.ErrorResponse(c, errors.New(helpers.InternalServerError))
		return
	}
	helpers.SuccessResponse(c, gin.H{"status": "success", "data": res}, 200)
}
