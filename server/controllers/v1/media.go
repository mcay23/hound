package v1

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/model"
	"github.com/mcay23/hound/sources"
	"github.com/mcay23/hound/view"

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

type IngestFileResponse struct {
	MediaFile *database.MediaFile `json:"file"`
}

type GetMetadataResponse struct {
	Metadata *database.VideoMetadata `json:"metadata"`
}

type GetTVEpisodesResponse struct {
	Episodes []database.MediaRecord `json:"episodes"`
}

func IngestFileHandler(c *gin.Context) {
	var body IngestFileRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to bind json: %w: %w", internal.BadRequestError, err))
		return
	}
	if body.MediaSource != sources.MediaSourceTMDB {
		internal.ErrorResponse(c, fmt.Errorf("invalid media source: %w", internal.BadRequestError))
		return
	}
	sourceID, err := strconv.Atoi(body.SourceID)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to convert source id to int: %w: %w", internal.BadRequestError, err))
		return
	}
	record, err := sources.UpsertMediaRecordTMDB(body.MediaType, sourceID)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to upsert media record: %w", err))
		return
	}
	infoHash := "12345"
	fileIdx := 1
	mediaFile, err := model.IngestFile(record, body.SeasonNumber, body.EpisodeNumber, &infoHash, &fileIdx, nil, body.FilePath, model.IngestTransferMove, database.FileOriginHoundManaged)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, IngestFileResponse{MediaFile: mediaFile}, 200)
}

func GetMetadataHandler(c *gin.Context) {
	uri := c.Query("uri")
	metadata, err := model.ProbeVideoFromURI(uri)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, GetMetadataResponse{Metadata: metadata}, 200)
}

func GetTVEpisodesHandler(c *gin.Context) {
	mediaSource, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get source id from params: %w: %w", internal.BadRequestError, err))
		return
	}
	sourceIDstr := strconv.Itoa(sourceID)
	episodeRecords, err := database.GetEpisodeMediaRecords(mediaSource, sourceIDstr, nil, nil)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get episode media records: %w", err))
		return
	}
	internal.SuccessResponse(c, GetTVEpisodesResponse{Episodes: episodeRecords}, 200)
}

// @Router /api/v1/ingest [get]
// @Summary Get Ingest Tasks
// @Tags Media
// @Accept json
// @Produce json
// @Param status query string false "Comma separated status"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} V1SuccessResponse{data=view.IngestTaskResponse}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
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
	limitNum, offsetNum, err := getLimitOffset(limit, offset)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	totalRecords, tasks, err := database.FindIngestTasksForStatus(statusSlice, limitNum, offsetNum)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to find ingest tasks for status: %w", err))
		return
	}
	response := view.IngestTaskResponse{
		TotalRecords: totalRecords,
		Limit:        limitNum,
		Offset:       offsetNum,
		Tasks:        tasks,
	}
	internal.SuccessResponse(c, response, 200)
}
