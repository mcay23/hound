package v1

import (
	"errors"
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/model"
	"hound/providers"
	"hound/sources"
	"log/slog"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TVSeasonDownloadRequest struct {
	database.IngestDownloadPreferences
	SkipDownloadedEpisodes *bool  `json:"skip_downloaded_episodes,omitempty"`
	EpisodesToDownload     *[]int `json:"episodes_to_download,omitempty"`
}

type DownloadResponse struct {
	Status string `json:"status" example:"started"`
}

// @Router /api/v1/download/{encodedString} [post]
// @Summary Download media file to server
// @Tags Download
// @Accept json
// @Produce json
// @Param encodedString path string true "Encoded Stream String - get from Query providers"
// @Success 200 {object} V1SuccessResponse{data=DownloadResponse}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func DownloadHandler(c *gin.Context) {
	streamDetails, err := providers.DecodeJsonStreamAES(c.Param("encodedString"))
	if err != nil || streamDetails == nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to parse encoded string for %s: %w", c.Param("encodedString"), err))
		return
	}
	if streamDetails.StreamProtocol == database.ProtocolFileHTTP {
		helpers.ErrorResponse(c, fmt.Errorf("this file should already be downloaded: %w: %w", helpers.BadRequestError, err))
		return
	}
	if streamDetails.MediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, fmt.Errorf("invalid media source for %s: %w: %w", streamDetails.MediaSource, helpers.BadRequestError, err))
		return
	}
	err = model.CreateIngestTaskDownload(streamDetails, nil, false)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to download torrent: %w", err))
		return
	}
	helpers.SuccessResponse(c, DownloadResponse{Status: "started"}, 200)
}

/*
Downloads a whole tv season based on the top result from the providers response
and the given preferences (infohash, etc.)
the source provider itself is resolved by the download worker when the episode
is picked up. The idea is this will naturally rate-limit calls made to external sources
*/
type skippedEpisode struct {
	EpisodeNumber int     `json:"episode_number"`
	Error         *string `json:"error,omitempty"`
}

type TVSeasonDownloadResponse struct {
	Status          string           `json:"status" example:"queued"`
	Season          int              `json:"season_number"`
	QueuedEpisodes  []int            `json:"queued_episodes"`
	SkippedEpisodes []skippedEpisode `json:"skipped_episodes"`
}

// @Router /api/v1/tv/{id}/season/{seasonNumber}/download [post]
// @Summary Download TV Season
// @Tags Download
// @Accept json
// @Produce json
// @Param id path int true "Media ID" example(tmdb-1234)
// @Param seasonNumber path int true "Season Number"
// @Param request body TVSeasonDownloadRequest false "Download Preferences"
// @Success 200 {object} V1SuccessResponse{data=TVSeasonDownloadResponse}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func DownloadTVSeasonHandler(c *gin.Context) {
	mediaSource, showID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, fmt.Errorf("request id param invalid: %w: %w", helpers.BadRequestError, err))
		return
	}
	seasonNumber, err := strconv.Atoi(c.Param("seasonNumber"))
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("invalid season number: %w: %w", helpers.BadRequestError, err))
		return
	}
	var request TVSeasonDownloadRequest
	if err := c.ShouldBindJSON(&request); err != nil && err.Error() != "EOF" {
		helpers.ErrorResponse(c, fmt.Errorf("invalid preferences body: %w: %w", helpers.BadRequestError, err))
		return
	}
	skipDownloaded := true
	if request.SkipDownloadedEpisodes != nil {
		skipDownloaded = *request.SkipDownloadedEpisodes
	}
	prefs := request.IngestDownloadPreferences
	if len(prefs.PreferenceList) > 0 {
		for _, pref := range prefs.PreferenceList {
			if pref.InfoHashPreference == nil && pref.StringMatchPreference == nil {
				helpers.ErrorResponse(c, fmt.Errorf("invalid preference list in request body: %w: %w", helpers.BadRequestError, err))
				return
			}
		}
	}
	seasonDetails, err := sources.GetTVSeasonTMDB(showID, seasonNumber)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to get tv season details: %w", err))
		return
	}
	if request.EpisodesToDownload != nil && len(*request.EpisodesToDownload) == 0 {
		helpers.ErrorResponse(c, fmt.Errorf("empty episodes_to_download passed, either fill this or omit this field to download all episodes: %w: %w", helpers.BadRequestError, err))
		return
	}
	queuedEpisodes := []int{}
	var skippedEpisodes []skippedEpisode
	for _, ep := range seasonDetails.Episodes {
		if request.EpisodesToDownload != nil {
			if !slices.Contains(*request.EpisodesToDownload, ep.EpisodeNumber) {
				ep := skippedEpisode{
					EpisodeNumber: ep.EpisodeNumber,
					Error:         nil,
				}
				skippedEpisodes = append(skippedEpisodes, ep)
				continue
			}
		}
		streamObj := &providers.StreamObjectFull{
			StreamMediaDetails: providers.StreamMediaDetails{
				MediaSource:   sources.MediaSourceTMDB,
				MediaType:     database.RecordTypeTVShow,
				SourceID:      strconv.Itoa(showID),
				SeasonNumber:  &seasonNumber,
				EpisodeNumber: &ep.EpisodeNumber,
			},
			StreamObject: providers.StreamObject{
				StreamProtocol: "",
			},
		}
		var prefsPtr *database.IngestDownloadPreferences
		// if prefs exist
		if len(prefs.PreferenceList) > 0 || prefs.ProviderProfileID > 0 {
			prefsPtr = &prefs
		}
		// if ignoreDownloaded is true, and file already exists, AlreadyExists error is returned
		err = model.CreateIngestTaskDownload(streamObj, prefsPtr, skipDownloaded)
		if err != nil {
			errMsg := err.Error()
			ep := skippedEpisode{
				EpisodeNumber: ep.EpisodeNumber,
				Error:         &errMsg,
			}
			skippedEpisodes = append(skippedEpisodes, ep)
			if errors.Is(err, helpers.AlreadyExistsError) {
				slog.Debug("Failed to queue episode: episode already exists " + strconv.Itoa(ep.EpisodeNumber))
				continue
			}
			slog.Debug("Failed to queue episode " + strconv.Itoa(ep.EpisodeNumber))
			continue
		} else {
			queuedEpisodes = append(queuedEpisodes, ep.EpisodeNumber)
		}
	}
	res := TVSeasonDownloadResponse{
		Status:         "queued",
		Season:         seasonNumber,
		QueuedEpisodes: queuedEpisodes,
	}
	res.SkippedEpisodes = skippedEpisodes
	helpers.SuccessResponse(c, res, 200)
}

type CancelIngestTaskResponse struct {
	IngestTaskID int    `json:"ingest_task_id"`
	Status       string `json:"status" example:"pending_cancel"`
}

// @Router /api/v1/ingest/{taskID}/cancel [post]
// @Summary Cancel an ingest/download task
// @Tags Download
// @Accept json
// @Produce json
// @Param taskID path int true "Task ID"
// @Success 200 {object} V1SuccessResponse{data=CancelIngestTaskResponse}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func CancelIngestTaskHandler(c *gin.Context) {
	taskIDStr := c.Param("taskID")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("invalid task_id: %w: %w", helpers.BadRequestError, err))
		return
	}
	task, err := database.GetIngestTask(database.IngestTask{IngestTaskID: int64(taskID)})
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to get task: %w", err))
		return
	}
	if task.Status != database.IngestStatusDownloading &&
		task.Status != database.IngestStatusPendingDownload {
		helpers.ErrorResponse(c, fmt.Errorf("only tasks that are downloading or pending_download can be canceled: %w", helpers.BadRequestError))
		return
	}
	updatedTask := database.IngestTask{
		IngestTaskID: int64(taskID),
		Status:       database.IngestStatusCanceled,
	}
	_, err = database.UpdateIngestTask(&updatedTask)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to update task: %w", err))
		return
	}
	helpers.SuccessResponse(c, CancelIngestTaskResponse{IngestTaskID: taskID, Status: "pending_cancel"}, 200)
}
