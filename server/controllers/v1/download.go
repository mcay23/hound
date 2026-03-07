package v1

import (
	"errors"
	"hound/database"
	"hound/helpers"
	"hound/model"
	"hound/model/providers"
	"hound/sources"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TVSeasonDownloadRequest struct {
	database.IngestDownloadPreferences
	SkipDownloadedEpisodes *bool `json:"skip_downloaded_episodes,omitempty"`
}

// This downloads the media file to the server, not the client
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
	helpers.SuccessResponse(c, gin.H{"status": "started"}, 200)
}

// Downloads a whole tv season based on the top result from the providers response
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
	queuedEpisodes := []int{}
	type skippedEpisode struct {
		EpisodeNumber int    `json:"episode_number"`
		Error         string `json:"error"`
	}
	var skippedEpisodes []skippedEpisode
	for _, ep := range seasonDetails.Episodes {
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
			ep := skippedEpisode{
				EpisodeNumber: ep.EpisodeNumber,
				Error:         err.Error(),
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
	helpers.SuccessResponse(c, gin.H{"status": "queued", "season": seasonNumber, "queued_episodes": queuedEpisodes, "skipped_episodes": skippedEpisodes}, 200)
}

// Cancel downloads
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
	helpers.SuccessResponse(c, gin.H{"ingest_task_id": taskID, "status": "pending_cancel"}, 200)
}
