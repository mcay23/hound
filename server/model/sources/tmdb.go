package sources

import (
	"encoding/json"
	"errors"
	tmdb "github.com/cyruzin/golang-tmdb"
	"hound/helpers"
	"hound/model/database"
	"os"
	"strconv"
	"time"
)

const (
	SourceTMDB string = "tmdb"
)

var tmdbClient *tmdb.Client
var tmdbTVGenres tmdb.GenreMovieList
var tmdbMovieGenres tmdb.GenreMovieList

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
}

/*
------------------------------
	TMDB TV SHOWS FUNCTIONS
------------------------------
 */

func GetTrendingTVShowsTMDB(page string) (*tmdb.Trending, error) {
	urlOptions := make(map[string]string)
	urlOptions["page"] = page
	shows, err := tmdbClient.GetTrending("tv", "week", urlOptions)
	if err != nil {
		return nil, err
	}
	return shows, nil
}

func SearchTVShowTMDB(query string) (*tmdb.SearchTVShowsResults, error) {
	shows, err := tmdbClient.GetSearchTVShow(query, nil)
	if err != nil {
		return nil, err
	}
	return shows.SearchTVShowsResults, nil
}

func GetTVShowFromIDTMDB(tmdbID int, options map[string]string) (*tmdb.TVDetails, error) {
	// TODO cache result
	tvShow, err := tmdbClient.GetTVDetails(tmdbID, options)
	if err != nil {
		return nil, helpers.LogErrorWithMessage(err, "Failed to get tv show details from tmdb")
	}
	return tvShow, nil
}

func GetTVShowIMDBID(tmdbID int, options map[string]string) (string, error) {
	externalIDs, err := tmdbClient.GetTVExternalIDs(tmdbID, options)
	if err != nil {
		return "", helpers.LogErrorWithMessage(err, "Failed to get tv show external ids from tmdb")
	}
	return externalIDs.IMDbID, nil
}

func GetTVSeasonTMDB(tmdbID int, seasonNumber int, options map[string]string) (*tmdb.TVSeasonDetails, error) {
	tvShow, err := tmdbClient.GetTVSeasonDetails(tmdbID, seasonNumber, options)
	if err != nil {
		return nil, helpers.LogErrorWithMessage(err, "Failed to get tv season details from tmdb")
	}
	return tvShow, nil
}

func AddTVShowToCollectionTMDB(username string, source string, sourceID int, collectionID *int64) error {
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		return err
	}
	if source != SourceTMDB {
		panic("Only tmdb source is allowed for now")
	}
	entry, err := GetLibraryObjectTMDB(database.MediaTypeTVShow, sourceID)
	if err != nil {
		return err
	}
	// insert record to internal library if not exists
	libraryID, err := database.AddRecordToInternalLibrary(entry)
	if err != nil {
		return err
	}
	// insert collection relation to collections table
	err = database.InsertCollectionRelation(userID, libraryID, collectionID)
	if err != nil {
		return err
	}
	return nil
}

func MarkTVSeasonAsWatchedTMDB(userID int64, libraryID int64, seasonNumber int, minEp int, maxEp int, date time.Time) error {
	var records []database.CommentRecord
	for i := minEp; i <= maxEp; i++ {
		tagData := "S" + strconv.Itoa(seasonNumber) + "E" + strconv.Itoa(i)
		records = append(records, database.CommentRecord{
			CommentType:  "history",
			UserID:       userID,
			LibraryID:    libraryID,
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
	urlOptions := make(map[string]string)
	urlOptions["page"] = page
	movies, err := tmdbClient.GetTrending("movie", "week", urlOptions)
	if err != nil {
		return nil, err
	}
	return movies, nil
}

func SearchMoviesTMDB(query string) (*tmdb.SearchMoviesResults, error) {
	shows, err := tmdbClient.GetSearchMovies(query, nil)
	if err != nil {
		return nil, err
	}
	return shows.SearchMoviesResults, nil
}

func GetMovieFromIDTMDB(tmdbID int, options map[string]string) (*tmdb.MovieDetails, error) {
	movie, err := tmdbClient.GetMovieDetails(tmdbID, options)
	if err != nil {
		return nil, helpers.LogErrorWithMessage(err, "Failed to get movie details from tmdb")
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
	entry, err := GetLibraryObjectTMDB(database.MediaTypeMovie, sourceID)
	if err != nil {
		return err
	}
	// insert record to internal library if not exists
	libraryID, err := database.AddRecordToInternalLibrary(entry)
	if err != nil {
		return err
	}
	// insert collection relation to collections table
	err = database.InsertCollectionRelation(userID, libraryID, collectionID)
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

func GetLibraryObjectTMDB(mediaType string, sourceID int) (*database.LibraryRecord, error) {
	var entry database.LibraryRecord
	if mediaType == database.MediaTypeTVShow {
		show, err := GetTVShowFromIDTMDB(sourceID, nil)
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
		temp := tmdb.GetImageURL(show.PosterPath, tmdb.W300)
		thumbnailURL := &temp
		if show.PosterPath == "" {
			thumbnailURL = nil
		}
		entry = database.LibraryRecord{
			MediaType:    database.MediaTypeTVShow,
			MediaSource:  SourceTMDB,
			SourceID:     strconv.Itoa(sourceID),
			MediaTitle:   show.Name,
			ReleaseDate:  show.FirstAirDate,
			Tags:         &tagsArray,
			Description:  []byte(show.Overview),
			FullData: 	  showJson,
			ThumbnailURL: thumbnailURL,
		}
		return &entry, nil
	} else if mediaType == database.MediaTypeMovie {
		movie, err := GetMovieFromIDTMDB(sourceID, nil)
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
		temp := tmdb.GetImageURL(movie.PosterPath, tmdb.W300)
		thumbnailURL := &temp
		if movie.PosterPath == "" {
			thumbnailURL = nil
		}
		entry = database.LibraryRecord{
			MediaType:    database.MediaTypeMovie,
			MediaSource:  SourceTMDB,
			SourceID:     strconv.Itoa(sourceID),
			MediaTitle:   movie.Title,
			ReleaseDate:  movie.ReleaseDate,
			Tags:         &tagsArray,
			Description:  []byte(movie.Overview),
			FullData: 	  movieJson,
			ThumbnailURL: thumbnailURL,
		}
		return &entry, nil
	}
	return nil, errors.New("invalid media type in call to GetLibraryObjectTMDB()")
}