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

// @Router /api/v1/movie/{id}/continue_watching [get]
// @Summary Get Movie Next Watch Action
// @ID get-movie-next-watch-action
// @Tags Watch Activity
// @Accept json
// @Produce json
// @Param id path int true "Media ID"
// @Success 200 {object} V1SuccessResponse{data=model.WatchAction}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetMovieNextWatchActionHandler(c *gin.Context) {
	handleGetNextWatchAction(c)
}

// @Router /api/v1/tv/{id}/continue_watching [get]
// @Summary Get TV Show Next Watch Action
// @ID get-tvshow-next-watch-action
// @Tags Watch Activity
// @Accept json
// @Produce json
// @Param id path int true "Media ID"
// @Success 200 {object} V1SuccessResponse{data=model.WatchAction}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetTVNextWatchActionHandler(c *gin.Context) {
	handleGetNextWatchAction(c)
}

func handleGetNextWatchAction(c *gin.Context) {
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
	userID, err := getUserIDFromContext(c)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	// if no watch action, we don't want to return error
	// but ideally need to check if no watch action vs. internal error
	watchAction, _ := model.GetNextWatchAction(userID, mediaType, mediaSource, strconv.Itoa(sourceID))
	internal.SuccessResponse(c, watchAction, 200)
}

// @Router /api/v1/continue_watching [get]
// @Summary Get Continue Watching
// @ID get-continue-watching
// @Tags Watch Activity
// @Accept json
// @Produce json
// @Success 200 {object} V1SuccessResponse{data=[]model.WatchAction}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetContinueWatchingHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	watchActions, err := model.GetContinueWatching(userID)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get continue watching: %w", err))
		return
	}
	internal.SuccessResponse(c, watchActions, 200)
}
