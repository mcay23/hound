package v1

import (
	"fmt"
	"strconv"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/model"
	"github.com/mcay23/hound/providers"
	"github.com/mcay23/hound/sources"
	"github.com/mcay23/hound/view"

	"github.com/gin-gonic/gin"
)

// @Router /api/v1/movie/{id}/media_files [get]
// @Summary Get Movie Media Files by ID
// @ID get-movie-media-files
// @Tags Media Files
// @Accept json
// @Produce json
// @Param id path int true "Movie ID"
// @Success 200 {object} V1SuccessResponse{data=providers.ProviderStreamsResponseObject}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetMovieMediaFilesHandler(c *gin.Context) {
	_, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get id param: %w: %w", internal.BadRequestError, err))
		return
	}
	streamObjects, err := providers.GetLocalStreamsForMovie(sourceID)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get local streams: %w", err))
		return
	}
	res := &providers.ProviderStreamsResponseObject{
		StreamMediaDetails: providers.StreamMediaDetails{
			MediaType:   database.MediaTypeMovie,
			MediaSource: sources.MediaSourceTMDB,
			SourceID:    strconv.Itoa(sourceID),
		},
		Providers: []*providers.ProviderStreamObject{
			{
				Provider: "Hound",
				Streams:  streamObjects,
			},
		},
	}
	internal.SuccessResponse(c, res, 200)
}

// @Router /api/v1/tv/{id}/media_files [get]
// @Summary Get TV Show Media Files by ID
// @ID get-tvshow-media-files
// @Tags Media Files
// @Accept json
// @Produce json
// @Param id path int true "TV Show ID"
// @Param season query int false "Season Number"
// @Param episode query int false "Episode Number"
// @Success 200 {object} V1SuccessResponse{data=providers.ProviderStreamsResponseObject}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetTVShowMediaFilesHandler(c *gin.Context) {
	_, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get id param: %w: %w", internal.BadRequestError, err))
		return
	}
	var seasonNumber *int
	if c.Query("season") != "" {
		s, err := strconv.Atoi(c.Query("season"))
		if err != nil {
			internal.ErrorResponse(c, fmt.Errorf("failed to get season param: %w: %w", internal.BadRequestError, err))
			return
		}
		seasonNumber = &s
	}
	var episodeNumber *int
	if c.Query("episode") != "" {
		e, err := strconv.Atoi(c.Query("episode"))
		if err != nil {
			internal.ErrorResponse(c, fmt.Errorf("failed to get episode param: %w: %w", internal.BadRequestError, err))
			return
		}
		episodeNumber = &e
	}
	streamObjects, err := providers.GetLocalStreamsForTVShow(sourceID, seasonNumber, episodeNumber)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get local streams: %w", err))
		return
	}
	// in regular flows, should be a cached call
	var epID *string
	if seasonNumber != nil && episodeNumber != nil {
		epDetails, err := sources.GetEpisodeTMDB(sourceID, *seasonNumber, *episodeNumber)
		if err == nil {
			idStr := strconv.Itoa(int(epDetails.ID))
			epID = &idStr
		}
	}
	res := &providers.ProviderStreamsResponseObject{
		StreamMediaDetails: providers.StreamMediaDetails{
			MediaType:       database.MediaTypeTVShow,
			MediaSource:     sources.MediaSourceTMDB,
			SourceID:        strconv.Itoa(sourceID),
			SeasonNumber:    seasonNumber,
			EpisodeNumber:   episodeNumber,
			EpisodeSourceID: epID,
		},
		Providers: []*providers.ProviderStreamObject{
			{
				Provider: "Hound",
				Streams:  streamObjects,
			},
		},
	}
	internal.SuccessResponse(c, res, 200)
}

// @Router /api/v1/media_files [get]
// @Summary Get all media file records
// @ID get-media-files
// @Tags Media Files
// @Accept json
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} V1SuccessResponse{data=view.MediaFilesResponse}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetMediaFilesHandler(c *gin.Context) {
	limit := c.Query("limit")
	offset := c.Query("offset")
	if limit == "" {
		limit = "100"
	}
	if offset == "" {
		offset = "0"
	}
	limitNum, offsetNum, err := getLimitOffset(limit, offset)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	totalRecords, files, err := database.GetMediaFiles(&limitNum, &offsetNum)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get media files: %w", err))
		return
	}
	res := &view.MediaFilesResponse{
		Files:        files,
		TotalRecords: totalRecords,
		Limit:        limitNum,
		Offset:       offsetNum,
	}
	internal.SuccessResponse(c, res, 200)
}

// @Router /api/v1/media_files/{id} [delete]
// @Summary Delete a media file
// @ID delete-media-file
// @Tags Media Files
// @Accept json
// @Produce json
// @Param id path int true "Media File ID"
// @Success 200 {object} V1SuccessResponse{data=object}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func DeleteMediaFileHandler(c *gin.Context) {
	mediaFileID := c.Param("id")
	if mediaFileID == "" {
		internal.ErrorResponse(c, fmt.Errorf("failed to get media file id param: %w", internal.BadRequestError))
		return
	}
	fileID, err := strconv.Atoi(mediaFileID)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get media file id param: %w: %w", internal.BadRequestError, err))
		return
	}
	err = model.DeleteMediaFile(fileID)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to delete media file: %w", err))
		return
	}
	internal.SuccessResponse(c, nil, 200)
}
