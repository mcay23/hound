package v1

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"hound/helpers"
	model "hound/model"
	"hound/model/database"
	"hound/model/sources"

	"github.com/gin-gonic/gin"
)

const scrobbleCacheTTL = 72 * time.Hour

// Only episode ids that belong to the same show should be inserted at the same time
type AddWatchHistoryPayload struct {
	EpisodeIDs []int   `json:"episode_ids" binding:"required"` // tmdb unique id for episode
	ActionType string  `json:"action_type" binding:"required,oneof=watch scrobble"`
	WatchedAt  *string `json:"watched_at"`
}

func GetWatchHistoryTVShowHandler(c *gin.Context) {
	username := c.GetHeader("X-Username")
	if username == "" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Username not found in header"))
		return
	}
	mediaSource, showID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.SourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error parsing source_id: "+c.Param("id")))
		return
	}
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting user id for watch history"))
		return
	}
	// get record
	has, _, err := database.GetMediaRecord(database.RecordTypeTVShow, mediaSource, strconv.Itoa(showID))
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting media record for source ID"))
		return
	}
	// exit early if show record doesn't exist, since this means no watch history
	if !has {
		helpers.SuccessResponse(c,
			gin.H{
				"status": "success",
				"data":   nil,
			}, 200)
	}
	// get all show_rewatches for user and source id
	events, err := database.GetWatchEventsFromSourceID(database.MediaTypeTVShow, mediaSource, strconv.Itoa(showID), userID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting watch events: "+c.Param("id")))
		return
	}
	helpers.SuccessResponse(c,
		gin.H{
			"status": "success",
			"data":   events,
		}, 200)
}

func AddWatchHistoryTVShowHandler(c *gin.Context) {
	username := c.GetHeader("X-Username")
	if username == "" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Username not found in header"))
		return
	}
	watchHistoryPayload := AddWatchHistoryPayload{}
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
	if err != nil || mediaSource != sources.SourceTMDB {
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
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting user id for watch history"))
		return
	}
	// 3. Get most current rewatch or create new rewatch if none
	rewatchRecord, err := database.GetActiveRewatchFromSourceID(database.MediaTypeTVShow, mediaSource, strconv.Itoa(showID), userID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting active rewatch: "+c.Param("id")))
		return
	}
	// add rewatch record if none exists
	if rewatchRecord == nil {
		rewatchRecord, err = database.InsertRewatchFromSourceID(database.MediaTypeTVShow, mediaSource, strconv.Itoa(showID), userID)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error creating rewatch record"))
			return
		}
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
	// create cache keys for scrobbles
	for _, episodeID := range episodeIDs {
		episodeIDStr := strconv.Itoa(episodeID)
		cacheKey := ""
		if actionType == database.ActionScrobble {
			cacheKey = fmt.Sprintf("watch_history:scrobble:%d:%s", rewatchRecord.UserID, episodeIDStr)
			var cached bool
			cacheHit, err := model.GetCache(cacheKey, &cached)
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
		showRewatchID := &rewatchRecord.ShowRewatchID
		pendingRecords = append(pendingRecords, database.WatchEventsRecord{
			ShowRewatchID: showRewatchID,
			RecordID:      int64(episodeID),
			WatchType:     actionType,
			WatchedAt:     watchTime,
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
			if _, err := model.SetCache(meta.CacheKey, true, scrobbleCacheTTL); err != nil {
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
