package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"hound/database"
	"hound/helpers"
	"net/http"
	"time"
)

const BASE_URL = "https://aiostreamsfortheweebs.midnightignite.me/stremio/5ebb4db5-dbc5-4b68-a9ad-574b926e04cf/eyJpIjoia2JBRE5paXlmdXR0c1oyMzd4SEh6dz09IiwiZSI6ImhjdWFoQ1hLTlNxcjdFOFNLYkJUWlE9PSIsInQiOiJhIn0"
const MANIFEST_PATH = "/manifest.json"
const TV_STREAMS_PATH = "/stream/series/%s:%d:%d.json"
const MOVIE_STREAMS_PATH = "/stream/movie/%s.json"

type StremioStreamBehaviorHints struct {
	BingeGroup *string `json:"bingeGroup,omitempty"`
	VideoHash  *string `json:"videoHash,omitempty"`
	Filename   *string `json:"filename,omitempty"`
	VideoSize  *int    `json:"videoSize,omitempty"` // size in bytes
}

// Pretty much everything is optional per Stremio docs,
// but url/infohash are required
// only url/infohash are supported for now
type StremioStreamObject struct {
	Name          *string                     `json:"name,omitempty"`
	Title         *string                     `json:"title,omitempty"`       // will be deprecated in stremio according to docs
	Description   *string                     `json:"description,omitempty"` // title will be replaced with description
	URL           *string                     `json:"url,omitempty"`
	InfoHash      *string                     `json:"infoHash,omitempty"`
	FileIdx       *int                        `json:"fileIdx,omitempty"`
	Sources       *[]string                   `json:"sources,omitempty"`
	BehaviorHints *StremioStreamBehaviorHints `json:"behaviorHints,omitempty"`
}

type StremioStreamResponse struct {
	Streams []StremioStreamObject `json:"streams,omitempty"`
}

func GetStremioStreams(mediaType string, imdbID string, seasonNumber int, episodeNumber int) (*StremioStreamResponse, error) {
	url := ""
	switch mediaType {
	case database.MediaTypeMovie:
		url = BASE_URL + fmt.Sprintf(MOVIE_STREAMS_PATH, imdbID)
	case database.MediaTypeTVShow:
		url = BASE_URL + fmt.Sprintf(TV_STREAMS_PATH, imdbID, seasonNumber, episodeNumber)
	default:
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid media type")
	}
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError),
			"Error querying stremio plugin"+resp.Status)
	}
	var streamResponse StremioStreamResponse
	if err := json.NewDecoder(resp.Body).Decode(&streamResponse); err != nil {
		return nil, err
	}
	return &streamResponse, nil
}
