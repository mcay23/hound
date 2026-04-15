package v1

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/model"
	"github.com/mcay23/hound/sources"
	"github.com/mcay23/hound/view"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/gin-gonic/gin"
)

// @Router /v1/movie/search [get]
// @Summary Search Movies
// @ID search-movies
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
		internal.ErrorResponse(c, fmt.Errorf("failed to search for movies: %w", err))
		return
	}
	internal.SuccessResponse(c, results, 200)
}

// @Router /v1/movie/{id} [get]
// @Summary Get Movie Details
// @ID get-movie-details
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
		internal.ErrorResponse(c, fmt.Errorf("failed to get source id from params: %w: %w", internal.BadRequestError, err))
		return
	}
	movieDetails, err := sources.GetMovieFromIDTMDB(sourceID)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	genreArray := database.ConvertGenres(sources.MediaSourceTMDB, database.MediaTypeMovie, movieDetails.Genres)
	logoURI := ""
	if len(movieDetails.Images.Logos) > 0 {
		logoURI = internal.GetTMDBImageURL(movieDetails.Images.Logos[0].FilePath, tmdb.W500)
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
		ThumbnailURI:     internal.GetTMDBImageURL(movieDetails.PosterPath, tmdb.W500),
		BackdropURI:      internal.GetTMDBImageURL(movieDetails.BackdropPath, tmdb.Original),
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
			ThumbnailURI: internal.GetTMDBImageURL(cast.ProfilePath, tmdb.W500),
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
				ThumbnailURI: internal.GetTMDBImageURL(crew.ProfilePath, tmdb.W500),
				Job:          "Director",
			})
		}
	}
	movieObject.Creators = &directorsArray
	internal.SuccessResponse(c, movieObject, 200)
}
