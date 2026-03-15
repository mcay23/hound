package v1

import (
	"fmt"
	"hound/database"
	"hound/helpers"
	"time"

	"github.com/gin-gonic/gin"
)

// @Router /api/v1/watch_stats [get]
// @Summary Get Watch Stats
// @Tags Watch Activity
// @Accept json
// @Produce json
// @Param start_time query string false "Start Time RFC3339"
// @Param end_time query string false "End Time RFC3339"
// @Success 200 {object} V1SuccessResponse{data=database.WatchStats}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetWatchStatsHandler(c *gin.Context) {
	userID, err := database.GetUserIDFromUsername(c.GetHeader("X-Username"))
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("error getting user id for username %s: %w",
			c.GetHeader("X-Username"), err))
		return
	}
	var startTime, endTime *time.Time
	if c.Query("start_time") != "" {
		t, err := time.Parse(time.RFC3339, c.Query("start_time"))
		if err != nil {
			helpers.ErrorResponse(c, fmt.Errorf("error parsing start_time %s (must be RFC3999): %w: %w",
				c.Query("start_time"), helpers.BadRequestError, err))
			return
		}
		startTime = &t
	}
	if c.Query("end_time") != "" {
		t, err := time.Parse(time.RFC3339, c.Query("end_time"))
		if err != nil {
			helpers.ErrorResponse(c, fmt.Errorf("error parsing end_time %s (must be RFC3999): %w: %w",
				c.Query("end_time"), helpers.BadRequestError, err))
			return
		}
		endTime = &t
	}
	stats, err := database.GetWatchStats(userID, startTime, endTime)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("error getting watch stats: %w", err))
		return
	}
	helpers.SuccessResponse(c, stats, 200)
}
