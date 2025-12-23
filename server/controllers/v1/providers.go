package v1

import (
	"errors"
	"hound/database"
	"hound/helpers"
	"hound/model"
	"hound/model/sources"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func DecodeTestHandler(c *gin.Context) {
	str := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjoie1wibWVkaWFfc291cmNlXCI6XCJ0bWRiXCIsXCJzb3VyY2VfaWRcIjozNzIwNTgsXCJtZWRpYV90eXBlXCI6XCJtb3ZpZVwiLFwiaW1kYl9pZFwiOlwidHQ1MzExNTE0XCIsXCJzZWFzb25cIjowLFwiZXBpc29kZVwiOjAsXCJhZGRvblwiOlwiVG9ycmVudGlvXCIsXCJjYWNoZWRcIjpcInRydWVcIixcInNlcnZpY2VcIjpcIlJEXCIsXCJwMnBcIjpcImRlYnJpZFwiLFwiaW5mb2hhc2hcIjpcIjcxZmVlMjkzZGMxMTdjNDg0ODcwMjljNmRjYjUwMzhkOTc0YTAyOTVcIixcImluZGV4ZXJcIjpcIlRvcnJlbnRHYWxheHlcIixcImZpbGVfbmFtZVwiOlwiWW91ci5OYW1lLjIwMTYuSkFQQU5FU0UuMTA4MHAuQmx1UmF5LkgyNjQuQUFDLVZYVC5tcDRcIixcImZvbGRlcl9uYW1lXCI6XCJJTURCIFRvcCAyNTAgLSAyMDI0IEVkaXRpb24gLSAxMDgwcCBCbHVSYXkgZVN1YnMgalpRXCIsXCJyZXNvbHV0aW9uXCI6XCIxMDgwcFwiLFwiZmlsZV9pZHhcIjotMSxcImZpbGVfc2l6ZVwiOjIxNzk2OTU5MDMsXCJyYW5rXCI6MTExNTAsXCJzZWVkZXJzXCI6NTI4LFwibGVlY2hlcnNcIjotMSxcInVybFwiOlwiaHR0cHM6Ly90b3JyZW50aW8uc3RyZW0uZnVuL3Jlc29sdmUvcmVhbGRlYnJpZC80RkhDTlBJVEhNQ1VDUkVHRDNETkNMNDVNNUpPV1RHQ0pMVkJGR1JFNEVBNEtYM1hNVVRRLzcxZmVlMjkzZGMxMTdjNDg0ODcwMjljNmRjYjUwMzhkOTc0YTAyOTUvbnVsbC82NTkvWW91ci5OYW1lLjIwMTYuSkFQQU5FU0UuMTA4MHAuQmx1UmF5LkgyNjQuQUFDLVZYVC5tcDRcIixcImVuY29kZWRfZGF0YVwiOlwiXCIsXCJkYXRhXCI6e1wiY29kZWNcIjpcImF2Y1wiLFwiYXVkaW9cIjpbXCJBQUNcIl0sXCJjaGFubmVsc1wiOltdLFwiY29udGFpbmVyXCI6XCJtcDRcIixcImxhbmd1YWdlc1wiOltcImphXCJdLFwiYml0X2RlcHRoXCI6XCJcIixcImhkclwiOltdfX0ifQ.RqCPlPNTk2BRPto2vqPHvI8nHgItOW4kNR-lKfRyXg0"
	obj, err := model.DecodeJsonStreamJWT(str)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "request id param invalid"+err.Error()))
		return
	}
	helpers.SuccessResponse(c, gin.H{"status": "success", "data": obj}, 200)
}

func ClearCacheHandler(c *gin.Context) {
	database.ClearCache()
	helpers.SuccessResponse(c, gin.H{"status": "success"}, 200)
}

func SearchProvidersHandler(c *gin.Context) {
	_, sourceID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "request id param invalid"+err.Error()))
		return
	}
	mediaType := ""
	imdbID := ""
	path := c.FullPath() // gives the registered route path like "/api/tv/:id"
	if strings.HasPrefix(path, "/api/v1/tv") {
		mediaType = database.MediaTypeTVShow
		imdbID, err = sources.GetTVShowIMDBID(sourceID)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Error retrieving TMDB imdb id"+err.Error()))
			return
		}
		// cannot find IMDB id
		// TODO other providers may allow searching for query, but for now through aiostreams, only imdb id search
		if imdbID == "" {
			res := map[string]interface{}{
				"results":    []interface{}{}, // empty array
				"media_type": mediaType,
				"message":    "No results found",
			}
			helpers.SuccessResponse(c, gin.H{"status": "success", "data": res}, 200)
			return
		}
	} else if strings.HasPrefix(path, "/api/v1/movie") {
		mediaType = database.MediaTypeMovie
		movie, err := sources.GetMovieFromIDTMDB(sourceID)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Error retrieving TMDB movie"+err.Error()))
			return
		}
		imdbID = movie.IMDbID
	}
	query := model.ProviderQueryObject{
		IMDbID:          imdbID,
		MediaType:       mediaType,
		MediaSource:     sources.MediaSourceTMDB,
		SourceID:        sourceID,
		Query:           "",
		Season:          0,
		Episode:         0,
		SourceEpisodeID: 0,
		EpisodeGroupID:  "",
	}
	if mediaType == database.MediaTypeTVShow {
		seasonNumber, err := strconv.Atoi(c.Query("season"))
		if err != nil || c.Query("season") == "" {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid season query param"+err.Error()))
		}
		episodeNumber, err := strconv.Atoi(c.Query("episode"))
		if err != nil || c.Query("episode") == "" {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid episode query param"+err.Error()))
		}
		query.EpisodeGroupID = c.Query("episode_group_id")
		// EVALUATE, THIS IS UNDEFINED BEHAVIOR
		// if query.EpisodeGroupID == "" {
		// 	query.EpisodeGroupID = "tvdb"
		// }
		query.Season = seasonNumber
		// For TV Shows, episodes are sometimes offset, eg. for show A, Season 2 starts at episode 20 instead of 1
		// Offset this negatively to normalize to S2E1 since this is how most providers work
		query.Episode = episodeNumber
		seasonData, err := sources.GetTVSeasonTMDB(query.SourceID, query.Season)
		if err != nil || len(seasonData.Episodes) < 1 {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Error retrieving TMDB season"+err.Error()))
			return
		}
		// no errors, continue set episode id
		for _, episode := range seasonData.Episodes {
			if episode.EpisodeNumber == query.Episode {
				query.SourceEpisodeID = int(episode.ID)
				break
			}
		}
		if query.SourceEpisodeID == 0 {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Episode ID not found for this query"))
			return
		}
		// normalize episode so every season starts from ep 1.
		query.Episode = episodeNumber - seasonData.Episodes[0].EpisodeNumber + 1
	}
	res, err := model.SearchProviders(query)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to search providers")
		helpers.ErrorResponse(c, errors.New(helpers.InternalServerError))
		return
	}
	helpers.SuccessResponse(c, gin.H{"status": "success", "data": res}, 200)
}
