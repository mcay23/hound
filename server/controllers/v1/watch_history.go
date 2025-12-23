package v1

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"hound/database"
	"hound/helpers"
	"hound/model/sources"
	"hound/view"

	"github.com/gin-gonic/gin"
)

const scrobbleCacheTTL = 48 * time.Hour

func GetWatchHistoryHandler(c *gin.Context) {
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
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting user id for watch history"))
		return
	}
	mediaSource, showID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error parsing source_id: "+c.Param("id")))
		return
	}
	rewatchRecords, err := database.GetRewatchesFromSourceID(recordType, mediaSource, strconv.Itoa(showID), userID)
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
	// Only episode ids that belong to the same show should be inserted at the same time
	type addWatchHistoryPayload struct {
		EpisodeIDs []int   `json:"episode_ids" binding:"required"` // tmdb unique id for episode
		ActionType string  `json:"action_type" binding:"required,oneof=watch scrobble"`
		RewatchID  *int64  `json:"rewatch_id"`
		WatchedAt  *string `json:"watched_at"`
	}
	watchHistoryPayload := addWatchHistoryPayload{}
	if err := c.ShouldBindJSON(&watchHistoryPayload); err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to bind watch history body: "+c.Param("id")))
		return
	}
	actionType := strings.ToLower(watchHistoryPayload.ActionType)
	if actionType != database.ActionWatch && actionType != database.ActionScrobble {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid action type"))
		return
	}
	watchTime := time.Now().UTC()
	if watchHistoryPayload.WatchedAt != nil && *watchHistoryPayload.WatchedAt != "" {
		parsed, err := time.Parse(time.RFC3339, *watchHistoryPayload.WatchedAt)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid watched_at timestamp"))
			return
		}
		watchTime = parsed.UTC()
	}
	// 1. Parse episode ids
	mediaSource, showID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error parsing source_id: "+c.Param("id")))
		return
	}
	// check for duplicated episode ids
	episodeSet := make(map[int]struct{})
	var episodeIDs []int
	for _, id := range watchHistoryPayload.EpisodeIDs {
		if _, exists := episodeSet[id]; exists {
			continue
		}
		episodeSet[id] = struct{}{}
		episodeIDs = append(episodeIDs, id)
	}
	if len(episodeIDs) == 0 {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "No valid episode ids provided"))
		return
	}
	// 2. Upsert show
	if _, err := sources.UpsertTVShowRecordTMDB(showID); err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error upserting tv show: "+c.Param("id")))
		return
	}
	// check if episode ids are in the database, and belong to the correct show
	episodeMap, invalidIDs, err := database.CheckShowEpisodesIDs(mediaSource, strconv.Itoa(showID), episodeIDs)
	if len(invalidIDs) > 0 {
		errorStr := ""
		for _, item := range invalidIDs {
			errorStr += strconv.Itoa(item) + ","
		}
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid Episode IDs found:"+errorStr))
		return
	}
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error checking episode ids for tv show:"+c.Param("id")))
		return
	}
	var targetRewatchID *int64
	if watchHistoryPayload.RewatchID != nil {
		targetRewatchID = watchHistoryPayload.RewatchID
		// check if rewatch_id exists for this user
		rewatchRecords, err := database.GetRewatchesFromSourceID(database.RecordTypeTVShow, mediaSource, strconv.Itoa(showID), userID)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting show rewatch records"))
			return
		}
		found := false
		for _, item := range rewatchRecords {
			if item.RewatchID == *targetRewatchID {
				found = true
				break
			}
		}
		if !found {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Could not find this rewatch ID in the database"))
			return
		}
	}
	// 3. Get most current rewatch or create new rewatch if none if rewatch payload is empty
	if targetRewatchID == nil {
		rewatchRecord, err := database.GetActiveRewatchFromSourceID(database.MediaTypeTVShow, mediaSource, strconv.Itoa(showID), userID)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting active rewatch: "+c.Param("id")))
			return
		}
		// add rewatch record if none exists
		if rewatchRecord == nil {
			rewatchRecord, err = database.InsertRewatchFromSourceID(database.MediaTypeTVShow, mediaSource,
				strconv.Itoa(showID), userID, time.Now().UTC())
			if err != nil {
				helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error creating rewatch record"))
				return
			}
		}
		targetRewatchID = &rewatchRecord.RewatchID
	}
	// 4. Filter cached scrobbles, since we don't want to accidentally double insert scrobbles
	// if they are inserted within X hours of each other. Watches are fine since it's manual
	type pendingInsert struct {
		EpisodeID    int
		EpisodeIDStr string
		CacheKey     string
	}
	pendingRecords := make([]database.WatchEventsRecord, 0, len(episodeIDs))
	pendingMetadata := make([]pendingInsert, 0, len(episodeIDs))
	skippedEpisodeIDs := []int{}
	// create cache keys for scrobbles to prevent accidental duplicate inserts
	for _, episodeID := range episodeIDs {
		episodeIDStr := strconv.Itoa(episodeID)
		cacheKey := ""
		if actionType == database.ActionScrobble {
			// realistically, scrobbles should only operate on the latest rewatch session
			cacheKey = fmt.Sprintf("watch_history:scrobble:userid-%d:rewatchid-%d:%d:%s:%s-%s", userID, targetRewatchID,
				database.RecordTypeEpisode, mediaSource, episodeIDStr)
			var cached bool
			cacheHit, err := database.GetCache(cacheKey, &cached)
			if err != nil {
				helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error checking scrobble cache"))
				return
			}
			// if cache hit and scrobble, skip insert
			if cacheHit {
				skippedEpisodeIDs = append(skippedEpisodeIDs, episodeID)
				continue
			}
		}
		int64Val, err := strconv.ParseInt(episodeMap[episodeIDStr], 10, 64)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error parsing episode id"))
			return
		}
		pendingRecords = append(pendingRecords, database.WatchEventsRecord{
			RewatchID: targetRewatchID,
			RecordID:  int64Val,
			WatchType: actionType,
			WatchedAt: watchTime,
		})
		pendingMetadata = append(pendingMetadata, pendingInsert{
			EpisodeID:    episodeID,
			EpisodeIDStr: episodeIDStr,
			CacheKey:     cacheKey,
		})
	}
	if len(pendingRecords) == 0 {
		response := gin.H{
			"status":               "success",
			"media_source":         mediaSource,
			"inserted_episode_ids": []int{},
		}
		if len(skippedEpisodeIDs) > 0 {
			response["skipped_episode_ids"] = skippedEpisodeIDs
		}
		helpers.SuccessResponse(c, response, 200)
		return
	}
	if err := database.BatchInsertWatchEvents(pendingRecords); err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error inserting watch events records"))
		return
	}
	// set idempotence cache for scrobbles, 72 hours
	// only set once inserts are successful
	insertedEpisodeIDs := make([]int, len(pendingMetadata))
	for idx, meta := range pendingMetadata {
		insertedEpisodeIDs[idx] = meta.EpisodeID
		if meta.CacheKey != "" {
			if _, err := database.SetCache(meta.CacheKey, true, scrobbleCacheTTL); err != nil {
				helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error caching scrobble entry"))
				return
			}
		}
	}
	response := gin.H{
		"status":               "success",
		"media_source":         mediaSource,
		"inserted_episode_ids": insertedEpisodeIDs,
	}
	if len(skippedEpisodeIDs) > 0 {
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
	rewatchRecord, err := database.InsertRewatchFromSourceID(database.MediaTypeTVShow, mediaSource, strconv.Itoa(showID), userID, startedAt)
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
	type addWatchHistoryPayload struct {
		ActionType string  `json:"action_type" binding:"required,oneof=watch scrobble"`
		WatchedAt  *string `json:"watched_at"`
	}
	watchHistoryPayload := addWatchHistoryPayload{}
	if err := c.ShouldBindJSON(&watchHistoryPayload); err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to bind watch history body: "+c.Param("id")))
		return
	}
	actionType := strings.ToLower(watchHistoryPayload.ActionType)
	if actionType != database.ActionWatch && actionType != database.ActionScrobble {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid action type"))
		return
	}
	watchTime := time.Now().UTC()
	if watchHistoryPayload.WatchedAt != nil && *watchHistoryPayload.WatchedAt != "" {
		parsed, err := time.Parse(time.RFC3339, *watchHistoryPayload.WatchedAt)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid watched_at timestamp"))
			return
		}
		watchTime = parsed.UTC()
	}
	// 1. Upsert movie record
	movieRecord, err := sources.UpsertMovieRecordTMDB(sourceID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error upserting media record for "+c.Param("id")))
		return
	}
	// 2. Get most current rewatch or create new rewatch if none exists
	rewatchRecord, err := database.GetActiveRewatchFromSourceID(database.MediaTypeMovie, mediaSource, strconv.Itoa(sourceID), userID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting active rewatch: "+c.Param("id")))
		return
	}
	// add rewatch record if none exists
	if rewatchRecord == nil {
		rewatchRecord, err = database.InsertRewatchFromSourceID(database.MediaTypeMovie, mediaSource,
			strconv.Itoa(sourceID), userID, time.Now().UTC())
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error creating rewatch record"))
			return
		}
	}
	watchEvent := database.WatchEventsRecord{
		RewatchID: &rewatchRecord.RewatchID,
		RecordID:  movieRecord.RecordID,
		WatchType: actionType,
		WatchedAt: watchTime,
	}
	if actionType == database.ActionScrobble {
		// check cache for recent scrobble
		cacheKey := fmt.Sprintf("watch_history:scrobble:userid-%d:rewatchid-%d:%s:%s-%d",
			userID, rewatchRecord.RewatchID, database.RecordTypeMovie, mediaSource, sourceID)
		var cached bool
		cacheHit, err := database.GetCache(cacheKey, &cached)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error checking scrobble cache"))
			return
		}
		// if cache hit, return without inserting
		if cacheHit {
			helpers.SuccessResponse(c, gin.H{
				"status":             "success",
				"media_source":       mediaSource,
				"action_type":        actionType,
				"inserted_source_id": nil,
			}, 200)
			return
		}
		// set cache for scrobbles to prevent accident duplicate inserts
		if _, err := database.SetCache(cacheKey, true, scrobbleCacheTTL); err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error caching scrobble entry"))
			return
		}
	}
	err = database.BatchInsertWatchEvents([]database.WatchEventsRecord{watchEvent})
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error inserting watch event to db: "+c.Param("id")))
		return
	}
	helpers.SuccessResponse(c, gin.H{
		"status":             "success",
		"media_source":       mediaSource,
		"action_type":        actionType,
		"inserted_source_id": sourceID,
	}, 200)
}
