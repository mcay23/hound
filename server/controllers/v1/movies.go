package v1

import (
	"errors"
	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/gin-gonic/gin"
	"hound/helpers"
	"hound/model/database"
	"hound/model/sources"
	"hound/view"
	"strconv"
	"strings"
)

func SearchMoviesHandler(c *gin.Context) {
	queryString := c.Query("query")
	results, err := SearchMoviesCore(queryString)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Error searching for tv show")
		helpers.ErrorResponse(c, err)
		return
	}
	helpers.SuccessResponse(c, results, 200)
}

func GetTrendingMoviesHandler(c *gin.Context) {
	// pagination locked for now
	results, err := sources.GetTrendingMoviesTMDB("1")
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Error getting popular tv shows")
		helpers.ErrorResponse(c, err)
		return
	}
	// convert url results
	var viewArray []view.LibraryItem
	for _, item := range results.Results {
		genreArray := sources.GetGenresMap(item.GenreIDs, database.MediaTypeMovie)
		thumbnailURL := GetTMDBImageURL(item.PosterPath, tmdb.W300)
		viewObject := view.LibraryItem{
			MediaType:    database.MediaTypeMovie,
			MediaSource:  sources.SourceTMDB,
			SourceID:     strconv.Itoa(int(item.ID)),
			MediaTitle:   item.OriginalTitle,
			ReleaseDate:  item.ReleaseDate,
			Description:  item.Overview,
			ThumbnailURL: &thumbnailURL,
			Tags:         genreArray,
			UserTags:     nil,
		}
		viewArray = append(viewArray, viewObject)
	}
	helpers.SuccessResponse(c, viewArray, 200)
}

func GetMovieFromIDHandler(c *gin.Context) {
	param := c.Param("id")
	split := strings.Split(param, "-")
	if len(split) != 2 {
		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
		return
	}
	id, err := strconv.ParseInt(split[1], 10, 64)
	// only accept tmdb ids for now
	if err != nil || split[0] != "tmdb" {
		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
		return
	}
	options := map[string]string{
		"append_to_response": "videos,watch/providers,credits,recommendations",
	}
	movieDetails, err := sources.GetMovieFromIDTMDB(int(id), options)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	// get profile, video urls
	for num, item := range movieDetails.Credits.MovieCredits.Cast {
		movieDetails.Credits.MovieCredits.Cast[num].ProfilePath = GetTMDBImageURL(item.ProfilePath, tmdb.W500)
	}
	for num, item := range movieDetails.Credits.MovieCredits.Crew {
		movieDetails.Credits.MovieCredits.Crew[num].ProfilePath = GetTMDBImageURL(item.ProfilePath, tmdb.W500)
	}
	returnObject := view.MovieFullObject{
		MediaSource:         sources.SourceTMDB,
		MediaType:           database.MediaTypeMovie,
		SourceID:            movieDetails.ID,
		MediaTitle:          movieDetails.Title,
		BackdropURL:         GetTMDBImageURL(movieDetails.BackdropPath, tmdb.Original),
		PosterURL:           GetTMDBImageURL(movieDetails.PosterPath, tmdb.W500),
		Budget:              movieDetails.Budget,
		Genres:              movieDetails.Genres,
		Homepage:            movieDetails.Homepage,
		IMDbID:              movieDetails.IMDbID,
		OriginalLanguage:    movieDetails.OriginalLanguage,
		OriginalTitle:       movieDetails.OriginalTitle,
		Overview:            movieDetails.Overview,
		Popularity:          movieDetails.Popularity,
		ProductionCompanies: movieDetails.ProductionCompanies,
		ReleaseDate:         movieDetails.ReleaseDate,
		Revenue:             movieDetails.Revenue,
		Runtime:             movieDetails.Runtime,
		Status:              movieDetails.Status,
		Tagline:             movieDetails.Tagline,
		VoteAverage:         movieDetails.VoteAverage,
		VoteCount:           movieDetails.VoteCount,
		MovieCredits:        movieDetails.Credits.MovieCredits,
		Videos:              movieDetails.Videos.MovieVideos,
		Recommendations:     movieDetails.Recommendations,
		WatchProviders:      movieDetails.WatchProviders,
	}
	helpers.SuccessResponse(c, returnObject, 200)
}

func SearchMoviesCore(queryString string) (*[]view.TMDBSearchResultObject, error) {
	results, err := sources.SearchMoviesTMDB(queryString)
	if err != nil {
		return nil, err
	}
	// convert url results
	var convertedResults []view.TMDBSearchResultObject
	for _, item := range results.Results {
		genreArray := sources.GetGenresMap(item.GenreIDs, database.MediaTypeTVShow)
		resultObject := view.TMDBSearchResultObject{
			MediaType:        database.MediaTypeMovie,
			MediaSource:      sources.SourceTMDB,
			OriginalName:     item.OriginalTitle,
			SourceID:         item.ID,
			MediaTitle:       item.Title,
			VoteCount:        item.VoteCount,
			VoteAverage:      item.VoteAverage,
			PosterURL:        GetTMDBImageURL(item.PosterPath, tmdb.W300),
			ReleaseDate:      item.ReleaseDate,
			Popularity:       item.Popularity,
			Genres:           genreArray,
			OriginalLanguage: item.OriginalLanguage,
			BackdropURL:      GetTMDBImageURL(item.BackdropPath, tmdb.Original),
			Overview:         item.Overview,
		}
		convertedResults = append(convertedResults, resultObject)
	}
	return &convertedResults, nil
}