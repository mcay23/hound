package v1

import (
	"hound/database"
	"hound/helpers"

	"github.com/gin-gonic/gin"
)

// @Router /api/v1/tv/genres [get]
// @Summary Get TV Show Genres
// @Tags TV Shows, Genres
// @Accept json
// @Produce json
// @Success 200 {object} V1SuccessResponse{data=[]database.GenreRecord}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetTVGenresHandler(c *gin.Context) {
	genres, err := database.GetGenresByType(database.MediaTypeTVShow)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to get TV genres"))
		return
	}
	helpers.SuccessResponse(c, genres, 200)
}

// @Router /api/v1/movie/genres [get]
// @Summary Get Movie Genres
// @Tags TV Shows, Genres
// @Accept json
// @Produce json
// @Success 200 {object} V1SuccessResponse{data=[]database.GenreRecord}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetMovieGenresHandler(c *gin.Context) {
	genres, err := database.GetGenresByType(database.MediaTypeMovie)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to get movie genres"))
		return
	}
	helpers.SuccessResponse(c, genres, 200)
}
