package v1

import (
	"errors"
	"hound/database"
	"hound/helpers"
	"hound/model"
	"hound/model/sources"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func SetPlaybackProgressHandler(c *gin.Context) {
	mediaType := database.MediaTypeMovie
	if strings.Contains(c.FullPath(), "/api/v1/tv/") {
		mediaType = database.MediaTypeTVShow
	} else if !strings.Contains(c.FullPath(), "/api/v1/movie/") {
		panic("Fatal error, invalid path for watch history")
	}
	username := c.GetHeader("X-Username")
	if username == "" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Username not found in header"))
		return
	}
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting user id for watch history"))
		return
	}
	mediaSource, sourceID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error parsing source_id: "+c.Param("id")))
		return
	}
	watchProgress := &model.WatchProgress{}
	if err := c.ShouldBindJSON(&watchProgress); err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error binding JSON for watch history"))
		return
	}
	if watchProgress.CurrentProgressSeconds < 300 {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Less than 5 minutes watched, skipping saving progress"))
		return
	}
	if watchProgress.TotalDurationSeconds < 60 {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"invalid param: Total duration is < 60 seconds, likely invalid file"))
		return
	}
	// if progress is > 85% of total duration or less than 5 minutes left, mark as watched
	setWatchCutoff := 0.85 * float64(watchProgress.TotalDurationSeconds)
	remainingSeconds := float64(watchProgress.TotalDurationSeconds) - float64(watchProgress.CurrentProgressSeconds)
	if float64(watchProgress.CurrentProgressSeconds) > setWatchCutoff || remainingSeconds < 300 {
		switch mediaType {
		case database.MediaTypeMovie:
			watchedAtString := time.Now().Format(time.RFC3339)
			watchHistoryPayload := model.WatchHistoryMoviePayload{
				ActionType: database.ActionScrobble,
				WatchedAt:  &watchedAtString,
			}
			_, err := model.CreateMovieWatchHistory(userID, mediaSource, sourceID, watchHistoryPayload)
			if err != nil {
				helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error creating watch history"))
				return
			}
			// delete watch progress
			_ = model.DeleteWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID), nil, nil, nil)
			helpers.SuccessResponse(c, gin.H{"status": "success", "watched": true}, 200)
			return
		case database.MediaTypeTVShow:
			if watchProgress.SeasonNumber == nil || watchProgress.EpisodeNumber == nil {
				helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
					"Invalid param: nil season_number or episode_number"))
				return
			}
			watchedAtString := time.Now().Format(time.RFC3339)
			// use season/episode pair instead of episode ids
			watchHistoryPayload := model.WatchHistoryTVShowPayload{
				EpisodeIDs:    nil,
				ActionType:    database.ActionScrobble,
				SeasonNumber:  watchProgress.SeasonNumber,
				EpisodeNumber: watchProgress.EpisodeNumber,
				RewatchID:     nil, // will autopopulate during creation
				WatchedAt:     &watchedAtString,
			}
			_, _, err = model.CreateTVShowWatchHistory(userID, mediaSource, sourceID, watchHistoryPayload)
			if err != nil {
				helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error creating watch history"))
				return
			}
			// delete watch progress
			_ = model.DeleteWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID),
				watchProgress.SeasonNumber, watchProgress.EpisodeNumber, nil)
			helpers.SuccessResponse(c, gin.H{"status": "success", "watched": true}, 200)
			return
		}
	}
	err = model.SetWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID), watchProgress)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error setting watch history"))
		return
	}
	helpers.SuccessResponse(c, gin.H{"status": "success", "watched": false}, 200)
}

func GetPlaybackProgressHandler(c *gin.Context) {
	mediaType := database.MediaTypeMovie
	if strings.Contains(c.FullPath(), "/api/v1/tv/") {
		mediaType = database.MediaTypeTVShow
	} else if !strings.Contains(c.FullPath(), "/api/v1/movie/") {
		panic("Fatal error, invalid path for watch history")
	}
	username := c.GetHeader("X-Username")
	if username == "" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Username not found in header"))
		return
	}
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting user id for watch history"))
		return
	}
	mediaSource, sourceID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error parsing source_id: "+c.Param("id")))
		return
	}
	if mediaType == database.MediaTypeMovie {
		watchProgress, err := model.GetWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID), nil)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting watch history"))
			return
		}
		if len(watchProgress) == 0 {
			helpers.SuccessResponse(c, gin.H{"status": "success", "data": nil}, 200)
			return
		}
		helpers.SuccessResponse(c, gin.H{"status": "success", "data": watchProgress[0]}, 200)
		return
	}
	// tv show case
	seasonNumber, err := strconv.Atoi(c.Param("seasonNumber"))
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error parsing season_number: "+c.Param("season_number")))
		return
	}
	watchProgress, err := model.GetWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID), &seasonNumber)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting watch history"))
		return
	}
	helpers.SuccessResponse(c, gin.H{"status": "success", "data": watchProgress}, 200)
}

func DeletePlaybackProgressHandler(c *gin.Context) {
	mediaType := database.MediaTypeMovie
	if strings.Contains(c.FullPath(), "/api/v1/tv/") {
		mediaType = database.MediaTypeTVShow
	} else if !strings.Contains(c.FullPath(), "/api/v1/movie/") {
		panic("Fatal error, invalid path for watch history")
	}
	username := c.GetHeader("X-Username")
	if username == "" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Username not found in header"))
		return
	}
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting user id for watch history"))
		return
	}
	mediaSource, sourceID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error parsing source_id: "+c.Param("id")))
		return
	}
	if mediaType == database.MediaTypeMovie {
		if err := model.DeleteWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID), nil, nil, nil); err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error deleting watch history"))
			return
		}
		helpers.SuccessResponse(c, gin.H{"status": "success"}, 200)
		return
	}
	// tv show case
	type deletePayload struct {
		SeasonNumber  *int `json:"season_number"`
		EpisodeNumber *int `json:"episode_number"`
	}
	var payload deletePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error binding JSON for watch history"))
		return
	}
	if err := model.DeleteWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID),
		payload.SeasonNumber, payload.EpisodeNumber, nil); err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error deleting some watch histories"))
		return
	}
	helpers.SuccessResponse(c, gin.H{"status": "success"}, 200)
}
