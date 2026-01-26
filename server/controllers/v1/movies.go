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

func GetMovieFromIDHandlerV2(c *gin.Context) {
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
	genreArray := []database.GenreObject{}
	for _, genre := range movieDetails.Genres {
		genreArray = append(genreArray, database.GenreObject{
			ID:   genre.ID,
			Name: genre.Name,
		})
	}
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
	logoURI := ""
	if len(movieDetails.Images.Logos) > 0 {
		logoURI = helpers.GetTMDBImageURL(movieDetails.Images.Logos[0].FilePath, tmdb.W500)
	}
	returnObject := view.MovieFullObject{
		MediaSource:         sources.MediaSourceTMDB,
		MediaType:           database.MediaTypeMovie,
		SourceID:            movieDetails.ID,
		MediaTitle:          movieDetails.Title,
		BackdropURI:         helpers.GetTMDBImageURL(movieDetails.BackdropPath, tmdb.Original),
		ThumbnailURI:        helpers.GetTMDBImageURL(movieDetails.PosterPath, tmdb.W500),
		LogoURI:             logoURI,
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
