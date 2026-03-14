package v1

import (
	"errors"
	"hound/database"
	"hound/helpers"
	"hound/model"
	"hound/model/providers"
	"hound/sources"
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
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError),
			"Failed to parse encoded string:"+c.Param("encodedString")))
		return
	}
	if streamDetails.StreamProtocol == database.ProtocolFileHTTP {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"This file should already be downloaded"))
		return
	}
	if streamDetails.MediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Invalid media source: "+streamDetails.MediaSource))
		return
	}
	err = model.CreateIngestTaskDownload(streamDetails, nil, false)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to download torrent"))
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
	Season          int              `json:"season"`
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
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "request id param invalid"+err.Error()))
		return
	}
	seasonNumber, err := strconv.Atoi(c.Param("seasonNumber"))
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid season number"))
		return
	}
	var request TVSeasonDownloadRequest
	if err := c.ShouldBindJSON(&request); err != nil && err.Error() != "EOF" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid preferences body"))
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
				helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
					"Invalid preference list in request body"))
				return
			}
		}
	}
	seasonDetails, err := sources.GetTVSeasonTMDB(showID, seasonNumber)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to get tv season details"))
		return
	}
	if request.EpisodesToDownload != nil && len(*request.EpisodesToDownload) == 0 {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Empty episodes_to_download passed, either fill this or omit this field to download all episodes"))
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
		if len(prefs.PreferenceList) > 0 {
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
			if err.Error() == helpers.AlreadyExists {
				_ = helpers.LogErrorWithMessage(err,
					"Failed to queue episode: episode already exists "+strconv.Itoa(ep.EpisodeNumber))
				continue
			}
			_ = helpers.LogErrorWithMessage(err, "Failed to queue episode "+strconv.Itoa(ep.EpisodeNumber))
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
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid task ID Param"))
		return
	}
	task, err := database.GetIngestTask(database.IngestTask{IngestTaskID: int64(taskID)})
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to get task"))
		return
	}
	if task.Status != database.IngestStatusDownloading &&
		task.Status != database.IngestStatusPendingDownload {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Only tasks that are downloading or pending_download can be canceled"))
		return
	}
	updatedTask := database.IngestTask{
		IngestTaskID: int64(taskID),
		Status:       database.IngestStatusCanceled,
	}
	_, err = database.UpdateIngestTask(&updatedTask)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to cancel task"))
		return
	}
	helpers.SuccessResponse(c, CancelIngestTaskResponse{IngestTaskID: taskID, Status: "pending_cancel"}, 200)
}
