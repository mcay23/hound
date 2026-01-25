package v1

import (
	"hound/database"
	"hound/helpers"
	"hound/sources"
	"hound/view"
	"strconv"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/gin-gonic/gin"
)

func GetTrendingTVShowsHandler(c *gin.Context) {
	// pagination locked for now
	results, err := sources.GetTrendingTVShowsTMDB("1")
	//results2, err := sources.GetTrendingTVShowsTMDB("2")
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error getting popular tv shows"))
		return
	}
	//results.Results = append(results.Results, results2.Results...)
	var viewArray []view.MediaCatalogObject
	for _, item := range results.Results {
		genreArray := sources.GetGenresMap(item.GenreIDs, database.MediaTypeTVShow)
		obj := view.MediaCatalogObject{
			MediaType:        database.MediaTypeTVShow,
			MediaSource:      sources.MediaSourceTMDB,
			SourceID:         strconv.Itoa(int(item.ID)),
			MediaTitle:       item.Name,
			OriginalName:     item.OriginalName,
			Overview:         item.Overview,
			VoteCount:        item.VoteCount,
			VoteAverage:      item.VoteAverage,
			Popularity:       item.Popularity,
			ThumbnailURL:     GetTMDBImageURL(item.PosterPath, tmdb.W300),
			BackdropURL:      GetTMDBImageURL(item.BackdropPath, tmdb.Original),
			ReleaseDate:      item.FirstAirDate,
			FirstAirDate:     item.FirstAirDate,
			Genres:           genreArray,
			OriginalLanguage: item.OriginalLanguage,
			OriginCountry:    item.OriginCountry,
		}
		viewArray = append(viewArray, obj)
	}
	helpers.SuccessResponse(c, viewArray, 200)
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
	var viewArray []view.MediaCatalogObject
	for _, item := range results.Results {
		genreArray := sources.GetGenresMap(item.GenreIDs, database.MediaTypeMovie)
		viewObject := view.MediaCatalogObject{
			MediaType:        database.MediaTypeMovie,
			MediaSource:      sources.MediaSourceTMDB,
			SourceID:         strconv.Itoa(int(item.ID)),
			MediaTitle:       item.Title,
			OriginalName:     item.OriginalTitle,
			Overview:         item.Overview,
			VoteCount:        item.VoteCount,
			VoteAverage:      item.VoteAverage,
			Popularity:       item.Popularity,
			ThumbnailURL:     GetTMDBImageURL(item.PosterPath, tmdb.W300),
			BackdropURL:      GetTMDBImageURL(item.BackdropPath, tmdb.Original),
			ReleaseDate:      item.ReleaseDate,
			FirstAirDate:     item.ReleaseDate,
			Genres:           genreArray,
			OriginalLanguage: item.OriginalLanguage,
			OriginCountry:    item.OriginCountry,
		}
		viewArray = append(viewArray, viewObject)
	}
	helpers.SuccessResponse(c, viewArray, 200)
}
