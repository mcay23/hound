package providers

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type ProviderQueryObject struct {
	IMDbID              *string  `json:"imdb_id"`      // starts with 'tt'
	MediaType           string   `json:"media_type"`   // movies or tvshows, etc.
	Season                 int   `json:"season"`
	Episode                int   `json:"episode"`
	Query				*string  `json:"search_query"`
	Params            []string   `json:"params"`
}

type ProviderQueryResponse struct {
	Status              bool    `json:"go_status"`
	IMDbID              string  `json:"imdb_id"`      // starts with 'tt'
	MediaType           string  `json:"media_type"`   // movies or tvshows, etc.

}

func InitializeProviders() {

}

func SearchProviders(query *ProviderQueryObject) (*map[string]interface{}, error) {
	if *query.IMDbID != "" {

	}
	cmd := exec.Command("python", "providers/torrentio.py")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error running script:", err)
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, err
	}
	return &result, nil
}