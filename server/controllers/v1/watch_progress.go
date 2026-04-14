package v1

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/model"
	"github.com/mcay23/hound/sources"

	"github.com/gin-gonic/gin"
)

type SetPlaybackProgressResponse struct {
	Watched bool `json:"watched" example:"false"`
}

// @Router /v1/movie/{id}/playback [post]
// @Summary Set Movie Playback Progress
// @ID set-movie-playback-progress
// @Tags Watch Progress
// @Accept json
// @Produce json
// @Param id path int true "Media ID" example(tmdb-1234)
// @Param body body model.WatchProgress true "Watch Progress Payload"
// @Success 200 {object} V1SuccessResponse{data=SetPlaybackProgressResponse}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func SetMoviePlaybackProgressHandler(c *gin.Context) {
	handleSetPlaybackProgress(c)
}

// @Router /v1/tv/{id}/playback [post]
// @Summary Set TV Playback Progress
// @ID set-tvshow-playback-progress
// @Tags Watch Progress
// @Accept json
// @Produce json
// @Param id path int true "Media ID" example(tmdb-1234)
// @Param body body model.WatchProgress true "Watch Progress Payload"
// @Success 200 {object} V1SuccessResponse{data=SetPlaybackProgressResponse}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func SetTVPlaybackProgressHandler(c *gin.Context) {
	handleSetPlaybackProgress(c)
}

func handleSetPlaybackProgress(c *gin.Context) {
	mediaType := database.MediaTypeMovie
	if strings.Contains(c.FullPath(), "/api/v1/tv/") {
		mediaType = database.MediaTypeTVShow
	} else if !strings.Contains(c.FullPath(), "/api/v1/movie/") {
		panic("Fatal error, invalid path for watch history")
	}
	userID, err := getUserIDFromContext(c)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	mediaSource, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		internal.ErrorResponse(c, fmt.Errorf("error parsing params %s: %w", c.Param("id"), err))
		return
	}
	watchProgress := &model.WatchProgress{}
	if err := c.ShouldBindJSON(&watchProgress); err != nil {
		internal.ErrorResponse(c, fmt.Errorf("error binding JSON for watch history: %w: %w", internal.BadRequestError, err))
		return
	}
	if watchProgress.TotalDurationSeconds < 60 {
		internal.ErrorResponse(c, fmt.Errorf("invalid param: total duration is < 60 seconds, likely invalid video: %w", internal.BadRequestError))
		return
	}
	if watchProgress.CurrentProgressSeconds < 120 {
		internal.ErrorResponse(c, fmt.Errorf("less than 2 minutes watched, skipping saving progress: %w", internal.BadRequestError))
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
				internal.ErrorResponse(c, internal.LogErrorWithMessage(err, "Error creating watch history"))
				return
			}
			// delete watch progress
			_ = model.DeleteWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID), nil, nil, nil)
			internal.SuccessResponse(c, SetPlaybackProgressResponse{Watched: true}, 200)
			return
		case database.MediaTypeTVShow:
			if watchProgress.SeasonNumber == nil || watchProgress.EpisodeNumber == nil {
				internal.ErrorResponse(c, fmt.Errorf("invalid param: nil season_number or episode_number: %w", internal.BadRequestError))
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
				internal.ErrorResponse(c, fmt.Errorf("error creating watch history: %w", err))
				return
			}
			// delete watch progress
			_ = model.DeleteWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID),
				watchProgress.SeasonNumber, watchProgress.EpisodeNumber, nil)
			internal.SuccessResponse(c, SetPlaybackProgressResponse{Watched: true}, 200)
			return
		}
	}
	// set client platform
	watchProgress.ClientPlatform = c.GetString("clientPlatform")
	// otherwise, continue to set watch progress
	err = model.SetWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID), watchProgress)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("error setting watch history: %w", err))
		return
	}
	internal.SuccessResponse(c, SetPlaybackProgressResponse{Watched: false}, 200)
}

// @Router /v1/movie/{id}/playback [get]
// @Summary Get Movie Playback Progress
// @ID get-movie-playback-progress
// @Tags Watch Progress
// @Accept json
// @Produce json
// @Param id path int true "Media ID" example(tmdb-1234)
// @Success 200 {object} V1SuccessResponse{data=model.WatchProgress}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetMoviePlaybackProgressHandler(c *gin.Context) {
	handleGetPlaybackProgress(c)
}

// @Router /v1/tv/{id}/season/{seasonNumber}/playback [get]
// @Summary Get TV Season Playback Progress
// @ID get-tvseason-playback-progress
// @Tags Watch Progress
// @Accept json
// @Produce json
// @Param id path int true "Media ID" example(tmdb-1234)
// @Param seasonNumber path int true "Season Number"
// @Success 200 {object} V1SuccessResponse{data=[]model.WatchProgress}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetTVSeasonPlaybackProgressHandler(c *gin.Context) {
	handleGetPlaybackProgress(c)
}

func handleGetPlaybackProgress(c *gin.Context) {
	mediaType := database.MediaTypeMovie
	if strings.Contains(c.FullPath(), "/api/v1/tv/") {
		mediaType = database.MediaTypeTVShow
	} else if !strings.Contains(c.FullPath(), "/api/v1/movie/") {
		panic("Fatal error, invalid path for watch history")
	}
	userID, err := getUserIDFromContext(c)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	mediaSource, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		internal.ErrorResponse(c, err)
		return
	}
	if mediaType == database.MediaTypeMovie {
		watchProgress, err := model.GetWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID), nil)
		if err != nil {
			internal.ErrorResponse(c, fmt.Errorf("error getting watch history: %w", err))
			return
		}
		if len(watchProgress) == 0 {
			internal.SuccessResponse(c, nil, 200)
			return
		}
		internal.SuccessResponse(c, watchProgress[0], 200)
		return
	}
	// tv show case
	seasonNumber, err := strconv.Atoi(c.Param("seasonNumber"))
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("error parsing season number %s: %w", c.Param("seasonNumber"), err))
		return
	}
	watchProgress, err := model.GetWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID), &seasonNumber)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("error getting watch history: %w", err))
		return
	}
	internal.SuccessResponse(c, watchProgress, 200)
}

type DeletePlaybackProgressPayload struct {
	SeasonNumber  *int `json:"season_number"`
	EpisodeNumber *int `json:"episode_number"`
}

// @Router /v1/movie/{id}/playback/delete [post]
// @Summary Delete Movie Playback Progress
// @ID delete-movie-playback-progress
// @Tags Watch Progress
// @Accept json
// @Produce json
// @Param id path int true "Media ID" example(tmdb-1234)
// @Success 200 {object} V1SuccessResponse{data=object}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func DeleteMoviePlaybackProgressHandler(c *gin.Context) {
	handleDeletePlaybackProgress(c)
}

// @Router /v1/tv/{id}/playback/delete [post]
// @Summary Delete TV Playback Progress
// @ID delete-tv-playback-progress
// @Tags Watch Progress
// @Accept json
// @Produce json
// @Param id path int true "Media ID" example(tmdb-1234)
// @Param body body DeletePlaybackProgressPayload false "Delete Payload (only for TV)"
// @Success 200 {object} V1SuccessResponse{data=object}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func DeleteTVPlaybackProgressHandler(c *gin.Context) {
	handleDeletePlaybackProgress(c)
}

func handleDeletePlaybackProgress(c *gin.Context) {
	mediaType := database.MediaTypeMovie
	if strings.Contains(c.FullPath(), "/api/v1/tv/") {
		mediaType = database.MediaTypeTVShow
	} else if !strings.Contains(c.FullPath(), "/api/v1/movie/") {
		panic("Fatal error, invalid path for watch history")
	}
	userID, err := getUserIDFromContext(c)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	mediaSource, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		internal.ErrorResponse(c, err)
		return
	}
	if mediaType == database.MediaTypeMovie {
		if err := model.DeleteWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID), nil, nil, nil); err != nil {
			internal.ErrorResponse(c, fmt.Errorf("error deleting watch history: %w", err))
			return
		}
		internal.SuccessResponse(c, nil, 200)
		return
	}
	// tv show case
	var payload DeletePlaybackProgressPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		internal.ErrorResponse(c, fmt.Errorf("error binding body for watch history: %w: %w", internal.BadRequestError, err))
		return
	}
	if err := model.DeleteWatchProgress(userID, mediaType, mediaSource, strconv.Itoa(sourceID),
		payload.SeasonNumber, payload.EpisodeNumber, nil); err != nil {
		internal.ErrorResponse(c, fmt.Errorf("error deleting watch history: %w", err))
		return
	}
	internal.SuccessResponse(c, nil, 200)
}
