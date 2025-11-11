package sources

import (
	"encoding/json"
	"errors"
	"fmt"
	"hound/helpers"
	"hound/model"
	"hound/model/database"
	"log/slog"
	"os"
	"strconv"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
)

const (
	SourceTMDB string = "tmdb"
)

var tmdbClient *tmdb.Client
var tmdbTVGenres tmdb.GenreMovieList
var tmdbMovieGenres tmdb.GenreMovieList

const trendingCacheDuration = 12 * time.Hour
const searchCacheDuration = 24 * time.Hour
const getCacheDuration = 2 * time.Hour

type TVShowObject struct {
	TMDBData  *tmdb.SearchTVShowsResults
	PosterURL string
}

type GenreObject struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func InitializeTMDB() {
	var err error
	tmdbClient, err = tmdb.InitV4(os.Getenv("TMDB_API_KEY"))
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to initialize tmdb client")
		panic(err)
	}
	tmdbClient.SetClientAutoRetry()
	err = populateTMDBTVGenres()
	if err != nil {
		panic(err)
	}
	err = populateTMDBMovieGenres()
	if err != nil {
		panic(err)
	}
	slog.Info("TMDB Initialized")
}

/*
------------------------------
	TMDB TV SHOWS FUNCTIONS
------------------------------
*/

func GetTrendingTVShowsTMDB(page string) (*tmdb.Trending, error) {
	cacheKey := "tmdb|" + database.MediaTypeTVShow + "|trending|page:" + page
	var cacheObject tmdb.Trending
	cacheExists, _ := model.GetCache(cacheKey, &cacheObject)
	if cacheExists {
		return &cacheObject, nil
	}
	urlOptions := make(map[string]string)
	urlOptions["page"] = page
	shows, err := tmdbClient.GetTrending("tv", "week", urlOptions)
	if err != nil {
		return nil, err
	}
	if shows != nil {
		_, _ = model.SetCache(cacheKey, shows, trendingCacheDuration)
	}
	return shows, nil
}

func SearchTVShowTMDB(query string) (*tmdb.SearchTVShowsResults, error) {
	cacheKey := "tmdb|" + database.MediaTypeTVShow + "|search|query:" + query
	var cacheObject tmdb.SearchTVShowsResults
	cacheExists, _ := model.GetCache(cacheKey, &cacheObject)
	if cacheExists {
		return &cacheObject, nil
	}
	shows, err := tmdbClient.GetSearchTVShow(query, nil)
	if err != nil {
		return nil, err
	}
	if shows != nil {
		_, _ = model.SetCache(cacheKey, shows, searchCacheDuration)
	}
	return shows.SearchTVShowsResults, nil
}

func GetTVShowFromIDTMDB(tmdbID int) (*tmdb.TVDetails, error) {
	cacheKey := fmt.Sprintf("tmdb|%s|get|tmdb-%d", database.MediaTypeTVShow, tmdbID)
	var cacheObject tmdb.TVDetails
	cacheExists, _ := model.GetCache(cacheKey, &cacheObject)
	if cacheExists {
		return &cacheObject, nil
	}
	// for now, remove ability to control append_to_response, just cache the complete
	// response for safety
	options := map[string]string{
		"append_to_response": "videos,watch/providers,credits,recommendations,images,external_ids",
	}
	tvShow, err := tmdbClient.GetTVDetails(tmdbID, options)
	if err != nil {
		return nil, helpers.LogErrorWithMessage(err, "Failed to get tv show details from tmdb")
	}
	if tvShow != nil {
		_, _ = model.SetCache(cacheKey, tvShow, getCacheDuration)
	}
	return tvShow, nil
}

func GetTVShowIMDBID(tmdbID int) (string, error) {
	cacheKey := fmt.Sprintf("tmdb|%s|get|tmdb-%d", database.MediaTypeTVShow, tmdbID)
	var cacheObject tmdb.TVDetails
	cacheExists, _ := model.GetCache(cacheKey, &cacheObject)
	if cacheExists {
		return cacheObject.TVExternalIDs.IMDbID, nil
	}
	externalIDs, err := tmdbClient.GetTVExternalIDs(tmdbID, nil)
	if err != nil {
		return "", helpers.LogErrorWithMessage(err, "Failed to get tv show external ids from tmdb")
	}
	return externalIDs.IMDbID, nil
}

func GetTVSeasonTMDB(tmdbID int, seasonNumber int) (*tmdb.TVSeasonDetails, error) {
	cacheKey := fmt.Sprintf("tmdb|%s|season|tmdb-%d|S%d", database.MediaTypeTVShow, tmdbID, seasonNumber)
	var cacheObject tmdb.TVSeasonDetails
	cacheExists, _ := model.GetCache(cacheKey, &cacheObject)
	if cacheExists {
		return &cacheObject, nil
	}
	season, err := tmdbClient.GetTVSeasonDetails(tmdbID, seasonNumber, nil)
	if err != nil {
		return nil, helpers.LogErrorWithMessage(err, "Failed to get tv season details from tmdb")
	}
	if season != nil {
		_, _ = model.SetCache(cacheKey, season, getCacheDuration)
	}
	return season, nil
}

func AddTVShowToCollectionTMDB(username string, source string, sourceID int, collectionID *int64) error {
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		return err
	}
	if source != SourceTMDB {
		panic("Only tmdb source is allowed for now")
	}
	entry, err := GetRecordObjectTMDB(database.MediaTypeTVShow, sourceID)
	if err != nil {
		return err
	}
	// insert record to internal library if not exists
	recordID, err := database.AddMediaRecord(entry)
	if err != nil {
		return err
	}
	// insert collection relation to collections table
	err = database.InsertCollectionRelation(userID, recordID, collectionID)
	if err != nil {
		return err
	}
	return nil
}

func MarkTVSeasonAsWatchedTMDB(userID int64, recordID int64, seasonNumber int, minEp int, maxEp int, date time.Time) error {
	var records []database.CommentRecord
	for i := minEp; i <= maxEp; i++ {
		tagData := "S" + strconv.Itoa(seasonNumber) + "E" + strconv.Itoa(i)
		records = append(records, database.CommentRecord{
			CommentType:  "history",
			UserID:       userID,
			RecordID:     recordID,
			IsPrivate:    true,
			CommentTitle: "",
			Comment:      nil,
			TagData:      tagData,
			StartDate:    date,
			EndDate:      date,
		})
	}
	return database.AddCommentsBatch(&records)
}

/*
------------------------------
	TMDB MOVIES FUNCTIONS
------------------------------
*/

func GetTrendingMoviesTMDB(page string) (*tmdb.Trending, error) {
	cacheKey := "tmdb|" + database.MediaTypeMovie + "|trending|page:" + page
	var cacheObject tmdb.Trending
	cacheExists, _ := model.GetCache(cacheKey, &cacheObject)
	if cacheExists {
		return &cacheObject, nil
	}
	urlOptions := make(map[string]string)
	urlOptions["page"] = page
	movies, err := tmdbClient.GetTrending("movie", "week", urlOptions)
	if err != nil {
		return nil, err
	}
	if movies != nil {
		_, _ = model.SetCache(cacheKey, movies, trendingCacheDuration)
	}
	return movies, nil
}

func SearchMoviesTMDB(query string) (*tmdb.SearchMoviesResults, error) {
	cacheKey := "tmdb|" + database.MediaTypeMovie + "|search|query:" + query
	var cacheObject tmdb.SearchMoviesResults
	cacheExists, _ := model.GetCache(cacheKey, &cacheObject)
	if cacheExists {
		return &cacheObject, nil
	}
	movies, err := tmdbClient.GetSearchMovies(query, nil)
	if err != nil {
		return nil, err
	}
	if movies != nil {
		_, _ = model.SetCache(cacheKey, movies, searchCacheDuration)
	}
	return movies.SearchMoviesResults, nil
}

func GetMovieFromIDTMDB(tmdbID int) (*tmdb.MovieDetails, error) {
	cacheKey := fmt.Sprintf("tmdb|%s|get|tmdb-%d", database.MediaTypeMovie, tmdbID)
	var cacheObject tmdb.MovieDetails
	cacheExists, _ := model.GetCache(cacheKey, &cacheObject)
	if cacheExists {
		return &cacheObject, nil
	}
	options := map[string]string{
		"append_to_response": "videos,watch/providers,credits,recommendations,images,external_ids",
	}
	movie, err := tmdbClient.GetMovieDetails(tmdbID, options)
	if err != nil {
		return nil, helpers.LogErrorWithMessage(err, "Failed to get movie details from tmdb")
	}
	if movie != nil {
		_, _ = model.SetCache(cacheKey, movie, getCacheDuration)
	}
	return movie, nil
}

func AddMovieToCollectionTMDB(username string, source string, sourceID int, collectionID *int64) error {
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		return err
	}
	if source != SourceTMDB {
		panic("Only tmdb source is allowed for now")
	}
	entry, err := GetRecordObjectTMDB(database.MediaTypeMovie, sourceID)
	if err != nil {
		return err
	}
	// insert record to internal library if not exists
	recordID, err := database.AddMediaRecord(entry)
	if err != nil {
		return err
	}
	// insert collection relation to collections table
	err = database.InsertCollectionRelation(userID, recordID, collectionID)
	if err != nil {
		return err
	}
	return nil
}

/*
------------------------------
	HELPERS
------------------------------
*/

func populateTMDBTVGenres() error {
	list, err := tmdbClient.GetGenreTVList(nil)
	if err != nil {
		return helpers.LogErrorWithMessage(err, "Failed to populate genre list (tmdb)")
	}
	tmdbTVGenres = *list
	return nil
}

func populateTMDBMovieGenres() error {
	list, err := tmdbClient.GetGenreMovieList(nil)
	if err != nil {
		return helpers.LogErrorWithMessage(err, "Failed to populate genre list (tmdb)")
	}
	tmdbMovieGenres = *list
	return nil
}

func GetGenresMap(genreIds []int64, mediaType string) *[]GenreObject {
	var genreList tmdb.GenreMovieList
	if mediaType == database.MediaTypeTVShow {
		genreList = tmdbTVGenres
	} else if mediaType == database.MediaTypeMovie {
		genreList = tmdbMovieGenres
	} else {
		_ = helpers.LogErrorWithMessage(errors.New("invalid param: mediaType"),
			"Invalid media type supplied to tmdb.GetGenresMap()")
		return nil
	}
	var ret []GenreObject
	for _, id := range genreIds {
		genreName := ""
		for _, obj := range genreList.Genres {
			if id == obj.ID {
				genreName = obj.Name
			}
		}
		// could not find id in map, possible new tmdb genre made?
		if genreName == "" {
			_ = populateTMDBTVGenres()
			_ = populateTMDBMovieGenres()
			// retry again
			for _, obj := range genreList.Genres {
				if id == obj.ID {
					genreName = obj.Name
				}
			}
		}
		insert := GenreObject{
			ID:   id,
			Name: genreName,
		}
		ret = append(ret, insert)
	}
	return &ret
}

// Returns an array
// For shows, seasons are injected as well
func GetRecordObjectTMDB(mediaType string, sourceID int) (*[]database.MediaRecord, error) {
	var entry database.MediaRecord
	if mediaType == database.MediaTypeTVShow {
		show, err := GetTVShowFromIDTMDB(sourceID)
		if err != nil {
			return nil, err
		}
		showJson, err := json.Marshal(show)
		if err != nil {
			return nil, err
		}
		// import tmdb genres
		var tagsArray []database.TagObject
		for _, genre := range show.Genres {
			tagsArray = append(tagsArray, database.TagObject{
				TagID:   genre.ID,
				TagName: genre.Name,
			})
		}
		// weird but works
		thumbnailURL := tmdb.GetImageURL(show.PosterPath, tmdb.W300)
		if show.PosterPath == "" {
			thumbnailURL = ""
		}
		entry = database.MediaRecord{
			RecordType:       mediaType,
			MediaSource:      SourceTMDB,
			SourceID:         strconv.Itoa(sourceID),
			MediaTitle:       show.Name,
			OriginalTitle:    show.OriginalName,
			OriginalLanguage: show.OriginalLanguage,
			OriginCountry:    show.OriginCountry,
			ReleaseDate:      show.FirstAirDate,
			Tags:             &tagsArray,
			Overview:         show.Overview,
			FullData:         showJson,
			ThumbnailURL:     thumbnailURL,
		}
		return &entry, nil
	} else if mediaType == database.MediaTypeMovie {
		movie, err := GetMovieFromIDTMDB(sourceID)
		if err != nil {
			return nil, err
		}
		movieJson, err := json.Marshal(movie)
		if err != nil {
			return nil, err
		}
		// import tmdb genres
		var tagsArray []database.TagObject
		for _, genre := range movie.Genres {
			tagsArray = append(tagsArray, database.TagObject{
				TagID:   genre.ID,
				TagName: genre.Name,
			})
		}
		// weird but works
		thumbnailURL := tmdb.GetImageURL(movie.PosterPath, tmdb.W300)
		if movie.PosterPath == "" {
			thumbnailURL = ""
		}
		entry = database.MediaRecord{
			RecordType:   mediaType,
			MediaSource:  SourceTMDB,
			SourceID:     strconv.Itoa(sourceID),
			MediaTitle:   movie.Title,
			ReleaseDate:  movie.ReleaseDate,
			Tags:         &tagsArray,
			Overview:     movie.Overview,
			FullData:     movieJson,
			ThumbnailURL: thumbnailURL,
		}
		return &entry, nil
	}
	return nil, errors.New("invalid media type in call to GetLibraryObjectTMDB()")
}

// create a tmdb movie record to be inserted to the internal library
func CreateMovieRecordEntryTMDB(sourceID int) (*database.MediaRecord, error) {
	movie, err := GetMovieFromIDTMDB(sourceID)
	if err != nil {
		return nil, err
	}
	movieJson, err := json.Marshal(movie)
	if err != nil {
		return nil, err
	}
	// import tmdb genres
	var tagsArray []database.TagObject
	for _, genre := range movie.Genres {
		tagsArray = append(tagsArray, database.TagObject{
			TagID:   genre.ID,
			TagName: genre.Name,
		})
	}
	// parse image keys -> links
	posterURL := tmdb.GetImageURL(movie.PosterPath, tmdb.W300)
	if movie.PosterPath == "" {
		posterURL = ""
	}
	backdropURL := tmdb.GetImageURL(movie.BackdropPath, tmdb.W1280)
	if movie.BackdropPath == "" {
		backdropURL = ""
	}
	entry := database.MediaRecord{
		RecordType:       database.MediaTypeMovie,
		MediaSource:      SourceTMDB,
		SourceID:         strconv.Itoa(sourceID),
		ParentID:         -1, // movie is top level, has no parent
		MediaTitle:       movie.Title,
		OriginalTitle:    movie.OriginalTitle,
		OriginalLanguage: movie.OriginalLanguage,
		OriginCountry:    movie.OriginCountry,
		ReleaseDate:      movie.ReleaseDate,
		SeasonNumber:     -1,
		EpisodeNumber:    -1,
		SortIndex:        -1, // not used for movies
		Status:           movie.Status,
		Overview:         movie.Overview,
		Duration:         movie.Runtime,
		Tags:             &tagsArray,
		ThumbnailURL:     posterURL,
		BackdropURL:      backdropURL,
		StillURL:         "", // don't use stills for movies
		FullData:         movieJson,
	}
	return &entry, nil
}
