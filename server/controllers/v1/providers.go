package v1

import (
	"fmt"
	"strconv"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/providers"
	"github.com/mcay23/hound/sources"

	"github.com/gin-gonic/gin"
)

type DecodeStreamRequest struct {
	EncodedData string `json:"encoded_data" binding:"required"`
}

// @Router /api/v1/decode [get]
// @Summary Decode Stream AES from Encoded Data
// @Description Use this to decode the stream object if you have the encoded data. Returns the same output as the providers endpoint for a stream
// @ID decode-stream
// @Tags Providers
// @Accept json
// @Produce json
// @Param request body DecodeStreamRequest true "Encoded Data"
// @Success 200 {object} V1SuccessResponse{data=providers.StreamObjectFull}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func DecodeStreamHandler(c *gin.Context) {
	var req DecodeStreamRequest
	if err := c.BindJSON(&req); err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to bind json: %w: %w", internal.BadRequestError, err))
		return
	}
	if req.EncodedData == "" {
		internal.ErrorResponse(c, fmt.Errorf("missing encoded_data: %w", internal.BadRequestError))
		return
	}
	obj, err := providers.DecodeJsonStreamAES(req.EncodedData)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to decode json stream aes: %w: %w", internal.BadRequestError, err))
		return
	}
	obj.URI = "<hidden>"
	internal.SuccessResponse(c, obj, 200)
}

func ClearCacheHandler(c *gin.Context) {
	database.ClearCache()
	internal.SuccessResponse(c, nil, 200)
}

// @Router /api/v1/tv/{id}/providers [get]
// @Summary Search Stream Providers for TV Show by ID
// @ID search-providers-tvshow
// @Tags Providers
// @Accept json
// @Produce json
// @Param id path string true "Media ID" example(tmdb-1234)
// @Param season query int true "Season Number"
// @Param episode query int true "Episode Number"
// @Param request_type query string false "request_stream or request_download"
// @Param provider_profile_id query int false "Provider Profile ID"
// @Param episode_group_id query string false "Episode Group ID"
// @Success 200 {object} V1SuccessResponse{data=providers.ProviderStreamsResponseObject}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func SearchProvidersTVHandler(c *gin.Context) {
	query, err := getProvidersQueryTV(c)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get providers query: %w", err))
		return
	}
	if query == nil {
		res := map[string]interface{}{
			"results":    []interface{}{}, // empty array
			"media_type": database.MediaTypeTVShow,
			"message":    "No results found",
		}
		internal.SuccessResponse(c, res, 200)
		return
	}
	results, err := providers.QueryProvidersStreams(*query)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to query providers: %w", err))
		return
	}
	internal.SuccessResponse(c, results, 200)
}

// @Router /api/v1/movie/{id}/providers [get]
// @Summary Search Stream Providers for Movie by ID
// @ID search-providers-movie
// @Tags Providers
// @Accept json
// @Produce json
// @Param id path string true "Media ID" example(tmdb-1234)
// @Param request_type query string false "request_stream or request_download"
// @Param provider_profile_id query int false "Provider Profile ID"
// @Success 200 {object} V1SuccessResponse{data=providers.ProviderStreamsResponseObject}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func SearchProvidersMovieHandler(c *gin.Context) {
	query, err := getProvidersQueryMovie(c)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get providers query: %w", err))
		return
	}
	if query == nil {
		res := map[string]interface{}{
			"results":    []interface{}{}, // empty array
			"media_type": database.MediaTypeMovie,
			"message":    "No results found",
		}
		internal.SuccessResponse(c, res, 200)
		return
	}
	results, err := providers.QueryProvidersStreams(*query)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to query providers: %w", err))
		return
	}
	internal.SuccessResponse(c, results, 200)
}

// @Router /api/v1/tv/{id}/subtitles [get]
// @Summary Search Subtitles for TV Show by ID
// @ID search-subtitles-tvshow
// @Tags Providers
// @Accept json
// @Produce json
// @Param id path string true "Media ID" example(tmdb-1234)
// @Param season query int true "Season Number"
// @Param episode query int true "Episode Number"
// @Param provider_profile_id query int false "Provider Profile ID"
// @Param episode_group_id query string false "Episode Group ID"
// @Success 200 {object} V1SuccessResponse{data=providers.ProviderSubtitlesResponseObject}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func SearchSubtitlesTVHandler(c *gin.Context) {
	query, err := getProvidersQueryTV(c)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get providers query: %w", err))
		return
	}
	if query == nil {
		res := map[string]interface{}{
			"results":    []interface{}{}, // empty array
			"media_type": database.MediaTypeTVShow,
			"message":    "No results found",
		}
		internal.SuccessResponse(c, res, 200)
		return
	}
	results, err := providers.QueryProvidersSubtitles(*query)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to query providers: %w", err))
		return
	}
	internal.SuccessResponse(c, results, 200)
}

// @Router /api/v1/movie/{id}/subtitles [get]
// @Summary Search Subtitles for Movies by ID
// @ID search-subtitles-movie
// @Tags Providers
// @Accept json
// @Produce json
// @Param id path string true "Media ID" example(tmdb-1234)
// @Param provider_profile_id query int false "Provider Profile ID"
// @Success 200 {object} V1SuccessResponse{data=providers.ProviderSubtitlesResponseObject}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func SearchSubtitlesMovieHandler(c *gin.Context) {
	query, err := getProvidersQueryMovie(c)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get providers query: %w", err))
		return
	}
	if query == nil {
		res := map[string]interface{}{
			"results":    []interface{}{}, // empty array
			"media_type": database.MediaTypeMovie,
			"message":    "No results found",
		}
		internal.SuccessResponse(c, res, 200)
		return
	}
	results, err := providers.QueryProvidersSubtitles(*query)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to query providers: %w", err))
		return
	}
	internal.SuccessResponse(c, results, 200)
}

func getProvidersQueryTV(c *gin.Context) (*providers.ProvidersQueryRequest, error) {
	_, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil {
		return nil, fmt.Errorf("failed to get source id from params: %w: %w", internal.BadRequestError, err)
	}
	// tmdb has imdb ids in the regular response for movies, but not for tv shows
	imdbID, err := sources.GetTVShowIMDBID(sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tv show imdb id: %w", err)
	}
	// cannot find IMDB id
	// TODO other providers may allow searching for query, but for now through aiostreams, only imdb id search
	if imdbID == "" {
		return nil, nil
	}
	seasonNumber, err := strconv.Atoi(c.Query("season"))
	if err != nil || c.Query("season") == "" {
		return nil, fmt.Errorf("failed to get season query param: %w: %w", internal.BadRequestError, err)
	}
	episodeNumber, err := strconv.Atoi(c.Query("episode"))
	if err != nil || c.Query("episode") == "" {
		return nil, fmt.Errorf("failed to get episode query param: %w: %w", internal.BadRequestError, err)
	}
	episode, err := sources.GetEpisodeTMDB(sourceID, seasonNumber, episodeNumber)
	if err != nil || episode == nil {
		return nil, fmt.Errorf("failed to get episode from tmdb: %w", err)
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
			return nil, fmt.Errorf("invalid provider profile id query param: %w: %w", internal.BadRequestError, err)
		}
		query.ProviderProfileID = &temp
	} else {
		query.ProviderProfileID = nil
	}
	return &query, nil
}

func getProvidersQueryMovie(c *gin.Context) (*providers.ProvidersQueryRequest, error) {
	_, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil {
		return nil, fmt.Errorf("failed to get source id from params: %w: %w", internal.BadRequestError, err)
	}
	movie, err := sources.GetMovieFromIDTMDB(sourceID)
	if err != nil || movie == nil {
		return nil, fmt.Errorf("failed to get movie from tmdb: %w", err)
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
			return nil, fmt.Errorf("invalid provider profile id query param: %w: %w", internal.BadRequestError, err)
		}
		query.ProviderProfileID = &temp
	} else {
		query.ProviderProfileID = nil
	}
	return &query, nil
}
