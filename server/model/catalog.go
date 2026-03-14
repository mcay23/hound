package model

import (
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/sources"
	"hound/view"
	"strconv"

	tmdb "github.com/cyruzin/golang-tmdb"
)

func GetInternalCatalog(catalogID string, page *int) ([]view.MediaRecordCatalog, error) {
	switch catalogID {
	case "trending-shows":
		return getTrendingTVShows(*page)
	case "trending-movies":
		return getTrendingMovies(*page)
	default:
		return nil, fmt.Errorf("invalid catalog id: %s: %w", catalogID, helpers.BadRequestError)
	}
}

func getTrendingTVShows(page int) ([]view.MediaRecordCatalog, error) {
	results, err := sources.GetTrendingTVShowsTMDB("1")
	if err != nil {
		return nil, fmt.Errorf("error getting popular tv shows: %w", err)
	}
	var viewArray []view.MediaRecordCatalog
	for _, item := range results.Results {
		genreArray := sources.GetGenresMap(item.GenreIDs, database.MediaTypeTVShow)
		obj := view.MediaRecordCatalog{
			MediaType:        database.MediaTypeTVShow,
			MediaSource:      sources.MediaSourceTMDB,
			SourceID:         strconv.Itoa(int(item.ID)),
			MediaTitle:       item.Name,
			OriginalTitle:    item.OriginalName,
			Overview:         item.Overview,
			VoteCount:        item.VoteCount,
			VoteAverage:      item.VoteAverage,
			Popularity:       item.Popularity,
			ThumbnailURI:     helpers.GetTMDBImageURL(item.PosterPath, tmdb.W300),
			BackdropURI:      helpers.GetTMDBImageURL(item.BackdropPath, tmdb.Original),
			ReleaseDate:      item.FirstAirDate,
			Genres:           genreArray,
			OriginalLanguage: item.OriginalLanguage,
			OriginCountry:    item.OriginCountry,
		}
		viewArray = append(viewArray, obj)
	}
	return viewArray, nil
}

func getTrendingMovies(page int) ([]view.MediaRecordCatalog, error) {
	results, err := sources.GetTrendingMoviesTMDB("1")
	if err != nil {
		return nil, fmt.Errorf("error getting popular movies: %w", err)
	}
	// convert url results
	var viewArray []view.MediaRecordCatalog
	for _, item := range results.Results {
		genreArray := sources.GetGenresMap(item.GenreIDs, database.MediaTypeMovie)
		viewObject := view.MediaRecordCatalog{
			MediaType:        database.MediaTypeMovie,
			MediaSource:      sources.MediaSourceTMDB,
			SourceID:         strconv.Itoa(int(item.ID)),
			MediaTitle:       item.Title,
			OriginalTitle:    item.OriginalTitle,
			Overview:         item.Overview,
			VoteCount:        item.VoteCount,
			VoteAverage:      item.VoteAverage,
			Popularity:       item.Popularity,
			ThumbnailURI:     helpers.GetTMDBImageURL(item.PosterPath, tmdb.W300),
			BackdropURI:      helpers.GetTMDBImageURL(item.BackdropPath, tmdb.Original),
			ReleaseDate:      item.ReleaseDate,
			Genres:           genreArray,
			OriginalLanguage: item.OriginalLanguage,
			OriginCountry:    item.OriginCountry,
		}
		viewArray = append(viewArray, viewObject)
	}
	return viewArray, nil
}
