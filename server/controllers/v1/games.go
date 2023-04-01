package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"hound/helpers"
	"hound/model/sources"
	"strconv"
	"strings"
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
	id, err := strconv.Atoi(split[1])
	// only accept tmdb ids for now
	if err != nil || split[0] != "igdb" {
		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
		return
	}
	results, err := sources.GetGameFromIDIGDB(id)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	helpers.SuccessResponse(c, results, 200)
}