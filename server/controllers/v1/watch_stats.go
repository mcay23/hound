package v1

import (
	"hound/database"
	"hound/helpers"

	"github.com/gin-gonic/gin"
)

func GetWatchStatsHandler(c *gin.Context) {
	userID, err := database.GetUserIDFromUsername(c.GetHeader("X-Username"))
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid user"))
		return
	}
	// startTime, err := time.Parse("2006-01-02 15:04:05", c.Query("start_time"))
	// if err != nil {
	// 	helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid start time"))
	// 	return
	// }
	// finishTime, err := time.Parse("2006-01-02 15:04:05", c.Query("finish_time"))
	// if err != nil {
	// 	helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid finish time"))
	// 	return
	// }
	stats, err := database.GetWatchStats(userID, nil, nil)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to get watch stats"))
		return
	}
	helpers.SuccessResponse(c, gin.H{"status": "success", "data": stats}, 200)
}
