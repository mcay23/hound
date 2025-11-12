package sources

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hound/helpers"
	"hound/model"
	"hound/model/database"
	"log/slog"
	"os"
	"strconv"
	"strings"
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
	// just grab the tv show from cache, by default external_ids are appended
	cacheKey := fmt.Sprintf("tmdb|%s|get|tmdb-%d", database.MediaTypeTVShow, tmdbID)
	var cacheObject tmdb.TVDetails
	cacheExists, _ := model.GetCache(cacheKey, &cacheObject)
	if cacheExists && cacheObject.TVExternalIDs.IMDbID != "" {
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
	recordID, err := database.UpsertMediaRecord(entry)
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
	recordID, err := database.UpsertMediaRecord(entry)
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

// Generate md5 hash from records
// Used to compare newly fetched data->internal library to see if there are changes to update/insert
// some flaws, credits/cast changes are not caught
// in the future, if the functionality to duplicate/copy a movie/show so we can make local changes exist
// update logic/hashing keys will need to change since this increases the risk of duplicate hashes
// hash key changes will also trigger updates all relevant records when fetched, which is potentially expensive
// additionalKey is appended at the end of the key before hashing, useful for season since its not specific enough
// to detect changes
func hashRecordTMDB(record database.MediaRecord, additionalKey string) string {
	var sb strings.Builder
	switch record.RecordType {
	case "movie":
		sb.WriteString(record.MediaSource)
		sb.WriteString(record.SourceID)
		sb.WriteString(record.MediaTitle)
		sb.WriteString(record.OriginalTitle)
		sb.WriteString(record.OriginalLanguage)
		sb.WriteString(record.ReleaseDate)
		sb.WriteString(record.Overview)
		sb.WriteString(string(record.Duration))
		sb.WriteString(record.ThumbnailURL)
		sb.WriteString(record.BackdropURL)
	case "tvshow":
		sb.WriteString(record.MediaSource)
		sb.WriteString(record.SourceID)
		sb.WriteString(record.MediaTitle)
		sb.WriteString(record.OriginalTitle)
		sb.WriteString(record.ReleaseDate)
		sb.WriteString(record.LastAirDate)
		sb.WriteString(record.NextAirDate)
		sb.WriteString(record.Status)
		sb.WriteString(record.Overview)
		sb.WriteString(record.ThumbnailURL)
		sb.WriteString(record.BackdropURL)
	case "season":
		sb.WriteString(record.MediaSource)
		sb.WriteString(record.SourceID) // tmdb seasonid
		sb.WriteString(string(record.SeasonNumber))
		sb.WriteString(record.Overview)
		sb.WriteString(record.ReleaseDate)
		sb.WriteString(record.ThumbnailURL)
		sb.WriteString(record.BackdropURL)
	case "episode":
		sb.WriteString(record.MediaSource)
		sb.WriteString(record.SourceID) // tmdb episodeid
		sb.WriteString(string(record.EpisodeNumber))
		sb.WriteString(record.MediaTitle) // episode title
		sb.WriteString(record.Overview)
		sb.WriteString(string(record.Duration))
		sb.WriteString(record.ReleaseDate) // air_date
		sb.WriteString(record.ThumbnailURL)
		sb.WriteString(record.StillURL)
	}
	hash := md5.Sum([]byte(sb.String() + additionalKey))
	return hex.EncodeToString(hash[:])
}

// create a tmdb movie record to be inserted to the internal library
func UpsertMovieRecordTMDB(sourceID int) (*database.MediaRecord, error) {
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
		RecordType:       database.RecordTypeMovie,
		MediaSource:      SourceTMDB,
		SourceID:         strconv.Itoa(sourceID),
		ParentID:         -1, // movie is top level, has no parent
		MediaTitle:       movie.Title,
		OriginalTitle:    movie.OriginalTitle,
		OriginalLanguage: movie.OriginalLanguage,
		OriginCountry:    movie.OriginCountry,
		ReleaseDate:      movie.ReleaseDate,
		LastAirDate:      movie.ReleaseDate,
		NextAirDate:      movie.ReleaseDate,
		SeasonNumber:     -1,
		EpisodeNumber:    -1,
		SortIndex:        -1, // not used for movies
		Status:           movie.Status,
		Overview:         movie.Overview,
		Duration:         movie.Runtime,
		ThumbnailURL:     posterURL,
		BackdropURL:      backdropURL,
		StillURL:         "", // don't use stills for movies
		Tags:             &tagsArray,
		UserTags:         nil,
		FullData:         movieJson,
	}
	entry.ContentHash = hashRecordTMDB(entry, "")
	return &entry, nil
}

/*
Create a tv show, season, and episode records
Eg. for a single show, this structure will be created
id: 1, parent_id: -1, "show", "game of thrones", ...
id: 2, parent_id: 1, "season", season_number: 1
id: 3, parent_id: 2, type: episode, episode_number: 1
id: 4, parent_id: 2, type: episode, episode_number: 2
id: 5, parent_id: 2, type: episode, episode_number: 3
*/
func UpsertTVShowRecordTMDB(sourceID int) ([]*database.MediaRecord, error) {
	showData, err := GetTVShowFromIDTMDB(sourceID)
	if err != nil {
		return nil, err
	}
	showJson, err := json.Marshal(showData)
	if err != nil {
		return nil, err
	}
	// import tmdb genres
	var tagsArray []database.TagObject
	for _, genre := range showData.Genres {
		tagsArray = append(tagsArray, database.TagObject{
			TagID:   genre.ID,
			TagName: genre.Name,
		})
	}
	posterURL := tmdb.GetImageURL(showData.PosterPath, tmdb.W300)
	if showData.PosterPath == "" {
		posterURL = ""
	}
	backdropURL := tmdb.GetImageURL(showData.BackdropPath, tmdb.W1280)
	if showData.BackdropPath == "" {
		backdropURL = ""
	}
	// construct show (parent)
	tvShowEntry := database.MediaRecord{
		RecordType:       database.RecordTypeTVShow,
		MediaSource:      SourceTMDB,
		SourceID:         strconv.Itoa(sourceID),
		ParentID:         -1, // show is top level, has no parent
		MediaTitle:       showData.Name,
		OriginalTitle:    showData.OriginalName,
		OriginalLanguage: showData.OriginalLanguage,
		OriginCountry:    showData.OriginCountry,
		ReleaseDate:      showData.FirstAirDate,
		LastAirDate:      showData.LastAirDate,
		NextAirDate:      showData.NextEpisodeToAir.AirDate,
		SeasonNumber:     -1,
		EpisodeNumber:    -1,
		SortIndex:        -1, // not used for shows
		Status:           showData.Status,
		Overview:         showData.Overview,
		Duration:         -1, // not used in tv show parent
		ThumbnailURL:     posterURL,
		BackdropURL:      backdropURL,
		StillURL:         "", // don't use stills for tv show parent
		Tags:             &tagsArray,
		UserTags:         nil,
		FullData:         showJson,
	}
	tvShowEntry.ContentHash = hashRecordTMDB(tvShowEntry, "")
	// create list of records to add
	recordEntries := []*database.MediaRecord{&tvShowEntry}
	// create record for each season
	for _, season := range showData.Seasons {
		// fetch each seasonData, season array from tv show doesn't append episodes
		seasonData, err := GetTVSeasonTMDB(sourceID, season.SeasonNumber)
		if err != nil {
			return nil, err
		}
		seasonJson, err := json.Marshal(seasonData)
		if err != nil {
			return nil, err
		}
		posterURL := tmdb.GetImageURL(seasonData.PosterPath, tmdb.W300)
		if showData.PosterPath == "" {
			posterURL = ""
		}
		seasonEntry := database.MediaRecord{
			RecordType:       database.RecordTypeSeason,
			MediaSource:      SourceTMDB,
			SourceID:         strconv.Itoa(int(seasonData.ID)),
			ParentID:         -1, // record_id of the parent show
			MediaTitle:       seasonData.Name,
			OriginalTitle:    seasonData.Name,
			OriginalLanguage: showData.OriginalLanguage, // inherit from show, probably don't need to
			OriginCountry:    showData.OriginCountry,
			ReleaseDate:      seasonData.AirDate,
			LastAirDate:      "",
			NextAirDate:      "",
			SeasonNumber:     seasonData.SeasonNumber,
			EpisodeNumber:    -1,
			SortIndex:        seasonData.SeasonNumber,
			Status:           "",
			Overview:         seasonData.Overview,
			Duration:         -1, // not used in season
			ThumbnailURL:     posterURL,
			BackdropURL:      "",
			StillURL:         "",  // don't use stills for season
			Tags:             nil, // just reuse
			UserTags:         nil,
			FullData:         seasonJson,
		}
		// add more hash info for seasons
		// number of episodes and latest air date should be sufficient
		seasonHashKey := ""
		if len(seasonData.Episodes) > 0 {
			seasonHashKey += strconv.Itoa(len(seasonData.Episodes))
			seasonHashKey += seasonData.Episodes[len(seasonData.Episodes)-1].AirDate
		}
		seasonHash := hashRecordTMDB(tvShowEntry, seasonHashKey)
		seasonEntry.ContentHash = seasonHash
		recordEntries = append(recordEntries, &seasonEntry)
		// create record for each episode in season
		for _, episode := range seasonData.Episodes {
			stillURL := tmdb.GetImageURL(episode.StillPath, tmdb.W1280)
			if episode.StillPath == "" {
				stillURL = ""
			}
			episodeEntry := database.MediaRecord{
				RecordType:       database.RecordTypeEpisode,
				MediaSource:      SourceTMDB,
				SourceID:         strconv.Itoa(int(episode.ID)),
				ParentID:         -1, // record_id of the parent season
				MediaTitle:       episode.Name,
				OriginalTitle:    episode.Name,
				OriginalLanguage: showData.OriginalLanguage, // inherit from show, probably don't need to
				OriginCountry:    showData.OriginCountry,
				ReleaseDate:      episode.AirDate,
				LastAirDate:      "",
				NextAirDate:      "",
				SeasonNumber:     seasonData.SeasonNumber,
				EpisodeNumber:    -1,
				SortIndex:        episode.SeasonNumber,
				Status:           "",
				Overview:         episode.Overview,
				Duration:         -1, // not used in season
				ThumbnailURL:     "",
				BackdropURL:      "",
				StillURL:         stillURL,
				Tags:             nil,
				UserTags:         nil,
				FullData:         showJson,
			}
			episodeEntry.ContentHash = hashRecordTMDB(episodeEntry, "")
			recordEntries = append(recordEntries, &episodeEntry)
		}
	}
	return recordEntries, nil
}
