package sources

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/mcay23/hound/database"
)

const MDBListListURL = `https://api.mdblist.com/lists/%s/%s/items?append_to_response=poster,description&unified=true&apikey=%s`

type MDBListItem struct {
	ID             int        `json:"id"`
	IDs            MDBListIDs `json:"ids"`
	MediaType      string     `json:"mediatype"` // movie or show
	Rank           int        `json:"rank"`
	Adult          int        `json:"adult"`
	Title          string     `json:"title"`
	Poster         string     `json:"poster"`
	Description    string     `json:"description"`
	Language       string     `json:"language"`
	SpokenLanguage string     `json:"spoken_language"`
	ReleaseDate    string     `json:"release_date"`
}

type MDBListIDs struct {
	TMDB *int64  `json:"tmdb,omitempty"`
	IMDB *string `json:"imdb,omitempty"`
}

var apiKey = "abc"

func GetMDBList(listOwner string, listName string) ([]MDBListItem, error) {
	cacheKey := fmt.Sprintf("mdblist|%s|%s", listOwner, listName)
	var cacheObject []MDBListItem
	found, _ := database.GetCache(cacheKey, &cacheObject)
	if found {
		return cacheObject, nil
	}
	url := fmt.Sprintf(
		MDBListListURL,
		url.PathEscape(listOwner),
		url.PathEscape(listName),
		apiKey,
	)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var items []MDBListItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}
	database.SetCache(cacheKey, items, time.Hour*24)
	return items, nil
}
