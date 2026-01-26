package model

import (
	"hound/database"
	"hound/helpers"
	"hound/sources"
	"hound/view"
	"strconv"

	tmdb "github.com/cyruzin/golang-tmdb"
)

func SearchMovies(queryString string) (*[]view.MediaRecordCatalog, error) {
	results, err := sources.SearchMoviesTMDB(queryString)
	if err != nil {
		return nil, err
	}
	// convert url results
	var convertedResults []view.MediaRecordCatalog
	for _, item := range results.Results {
		genreArray := sources.GetGenresMap(item.GenreIDs, database.MediaTypeMovie)
		resultObject := view.MediaRecordCatalog{
			RecordType:       database.MediaTypeMovie,
			MediaSource:      sources.MediaSourceTMDB,
			SourceID:         strconv.Itoa(int(item.ID)),
			OriginalTitle:    item.OriginalTitle,
			MediaTitle:       item.Title,
			VoteCount:        item.VoteCount,
			VoteAverage:      item.VoteAverage,
			ThumbnailURI:     helpers.GetTMDBImageURL(item.PosterPath, tmdb.W300),
			ReleaseDate:      item.ReleaseDate,
			Popularity:       item.Popularity,
			Genres:           genreArray,
			OriginalLanguage: item.OriginalLanguage,
			BackdropURI:      helpers.GetTMDBImageURL(item.BackdropPath, tmdb.Original),
			Overview:         item.Overview,
		}
		convertedResults = append(convertedResults, resultObject)
	}
	return &convertedResults, nil
}

func SearchTVShows(queryString string) (*[]view.MediaRecordCatalog, error) {
	results, err := sources.SearchTVShowTMDB(queryString)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Error searching for tv show")
		return nil, err
	}
	// convert url results
	var convertedResults []view.MediaRecordCatalog
	for _, item := range results.Results {
		genreArray := sources.GetGenresMap(item.GenreIDs, database.MediaTypeTVShow)
		resultObject := view.MediaRecordCatalog{
			MediaSource:      sources.MediaSourceTMDB,
			RecordType:       database.MediaTypeTVShow,
			SourceID:         strconv.Itoa(int(item.ID)),
			MediaTitle:       item.Name,
			OriginalTitle:    item.OriginalName,
			VoteCount:        item.VoteCount,
			VoteAverage:      item.VoteAverage,
			ThumbnailURI:     helpers.GetTMDBImageURL(item.PosterPath, tmdb.W300),
			ReleaseDate:      item.FirstAirDate,
			Popularity:       item.Popularity,
			Genres:           genreArray,
			OriginalLanguage: item.OriginalLanguage,
			BackdropURI:      helpers.GetTMDBImageURL(item.BackdropPath, tmdb.Original),
			Overview:         item.Overview,
			OriginCountry:    item.OriginCountry,
		}
		convertedResults = append(convertedResults, resultObject)
	}
	return &convertedResults, nil
}
