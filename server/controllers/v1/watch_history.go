package v1

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"hound/database"
	"hound/helpers"
	"hound/model"
	"hound/model/sources"
	"hound/view"

	"github.com/gin-gonic/gin"
)

func GetWatchHistoryHandler(c *gin.Context) {
	recordType := database.RecordTypeMovie
	if strings.Contains(c.FullPath(), "/api/v1/tv/") {
		recordType = database.RecordTypeTVShow
	} else if !strings.Contains(c.FullPath(), "/api/v1/movie/") {
		// this shouldn't happen
		panic("Fatal error, invalid path for watch history")
	}
	username := c.GetHeader("X-Username")
	if username == "" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Username not found in header"))
		return
	}
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting user id for watch history"))
		return
	}
	mediaSource, parentSourceID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error parsing source_id: "+c.Param("id")))
		return
	}
	rewatchRecords, err := database.GetRewatchesFromSourceID(recordType, mediaSource, strconv.Itoa(parentSourceID), userID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting show rewatch records"))
		return
	}
	// exit early if rewatch record doesn't exist, since this means no watch history
	if len(rewatchRecords) == 0 {
		helpers.SuccessResponse(c,
			gin.H{
				"status": "success",
				"data":   nil,
			}, 200)
		return
	}
	var targetSeason *int
	if c.Param("seasonNumber") != "" {
		if recordType != database.RecordTypeTVShow {
			errMsg := "Season number is only valid for tv shows"
			helpers.ErrorResponseWithMessage(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), errMsg), errMsg)
			return
		}
		temp, err := strconv.Atoi(c.Param("seasonNumber"))
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error parsing target season"))
			return
		}
		targetSeason = &temp
	}
	var rewatchObjects []*view.MediaRewatchRecordWatchEvents
	for _, rewatchRecord := range rewatchRecords {
		watchEvents, err := database.GetWatchEventsFromRewatchID(rewatchRecord.RewatchID, targetSeason)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting watch events from rewatch id"))
			return
		}
		rewatchObjects = append(rewatchObjects, &view.MediaRewatchRecordWatchEvents{
			RewatchRecord: *rewatchRecord,
			TargetSeason:  targetSeason,
			WatchEvents:   watchEvents,
		})
	}
	helpers.SuccessResponse(c,
		gin.H{
			"status": "success",
			"data":   rewatchObjects,
		}, 200)
}

/*
TV Show Watch History Handlers
*/
func AddWatchHistoryTVShowHandler(c *gin.Context) {
	username := c.GetHeader("X-Username")
	if username == "" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Username not found in header"))
		return
	}
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting user id"))
		return
	}
	mediaSource, showID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error parsing source_id: "+c.Param("id")))
		return
	}
	// Only episode ids that belong to the same show should be inserted at the same time
	watchHistoryPayload := model.WatchHistoryTVShowPayload{}
	if err := c.ShouldBindJSON(&watchHistoryPayload); err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to bind watch history body: "+c.Param("id")))
		return
	}
	insertedEpisodeIDs, skippedEpisodeIDs, err := model.CreateTVShowWatchHistory(userID, mediaSource, showID, watchHistoryPayload)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error creating watch history"))
		return
	}
	response := gin.H{
		"status":               "success",
		"media_source":         mediaSource,
		"inserted_episode_ids": insertedEpisodeIDs,
	}
	if len(*skippedEpisodeIDs) > 0 {
		response["skipped_episode_ids"] = skippedEpisodeIDs
	}
	helpers.SuccessResponse(c, response, 200)
}

func DeleteWatchHistoryHandler(c *gin.Context) {
	recordType := database.RecordTypeMovie
	if strings.Contains(c.FullPath(), "/tv/") {
		recordType = database.RecordTypeTVShow
	} else if !strings.Contains(c.FullPath(), "/movie/") {
		// this shouldn't happen
		panic("Fatal error, invalid path for watch history")
	}
	username := c.GetHeader("X-Username")
	if username == "" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Username not found in header"))
		return
	}
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting user id"))
		return
	}
	// Only episode ids that belong to the same show should be inserted at the same time
	type deleteWatchHistoryPayload struct {
		WatchEventIDs []int64 `json:"watch_event_ids" binding:"required"`
	}
	payload := deleteWatchHistoryPayload{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to bind watch history body: "+c.Param("id")))
		return
	}
	// get record id from show source id
	mediaSource, showID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error parsing source_id: "+c.Param("id")))
		return
	}
	has, record, err := database.GetMediaRecord(recordType, mediaSource, strconv.Itoa(showID))
	if !has || err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting media record"))
		return
	}
	if err := database.BatchDeleteWatchEvents(payload.WatchEventIDs, userID, int(record.RecordID)); err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error deleting watch history records"))
		return
	}
	helpers.SuccessResponse(c, gin.H{"status": "success"}, 200)
}

// Create new rewatch for tv show
// should be user trigerred
func AddTVShowRewatchHandler(c *gin.Context) {
	username := c.GetHeader("X-Username")
	if username == "" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Username not found in header"))
		return
	}
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting user id for watch history"))
		return
	}
	mediaSource, showID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error parsing source_id: "+c.Param("id")))
		return
	}
	startedAt := time.Now().UTC()
	// supplying a body is optional
	if c.Request.ContentLength != 0 {
		type addRewatchPayload struct {
			StartedAt string `json:"rewatch_started_at"`
		}
		rewatchPayload := addRewatchPayload{}
		if err := c.ShouldBindJSON(&rewatchPayload); err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to bind watch history body: "+c.Param("id")))
			return
		}
		if rewatchPayload.StartedAt != "" {
			parsed, err := time.Parse(time.RFC3339, rewatchPayload.StartedAt)
			if err != nil {
				helpers.ErrorResponseWithMessage(c, err, "Error parsing rewatch_started_at, must be RFC3339 string")
				return
			}
			startedAt = parsed
		}
	}
	rewatchRecord, err := model.InsertRewatchFromSourceID(database.MediaTypeTVShow, mediaSource, strconv.Itoa(showID), userID, startedAt)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error creating rewatch record"))
		return
	}
	helpers.SuccessResponse(c,
		gin.H{
			"status": "success",
			"data":   rewatchRecord,
		}, 200)
}

/*
Movie Watch Handlers
*/

// for movies, only a single rewatch is supported
func AddWatchHistoryMovieHandler(c *gin.Context) {
	username := c.GetHeader("X-Username")
	if username == "" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Username not found in header"))
		return
	}
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting user id"))
		return
	}
	mediaSource, sourceID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error parsing source_id: "+c.Param("id")))
		return
	}
	watchHistoryPayload := model.WatchHistoryMoviePayload{}
	if err := c.ShouldBindJSON(&watchHistoryPayload); err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to bind watch history body: "+c.Param("id")))
		return
	}
	insertedSourceID, err := model.CreateMovieWatchHistory(userID, mediaSource, sourceID, watchHistoryPayload)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	helpers.SuccessResponse(c, gin.H{
		"status":             "success",
		"media_source":       mediaSource,
		"action_type":        strings.ToLower(watchHistoryPayload.ActionType),
		"inserted_source_id": insertedSourceID,
	}, 200)
}
