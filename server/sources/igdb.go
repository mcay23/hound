package sources

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"hound/database"
	"hound/helpers"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	SourceIGDB              = "igdb"
	OAuthPath               = "https://id.twitch.tv/oauth2/token?client_id=%s&client_secret=%s&grant_type=client_credentials"
	IGDBGamesAPIPath        = "https://api.igdb.com/v4/games"
	IGDBImagePath           = "https://images.igdb.com/igdb/image/upload/t_%s/%s.jpg"
	YoutubePath             = "https://www.youtube.com/watch?v=%s"
	IGDBAccessTokenCacheKey = "IGDB-access-token"
)

const (
	IGDBImageCover    = "cover_big"
	IGDBImageOriginal = "original"
)

type IGDBAuthenticateResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type IGDBSearchObject struct {
	ID          int    `json:"id"`
	MediaTitle  string `json:"media_title"`
	MediaType   string `json:"media_type"`
	MediaSource string `json:"media_source"`
	SourceID    int    `json:"source_id"`
	PosterURL   string `json:"poster_url"`
	Cover       struct {
		ID       int    `json:"id"`
		ImageURL string `json:"image_url"`
		ImageID  string `json:"image_id"`
	} `json:"cover"`
	FirstReleaseDate int    `json:"first_release_date"`
	ReleaseDate      string `json:"release_date"`
	Genres           []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	Name      string `json:"name"`
	Platforms []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"platforms"`
}

type IGDBGameObject struct {
	ID          int    `json:"id"`
	MediaTitle  string `json:"media_title"`
	MediaType   string `json:"media_type"`
	MediaSource string `json:"media_source"`
	SourceID    int    `json:"source_id"`
	Artworks    []struct {
		ID       int    `json:"id"`
		Animated bool   `json:"animated"`
		Height   int    `json:"height"`
		ImageID  string `json:"image_id"`
		ImageURL string `json:"image_url"`
		Width    int    `json:"width"`
	} `json:"artworks"`
	PosterURL string `json:"poster_url"`
	Cover     struct {
		ID       int    `json:"id"`
		ImageID  string `json:"image_id"`
		ImageURL string `json:"image_url"`
	} `json:"cover"`
	DLCs []struct {
		ID    int `json:"id"`
		Cover struct {
			ID       int    `json:"id"`
			ImageID  string `json:"image_id"`
			ImageURL string `json:"image_url"`
		} `json:"cover"`
		Name         string `json:"name"`
		Summary      string `json:"summary"`
		ReleaseDates []struct {
			ID       int    `json:"id"`
			Date     int    `json:"date"`
			Human    string `json:"human"`
			Platform struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"platform"`
		} `json:"release_dates"`
	} `json:"dlcs"`
	FirstReleaseDate int    `json:"first_release_date"`
	ReleaseDate      string `json:"release_date"`
	GameModes        []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"game_modes"`
	Genres []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	InvolvedCompanies []struct {
		ID      int `json:"id"`
		Company struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"company"`
		Developer bool `json:"developer"`
		Publisher bool `json:"publisher"`
	} `json:"involved_companies"`
	Name      string `json:"name"`
	Platforms []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"platforms"`
	ReleaseDates []struct {
		ID       int    `json:"id"`
		Date     int    `json:"date"`
		Human    string `json:"human"`
		Platform struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"platform"`
	} `json:"release_dates"`
	Screenshots []struct {
		ID       int    `json:"id"`
		Height   int    `json:"height"`
		ImageID  string `json:"image_id"`
		ImageURL string `json:"image_url"`
		Width    int    `json:"width"`
	} `json:"screenshots"`
	SimilarGames []struct {
		ID    int `json:"id"`
		Cover struct {
			ID       int    `json:"id"`
			ImageID  string `json:"image_id"`
			ImageURL string `json:"image_url"`
		} `json:"cover"`
		Name string `json:"name"`
	} `json:"similar_games"`
	Storyline string `json:"storyline"`
	Summary   string `json:"summary"`
	Themes    []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"themes"`
	Videos []struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		VideoID string `json:"video_id"`
		Key     string `json:"key"`
	} `json:"videos"`
	Websites []struct {
		ID       int    `json:"id"`
		Category int    `json:"category"`
		URL      string `json:"url"`
	} `json:"websites"`
}

type IGDBGamesResultsResponseObject []IGDBGameObject

type IGDBSearchResultObject []IGDBSearchObject

var igdbClient = &http.Client{Timeout: 10 * time.Second}

func getAccessToken(forceRefresh bool) string {
	// get from ttl cache, return if it exists
	// if force refresh is true, refresh token
	var token string
	cacheExists, _ := database.GetCache(IGDBAccessTokenCacheKey, &token)
	if cacheExists && !forceRefresh {
		return token
	}
	// if token not in cache, request a new one
	url := fmt.Sprintf(OAuthPath, os.Getenv("IGDB_CLIENT_ID"),
		os.Getenv("IGDB_CLIENT_SECRET"))
	r, err := http.NewRequest("POST", url, nil)
	if err != nil {
		panic(err)
	}
	res, err := igdbClient.Do(r)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		resp := &IGDBAuthenticateResponse{}
		err := json.NewDecoder(res.Body).Decode(resp)
		if err != nil {
			panic(err)
		}
		// expire 1 min earlier
		_, err = database.SetCache(IGDBAccessTokenCacheKey, resp.AccessToken, time.Second*time.Duration(resp.ExpiresIn-60))
		if err != nil {
			panic(err)
		}
		return resp.AccessToken
	}
	panic(errors.New("non-200 response received from igdb oauth flow"))
}

func queryIGDBGames(body string) ([]byte, error) {
	r, err := http.NewRequest("POST", IGDBGamesAPIPath, bytes.NewBufferString(body))
	if err != nil {
		panic(err)
	}
	r.Header.Set("Client-ID", os.Getenv("IGDB_CLIENT_ID"))
	// call getAccessToken
	r.Header.Set("Authorization", "Bearer "+getAccessToken(false))
	res, err := igdbClient.Do(r)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New("panic querying igdb")
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	// let functions unmarshal themselves
	return b, nil
}

func SearchGameIGDB(query string) (IGDBSearchResultObject, error) {
	// construct query string
	requestBody := `search "` + query + `"; fields name, platforms.name, cover.image_id, status, genres.name, first_release_date; limit 20; where version_parent = null & category=(0,3,4,8,9,10);`
	b, err := queryIGDBGames(requestBody)
	if err != nil {
		return nil, err
	}
	// unmarshal response
	var games IGDBSearchResultObject
	err = json.Unmarshal(b, &games)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	// get image urls, set response params
	for num, game := range games {
		games[num].MediaTitle = game.Name
		games[num].MediaType = database.MediaTypeGame
		games[num].MediaSource = SourceIGDB
		games[num].SourceID = game.ID
		games[num].Cover.ImageURL = getIGDBImageURL(game.Cover.ImageID, IGDBImageCover)
		games[num].PosterURL = getIGDBImageURL(game.Cover.ImageID, IGDBImageCover)
		if game.FirstReleaseDate != 0 {
			games[num].ReleaseDate = time.Unix(int64(game.FirstReleaseDate), 0).Format("2006-01-02")
		}
	}
	return games, nil
}

func GetGameFromIDIGDB(igdbID int) (*IGDBGameObject, error) {
	// construct query string
	requestBody := `where id=` + strconv.Itoa(igdbID) + `; fields name, platforms.name, screenshots.animated, first_release_date, screenshots.image_id, screenshots.height, screenshots.width, cover.image_id, similar_games.name, similar_games.cover.image_id, storyline, summary, websites.url, websites.category, videos.name, videos.video_id, game_modes.name, involved_companies.developer, involved_companies.publisher, involved_companies.company.name, genres.name, artworks.animated, artworks.image_id, artworks.height, artworks.width, dlcs.name, dlcs.summary, dlcs.release_dates.human, dlcs.release_dates.platform.name, dlcs.release_dates.date, dlcs.cover.image_id, release_dates.human, release_dates.platform.name, release_dates.date, artworks.*; limit 20;`
	b, err := queryIGDBGames(requestBody)
	if err != nil {
		return nil, err
	}
	// unmarshal response
	var response IGDBGamesResultsResponseObject
	err = json.Unmarshal(b, &response)
	if err != nil {
		return nil, err
	}
	if len(response) == 0 {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "No game found with this igdbID")
	}
	game := response[0]
	// get image urls
	game.Cover.ImageURL = getIGDBImageURL(game.Cover.ImageID, IGDBImageCover)
	game.PosterURL = getIGDBImageURL(game.Cover.ImageID, IGDBImageCover)
	game.MediaTitle = game.Name
	game.MediaType = database.MediaTypeGame
	game.MediaSource = SourceIGDB
	game.SourceID = game.ID
	if game.FirstReleaseDate != 0 {
		game.ReleaseDate = time.Unix(int64(game.FirstReleaseDate), 0).Format("2006-01-02")
	}
	for num, artwork := range game.Artworks {
		game.Artworks[num].ImageURL = getIGDBImageURL(artwork.ImageID, IGDBImageOriginal)
	}
	for num, screenshot := range game.Screenshots {
		game.Screenshots[num].ImageURL = getIGDBImageURL(screenshot.ImageID, IGDBImageOriginal)
	}
	for num, dlc := range game.DLCs {
		game.DLCs[num].Cover.ImageURL = getIGDBImageURL(dlc.Cover.ImageID, IGDBImageCover)
	}
	for num, similarGame := range game.SimilarGames {
		game.SimilarGames[num].Cover.ImageURL = getIGDBImageURL(similarGame.Cover.ImageID, IGDBImageCover)
	}
	for num, videos := range game.Videos {
		game.Videos[num].Key = videos.VideoID
	}
	return &game, nil
}

func AddGameToCollectionIGDB(username string, source string, sourceID int, collectionID *int64) error {
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		return err
	}
	if source != SourceIGDB {
		panic("Only igdb source is allowed for now")
	}
	entry, err := GetRecordObjectIGDB(sourceID)
	if err != nil {
		return err
	}
	// insert record to internal library if not exists
	err = database.UpsertMediaRecord(entry)
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

func GetRecordObjectIGDB(igdbID int) (*database.MediaRecord, error) {
	game, err := GetGameFromIDIGDB(igdbID)
	if err != nil {
		return nil, err
	}
	// add item to internal library if not there
	releaseDate := ""
	if game.FirstReleaseDate != 0 {
		releaseDate = time.Unix(int64(game.FirstReleaseDate), 0).Format("2006-01-02")
	}
	gameJson, err := json.Marshal(game)
	if err != nil {
		return nil, err
	}
	// import igdb genres
	var tagsArray []database.GenreObject
	for _, genre := range game.Genres {
		tagsArray = append(tagsArray, database.GenreObject{
			ID:   int64(genre.ID),
			Name: genre.Name,
		})
	}
	record := database.MediaRecord{
		RecordType:   database.MediaTypeGame,
		MediaSource:  SourceIGDB,
		SourceID:     strconv.Itoa(game.SourceID),
		MediaTitle:   game.MediaTitle,
		ReleaseDate:  releaseDate,
		Overview:     game.Summary,
		FullData:     gameJson,
		ThumbnailURL: game.Cover.ImageURL,
		Genres:       tagsArray,
		UserTags:     nil,
	}
	return &record, nil
}

func getIGDBImageURL(imageID string, imageType string) string {
	if imageID == "" {
		return ""
	}
	return fmt.Sprintf(IGDBImagePath, imageType, imageID)
}
