package v1

import (
	"errors"
	"hound/database"
	"hound/helpers"
	"hound/model"
	"hound/sources"
	"hound/view"
	"strconv"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/gin-gonic/gin"
)

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

func GetMovieFromIDHandler(c *gin.Context) {
	mediaSource, sourceID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "request id param invalid"+err.Error()))
		return
	}
	movieDetails, err := sources.GetMovieFromIDTMDB(sourceID)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	// get profile, video urls
	if movieDetails.Credits.MovieCredits != nil {
		for num, item := range movieDetails.Credits.MovieCredits.Cast {
			movieDetails.Credits.MovieCredits.Cast[num].ProfilePath = helpers.GetTMDBImageURL(item.ProfilePath, tmdb.W500)
		}
		for num, item := range movieDetails.Credits.MovieCredits.Crew {
			movieDetails.Credits.MovieCredits.Crew[num].ProfilePath = helpers.GetTMDBImageURL(item.ProfilePath, tmdb.W500)
		}
	}
	logoURL := ""
	if len(movieDetails.Images.Logos) > 0 {
		logoURL = helpers.GetTMDBImageURL(movieDetails.Images.Logos[0].FilePath, tmdb.W500)
	}
	returnObject := view.MovieFullObject{
		MediaSource:         sources.MediaSourceTMDB,
		MediaType:           database.MediaTypeMovie,
		SourceID:            movieDetails.ID,
		MediaTitle:          movieDetails.Title,
		BackdropURL:         helpers.GetTMDBImageURL(movieDetails.BackdropPath, tmdb.Original),
		PosterURL:           helpers.GetTMDBImageURL(movieDetails.PosterPath, tmdb.W500),
		LogoURL:             logoURL,
		Budget:              movieDetails.Budget,
		Genres:              &movieDetails.Genres,
		Homepage:            movieDetails.Homepage,
		IMDbID:              movieDetails.IMDbID,
		OriginalLanguage:    movieDetails.OriginalLanguage,
		OriginalTitle:       movieDetails.OriginalTitle,
		Overview:            movieDetails.Overview,
		Popularity:          movieDetails.Popularity,
		ProductionCompanies: &movieDetails.ProductionCompanies,
		ReleaseDate:         movieDetails.ReleaseDate,
		Revenue:             movieDetails.Revenue,
		Runtime:             movieDetails.Runtime,
		Status:              movieDetails.Status,
		Tagline:             movieDetails.Tagline,
		VoteAverage:         movieDetails.VoteAverage,
		VoteCount:           movieDetails.VoteCount,
		MovieCredits:        movieDetails.Credits.MovieCredits,
		Videos:              movieDetails.MovieVideosAppend.Videos,
		Recommendations:     movieDetails.Recommendations,
		WatchProviders:      movieDetails.WatchProviders,
		ExternalIDs:         movieDetails.MovieExternalIDs,
	}
	_, record, err := database.GetMediaRecord(database.MediaTypeMovie, sources.MediaSourceTMDB, strconv.Itoa(int(movieDetails.ID)))
	if err == nil && record != nil {
		commentType := c.Query("type")
		comments, err := GetCommentsCore(c.GetHeader("X-Username"), record.RecordID, &commentType)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Error retrieving comments"))
			return
		}
		returnObject.Comments = comments
	}
	helpers.SuccessResponse(c, returnObject, 200)
}
