package v1

import (
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/model"
	"hound/sources"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type SetPlaybackProgressResponse struct {
	Watched bool `json:"watched" example:"false"`
}

// @Router /v1/movie/{id}/playback [post]
// @Router /v1/tv/{id}/playback [post]
// @Summary Set Playback Progress
// @Tags Watch Progress
// @Accept json
// @Produce json
// @Param id path int true "Media ID" example(tmdb-1234)
// @Param body body model.WatchProgress true "Watch Progress Payload"
// @Success 200 {object} V1SuccessResponse{data=SetPlaybackProgressResponse}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func SetPlaybackProgressHandler(c *gin.Context) {
	mediaType := database.MediaTypeMovie
	if strings.Contains(c.FullPath(), "/api/v1/tv/") {
		mediaType = database.MediaTypeTVShow
	} else if !strings.Contains(c.FullPath(), "/api/v1/movie/") {
		panic("Fatal error, invalid path for watch history")
	}
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
		helpers.ErrorResponse(c, fmt.Errorf("error parsing params %s: %w", c.Param("id"), err))
		return
	}
	watchProgress := &model.WatchProgress{}
	if err := c.ShouldBindJSON(&watchProgress); err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("error binding JSON for watch history: %w: %w", helpers.BadRequestError, err))
		return
	}
	if watchProgress.TotalDurationSeconds < 60 {
		helpers.ErrorResponse(c, fmt.Errorf("invalid param: total duration is < 60 seconds, likely invalid video: %w", helpers.BadRequestError))
		return
	}
	if watchProgress.CurrentProgressSeconds < 120 {
		helpers.ErrorResponse(c, fmt.Errorf("less than 2 minutes watched, skipping saving progress: %w", helpers.BadRequestError))
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
			helpers.SuccessResponse(c, SetPlaybackProgressResponse{Watched: true}, 200)
			return
		case database.MediaTypeTVShow:
			if watchProgress.SeasonNumber == nil || watchProgress.EpisodeNumber == nil {
				helpers.ErrorResponse(c, fmt.Errorf("invalid param: nil season_number or episode_number: %w", helpers.BadRequestError))
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
				helpers.ErrorResponse(c, fmt.Errorf("error creating watch history: %w", err))
				return
			}
			// delete watch progress
			_ = model.DeleteWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID),
				watchProgress.SeasonNumber, watchProgress.EpisodeNumber, nil)
			helpers.SuccessResponse(c, SetPlaybackProgressResponse{Watched: true}, 200)
			return
		}
	}
	// set client platform
	watchProgress.ClientPlatform = c.GetHeader("X-Client-Platform")
	// otherwise, continue to set watch progress
	err = model.SetWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID), watchProgress)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("error setting watch history: %w", err))
		return
	}
	helpers.SuccessResponse(c, SetPlaybackProgressResponse{Watched: false}, 200)
}

// @Router /v1/movie/{id}/playback [get]
// @Router /v1/tv/{id}/season/{seasonNumber}/playback [get]
// @Summary Get Playback Progress
// @Tags Watch Progress
// @Accept json
// @Produce json
// @Param id path int true "Media ID" example(tmdb-1234)
// @Param seasonNumber path int false "Season Number"
// @Success 200 {object} V1SuccessResponse{data=[]model.WatchProgress}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetPlaybackProgressHandler(c *gin.Context) {
	mediaType := database.MediaTypeMovie
	if strings.Contains(c.FullPath(), "/api/v1/tv/") {
		mediaType = database.MediaTypeTVShow
	} else if !strings.Contains(c.FullPath(), "/api/v1/movie/") {
		panic("Fatal error, invalid path for watch history")
	}
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
		helpers.ErrorResponse(c, err)
		return
	}
	if mediaType == database.MediaTypeMovie {
		watchProgress, err := model.GetWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID), nil)
		if err != nil {
			helpers.ErrorResponse(c, fmt.Errorf("error getting watch history: %w", err))
			return
		}
		if len(watchProgress) == 0 {
			helpers.SuccessResponse(c, nil, 200)
			return
		}
		helpers.SuccessResponse(c, watchProgress[0], 200)
		return
	}
	// tv show case
	seasonNumber, err := strconv.Atoi(c.Param("seasonNumber"))
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("error parsing season number %s: %w", c.Param("seasonNumber"), err))
		return
	}
	watchProgress, err := model.GetWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID), &seasonNumber)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("error getting watch history: %w", err))
		return
	}
	helpers.SuccessResponse(c, watchProgress, 200)
}

type DeletePlaybackProgressPayload struct {
	SeasonNumber  *int `json:"season_number"`
	EpisodeNumber *int `json:"episode_number"`
}

// @Router /v1/movie/{id}/playback/delete [post]
// @Router /v1/tv/{id}/playback/delete [post]
// @Summary Delete Playback Progress
// @Tags Watch Progress
// @Accept json
// @Produce json
// @Param id path int true "Media ID" example(tmdb-1234)
// @Param body body DeletePlaybackProgressPayload false "Delete Payload (only for TV)"
// @Success 200 {object} V1SuccessResponse{data=object}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func DeletePlaybackProgressHandler(c *gin.Context) {
	mediaType := database.MediaTypeMovie
	if strings.Contains(c.FullPath(), "/api/v1/tv/") {
		mediaType = database.MediaTypeTVShow
	} else if !strings.Contains(c.FullPath(), "/api/v1/movie/") {
		panic("Fatal error, invalid path for watch history")
	}
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
		helpers.ErrorResponse(c, err)
		return
	}
	if mediaType == database.MediaTypeMovie {
		if err := model.DeleteWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID), nil, nil, nil); err != nil {
			helpers.ErrorResponse(c, fmt.Errorf("error deleting watch history: %w", err))
			return
		}
		helpers.SuccessResponse(c, nil, 200)
		return
	}
	// tv show case
	var payload DeletePlaybackProgressPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("error binding body for watch history: %w: %w", helpers.BadRequestError, err))
		return
	}
	if err := model.DeleteWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID),
		payload.SeasonNumber, payload.EpisodeNumber, nil); err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("error deleting watch history: %w", err))
		return
	}
	helpers.SuccessResponse(c, nil, 200)
}
