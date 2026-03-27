package v1

import (
	"fmt"
	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/helpers"
	"github.com/mcay23/hound/providers"
	"github.com/mcay23/hound/sources"
	"strconv"

	"github.com/gin-gonic/gin"
)

func DecodeTestHandler(c *gin.Context) {
	str := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjoie1wibWVkaWFfc291cmNlXCI6XCJ0bWRiXCIsXCJzb3VyY2VfaWRcIjozNzIwNTgsXCJtZWRpYV90eXBlXCI6XCJtb3ZpZVwiLFwiaW1kYl9pZFwiOlwidHQ1MzExNTE0XCIsXCJzZWFzb25cIjowLFwiZXBpc29kZVwiOjAsXCJhZGRvblwiOlwiVG9ycmVudGlvXCIsXCJjYWNoZWRcIjpcInRydWVcIixcInNlcnZpY2VcIjpcIlJEXCIsXCJwMnBcIjpcImRlYnJpZFwiLFwiaW5mb2hhc2hcIjpcIjcxZmVlMjkzZGMxMTdjNDg0ODcwMjljNmRjYjUwMzhkOTc0YTAyOTVcIixcImluZGV4ZXJcIjpcIlRvcnJlbnRHYWxheHlcIixcImZpbGVfbmFtZVwiOlwiWW91ci5OYW1lLjIwMTYuSkFQQU5FU0UuMTA4MHAuQmx1UmF5LkgyNjQuQUFDLVZYVC5tcDRcIixcImZvbGRlcl9uYW1lXCI6XCJJTURCIFRvcCAyNTAgLSAyMDI0IEVkaXRpb24gLSAxMDgwcCBCbHVSYXkgZVN1YnMgalpRXCIsXCJyZXNvbHV0aW9uXCI6XCIxMDgwcFwiLFwiZmlsZV9pZHhcIjotMSxcImZpbGVfc2l6ZVwiOjIxNzk2OTU5MDMsXCJyYW5rXCI6MTExNTAsXCJzZWVkZXJzXCI6NTI4LFwibGVlY2hlcnNcIjotMSxcInVybFwiOlwiaHR0cHM6Ly90b3JyZW50aW8uc3RyZW0uZnVuL3Jlc29sdmUvcmVhbGRlYnJpZC80RkhDTlBJVEhNQ1VDUkVHRDNETkNMNDVNNUpPV1RHQ0pMVkJGR1JFNEVBNEtYM1hNVVRRLzcxZmVlMjkzZGMxMTdjNDg0ODcwMjljNmRjYjUwMzhkOTc0YTAyOTUvbnVsbC82NTkvWW91ci5OYW1lLjIwMTYuSkFQQU5FU0UuMTA4MHAuQmx1UmF5LkgyNjQuQUFDLVZYVC5tcDRcIixcImVuY29kZWRfZGF0YVwiOlwiXCIsXCJkYXRhXCI6e1wiY29kZWNcIjpcImF2Y1wiLFwiYXVkaW9cIjpbXCJBQUNcIl0sXCJjaGFubmVsc1wiOltdLFwiY29udGFpbmVyXCI6XCJtcDRcIixcImxhbmd1YWdlc1wiOltcImphXCJdLFwiYml0X2RlcHRoXCI6XCJcIixcImhkclwiOltdfX0ifQ.RqCPlPNTk2BRPto2vqPHvI8nHgItOW4kNR-lKfRyXg0"
	obj, err := providers.DecodeJsonStreamAES(str)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to decode json stream aes: %w: %w", helpers.BadRequestError, err))
		return
	}
	helpers.SuccessResponse(c, obj, 200)
}

func ClearCacheHandler(c *gin.Context) {
	database.ClearCache()
	helpers.SuccessResponse(c, nil, 200)
}

// @Router /api/v1/tv/{id}/providers [get]
// @Summary Search Stream Providers for TV Shows by ID
// @Tags Providers
// @Accept json
// @Produce json
// @Param id path string true "Media ID" example(tmdb-1234)
// @Param season query int true "Season Number"
// @Param episode query int true "Episode Number"
// @Param provider_profile_id query int false "Provider Profile ID"
// @Param episode_group_id query string false "Episode Group ID"
// @Success 200 {object} V1SuccessResponse{data=providers.ProviderResponseObject}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func SearchProvidersTVHandler(c *gin.Context) {
	_, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to get source id from params: %w: %w", helpers.BadRequestError, err))
		return
	}
	imdbID, err := sources.GetTVShowIMDBID(sourceID)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to get tv show imdb id: %w", err))
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
		helpers.SuccessResponse(c, res, 200)
		return
	}
	seasonNumber, err := strconv.Atoi(c.Query("season"))
	if err != nil || c.Query("season") == "" {
		helpers.ErrorResponse(c, fmt.Errorf("failed to get season query param: %w: %w", helpers.BadRequestError, err))
		return
	}
	episodeNumber, err := strconv.Atoi(c.Query("episode"))
	if err != nil || c.Query("episode") == "" {
		helpers.ErrorResponse(c, fmt.Errorf("failed to get episode query param: %w: %w", helpers.BadRequestError, err))
		return
	}
	episode, err := sources.GetEpisodeTMDB(sourceID, seasonNumber, episodeNumber)
	if err != nil || episode == nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to get episode from tmdb: %w", err))
		return
	}
	// slightly hacky, revise if we add more types
	requestType := c.Query("request_type")
	if requestType != providers.ProviderRequestDownload {
		requestType = providers.ProviderRequestStream
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
		RequestType:     requestType,
	}
	// if not supplied, will use defaults
	providerQuery := c.Query("provider_profile_id")
	if providerQuery != "" {
		temp, err := strconv.Atoi(c.Query("provider_profile_id"))
		if err != nil {
			helpers.ErrorResponse(c, fmt.Errorf("invalid provider profile id query param: %w: %w", helpers.BadRequestError, err))
			return
		}
		query.ProviderProfileID = &temp
	} else {
		query.ProviderProfileID = nil
	}
	results, err := providers.QueryProviders(query)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to query providers: %w", err))
		return
	}
	helpers.SuccessResponse(c, results, 200)
}

// @Router /api/v1/movie/{id}/providers [get]
// @Summary Search Stream Providers for Movies
// @Tags Providers
// @Accept json
// @Produce json
// @Param id path string true "Media ID" example(tmdb-1234)
// @Param provider_profile_id query int false "Provider Profile ID"
// @Success 200 {object} V1SuccessResponse{data=providers.ProviderResponseObject}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func SearchProvidersMovieHandler(c *gin.Context) {
	_, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to get source id from params: %w: %w", helpers.BadRequestError, err))
		return
	}
	movie, err := sources.GetMovieFromIDTMDB(sourceID)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to get movie from tmdb: %w", err))
		return
	}
	// slightly hacky, revise if we add more types
	requestType := c.Query("request_type")
	if requestType != providers.ProviderRequestDownload {
		requestType = providers.ProviderRequestStream
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
		RequestType:     requestType,
	}
	// if not supplied, will use defaults
	providerQuery := c.Query("provider_profile_id")
	if providerQuery != "" {
		temp, err := strconv.Atoi(c.Query("provider_profile_id"))
		if err != nil {
			helpers.ErrorResponse(c, fmt.Errorf("invalid provider profile id query param: %w: %w", helpers.BadRequestError, err))
			return
		}
		query.ProviderProfileID = &temp
	} else {
		query.ProviderProfileID = nil
	}
	results, err := providers.QueryProviders(query)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to query providers: %w", err))
		return
	}
	helpers.SuccessResponse(c, results, 200)
}
