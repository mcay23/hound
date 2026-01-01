package v1

import (
	"errors"
	"hound/database"
	"hound/helpers"
	"hound/model"
	"hound/model/sources"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetNextWatchAction(c *gin.Context) {
	mediaType := ""
	path := c.FullPath()
	if strings.HasPrefix(path, "/api/v1/tv") {
		mediaType = database.MediaTypeTVShow
	} else if strings.HasPrefix(path, "/api/v1/movie") {
		mediaType = database.MediaTypeMovie
	}
	mediaSource, sourceID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "request id param invalid"+err.Error()))
		return
	}
	username := c.GetHeader("X-Username")
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid user"))
		return
	}
	watchAction, err := model.GetNextWatchAction(userID, mediaType, mediaSource, strconv.Itoa(sourceID))
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting next watch action"))
		return
	}
	helpers.SuccessResponse(c, gin.H{"status": "success", "watch_action": watchAction}, 200)
}
