package sources

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hound/database"
	"hound/helpers"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
)

const (
	MediaSourceTMDB string = "tmdb"
)

var tmdbClient *tmdb.Client
var tmdbTVGenres tmdb.GenreMovieList
var tmdbMovieGenres tmdb.GenreMovieList

const trendingCacheTTL = 12 * time.Hour
const searchCacheTTL = 24 * time.Hour
const getCacheTTL = 2 * time.Hour

// defined anonymously in tmdb, so we redefine
type TMDBEpisode struct {
	AirDate        string `json:"air_date"`
	EpisodeNumber  int    `json:"episode_number"`
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Overview       string `json:"overview"`
	ProductionCode string `json:"production_code"`
	Runtime        int    `json:"runtime"`
	SeasonNumber   int    `json:"season_number"`
	ShowID         int64  `json:"show_id"`
	StillPath      string `json:"still_path"`
}

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
	tmdbClient.SetClientConfig(http.Client{
		Timeout: time.Second * 30,
	})
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
	cacheExists, _ := database.GetCache(cacheKey, &cacheObject)
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
		_, _ = database.SetCache(cacheKey, shows, trendingCacheTTL)
	}
	return shows, nil
}

func SearchTVShowTMDB(query string) (*tmdb.SearchTVShowsResults, error) {
	cacheKey := "tmdb|" + database.MediaTypeTVShow + "|search|query:" + query
	var cacheObject tmdb.SearchTVShowsResults
	cacheExists, _ := database.GetCache(cacheKey, &cacheObject)
	if cacheExists {
		return &cacheObject, nil
	}
	shows, err := tmdbClient.GetSearchTVShow(query, nil)
	if err != nil {
		return nil, err
	}
	if shows != nil {
		_, _ = database.SetCache(cacheKey, shows, searchCacheTTL)
	}
	return shows.SearchTVShowsResults, nil
}

func GetTVShowFromIDTMDB(tmdbID int) (*tmdb.TVDetails, error) {
	cacheKey := fmt.Sprintf("tmdb|%s|get|tmdb-%d", database.MediaTypeTVShow, tmdbID)
	var cacheObject tmdb.TVDetails
	cacheExists, _ := database.GetCache(cacheKey, &cacheObject)
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
		_, _ = database.SetCache(cacheKey, tvShow, getCacheTTL)
	}
	return tvShow, nil
}

func GetTVShowIMDBID(tmdbID int) (string, error) {
	// just grab the tv show from cache, by default external_ids are appended
	cacheKey := fmt.Sprintf("tmdb|%s|get|tmdb-%d", database.MediaTypeTVShow, tmdbID)
	var cacheObject tmdb.TVDetails
	cacheExists, _ := database.GetCache(cacheKey, &cacheObject)
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
	cacheExists, _ := database.GetCache(cacheKey, &cacheObject)
	if cacheExists {
		return &cacheObject, nil
	}
	season, err := tmdbClient.GetTVSeasonDetails(tmdbID, seasonNumber, nil)
	if err != nil {
		return nil, helpers.LogErrorWithMessage(err, "Failed to get tv season details from tmdb")
	}
	if season == nil {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError),
			"failed to get tv season details from tmdb: season is nil")
	}
	_, _ = database.SetCache(cacheKey, season, getCacheTTL)
	return season, nil
}

func GetEpisodeTMDB(tmdbID int, seasonNumber int, episodeNumber int) (*TMDBEpisode, error) {
	// cached call, should be fast under normal flow
	season, err := GetTVSeasonTMDB(tmdbID, seasonNumber)
	if err != nil {
		return nil, helpers.LogErrorWithMessage(err, "failed to get season")
	}
	for _, episode := range season.Episodes {
		if episode.EpisodeNumber == episodeNumber {
			// tmdb package episode is anonymous struct, so we make our own
			tmdbEpisode := TMDBEpisode{
				AirDate:        episode.AirDate,
				EpisodeNumber:  episode.EpisodeNumber,
				ID:             episode.ID,
				Name:           episode.Name,
				Overview:       episode.Overview,
				ProductionCode: episode.ProductionCode,
				SeasonNumber:   episode.SeasonNumber,
				ShowID:         episode.ShowID,
				StillPath:      episode.StillPath,
			}
			return &tmdbEpisode, nil
		}
	}
	return nil, nil
}

func GetTVEpisodeGroupsTMDB(tmdbID int) (*tmdb.TVEpisodeGroups, error) {
	cacheKey := fmt.Sprintf("tmdb|%s|episode_groups|tmdb-%d", database.MediaTypeTVShow, tmdbID)
	var cacheObject tmdb.TVEpisodeGroups
	cacheExists, _ := database.GetCache(cacheKey, &cacheObject)
	if cacheExists {
		return &cacheObject, nil
	}
	episodeGroups, err := tmdbClient.GetTVEpisodeGroups(tmdbID, nil)
	if err != nil {
		return nil, err
	}
	if episodeGroups != nil {
		_, _ = database.SetCache(cacheKey, episodeGroups, getCacheTTL)
	}
	return episodeGroups, err
}

func GetTVEpisodeGroupsDetailsTMDB(tmdbEpisodeGroupID string) (*tmdb.TVEpisodeGroupsDetails, error) {
	cacheKey := fmt.Sprintf("tmdb|%s|episode_groups_details|tmdb-%s", database.MediaTypeTVShow, tmdbEpisodeGroupID)
	var cacheObject tmdb.TVEpisodeGroupsDetails
	cacheExists, _ := database.GetCache(cacheKey, &cacheObject)
	if cacheExists {
		return &cacheObject, nil
	}
	episodeGroupDetails, err := tmdbClient.GetTVEpisodeGroupsDetails(tmdbEpisodeGroupID, nil)
	if err != nil {
		return nil, err
	}
	if episodeGroupDetails != nil {
		_, _ = database.SetCache(cacheKey, episodeGroupDetails, getCacheTTL)
	}
	return episodeGroupDetails, err
}

func AddTVShowToCollectionTMDB(username string, source string, sourceID int, collectionID *int64) error {
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		return err
	}
	if source != MediaSourceTMDB {
		panic("Only tmdb source is allowed for now")
	}
	// this is quite expensive since by default all seasons and episodes are fetched and inserted
	// but upsert returns after inserting the first season, the rest are concurrently added
	record, err := UpsertTVShowRecordTMDB(sourceID)
	if err != nil {
		return err
	}
	// insert collection relation to collections table
	err = database.InsertCollectionRelation(userID, record.RecordID, collectionID)
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
	cacheExists, _ := database.GetCache(cacheKey, &cacheObject)
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
		_, _ = database.SetCache(cacheKey, movies, trendingCacheTTL)
	}
	return movies, nil
}

func SearchMoviesTMDB(query string) (*tmdb.SearchMoviesResults, error) {
	cacheKey := "tmdb|" + database.MediaTypeMovie + "|search|query:" + query
	var cacheObject tmdb.SearchMoviesResults
	cacheExists, _ := database.GetCache(cacheKey, &cacheObject)
	if cacheExists {
		return &cacheObject, nil
	}
	movies, err := tmdbClient.GetSearchMovies(query, nil)
	if err != nil {
		return nil, err
	}
	if movies != nil {
		_, _ = database.SetCache(cacheKey, movies, searchCacheTTL)
	}
	return movies.SearchMoviesResults, nil
}

func GetMovieFromIDTMDB(tmdbID int) (*tmdb.MovieDetails, error) {
	cacheKey := fmt.Sprintf("tmdb|%s|get|tmdb-%d", database.MediaTypeMovie, tmdbID)
	var cacheObject tmdb.MovieDetails
	cacheExists, _ := database.GetCache(cacheKey, &cacheObject)
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
		_, _ = database.SetCache(cacheKey, movie, getCacheTTL)
	}
	return movie, nil
}

func AddMovieToCollectionTMDB(username string, source string, sourceID int, collectionID *int64) error {
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		return err
	}
	if source != MediaSourceTMDB {
		panic("Only tmdb source is allowed for now")
	}
	entry, err := UpsertMovieRecordTMDB(sourceID)
	if err != nil {
		return err
	}
	// insert collection relation to collections table
	err = database.InsertCollectionRelation(userID, entry.RecordID, collectionID)
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

/*
Generate md5 hash from records
Used to compare newly fetched data->internal library to see if there are changes to update/insert
some flaws, credits/cast changes are not caught
in the future, if the functionality to duplicate/copy a movie/show so we can make local changes exist
update logic/hashing keys will need to change since this increases the risk of duplicate hashes
hash key changes will also trigger updates all relevant records when fetched, which is potentially expensive
additionalKey is appended at the end of the key before hashing, useful for season since its not specific enough
to detect changes
*/
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
		if record.SeasonNumber != nil {
			sb.WriteString(strconv.Itoa(*record.SeasonNumber))
		}
		sb.WriteString(record.Overview)
		sb.WriteString(record.ReleaseDate)
		sb.WriteString(record.ThumbnailURL)
		sb.WriteString(record.BackdropURL)
	case "episode":
		sb.WriteString(record.MediaSource)
		sb.WriteString(record.SourceID) // tmdb episodeid
		if record.EpisodeNumber != nil {
			sb.WriteString(strconv.Itoa(*record.EpisodeNumber))
		}
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

// simple helper function
func UpsertMediaRecordTMDB(mediaType string, sourceID int) (*database.MediaRecord, error) {
	switch mediaType {
	case database.MediaTypeMovie:
		return UpsertMovieRecordTMDB(sourceID)
	case database.MediaTypeTVShow:
		return UpsertTVShowRecordTMDB(sourceID)
	default:
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid media type")
	}
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
		MediaSource:      MediaSourceTMDB,
		SourceID:         strconv.Itoa(sourceID),
		ParentID:         nil, // movie is top level, has no parent
		MediaTitle:       movie.Title,
		OriginalTitle:    movie.OriginalTitle,
		OriginalLanguage: movie.OriginalLanguage,
		OriginCountry:    movie.OriginCountry,
		ReleaseDate:      movie.ReleaseDate,
		LastAirDate:      movie.ReleaseDate,
		NextAirDate:      movie.ReleaseDate,
		SeasonNumber:     nil,
		EpisodeNumber:    nil,
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
	err = database.UpsertMediaRecord(&entry)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// Triggers a full update attempt
// but quits early if hash matches
// first call/update is expensive since it fetches all seasons and episodes
func UpsertTVShowRecordTMDB(showSourceID int) (*database.MediaRecord, error) {
	// create show records
	showData, err := GetTVShowFromIDTMDB(showSourceID)
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
		MediaSource:      MediaSourceTMDB,
		SourceID:         strconv.Itoa(showSourceID),
		ParentID:         nil, // show is top level, has no parent
		MediaTitle:       showData.Name,
		OriginalTitle:    showData.OriginalName,
		OriginalLanguage: showData.OriginalLanguage,
		OriginCountry:    showData.OriginCountry,
		ReleaseDate:      showData.FirstAirDate,
		LastAirDate:      showData.LastAirDate,
		NextAirDate:      showData.NextEpisodeToAir.AirDate,
		SeasonNumber:     nil,
		EpisodeNumber:    nil,
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
	// include next/last episode data to hash
	// so refresh is more likely to trigger for new episodes
	tvShowEntry.ContentHash = hashRecordTMDB(tvShowEntry,
		showData.LastEpisodeToAir.AirDate+
			showData.LastEpisodeToAir.Name+
			showData.LastEpisodeToAir.Overview+
			showData.LastEpisodeToAir.StillPath+
			showData.NextEpisodeToAir.AirDate+
			showData.NextEpisodeToAir.Name+
			showData.NextEpisodeToAir.StillPath+
			showData.NextEpisodeToAir.Overview)
	// start session
	session := database.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return nil, helpers.LogErrorWithMessage(err,
			"UpsertTVShowRecordTMDB(): Failed to start xorm session")
	}
	// upsert the root level entry
	affected, err := database.UpsertMediaRecordsTrx(session, &tvShowEntry)
	if err != nil {
		return nil, err
	}
	// we get here since xorm.Update doesn't get recordID automatically
	has, showRecord, err := database.GetMediaRecordTrx(session, database.RecordTypeTVShow, MediaSourceTMDB,
		strconv.Itoa(showSourceID))
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"No Media Record Found for "+database.RecordTypeTVShow+":"+MediaSourceTMDB+"-"+strconv.Itoa(showSourceID))
	}
	// hash same, no update/insert
	if !affected {
		return showRecord, nil
	}
	// show hash changed, preload seasons to the cache
	_ = PrefetchSeasons(showSourceID)
	// batch insert all episodes later
	episodeRecords := []*database.MediaRecord{}
	for _, season := range showData.Seasons {
		// create season records
		seasonData, err := GetTVSeasonTMDB(showSourceID, season.SeasonNumber)
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
			MediaSource:      MediaSourceTMDB,
			SourceID:         strconv.Itoa(int(seasonData.ID)),
			ParentID:         &showRecord.RecordID, // record_id of the parent show
			MediaTitle:       seasonData.Name,
			OriginalTitle:    seasonData.Name,
			OriginalLanguage: showData.OriginalLanguage, // inherit from show, probably don't need to
			OriginCountry:    showData.OriginCountry,
			ReleaseDate:      seasonData.AirDate,
			LastAirDate:      "",
			NextAirDate:      "",
			SeasonNumber:     &seasonData.SeasonNumber,
			EpisodeNumber:    nil,
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
		seasonHash := hashRecordTMDB(seasonEntry, seasonHashKey)
		seasonEntry.ContentHash = seasonHash
		// upsert the season entry
		affected, err = database.UpsertMediaRecordsTrx(session, &seasonEntry)
		if err != nil {
			return nil, err
		}
		// skip if no change
		if !affected {
			continue
		}
		// get season so we know the parent ID
		has, seasonRecord, err := database.GetMediaRecordTrx(session, database.RecordTypeSeason, MediaSourceTMDB,
			strconv.Itoa(int(seasonData.ID)))
		if err != nil {
			return nil, err
		}
		if !has {
			return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"No Media Record Found for "+database.RecordTypeTVShow+":"+MediaSourceTMDB+"-"+strconv.Itoa(showSourceID))
		}
		if seasonRecord == nil || seasonRecord.ParentID == nil {
			return nil, fmt.Errorf("UpsertTVShowRecordTMDB(): season record is nil or has no parent id")
		}
		// upsert all children
		for _, episode := range seasonData.Episodes {
			stillURL := tmdb.GetImageURL(episode.StillPath, tmdb.W1280)
			if episode.StillPath == "" {
				stillURL = ""
			}
			seasonNum := seasonData.SeasonNumber
			episodeNum := episode.EpisodeNumber
			episodeEntry := database.MediaRecord{
				RecordType:       database.RecordTypeEpisode,
				MediaSource:      MediaSourceTMDB,
				SourceID:         strconv.Itoa(int(episode.ID)),
				ParentID:         &seasonRecord.RecordID, // record_id of the season
				MediaTitle:       episode.Name,
				OriginalTitle:    episode.Name,
				OriginalLanguage: showData.OriginalLanguage, // inherit from show, probably don't need to
				OriginCountry:    showData.OriginCountry,
				ReleaseDate:      episode.AirDate,
				LastAirDate:      "",
				NextAirDate:      "",
				SeasonNumber:     &seasonNum,
				EpisodeNumber:    &episodeNum,
				SortIndex:        episode.EpisodeNumber,
				Status:           "",
				Overview:         episode.Overview,
				Duration:         episode.Runtime, // not used in season
				ThumbnailURL:     "",
				BackdropURL:      "",
				StillURL:         stillURL,
				Tags:             nil,
				UserTags:         nil,
				AncestorID:       &showRecord.RecordID,
				FullData:         showJson,
			}
			episodeEntry.ContentHash = hashRecordTMDB(episodeEntry, "")
			episodeRecords = append(episodeRecords, &episodeEntry)
		}
	}
	err = database.BatchUpsertMediaRecords(session, episodeRecords)
	if err != nil {
		return nil, err
	}
	// only commit if everything succeeds
	session.Commit()
	return showRecord, nil
}

// prefetches all seasons for the show and stores it in the cache
// ideally, we would use tmdb append_to_response = season/1,season/2,...
// but seems like the tmdb go library doesn't marshalling this info yet
func PrefetchSeasons(sourceID int) error {
	// very likely cached, should be fine
	show, err := GetTVShowFromIDTMDB(sourceID)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	for _, season := range show.Seasons {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// auto caches
			_, _ = GetTVSeasonTMDB(sourceID, season.SeasonNumber)
		}()
	}
	wg.Wait()
	return nil
}
