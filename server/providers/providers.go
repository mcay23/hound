package providers

import (
	"encoding/json"
	"fmt"
	"hound/helpers"
	"hound/view"
	"os/exec"
	"strconv"
)

type ProviderQueryObject struct {
	MediaSource			string	 `json:"media_source"`			 // eg. tmdb
	SourceID			   int	 `json:"source_id"`
	IMDbID              string   `json:"imdb_id,omitempty"`      // starts with 'tt'
	MediaType           string   `json:"media_type"`   // movies or tvshows, etc.
	Season                 int   `json:"season,omitempty"`
	Episode                int   `json:"episode,omitempty"`
	Query				string   `json:"search_query,omitempty"`
	Params            []string   `json:"params"`
}

type ProviderQueryResponse struct {
	Status            string                    `json:"go_status"`
	Provider          string                    `json:"provider"`
	IMDbID            string                    `json:"imdb_id"`      // starts with 'tt'
	MediaType		  string                    `json:"media_type"`   // movies or tvshows, etc.
	Season            int                       `json:"season,omitempty"`
	Episode           int                       `json:"episode,omitempty"`
	Streams           *[]map[string]interface{}   `json:"streams"`
}

func InitializeProviders() {

}

/*
	For now, works with aiostreams, only one provider
 */
func SearchProviders(query ProviderQueryObject) (*view.ProvidersResponseObject, error) {
	cmd := exec.Command("python", "providers/aiostreams.py",
			"--connection_string", "http://localhost:3500|08f3f1c0-12ff-439d-81d1-962ac3dfe620|abcdef",
			"--imdb_id", query.IMDbID,
			"--media_type", query.MediaType,
			"--season", strconv.Itoa(query.Season),
			"--episode", strconv.Itoa(query.Episode),
		)
	out, err := cmd.CombinedOutput()
	if err != nil {
		helpers.LogErrorWithMessage(err, fmt.Sprintf("error: %v\n%s", err, string(out)))
		return nil, err
	}
	var obj view.ProviderObject
	if err := json.Unmarshal(out, &obj); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, err
	}
	result := view.ProvidersResponseObject{
		MediaSource: query.MediaSource,
		SourceID:    query.SourceID,
		MediaType:   query.MediaType,
		IMDbID:      query.IMDbID,
		Season:      query.Season,
		Episode:     query.Episode,
		Providers:   &[]view.ProviderObject{obj},
	}
	result.IMDbID = query.IMDbID
	result.MediaType = query.MediaType
	result.Season = query.Season
	result.Episode = query.Episode
	return &result, nil
}