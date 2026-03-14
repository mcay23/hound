package v1

import (
	"errors"
	"hound/database"
	"hound/helpers"
	"hound/model"
	"hound/sources"
	"hound/view"
	"strconv"
	"strings"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/gin-gonic/gin"
)

// @Router /v1/movie/search [get]
// @Summary Search Movies
// @Tags Movie, Search
// @Accept json
// @Produce json
// @Param query query string true "Search Query"
// @Success 200 {object} V1SuccessResponse{data=[]view.MediaRecordCatalog}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func SearchMoviesHandler(c *gin.Context) {
	queryString := c.Query("query")
	results, err := model.SearchMovies(queryString)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Error searching for tv show")
		helpers.ErrorResponse(c, err)
		return
	}
	helpers.SuccessResponse(c, results, 200)
}

// @Router /v1/movie/{id} [get]
// @Summary Get Movie Details
// @Tags Movie
// @Accept json
// @Produce json
// @Param id path string true "Media ID" example(tmdb-1234)
// @Success 200 {object} V1SuccessResponse{data=view.MediaRecordCatalog}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetMovieFromIDHandler(c *gin.Context) {
	mediaSource, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "request id param invalid"+err.Error()))
		return
	}
	movieDetails, err := sources.GetMovieFromIDTMDB(sourceID)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	genreArray := database.ConvertGenres(sources.MediaSourceTMDB, database.MediaTypeMovie, movieDetails.Genres)
	logoURI := ""
	if len(movieDetails.Images.Logos) > 0 {
		logoURI = helpers.GetTMDBImageURL(movieDetails.Images.Logos[0].FilePath, tmdb.W500)
	}
	movieObject := view.MediaRecordCatalog{
		MediaType:        database.RecordTypeMovie,
		MediaSource:      sources.MediaSourceTMDB,
		SourceID:         strconv.Itoa(int(sourceID)),
		MediaTitle:       movieDetails.Title,
		OriginalTitle:    movieDetails.OriginalTitle,
		Overview:         movieDetails.Overview,
		VoteCount:        movieDetails.VoteCount,
		VoteAverage:      movieDetails.VoteAverage,
		Popularity:       movieDetails.Popularity,
		ReleaseDate:      movieDetails.ReleaseDate,
		Duration:         movieDetails.Runtime,
		Status:           movieDetails.Status,
		Genres:           genreArray,
		OriginalLanguage: movieDetails.OriginalLanguage,
		ThumbnailURI:     helpers.GetTMDBImageURL(movieDetails.PosterPath, tmdb.W500),
		BackdropURI:      helpers.GetTMDBImageURL(movieDetails.BackdropPath, tmdb.Original),
		LogoURI:          logoURI,
		OriginCountry:    movieDetails.OriginCountry,
	}
	castArray := []view.Credit{}
	for _, cast := range movieDetails.Credits.MovieCredits.Cast {
		castArray = append(castArray, view.Credit{
			MediaSource:  sources.MediaSourceTMDB,
			SourceID:     strconv.Itoa(int(cast.ID)),
			CreditID:     cast.CreditID,
			Name:         cast.Name,
			OriginalName: cast.OriginalName,
			Character:    &cast.Character,
			ThumbnailURI: helpers.GetTMDBImageURL(cast.ProfilePath, tmdb.W500),
		})
	}
	movieObject.Cast = &castArray
	directorsArray := []view.Credit{}
	for _, crew := range movieDetails.Credits.MovieCredits.Crew {
		if strings.ToLower(crew.Job) == "director" {
			directorsArray = append(directorsArray, view.Credit{
				MediaSource:  sources.MediaSourceTMDB,
				SourceID:     strconv.Itoa(int(crew.ID)),
				CreditID:     crew.CreditID,
				Name:         crew.Name,
				OriginalName: crew.OriginalName,
				ThumbnailURI: helpers.GetTMDBImageURL(crew.ProfilePath, tmdb.W500),
				Job:          "Director",
			})
		}
	}
	movieObject.Creators = &directorsArray
	helpers.SuccessResponse(c, movieObject, 200)
}
