package model

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
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

// used by recommendations, mdblist
type catalogLock struct {
	locks sync.Map
}

func (m *catalogLock) getLock(key string) *sync.Mutex {
	lock, _ := m.locks.LoadOrStore(key, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

// TMDB catalogs are cached, internal isn't since db data can change (eg. new download)
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
	case database.CatalogSourceMDBList:
		return GetMDBListCatalog(catalogID, MaxItemsPerHomeRow)
	default:
		return nil, fmt.Errorf("invalid catalog type: %s: %w", catalogSource, internal.BadRequestError)
	}
}

func GetTMDBCatalog(catalogID string) ([]view.MediaRecordCatalog, error) {
	switch catalogID {
	case "trending-movies":
		return getTrendingMovies()
	case "trending-shows":
		return getTrendingTVShows()
	}
	if catalog, ok := getTMDBCatalogDefinition(catalogID); ok {
		switch catalog.MediaType {
		case database.MediaTypeMovie:
			return getDiscoverMovies(catalog.DiscoverType, catalog.DiscoverQuery)
		case database.MediaTypeTVShow:
			return getDiscoverTVShows(catalog.DiscoverType, catalog.DiscoverQuery)
		}
	}
	return nil, fmt.Errorf("invalid catalog id: %s: %w", catalogID, internal.BadRequestError)
}

func GetInternalCatalog(userID int64, catalogID string) ([]view.MediaRecordCatalog, error) {
	// if these are updated, catalog_definitions.go should also be
	// TODO find a better way to define once
	switch catalogID {
	case "hound-library":
		return getHoundLibraryRecords(MaxItemsPerHomeRow, 0, "", nil)
	case "hound-library-shows":
		return getHoundLibraryRecords(MaxItemsPerHomeRow, 0, database.MediaTypeTVShow, nil)
	case "hound-library-movies":
		return getHoundLibraryRecords(MaxItemsPerHomeRow, 0, database.MediaTypeMovie, nil)
	case "hound-recent-collection":
		return getHoundRecentRecords(userID, MaxItemsPerHomeRow)
	case "hound-recommended-movies":
		return GetUserRecommendations(userID, database.MediaTypeMovie, MaxItemsPerHomeRow)
	case "hound-recommended-shows":
		return GetUserRecommendations(userID, database.MediaTypeTVShow, MaxItemsPerHomeRow)
	default:
		return nil, fmt.Errorf("invalid catalog id: %s: %w", catalogID, internal.BadRequestError)
	}
}

func getTrendingTVShows() ([]view.MediaRecordCatalog, error) {
	results, err := sources.GetTrendingTVShowsTMDB("1")
	if err != nil {
		return nil, fmt.Errorf("error getting popular tv shows: %w", err)
	}
	viewArray := []view.MediaRecordCatalog{}
	for _, item := range results.Results {
		genreArray := sources.GetGenresMapFromTMDBIDs(item.GenreIDs, database.MediaTypeTVShow)
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

func getTrendingMovies() ([]view.MediaRecordCatalog, error) {
	results, err := sources.GetTrendingMoviesTMDB("1")
	if err != nil {
		return nil, fmt.Errorf("error getting popular movies: %w", err)
	}
	// convert url results
	viewArray := []view.MediaRecordCatalog{}
	for _, item := range results.Results {
		genreArray := sources.GetGenresMapFromTMDBIDs(item.GenreIDs, database.MediaTypeMovie)
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
		genreArray := sources.GetGenresMapFromTMDBIDs(item.GenreIDs, database.MediaTypeMovie)
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
		genreArray := sources.GetGenresMapFromTMDBIDs(item.GenreIDs, database.MediaTypeTVShow)
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

func getHoundRecentRecords(userID int64, limit int) ([]view.MediaRecordCatalog, error) {
	records, err := database.GetRecentCollectionRecords(userID, limit)
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

var (
	mdbListCacheKey       = "mdb_list|%s|%d"
	mdbListLock           catalogLock
	mdbListRefreshingLock catalogLock
	mdbListStaleExpiry    = 14 * 24 * time.Hour
)

// This is network expensive, first call will be slow
func GetMDBListCatalog(listID string, limit int) ([]view.MediaRecordCatalog, error) {
	cacheKey := fmt.Sprintf(mdbListCacheKey, listID, limit)
	lock := mdbListLock.getLock(cacheKey)
	lock.Lock()
	defer lock.Unlock()
	var cachedObject []view.MediaRecordCatalog
	found, _ := database.GetCache(cacheKey, &cachedObject)
	if found {
		expiryKey := cacheKey + "|expiry"
		notExpired, _ := database.GetCache(expiryKey, nil)
		if notExpired {
			return cachedObject, nil
		}
		// only allow one refresh to run at a time
		refreshLock := mdbListRefreshingLock.getLock(cacheKey)
		if refreshLock.TryLock() {
			go func() {
				defer refreshLock.Unlock()
				_, err := getMDBListCatalogInternal(listID, limit)
				if err != nil {
					slog.Error("failed refreshing mdb list catalog", "listID", listID, "limit", limit, "err", err)
				}
			}()
		}
		return cachedObject, nil
	}
	return getMDBListCatalogInternal(listID, limit)
}

func getMDBListCatalogInternal(listID string, limit int) ([]view.MediaRecordCatalog, error) {
	// parse listID
	parts := strings.Split(listID, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid list id: %s: %w", listID, internal.BadRequestError)
	}
	results, err := sources.GetMDBList(parts[0], parts[1])
	if err != nil {
		return nil, fmt.Errorf("error getting mdb list: %w", err)
	}
	clamp := limit
	if len(results) < clamp {
		clamp = len(results)
	}
	results = results[:clamp]
	// get tmdb objects, necessary to grab backdrops
	// if an error is encountered during network calls, skip and continue
	// and don't save cache so it's fetched the next time
	viewArray := []view.MediaRecordCatalog{}
	fetchError := false
	for _, item := range results {
		// if an item doesn't have a tmdb id, create a generic object
		if item.IDs.TMDB == nil {
			viewArray = append(viewArray, *createGenericCatalogObjectfromMDBList(item))
			continue
		}
		// mdblist uses 'movie' and 'show' for media types
		if item.MediaType == database.MediaTypeMovie {
			details, err := sources.GetMovieFromIDTMDB(int(*item.IDs.TMDB))
			if err != nil || details == nil {
				fetchError = true
				viewArray = append(viewArray, *createGenericCatalogObjectfromMDBList(item))
				continue
			}
			viewArray = append(viewArray, *createTMDBMovieCatalogObject(details))
		} else if item.MediaType == "show" {
			details, err := sources.GetTVShowFromIDTMDB(int(*item.IDs.TMDB))
			if err != nil || details == nil {
				fetchError = true
				viewArray = append(viewArray, *createGenericCatalogObjectfromMDBList(item))
				continue
			}
			viewArray = append(viewArray, *createTMDBShowCatalogObject(details))
		} else {
			fetchError = true
		}
	}
	if !fetchError {
		// cache the results for 14 days, but mark for update in 24 hours
		cacheKey := fmt.Sprintf(mdbListCacheKey, listID, limit)
		database.SetCache(cacheKey, viewArray, mdbListStaleExpiry)
		database.SetCache(cacheKey+"|expiry", 1, HomeRowRefreshTime)
	}
	return viewArray, nil
}

// if fetching from tmdb fails, create a generic object. This loses genre, backdrop data
func createGenericCatalogObjectfromMDBList(item sources.MDBListItem) *view.MediaRecordCatalog {
	return &view.MediaRecordCatalog{
		MediaType:        item.MediaType,
		MediaSource:      sources.MediaSourceTMDB,
		SourceID:         strconv.Itoa(int(*item.IDs.TMDB)),
		MediaTitle:       item.Title,
		OriginalTitle:    item.Title,
		Overview:         item.Description,
		ThumbnailURI:     item.Poster,
		ReleaseDate:      item.ReleaseDate,
		OriginalLanguage: item.Language,
		OriginCountry:    []string{strings.ToUpper(item.Country)},
	}
}

func createTMDBMovieCatalogObject(item *tmdb.MovieDetails) *view.MediaRecordCatalog {
	if item == nil {
		return nil
	}
	genreIDs := []int64{}
	for _, genre := range item.Genres {
		genreIDs = append(genreIDs, genre.ID)
	}
	logoURI := ""
	if len(item.Images.Logos) > 0 {
		logoURI = internal.GetTMDBImageURL(item.Images.Logos[0].FilePath, tmdb.W500)
	}
	genreArray := sources.GetGenresMapFromTMDBIDs(genreIDs, database.MediaTypeMovie)
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
		LogoURI:          logoURI,
		ReleaseDate:      item.ReleaseDate,
		Duration:         item.Runtime,
		Genres:           genreArray,
		OriginalLanguage: item.OriginalLanguage,
		OriginCountry:    item.OriginCountry,
	}
	return &viewObject
}

func createTMDBShowCatalogObject(item *tmdb.TVDetails) *view.MediaRecordCatalog {
	if item == nil {
		return nil
	}
	genreIDs := []int64{}
	for _, genre := range item.Genres {
		genreIDs = append(genreIDs, genre.ID)
	}
	genreArray := sources.GetGenresMapFromTMDBIDs(genreIDs, database.MediaTypeTVShow)
	duration := 0
	if len(item.EpisodeRunTime) > 0 {
		duration = item.EpisodeRunTime[0]
	}
	logoURI := ""
	if len(item.Images.Logos) > 0 {
		logoURI = internal.GetTMDBImageURL(item.Images.Logos[0].FilePath, tmdb.W500)
	}
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
		LogoURI:          logoURI,
		ReleaseDate:      item.FirstAirDate,
		LastAirDate:      item.LastAirDate,
		Status:           item.Status,
		Duration:         duration,
		Genres:           genreArray,
		OriginalLanguage: item.OriginalLanguage,
		OriginCountry:    item.OriginCountry,
	}
	return &viewObject
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
