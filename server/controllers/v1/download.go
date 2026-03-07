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
	err = model.CreateIngestTaskDownload(streamDetails)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to download torrent"))
		return
	}
	helpers.SuccessResponse(c, gin.H{"status": "started"}, 200)
}

// Downloads a whole tv season based on the top result from the providers response

func DownloadTVSeasonHandler(c *gin.Context) {

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
