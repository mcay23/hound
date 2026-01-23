package v1

import (
	"errors"
	"hound/database"
	"hound/helpers"
	"hound/model"
	"hound/model/sources"
	"hound/view"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type IngestFileRequest struct {
	MediaType       string `json:"media_type" binding:"required"` // tvshow/movie, not episode
	MediaSource     string `json:"media_source"`
	SourceID        string `json:"source_id"` // parent source id of show/movie
	SeasonNumber    *int   `json:"season_number"`
	EpisodeNumber   *int   `json:"episode_number"`
	EpisodeSourceID string `json:"episode_source_id"` // source id of episode
	FilePath        string `json:"file_path" binding:"required"`
}

func IngestFileHandler(c *gin.Context) {
	var body IngestFileRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to bind JSON"))
		return
	}
	if body.MediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Invalid media source, only tmdb is supported at the current time"))
		return
	}
	sourceID, err := strconv.Atoi(body.SourceID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to convert source id to int"))
		return
	}
	record, err := sources.UpsertMediaRecordTMDB(body.MediaType, sourceID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to upsert media record"))
		return
	}
	infoHash := "12345"
	fileIdx := 1
	mediaFile, err := model.IngestFile(record, body.SeasonNumber, body.EpisodeNumber, &infoHash, &fileIdx, nil, body.FilePath)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to ingest file"))
		return
	}
	helpers.SuccessResponse(c, gin.H{
		"file": mediaFile,
	}, 200)
}

func GetMetadataHandler(c *gin.Context) {
	uri := c.Query("uri")
	metadata, err := model.ProbeVideoFromURI(uri)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	helpers.SuccessResponse(c, gin.H{"metadata": metadata}, 200)
}

func GetTVEpisodesHandler(c *gin.Context) {
	mediaSource, sourceID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "request id param invalid"+err.Error()))
		return
	}
	sourceIDstr := strconv.Itoa(sourceID)
	episodeRecords, err := database.GetEpisodeMediaRecords(mediaSource, sourceIDstr, nil, nil)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to get episodes"))
		return
	}
	helpers.SuccessResponse(c, gin.H{"episodes": episodeRecords}, 200)
}

func GetIngestTasksHandler(c *gin.Context) {
	status := c.Query("status")
	statusSlice := strings.Split(status, ",")
	if status == "" {
		statusSlice = []string{}
	}
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
	totalRecords, tasks, err := database.FindIngestTasksForStatus(statusSlice, limitNum, offsetNum)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to get downloads"))
		return
	}
	response := view.IngestTaskResponse{
		TotalRecords: totalRecords,
		Limit:        limitNum,
		Offset:       offsetNum,
		Tasks:        tasks,
	}
	helpers.SuccessResponse(c, response, 200)
}
