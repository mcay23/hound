package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/sources"
	"github.com/mcay23/hound/view"

	tmdb "github.com/cyruzin/golang-tmdb"
)

const (
	TMDBCatalogRefreshTime = 24 * time.Hour
)

func GetCatalog(catalogSource string, catalogID string, userID int64) ([]view.MediaRecordCatalog, error) {
	switch catalogSource {
	case database.CatalogSourceTMDB:
		cacheKey := fmt.Sprintf("tmdb_catalog|%s", catalogID)
		var cachedObject []view.MediaRecordCatalog
		found, _ := database.GetCache(cacheKey, &cachedObject)
		if found {
			return cachedObject, nil
		}
		result, err := GetTMDBCatalog(catalogID)
		if err != nil {
			return nil, err
		}
		database.SetCache(cacheKey, result, TMDBCatalogRefreshTime)
		return result, nil
	case database.CatalogSourceInternal:
		return GetInternalCatalog(userID, catalogID)
	default:
		return nil, fmt.Errorf("invalid catalog type: %s: %w", catalogSource, internal.BadRequestError)
	}
}

func GetTMDBCatalog(catalogID string) ([]view.MediaRecordCatalog, error) {
	switch catalogID {
	case "trending-shows":
		return getTrendingTVShows(1)
	case "trending-movies":
		return getTrendingMovies(1)
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

func GetInternalCatalog(userID int64, catalogID string) ([]view.MediaRecordCatalog, error) {
	switch catalogID {
	case "hound-library":
		return getHoundLibraryRecords(MaxItemsPerHomeRow, 0, "", nil)
	case "hound-library-shows":
		return getHoundLibraryRecords(MaxItemsPerHomeRow, 0, database.MediaTypeTVShow, nil)
	case "hound-library-movies":
		return getHoundLibraryRecords(MaxItemsPerHomeRow, 0, database.MediaTypeMovie, nil)
	case "hound-recent-collection":
		return getHoundRecentRecords(userID)
	default:
		return nil, fmt.Errorf("invalid catalog id: %s: %w", catalogID, internal.BadRequestError)
	}
}

func getTrendingTVShows(page int) ([]view.MediaRecordCatalog, error) {
	results, err := sources.GetTrendingTVShowsTMDB("1")
	if err != nil {
		return nil, fmt.Errorf("error getting popular tv shows: %w", err)
	}
	viewArray := []view.MediaRecordCatalog{}
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
	viewArray := []view.MediaRecordCatalog{}
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
	viewArray := []view.MediaRecordCatalog{}
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
	viewArray := []view.MediaRecordCatalog{}
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

func getHoundRecentRecords(userID int64) ([]view.MediaRecordCatalog, error) {
	records, err := database.GetRecentCollectionRecords(userID, MaxItemsPerHomeRow)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent collection records: %w: %w", internal.InternalServerError, err)
	}
	var viewArray []view.MediaRecordCatalog
	for _, item := range records {
		viewObject := CreateMediaRecordCatalogObject(item)
		viewArray = append(viewArray, viewObject)
	}
	return viewArray, nil
}

func getHoundLibraryRecords(limit int, offset int, mediaType string, genreIDs []int64) ([]view.MediaRecordCatalog, error) {
	records, _, err := database.GetDownloadedParentRecords("recent", limit, offset, mediaType, genreIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get downloaded records: %w", err)
	}
	var viewArray []view.MediaRecordCatalog
	for _, item := range records {
		viewObject := CreateMediaRecordCatalogObject(item)
		viewArray = append(viewArray, viewObject)
	}
	return viewArray, nil
}

func CreateMediaRecordCatalogObject(record database.MediaRecordGroup) view.MediaRecordCatalog {
	return view.MediaRecordCatalog{
		MediaType:        record.RecordType,
		MediaSource:      record.MediaSource,
		SourceID:         record.SourceID,
		MediaTitle:       record.MediaTitle,
		OriginalTitle:    record.OriginalTitle,
		Status:           record.Status,
		Overview:         record.Overview,
		Duration:         record.Duration,
		ReleaseDate:      record.ReleaseDate,
		LastAirDate:      record.LastAirDate,
		NextAirDate:      record.NextAirDate,
		SeasonNumber:     record.SeasonNumber,
		EpisodeNumber:    record.EpisodeNumber,
		ThumbnailURI:     record.ThumbnailURI,
		BackdropURI:      record.BackdropURI,
		Genres:           record.Genres,
		OriginalLanguage: record.OriginalLanguage,
		OriginCountry:    record.OriginCountry,
	}
}
