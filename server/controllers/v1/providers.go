package v1

import (
	"errors"
	"hound/database"
	"hound/helpers"
	"hound/model/providers"
	"hound/model/sources"
	"strconv"

	"github.com/gin-gonic/gin"
)

func DecodeTestHandler(c *gin.Context) {
	str := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjoie1wibWVkaWFfc291cmNlXCI6XCJ0bWRiXCIsXCJzb3VyY2VfaWRcIjozNzIwNTgsXCJtZWRpYV90eXBlXCI6XCJtb3ZpZVwiLFwiaW1kYl9pZFwiOlwidHQ1MzExNTE0XCIsXCJzZWFzb25cIjowLFwiZXBpc29kZVwiOjAsXCJhZGRvblwiOlwiVG9ycmVudGlvXCIsXCJjYWNoZWRcIjpcInRydWVcIixcInNlcnZpY2VcIjpcIlJEXCIsXCJwMnBcIjpcImRlYnJpZFwiLFwiaW5mb2hhc2hcIjpcIjcxZmVlMjkzZGMxMTdjNDg0ODcwMjljNmRjYjUwMzhkOTc0YTAyOTVcIixcImluZGV4ZXJcIjpcIlRvcnJlbnRHYWxheHlcIixcImZpbGVfbmFtZVwiOlwiWW91ci5OYW1lLjIwMTYuSkFQQU5FU0UuMTA4MHAuQmx1UmF5LkgyNjQuQUFDLVZYVC5tcDRcIixcImZvbGRlcl9uYW1lXCI6XCJJTURCIFRvcCAyNTAgLSAyMDI0IEVkaXRpb24gLSAxMDgwcCBCbHVSYXkgZVN1YnMgalpRXCIsXCJyZXNvbHV0aW9uXCI6XCIxMDgwcFwiLFwiZmlsZV9pZHhcIjotMSxcImZpbGVfc2l6ZVwiOjIxNzk2OTU5MDMsXCJyYW5rXCI6MTExNTAsXCJzZWVkZXJzXCI6NTI4LFwibGVlY2hlcnNcIjotMSxcInVybFwiOlwiaHR0cHM6Ly90b3JyZW50aW8uc3RyZW0uZnVuL3Jlc29sdmUvcmVhbGRlYnJpZC80RkhDTlBJVEhNQ1VDUkVHRDNETkNMNDVNNUpPV1RHQ0pMVkJGR1JFNEVBNEtYM1hNVVRRLzcxZmVlMjkzZGMxMTdjNDg0ODcwMjljNmRjYjUwMzhkOTc0YTAyOTUvbnVsbC82NTkvWW91ci5OYW1lLjIwMTYuSkFQQU5FU0UuMTA4MHAuQmx1UmF5LkgyNjQuQUFDLVZYVC5tcDRcIixcImVuY29kZWRfZGF0YVwiOlwiXCIsXCJkYXRhXCI6e1wiY29kZWNcIjpcImF2Y1wiLFwiYXVkaW9cIjpbXCJBQUNcIl0sXCJjaGFubmVsc1wiOltdLFwiY29udGFpbmVyXCI6XCJtcDRcIixcImxhbmd1YWdlc1wiOltcImphXCJdLFwiYml0X2RlcHRoXCI6XCJcIixcImhkclwiOltdfX0ifQ.RqCPlPNTk2BRPto2vqPHvI8nHgItOW4kNR-lKfRyXg0"
	obj, err := providers.DecodeJsonStreamAES(str)
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

func SearchProvidersTVShowsHandler(c *gin.Context) {
	_, sourceID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "request id param invalid"+err.Error()))
		return
	}
	imdbID, err := sources.GetTVShowIMDBID(sourceID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Error retrieving TMDB imdb id"+err.Error()))
		return
	}
	// cannot find IMDB id
	// TODO other providers may allow searching for query, but for now through aiostreams, only imdb id search
	if imdbID == "" {
		res := map[string]interface{}{
			"results":    []interface{}{}, // empty array
			"media_type": database.MediaTypeTVShow,
			"message":    "No results found",
		}
		helpers.SuccessResponse(c, gin.H{"status": "success", "data": res}, 200)
		return
	}
	seasonNumber, err := strconv.Atoi(c.Query("season"))
	if err != nil || c.Query("season") == "" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Invalid season query param"+err.Error()))
	}
	episodeNumber, err := strconv.Atoi(c.Query("episode"))
	if err != nil || c.Query("episode") == "" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Invalid episode query param"+err.Error()))
	}
	episode, err := sources.GetEpisodeTMDB(sourceID, seasonNumber, episodeNumber)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError),
			"Error retrieving TMDB episode"+err.Error()))
		return
	}
	sourceEpisodeIDstr := strconv.Itoa(int(episode.ID))
	query := providers.ProvidersQueryRequest{
		IMDbID:          imdbID,
		MediaType:       database.MediaTypeTVShow,
		MediaSource:     sources.MediaSourceTMDB,
		SourceID:        strconv.Itoa(sourceID),
		SeasonNumber:    &seasonNumber,
		EpisodeNumber:   &episodeNumber,
		EpisodeSourceID: &sourceEpisodeIDstr,
		EpisodeGroupID:  c.Query("episode_group_id"),
	}
	results, err := providers.QueryProviders(query)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError),
			"Error retrieving Stremio streams"+err.Error()))
		return
	}
	helpers.SuccessResponse(c, gin.H{"status": "success", "data": results}, 200)
}

func SearchProvidersMovieHandler(c *gin.Context) {
	_, sourceID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"request id param invalid"+err.Error()))
		return
	}
	movie, err := sources.GetMovieFromIDTMDB(sourceID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError),
			"Error retrieving TMDB movie"+err.Error()))
		return
	}
	query := providers.ProvidersQueryRequest{
		IMDbID:          movie.IMDbID,
		MediaType:       database.MediaTypeMovie,
		MediaSource:     sources.MediaSourceTMDB,
		SourceID:        strconv.Itoa(sourceID),
		SeasonNumber:    nil,
		EpisodeNumber:   nil,
		EpisodeSourceID: nil,
		EpisodeGroupID:  "",
	}
	res, err := providers.QueryProviders(query)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to search providers")
		helpers.ErrorResponse(c, errors.New(helpers.InternalServerError))
		return
	}
	helpers.SuccessResponse(c, gin.H{"status": "success", "data": res}, 200)
}
