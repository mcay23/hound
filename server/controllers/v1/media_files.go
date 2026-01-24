package v1

import (
	"errors"
	"hound/database"
	"hound/helpers"
	"hound/model"
	"hound/model/providers"
	"hound/model/sources"
	"hound/view"
	"strconv"

	"github.com/gin-gonic/gin"
)

func SearchMovieMediaFilesHandler(c *gin.Context) {
	_, sourceID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"request id param invalid"+err.Error()))
		return
	}
	streamObjects, err := providers.GetLocalStreamsForMovie(sourceID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to get local streams"))
		return
	}
	res := &providers.ProviderResponseObject{
		StreamMediaDetails: providers.StreamMediaDetails{
			MediaType:   database.MediaTypeMovie,
			MediaSource: sources.MediaSourceTMDB,
			SourceID:    strconv.Itoa(sourceID),
		},
		Providers: []*providers.ProviderObject{
			{
				Provider: "Hound",
				Streams:  streamObjects,
			},
		},
	}
	helpers.SuccessResponse(c, res, 200)
}

func SearchTVShowMediaFilesHandler(c *gin.Context) {
	_, sourceID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"request id param invalid"+err.Error()))
		return
	}
	seasonNumber, err := strconv.Atoi(c.Query("season"))
	if err != nil || c.Query("season") == "" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Invalid season query param"+err.Error()))
		return
	}
	episodeNumber, err := strconv.Atoi(c.Query("episode"))
	if err != nil || c.Query("episode") == "" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Invalid episode query param"+err.Error()))
		return
	}
	streamObjects, err := providers.GetLocalStreamsForTVShow(sourceID, seasonNumber, episodeNumber)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to get local streams"))
		return
	}
	// in regular flows, should be a cached call
	epDetails, err := sources.GetEpisodeTMDB(sourceID, seasonNumber, episodeNumber)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to get episode record"))
		return
	}
	epID := strconv.Itoa(int(epDetails.ID))
	res := &providers.ProviderResponseObject{
		StreamMediaDetails: providers.StreamMediaDetails{
			MediaType:       database.MediaTypeTVShow,
			MediaSource:     sources.MediaSourceTMDB,
			SourceID:        strconv.Itoa(sourceID),
			SeasonNumber:    &seasonNumber,
			EpisodeNumber:   &episodeNumber,
			EpisodeSourceID: &epID,
		},
		Providers: []*providers.ProviderObject{
			{
				Provider: "Hound",
				Streams:  streamObjects,
			},
		},
	}
	helpers.SuccessResponse(c, res, 200)
}

func GetMediaFilesHandler(c *gin.Context) {
	limit := c.Query("limit")
	offset := c.Query("offset")
	if limit == "" {
		limit = "100"
	}
	if offset == "" {
		offset = "0"
	}
	limitNum, err := strconv.Atoi(limit)
	if err != nil {
		helpers.LogErrorWithMessage(err, "Invalid limit query param")
	}
	offsetNum, err := strconv.Atoi(offset)
	if err != nil {
		helpers.LogErrorWithMessage(err, "Invalid offset query param")
	}
	totalRecords, files, err := database.GetMediaFiles(&limitNum, &offsetNum)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to get media files"))
		return
	}
	res := &view.MediaFilesResponse{
		Files:        files,
		TotalRecords: totalRecords,
		Limit:        limitNum,
		Offset:       offsetNum,
	}
	helpers.SuccessResponse(c, res, 200)
}

func DeleteMediaFileHandler(c *gin.Context) {
	mediaFileID := c.Param("id")
	if mediaFileID == "" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Media file ID not provided"))
		return
	}
	fileID, err := strconv.Atoi(mediaFileID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid media file ID"))
		return
	}
	err = model.DeleteMediaFile(fileID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to delete media file"))
		return
	}
	helpers.SuccessResponse(c, nil, 200)
}
