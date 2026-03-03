package v1

import (
	"hound/database"
	"hound/helpers"

	"github.com/gin-gonic/gin"
)

func GetTVGenresHandler(c *gin.Context) {
	genres, err := database.GetGenresByType(database.MediaTypeTVShow)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to get TV genres"))
		return
	}
	helpers.SuccessResponse(c, genres, 200)
}

func GetMovieGenresHandler(c *gin.Context) {
	genres, err := database.GetGenresByType(database.MediaTypeMovie)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to get movie genres"))
		return
	}
	helpers.SuccessResponse(c, genres, 200)
}
