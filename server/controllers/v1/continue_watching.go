package v1

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/model"
	"github.com/mcay23/hound/sources"

	"github.com/gin-gonic/gin"
)

// @Router /api/v1/tv/{id}/continue_watching [get]
// @Router /api/v1/movie/{id}/continue_watching [get]
// @Summary Get next watch action for a media
// @Tags Watch Activity
// @Accept json
// @Produce json
// @Param id path int true "Media ID"
// @Success 200 {object} V1SuccessResponse{data=model.WatchAction}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetNextWatchActionHandler(c *gin.Context) {
	mediaType := ""
	path := c.FullPath()
	if strings.HasPrefix(path, "/api/v1/tv") {
		mediaType = database.MediaTypeTVShow
	} else if strings.HasPrefix(path, "/api/v1/movie") {
		mediaType = database.MediaTypeMovie
	}
	mediaSource, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		internal.ErrorResponse(c, fmt.Errorf("request id param invalid: %w: %w", internal.BadRequestError, err))
		return
	}
	username := c.GetHeader("X-Username")
	userID, err := database.GetUserIDFromUsername(username)
	// if no watch action, we don't want to return error
	// but ideally need to check if no watch action vs. internal error
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("invalid user: %w: %w", internal.BadRequestError, err))
		return
	}
	watchAction, _ := model.GetNextWatchAction(userID, mediaType, mediaSource, strconv.Itoa(sourceID))
	internal.SuccessResponse(c, watchAction, 200)
}

// @Router /api/v1/continue_watching [get]
// @Summary Get continue watching list
// @Tags Watch Activity
// @Accept json
// @Produce json
// @Success 200 {object} V1SuccessResponse{data=[]model.WatchAction}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetContinueWatchingHandler(c *gin.Context) {
	username := c.GetHeader("X-Username")
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("invalid user: %w: %w", internal.BadRequestError, err))
		return
	}
	watchActions, err := model.GetContinueWatching(userID)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get continue watching: %w", err))
		return
	}
	internal.SuccessResponse(c, watchActions, 200)
}
