package model

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/sources"
	"github.com/mcay23/hound/view"

	tmdb "github.com/cyruzin/golang-tmdb"
)

func GetInternalCatalog(catalogID string, page *int) ([]view.MediaRecordCatalog, error) {
	switch catalogID {
	case "trending-shows":
		return getTrendingTVShows(*page)
	case "trending-movies":
		return getTrendingMovies(*page)
	case "netflix-movies":
		return getDiscoverMovies(sources.DiscoverTypeWatchProvider, sources.TMDBProviderNetflix)
	case "netflix-shows":
		return getDiscoverTVShows(sources.DiscoverTypeWatchProvider, sources.TMDBProviderNetflix)
	case "disney-plus-movies":
		return getDiscoverMovies(sources.DiscoverTypeWatchProvider, sources.TMDBProviderDisneyPlus)
	case "disney-plus-shows":
		return getDiscoverTVShows(sources.DiscoverTypeWatchProvider, sources.TMDBProviderDisneyPlus)
	case "hbo-max-movies":
		return getDiscoverMovies(sources.DiscoverTypeWatchProvider, sources.TMDBProviderHBOMax)
	case "hbo-max-shows":
		return getDiscoverTVShows(sources.DiscoverTypeWatchProvider, sources.TMDBProviderHBOMax)
	case "apple-tv-movies":
		return getDiscoverMovies(sources.DiscoverTypeWatchProvider, sources.TMDBProviderAppleTV)
	case "apple-tv-shows":
		return getDiscoverTVShows(sources.DiscoverTypeWatchProvider, sources.TMDBProviderAppleTV)
	case "amazon-prime-movies":
		return getDiscoverMovies(sources.DiscoverTypeWatchProvider, sources.TMDBProviderAmazonVideo)
	case "amazon-prime-shows":
		return getDiscoverTVShows(sources.DiscoverTypeWatchProvider, sources.TMDBProviderAmazonVideo)
	case "paramount-movies":
		return getDiscoverMovies(sources.DiscoverTypeWatchProvider, sources.TMDBProviderParamount)
	case "paramount-shows":
		return getDiscoverTVShows(sources.DiscoverTypeWatchProvider, sources.TMDBProviderParamount)
	default:
		// genre ids should be hound's internal ids, not tmdb's
		if strings.Contains(catalogID, "genre-movies") {
			parts := strings.Split(catalogID, "genre-movies-")
			if len(parts) < 2 {
				return nil, fmt.Errorf("invalid catalog id: %s: %w", catalogID, internal.BadRequestError)
			}
			genreID := parts[1]
			return getDiscoverMovies(sources.DiscoverTypeGenre, genreID)
		}
		if strings.Contains(catalogID, "genre-shows") {
			parts := strings.Split(catalogID, "genre-shows-")
			if len(parts) < 2 {
				return nil, fmt.Errorf("invalid catalog id: %s: %w", catalogID, internal.BadRequestError)
			}
			genreID := parts[1]
			return getDiscoverTVShows(sources.DiscoverTypeGenre, genreID)
		}
		return nil, fmt.Errorf("invalid catalog id: %s: %w", catalogID, internal.BadRequestError)
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
			ThumbnailURI:     internal.GetTMDBImageURL(item.PosterPath, tmdb.W300),
			BackdropURI:      internal.GetTMDBImageURL(item.BackdropPath, tmdb.Original),
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
			ThumbnailURI:     internal.GetTMDBImageURL(item.PosterPath, tmdb.W300),
			BackdropURI:      internal.GetTMDBImageURL(item.BackdropPath, tmdb.Original),
			ReleaseDate:      item.ReleaseDate,
			Genres:           genreArray,
			OriginalLanguage: item.OriginalLanguage,
			OriginCountry:    item.OriginCountry,
		}
		viewArray = append(viewArray, viewObject)
	}
	return viewArray, nil
}

func getDiscoverMovies(discoverType string, query string) ([]view.MediaRecordCatalog, error) {
	results, err := sources.TMDBMovieDiscover(discoverType, query)
	if err != nil {
		return nil, fmt.Errorf("error getting discover movies: %w", err)
	}
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
			ThumbnailURI:     internal.GetTMDBImageURL(item.PosterPath, tmdb.W300),
			BackdropURI:      internal.GetTMDBImageURL(item.BackdropPath, tmdb.Original),
			ReleaseDate:      item.ReleaseDate,
			Genres:           genreArray,
			OriginalLanguage: item.OriginalLanguage,
		}
		viewArray = append(viewArray, viewObject)
	}
	return viewArray, nil
}

func getDiscoverTVShows(discoverType string, query string) ([]view.MediaRecordCatalog, error) {
	results, err := sources.TMDBTVShowDiscover(discoverType, query)
	if err != nil {
		return nil, fmt.Errorf("error getting discover tv shows: %w", err)
	}
	var viewArray []view.MediaRecordCatalog
	for _, item := range results.Results {
		genreArray := sources.GetGenresMap(item.GenreIDs, database.MediaTypeTVShow)
		viewObject := view.MediaRecordCatalog{
			MediaType:        database.MediaTypeTVShow,
			MediaSource:      sources.MediaSourceTMDB,
			SourceID:         strconv.Itoa(int(item.ID)),
			MediaTitle:       item.Name,
			OriginalTitle:    item.OriginalName,
			Overview:         item.Overview,
			VoteCount:        item.VoteCount,
			VoteAverage:      item.VoteAverage,
			Popularity:       item.Popularity,
			ThumbnailURI:     internal.GetTMDBImageURL(item.PosterPath, tmdb.W300),
			BackdropURI:      internal.GetTMDBImageURL(item.BackdropPath, tmdb.Original),
			ReleaseDate:      item.FirstAirDate,
			Genres:           genreArray,
			OriginalLanguage: item.OriginalLanguage,
			OriginCountry:    item.OriginCountry,
		}
		viewArray = append(viewArray, viewObject)
	}
	return viewArray, nil
}
