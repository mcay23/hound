package v1

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"hound/database"
	"hound/helpers"
	"hound/model"
	"hound/sources"
	"hound/view"

	"github.com/gin-gonic/gin"
)

// @Router /v1/watch_activity [get]
// @Summary Get Watch Activity for User
// @Tags Watch Activity
// @Accept json
// @Produce json
// @Param limit query int false "Limit - Defaults at 500"
// @Param offset query int false "Offset"
// @Param start_time query string false "Start time in RFC3339 format" example(2026-03-13T10:20:30Z)
// @Param end_time query string false "End time in RFC3339 format" example(2026-03-13T10:20:30Z)
// @Success 200 {object} V1SuccessResponse{data=view.WatchActivityResponse}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetWatchActivityHandler(c *gin.Context) {
	username := c.GetHeader("X-Username")
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("error getting user id for username %s: %w", username, err))
		return
	}
	limitQuery := c.DefaultQuery("limit", "500")
	offsetQuery := c.DefaultQuery("offset", "0")
	startQuery := c.Query("start_time")
	endQuery := c.Query("end_time")
	limit, offset, err := getLimitOffset(limitQuery, offsetQuery)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	var startTime, endTime *time.Time
	if startQuery != "" {
		t, err := time.Parse(time.RFC3339, startQuery)
		if err != nil {
			helpers.ErrorResponse(c, fmt.Errorf("error parsing start time: %w (must be RFC3339): %w", helpers.BadRequestError, err))
			return
		}
		startTime = &t
	}
	if endQuery != "" {
		t, err := time.Parse(time.RFC3339, endQuery)
		if err != nil {
			helpers.ErrorResponse(c, fmt.Errorf("error parsing end time: %w (must be RFC3339): %w", helpers.BadRequestError, err))
			return
		}
		endTime = &t
	}
	activity, total, err := database.GetWatchActivity(userID, startTime, endTime, limit, offset)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	res := view.WatchActivityResponse{
		WatchActivity: activity,
		Limit:         limit,
		Offset:        offset,
		TotalRecords:  total,
	}
	helpers.SuccessResponse(c, res, 200)
}

// @Router /v1/tv/{id}/history [get]
// @Router /v1/tv/{id}/season/{seasonNumber}/history [get]
// @Summary Get TV Show Watch History
// @Tags Watch History
// @Accept json
// @Produce json
// @Param id path int true "Media ID" example(tmdb-1234)
// @Param seasonNumber query int false "Season number"
// @Success 200 {object} V1SuccessResponse{data=[]view.MediaRewatchRecordWatchEvents}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetWatchHistoryTVHandler(c *gin.Context) {
	handleGetWatchHistory(c, database.RecordTypeTVShow)
}

// @Router /v1/movie/{id}/history [get]
// @Summary Get Movie Watch History
// @Tags Watch History
// @Accept json
// @Produce json
// @Param id path int true "Media ID" example(tmdb-1234)
// @Success 200 {object} V1SuccessResponse{data=[]view.MediaRewatchRecordWatchEvents}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetWatchHistoryMovieHandler(c *gin.Context) {
	handleGetWatchHistory(c, database.RecordTypeMovie)
}

func handleGetWatchHistory(c *gin.Context, recordType string) {
	username := c.GetHeader("X-Username")
	if username == "" {
		helpers.ErrorResponse(c, fmt.Errorf("X-Username not found in header: %w", helpers.BadRequestError))
		return
	}
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("error getting user id for username %s: %w", username, err))
		return
	}
	mediaSource, parentSourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, fmt.Errorf("error parsing id: %w: %w", helpers.BadRequestError, err))
		return
	}
	rewatchRecords, err := database.GetRewatchesFromSourceID(recordType, mediaSource, strconv.Itoa(parentSourceID), userID)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	// exit early if rewatch record doesn't exist, since this means no watch history
	if len(rewatchRecords) == 0 {
		helpers.SuccessResponse(c, nil, 200)
		return
	}
	var targetSeason *int
	if c.Param("seasonNumber") != "" {
		if recordType != database.RecordTypeTVShow {
			helpers.ErrorResponse(c, fmt.Errorf("season number passed for non-tvshow: %w", helpers.BadRequestError))
			return
		}
		temp, err := strconv.Atoi(c.Param("seasonNumber"))
		if err != nil {
			helpers.ErrorResponse(c, fmt.Errorf("error parsing season number: %w: %w", helpers.BadRequestError, err))
			return
		}
		targetSeason = &temp
	}
	var rewatchObjects []*view.MediaRewatchRecordWatchEvents
	for _, rewatchRecord := range rewatchRecords {
		watchEvents, err := database.GetWatchEventsFromRewatchID(rewatchRecord.RewatchID, targetSeason)
		if err != nil {
			helpers.ErrorResponse(c, fmt.Errorf("error getting watch events from rewatch id: %w: %w", helpers.BadRequestError, err))
			return
		}
		rewatchObjects = append(rewatchObjects, &view.MediaRewatchRecordWatchEvents{
			RewatchRecord: *rewatchRecord,
			TargetSeason:  targetSeason,
			WatchEvents:   watchEvents,
		})
	}
	helpers.SuccessResponse(c, rewatchObjects, 200)
}

type AddWatchHistoryTVResponse struct {
	MediaSource        string `json:"media_source"`
	InsertedEpisodeIDs *[]int `json:"inserted_episode_ids,omitempty"`
	SkippedEpisodeIDs  *[]int `json:"skipped_episode_ids,omitempty"`
}

// @Router /v1/tv/{id}/history [post]
// @Summary Add TV Show Watch History
// @Tags Watch History
// @Accept json
// @Produce json
// @Param id path int true "Media ID" example(tmdb-1234)
// @Param body body model.WatchHistoryTVShowPayload true "Watch History Payload"
// @Success 200 {object} V1SuccessResponse{data=AddWatchHistoryTVResponse}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func AddWatchHistoryTVHandler(c *gin.Context) {
	username := c.GetHeader("X-Username")
	if username == "" {
		helpers.ErrorResponse(c, fmt.Errorf("X-Username not found in header: %w", helpers.BadRequestError))
		return
	}
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("error getting user id for username %s: %w", username, err))
		return
	}
	mediaSource, showID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, fmt.Errorf("error parsing id param: %w: %w", helpers.BadRequestError, err))
		return
	}
	// Only episode ids that belong to the same show should be inserted at the same time
	watchHistoryPayload := model.WatchHistoryTVShowPayload{}
	if err := c.ShouldBindJSON(&watchHistoryPayload); err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to bind watch history body: %w: %w", helpers.BadRequestError, err))
		return
	}
	insertedEpisodeIDs, skippedEpisodeIDs, err :=
		model.CreateTVShowWatchHistory(userID, mediaSource, showID, watchHistoryPayload)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	response := AddWatchHistoryTVResponse{
		MediaSource:        mediaSource,
		InsertedEpisodeIDs: insertedEpisodeIDs,
	}
	if *skippedEpisodeIDs != nil && len(*skippedEpisodeIDs) > 0 {
		response.SkippedEpisodeIDs = skippedEpisodeIDs
	}
	helpers.SuccessResponse(c, response, 200)
}

// @Router /tv/{id}/history/delete [post]
// @Summary Delete TV Show Watch History
// @Tags Watch History
// @Accept json
// @Produce json
// @Param id path int true "Media ID" example(tmdb-1234)
// @Param body body DeleteWatchHistoryPayload true "Watch Event IDs to delete"
// @Success 200 {object} V1SuccessResponse{data=object}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func DeleteWatchHistoryTVHandler(c *gin.Context) {
	handleDeleteWatchHistory(c, database.RecordTypeTVShow)
}

// @Router /movie/{id}/history/delete [post]
// @Summary Delete Movie Watch History
// @Tags Watch History
// @Accept json
// @Produce json
// @Param id path int true "Media ID" example(tmdb-1234)
// @Param body body DeleteWatchHistoryPayload true "Watch Event IDs to delete"
// @Success 200 {object} V1SuccessResponse{data=object}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func DeleteWatchHistoryMovieHandler(c *gin.Context) {
	handleDeleteWatchHistory(c, database.RecordTypeMovie)
}

type DeleteWatchHistoryPayload struct {
	WatchEventIDs []int64 `json:"watch_event_ids" binding:"required"`
}

func handleDeleteWatchHistory(c *gin.Context, recordType string) {
	username := c.GetHeader("X-Username")
	if username == "" {
		helpers.ErrorResponse(c, fmt.Errorf("X-Username not found in header: %w", helpers.BadRequestError))
		return
	}
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("error getting user id for username %s: %w", username, err))
		return
	}
	// Only episode ids that belong to the same show should be inserted at the same time
	payload := DeleteWatchHistoryPayload{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to bind watch history body: %w: %w", helpers.BadRequestError, err))
		return
	}
	// get record id from source id
	mediaSource, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, err)
		return
	}
	has, record, err := database.GetMediaRecord(recordType, mediaSource, strconv.Itoa(sourceID))
	if !has || err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("error getting media record for %s-%s: %w", mediaSource, sourceID, err))
		return
	}
	if err := database.BatchDeleteWatchEvents(payload.WatchEventIDs, userID, int(record.RecordID)); err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("error deleting watch history records for %s-%s: %w", mediaSource, sourceID, err))
		return
	}
	helpers.SuccessResponse(c, nil, 200)
}

// @Router /v1/tv/{id}/history/rewatch [post]
// @Summary Create TV Show Rewatch
// @Description Create new rewatch for tv show. This archives the previous watches, so user's can start fresh.
// @Tags Watch History
// @Accept json
// @Produce json
// @Param id path int true "Media ID" example(tmdb-1234)
// @Success 200 {object} V1SuccessResponse{data=database.RewatchRecord}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func AddTVShowRewatchHandler(c *gin.Context) {
	username := c.GetHeader("X-Username")
	if username == "" {
		helpers.ErrorResponse(c, fmt.Errorf("X-Username not found in header: %w", helpers.BadRequestError))
		return
	}
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("error getting user id for username %s: %w", username, err))
		return
	}
	mediaSource, showID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, err)
		return
	}
	startedAt := time.Now().UTC()
	// for now, we don't support custom startedAt, evaluate in the future if this might be needed
	// supplying a body is optional
	// if c.Request.ContentLength != 0 {
	// 	type addRewatchPayload struct {
	// 		StartedAt string `json:"rewatch_started_at"`
	// 	}
	// 	rewatchPayload := addRewatchPayload{}
	// 	if err := c.ShouldBindJSON(&rewatchPayload); err != nil {
	// 		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to bind watch history body: "+c.Param("id")))
	// 		return
	// 	}
	// 	if rewatchPayload.StartedAt != "" {
	// 		parsed, err := time.Parse(time.RFC3339, rewatchPayload.StartedAt)
	// 		if err != nil {
	// 			helpers.ErrorResponseWithMessage(c, err, "Error parsing rewatch_started_at, must be RFC3339 string")
	// 			return
	// 		}
	// 		startedAt = parsed
	// 	}
	// }
	rewatchRecord, err := model.InsertRewatchFromSourceID(database.MediaTypeTVShow, mediaSource,
		strconv.Itoa(showID), userID, startedAt)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	helpers.SuccessResponse(c, rewatchRecord, 200)
}

type AddWatchHistoryMovieResponse struct {
	MediaSource      string `json:"media_source"`
	ActionType       string `json:"action_type"`
	InsertedSourceID *int   `json:"inserted_source_id"`
}

// @Router /v1/movie/{id}/history [post]
// @Summary Add Movie Watch History
// @Tags Watch History
// @Accept json
// @Produce json
// @Param id path int true "Media ID" example(tmdb-1234)
// @Param body body model.WatchHistoryMoviePayload true "Watch History Payload"
// @Success 200 {object} V1SuccessResponse{data=AddWatchHistoryMovieResponse}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func AddWatchHistoryMovieHandler(c *gin.Context) {
	username := c.GetHeader("X-Username")
	if username == "" {
		helpers.ErrorResponse(c, fmt.Errorf("X-Username not found in header: %w", helpers.BadRequestError))
		return
	}
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("error getting user id for username %s: %w", username, err))
		return
	}
	mediaSource, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, fmt.Errorf("error param %s: %w", c.Param("id"), err))
		return
	}
	watchHistoryPayload := model.WatchHistoryMoviePayload{}
	if err := c.ShouldBindJSON(&watchHistoryPayload); err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to bind watch history body: %w: %w", helpers.BadRequestError, err))
		return
	}
	insertedSourceID, err := model.CreateMovieWatchHistory(userID, mediaSource, sourceID, watchHistoryPayload)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	helpers.SuccessResponse(c, AddWatchHistoryMovieResponse{
		MediaSource:      mediaSource,
		ActionType:       strings.ToLower(watchHistoryPayload.ActionType),
		InsertedSourceID: insertedSourceID,
	}, 200)
}
