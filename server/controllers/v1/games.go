package v1

import (
	"errors"
	"hound/helpers"
	"hound/model/database"
	"hound/model/sources"
	"hound/view"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func SearchGamesHandler(c *gin.Context) {
	queryString := c.Query("query")
	results, err := sources.SearchGameIGDB(queryString)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to search for tv show")
		helpers.ErrorResponse(c, err)
		return
	}
	helpers.SuccessResponse(c, results, 200)
}

func GetGameFromIDHandler(c *gin.Context) {
	param := c.Param("id")
	split := strings.Split(param, "-")
	if len(split) != 2 {
		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
		return
	}
	sourceID, err := strconv.Atoi(split[1])
	// only accept tmdb ids for now
	if err != nil || split[0] != "igdb" {
		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
		return
	}
	result, err := sources.GetGameFromIDIGDB(sourceID)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	resultView := view.GameFullObject{
		IGDBGameObject: result,
	}
	recordID, err := database.GetRecordID(database.MediaTypeGame, sources.SourceIGDB, strconv.Itoa(sourceID))
	if err == nil {
		commentType := c.Query("type")
		comments, err := GetCommentsCore(c.GetHeader("X-Username"), *recordID, &commentType)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Error retrieving comments"))
			return
		}
		resultView.Comments = comments
	}
	helpers.SuccessResponse(c, resultView, 200)
}
