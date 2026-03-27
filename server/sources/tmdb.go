package sources

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/helpers"
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
var tmdbTVGenreInternalIDs = map[int64]int64{}
var tmdbMovieGenreInternalIDs = map[int64]int64{}

const trendingCacheTTL = 12 * time.Hour
const searchCacheTTL = 24 * time.Hour
const getCacheTTL = 30 * time.Minute

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

func InitializeTMDB() {
	var err error
	apiKey := os.Getenv("TMDB_API_KEY")
	// if user doesn't use their own api key, use the default one
	// as I understand, this is allowed by the devs
	// jellyfin, etc. uses a single api key for all their users
	// however, if rate-limiting were to be added to tmdb apis, user's
	// should set their own key
	if apiKey == "" {
		apiKey = "eyJhbGciOiJIUzI1NiJ9.eyJhdWQiOiJmMDZkMDdiOTk1ZmY2NjQxNjY0OWMzNjA4YzllMGE2NyIsIm5iZiI6MTYzMjc1Nzg2MS4zMDA5OTk5LCJzdWIiOiI2MTUxZTg2NTFjNjM1YjAwMmExMGNmNTciLCJzY29wZXMiOlsiYXBpX3JlYWQiXSwidmVyc2lvbiI6MX0.xTmUqmj38ElN1n0UWfsaL-1IJ46SAhCd1WtBD_Of_2A"
	}
	tmdbClient, err = tmdb.InitV4(apiKey)
	if err != nil {
		slog.Error("Failed to initialize tmdb client", "error", err)
		panic(err)
	}
	tmdbClient.SetClientAutoRetry()
	tmdbClient.SetClientConfig(http.Client{
		Timeout: time.Second * 30,
	})
	/*
		genres are loaded to memory at startup for fast access
		however, if server is running and tmdb adds a new genre,
		media_record upserts with the new genre will not be inserted
		this is a pretty niche case as genres seem rarely added, so we don't
		handle repopulating genres at runtime for simplicity
	*/
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
		return nil, fmt.Errorf("failed to get trending tv shows from tmdb for page %s: %w", page, err)
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
		return nil, fmt.Errorf("failed to search tv show from tmdb for query %s: %w", query, err)
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
		"append_to_response": "videos,watch/providers,credits,recommendations,images,external_ids,alternative_titles",
	}
	tvShow, err := tmdbClient.GetTVDetails(tmdbID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get tv show details from tmdb for source_id %d: %w", tmdbID, err)
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
		return "", fmt.Errorf("failed to get tv show external ids from tmdb for source_id %d: %w", tmdbID, err)
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
		return nil, fmt.Errorf("failed to get tv season details from tmdb for source_id %d, season_number %d: %w", tmdbID, seasonNumber, err)
	}
	if season == nil {
		return nil, fmt.Errorf("failed to get tv season details from tmdb: season is nil: %w", helpers.InternalServerError)
	}
	_, _ = database.SetCache(cacheKey, season, getCacheTTL)
	return season, nil
}

func GetEpisodeTMDB(tmdbID int, seasonNumber int, episodeNumber int) (*TMDBEpisode, error) {
	// cached call, should be fast under normal flow
	season, err := GetTVSeasonTMDB(tmdbID, seasonNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get season %d for show %d: %w", seasonNumber, tmdbID, err)
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
		return nil, fmt.Errorf("failed to get tv episode groups from tmdb for source_id %d: %w", tmdbID, err)
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
		return nil, fmt.Errorf("failed to get episode group details from tmdb for episode group id %s: %w", tmdbEpisodeGroupID, err)
	}
	if episodeGroupDetails != nil {
		_, _ = database.SetCache(cacheKey, episodeGroupDetails, getCacheTTL)
	}
	return episodeGroupDetails, err
}

func AddTVShowToCollectionTMDB(username string, source string, sourceID int, collectionID int64) error {
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
		return fmt.Errorf("failed to upsert tv show record: %w", err)
	}
	// insert collection relation to collections table
	err = database.InsertCollectionRelation(userID, record.RecordID, collectionID)
	if err != nil {
		return fmt.Errorf("failed to insert collection relation: %w", err)
	}
	return nil
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
		return nil, fmt.Errorf("failed to get trending movies from tmdb for page %s: %w", page, err)
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
		return nil, fmt.Errorf("failed to search movies from tmdb for query %s: %w", query, err)
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
		"append_to_response": "videos,watch/providers,credits,recommendations,images,external_ids,alternative_titles",
	}
	movie, err := tmdbClient.GetMovieDetails(tmdbID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get movie details from tmdb for id %d: %w", tmdbID, err)
	}
	if movie != nil {
		_, _ = database.SetCache(cacheKey, movie, getCacheTTL)
	}
	return movie, nil
}

func AddMovieToCollectionTMDB(username string, source string, sourceID int, collectionID int64) error {
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		return err
	}
	if source != MediaSourceTMDB {
		panic("Only tmdb source is allowed for now")
	}
	entry, err := UpsertMovieRecordTMDB(sourceID)
	if err != nil {
		return fmt.Errorf("failed to upsert movie record: %w", err)
	}
	// insert collection relation to collections table
	err = database.InsertCollectionRelation(userID, entry.RecordID, collectionID)
	if err != nil {
		return fmt.Errorf("failed to insert collection relation: %w", err)
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
		return fmt.Errorf("failed to populate genre list (tmdb): %w", err)
	}
	tmdbTVGenres = *list
	genreRecords := make([]database.GenreObject, 0, len(list.Genres))
	for _, genre := range list.Genres {
		genreRecords = append(genreRecords, database.GenreObject{
			SourceID:    genre.ID,
			Genre:       genre.Name,
			MediaType:   database.MediaTypeTVShow,
			MediaSource: MediaSourceTMDB,
		})
	}
	mapping, err := database.UpsertGenres(MediaSourceTMDB, database.MediaTypeTVShow, genreRecords)
	if err != nil {
		return fmt.Errorf("failed to sync tv genres to database: %w", err)
	}
	tmdbTVGenreInternalIDs = mapping
	return nil
}

func populateTMDBMovieGenres() error {
	list, err := tmdbClient.GetGenreMovieList(nil)
	if err != nil {
		return fmt.Errorf("failed to populate genre list (tmdb): %w", err)
	}
	tmdbMovieGenres = *list
	genreRecords := make([]database.GenreObject, 0, len(list.Genres))
	for _, genre := range list.Genres {
		genreRecords = append(genreRecords, database.GenreObject{
			SourceID:    genre.ID,
			Genre:       genre.Name,
			MediaType:   database.MediaTypeMovie,
			MediaSource: MediaSourceTMDB,
		})
	}
	mapping, err := database.UpsertGenres(MediaSourceTMDB, database.MediaTypeMovie, genreRecords)
	if err != nil {
		return fmt.Errorf("failed to sync movie genres to database: %w", err)
	}
	tmdbMovieGenreInternalIDs = mapping
	return nil
}

// checks if genres are missing from the memory mapping
// should be a rare case, but should be handled
func resolveTMDBGenreInternalIDs(mediaType string, genres []database.GenreObject) ([]int64, []int64, error) {
	if len(genres) == 0 {
		return nil, nil, nil
	}
	var src map[int64]int64
	switch mediaType {
	case database.MediaTypeTVShow:
		src = tmdbTVGenreInternalIDs
	case database.MediaTypeMovie:
		src = tmdbMovieGenreInternalIDs
	default:
		return nil, nil, fmt.Errorf("invalid media type for genre mapping: %w", helpers.BadRequestError)
	}
	ret := make([]int64, 0, len(genres))
	missing := make([]int64, 0)
	for _, genre := range genres {
		internalID, ok := src[genre.SourceID]
		if !ok {
			missing = append(missing, genre.SourceID)
			continue
		}
		ret = append(ret, internalID)
	}
	return ret, missing, nil
}

func GetGenresMap(genreIds []int64, mediaType string) []database.GenreObject {
	var ret []database.GenreObject
	for _, id := range genreIds {
		cached := database.GetGenreFromCache(MediaSourceTMDB, mediaType, id)
		// if genre is missing, skip it for now, don't want to handle refetching
		// since possible race conditions? This should be really rare
		if cached == nil {
			continue
			// missing genre, reload? (rare)
			// _ = populateTMDBTVGenres()
			// _ = populateTMDBMovieGenres()
			// cached = database.GetGenreFromCache(MediaSourceTMDB, mediaType, id)
		}
		ret = append(ret, database.GenreObject{
			GenreID:     cached.GenreID,
			Genre:       cached.Genre,
			MediaType:   cached.MediaType,
			MediaSource: cached.MediaSource,
			SourceID:    cached.SourceID,
		})
	}
	return ret
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
		sb.WriteString(fmt.Sprintf("%d", record.Duration))
		sb.WriteString(record.ThumbnailURI)
		sb.WriteString(record.BackdropURI)
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
		sb.WriteString(record.ThumbnailURI)
		sb.WriteString(record.BackdropURI)
	case "season":
		sb.WriteString(record.MediaSource)
		sb.WriteString(record.SourceID) // tmdb seasonid
		if record.SeasonNumber != nil {
			sb.WriteString(strconv.Itoa(*record.SeasonNumber))
		}
		sb.WriteString(record.Overview)
		sb.WriteString(record.ReleaseDate)
		sb.WriteString(record.ThumbnailURI)
		sb.WriteString(record.BackdropURI)
	case "episode":
		sb.WriteString(record.MediaSource)
		sb.WriteString(record.SourceID) // tmdb episodeid
		if record.EpisodeNumber != nil {
			sb.WriteString(strconv.Itoa(*record.EpisodeNumber))
		}
		sb.WriteString(record.MediaTitle) // episode title
		sb.WriteString(record.Overview)
		sb.WriteString(fmt.Sprintf("%d", record.Duration))
		sb.WriteString(record.ReleaseDate) // air_date
		sb.WriteString(record.ThumbnailURI)
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
		return nil, fmt.Errorf("invalid media_type %s: %w", mediaType, helpers.BadRequestError)
	}
}

// create a tmdb movie record to be inserted to the internal library
func UpsertMovieRecordTMDB(sourceID int) (*database.MediaRecord, error) {
	movie, err := GetMovieFromIDTMDB(sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get movie from tmdb: %w", err)
	}
	movieJson, err := json.Marshal(movie)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal movie: %w", err)
	}
	// import tmdb genres
	genreArray := database.ConvertGenres(MediaSourceTMDB, database.MediaTypeMovie, movie.Genres)
	// parse image keys -> links
	thumbnailURI := tmdb.GetImageURL(movie.PosterPath, tmdb.W300)
	if movie.PosterPath == "" {
		thumbnailURI = ""
	}
	backdropURI := tmdb.GetImageURL(movie.BackdropPath, tmdb.W1280)
	if movie.BackdropPath == "" {
		backdropURI = ""
	}
	logoURI := ""
	if len(movie.Images.Logos) > 0 {
		logoURI = tmdb.GetImageURL(movie.Images.Logos[0].FilePath, tmdb.W500)
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
		ThumbnailURI:     thumbnailURI,
		BackdropURI:      backdropURI,
		LogoURI:          logoURI,
		Genres:           genreArray,
		Tags:             nil,
		FullData:         movieJson,
	}
	entry.ContentHash = hashRecordTMDB(entry, "")
	session := database.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return nil, fmt.Errorf("failed to start xorm session: %w", err)
	}
	affected, err := database.UpsertMediaRecordsTrx(session, &entry)
	if err != nil {
		session.Rollback()
		return nil, fmt.Errorf("failed to upsert media record: %w", err)
	}
	has, record, err := database.GetMediaRecordTrx(session, database.RecordTypeMovie, MediaSourceTMDB, strconv.Itoa(sourceID))
	if err != nil {
		session.Rollback()
		return nil, fmt.Errorf("failed to get media record: %w", err)
	}
	if !has || record == nil {
		session.Rollback()
		return nil, fmt.Errorf("failed to get movie media record after upsert: %w", helpers.InternalServerError)
	}
	if affected {
		internalGenreIDs, missingGenreIDs, err := resolveTMDBGenreInternalIDs(database.MediaTypeMovie, genreArray)
		if err != nil {
			session.Rollback()
			return nil, fmt.Errorf("failed to resolve genre internal ids: %w", err)
		}
		if len(missingGenreIDs) > 0 {
			slog.Debug("Skipping unknown movie genre ids", "sourceID", sourceID, "missingIDs", missingGenreIDs)
		}
		if err := database.ReplaceMediaRecordGenresByIDsTrx(session, record.RecordID, internalGenreIDs); err != nil {
			session.Rollback()
			return nil, fmt.Errorf("failed to replace media record genres by ids: %w", err)
		}
	}
	if err := session.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit xorm session: %w", err)
	}
	return record, nil
}

// Triggers a full update attempt
// but quits early if hash matches
// first call/update is expensive since it fetches all seasons and episodes
func UpsertTVShowRecordTMDB(showSourceID int) (*database.MediaRecord, error) {
	// create show records
	showData, err := GetTVShowFromIDTMDB(showSourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tv show from tmdb: %w", err)
	}
	showJson, err := json.Marshal(showData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tv show: %w", err)
	}
	// import tmdb genres
	genreArray := database.ConvertGenres(MediaSourceTMDB, database.MediaTypeTVShow, showData.Genres)
	thumbnailURI := tmdb.GetImageURL(showData.PosterPath, tmdb.W300)
	if showData.PosterPath == "" {
		thumbnailURI = ""
	}
	backdropURI := tmdb.GetImageURL(showData.BackdropPath, tmdb.W1280)
	if showData.BackdropPath == "" {
		backdropURI = ""
	}
	logoURI := ""
	if len(showData.Images.Logos) > 0 {
		logoURI = tmdb.GetImageURL(showData.Images.Logos[0].FilePath, tmdb.W500)
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
		ThumbnailURI:     thumbnailURI,
		BackdropURI:      backdropURI,
		LogoURI:          logoURI,
		Genres:           genreArray,
		Tags:             nil,
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
		return nil, fmt.Errorf("failed to start xorm session: %w", err)
	}
	// upsert the root level entry
	affected, err := database.UpsertMediaRecordsTrx(session, &tvShowEntry)
	if err != nil {
		session.Rollback()
		return nil, fmt.Errorf("failed to upsert media records trx: %w", err)
	}
	// we get here since xorm.Update doesn't get recordID automatically
	has, showRecord, err := database.GetMediaRecordTrx(session, database.RecordTypeTVShow, MediaSourceTMDB,
		strconv.Itoa(showSourceID))
	if err != nil {
		session.Rollback()
		return nil, fmt.Errorf("failed to get media record trx: %w", err)
	}
	if !has {
		session.Rollback()
		return nil, fmt.Errorf("no media record found for record_type %s, media_source %s, source_id %d: %w", database.RecordTypeTVShow, MediaSourceTMDB, showSourceID, helpers.NotFoundError)
	}
	// hash same, no update/insert
	if !affected {
		session.Commit()
		return showRecord, nil
	}
	internalGenreIDs, missingGenreIDs, err := resolveTMDBGenreInternalIDs(database.MediaTypeTVShow, genreArray)
	if err != nil {
		session.Rollback()
		return nil, fmt.Errorf("failed to resolve genre internal ids: %w", err)
	}
	if len(missingGenreIDs) > 0 {
		slog.Debug("Skipping unknown tv genre ids", "sourceID", showSourceID, "missingIDs", missingGenreIDs)
	}
	if err := database.ReplaceMediaRecordGenresByIDsTrx(session, showRecord.RecordID, internalGenreIDs); err != nil {
		session.Rollback()
		return nil, fmt.Errorf("failed to replace media record genres by ids trx: %w", err)
	}
	// show hash changed, preload seasons to the cache
	_ = PrefetchSeasons(showSourceID)
	// batch insert all episodes later
	episodeRecords := []*database.MediaRecord{}
	for _, season := range showData.Seasons {
		// create season records
		seasonData, err := GetTVSeasonTMDB(showSourceID, season.SeasonNumber)
		if err != nil {
			session.Rollback()
			return nil, fmt.Errorf("failed to get tv season tmdb: %w", err)
		}
		seasonJson, err := json.Marshal(seasonData)
		if err != nil {
			session.Rollback()
			return nil, fmt.Errorf("failed to marshal tv season: %w", err)
		}
		thumbnailURI := tmdb.GetImageURL(seasonData.PosterPath, tmdb.W300)
		if showData.PosterPath == "" {
			thumbnailURI = ""
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
			ThumbnailURI:     thumbnailURI,
			BackdropURI:      "",
			Genres:           nil, // just reuse
			Tags:             nil,
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
			session.Rollback()
			return nil, fmt.Errorf("failed to upsert media records trx: %w", err)
		}
		// skip if no change
		if !affected {
			continue
		}
		// get season so we know the parent ID
		has, seasonRecord, err := database.GetMediaRecordTrx(session, database.RecordTypeSeason, MediaSourceTMDB,
			strconv.Itoa(int(seasonData.ID)))
		if err != nil {
			session.Rollback()
			return nil, fmt.Errorf("failed to get media record trx: %w", err)
		}
		if !has {
			session.Rollback()
			return nil, fmt.Errorf("no media record found for record_type %s, media_source %s, source_id %d: %w", database.RecordTypeSeason, MediaSourceTMDB, seasonData.ID, helpers.NotFoundError)
		}
		if seasonRecord == nil || seasonRecord.ParentID == nil {
			session.Rollback()
			return nil, fmt.Errorf("season record is nil or has no parent id: %w", helpers.InternalServerError)
		}
		// upsert all children
		for _, episode := range seasonData.Episodes {
			thumbnailURL := tmdb.GetImageURL(episode.StillPath, tmdb.W1280)
			if episode.StillPath == "" {
				thumbnailURL = ""
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
				ThumbnailURI:     thumbnailURL,
				BackdropURI:      "",
				Genres:           nil,
				Tags:             nil,
				AncestorID:       &showRecord.RecordID,
				FullData:         showJson,
			}
			episodeEntry.ContentHash = hashRecordTMDB(episodeEntry, "")
			episodeRecords = append(episodeRecords, &episodeEntry)
		}
	}
	err = database.BatchUpsertMediaRecords(session, episodeRecords)
	if err != nil {
		session.Rollback()
		return nil, fmt.Errorf("failed to batch upsert media records: %w", err)
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
