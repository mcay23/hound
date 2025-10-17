package model

import (
	"encoding/json"
	"fmt"
	"hound/helpers"
	"hound/model/database"
	"os/exec"
	"strconv"
	"time"
)

// for calling the python scripts
type ProviderQueryObject struct {
	MediaSource string   `json:"media_source"` // eg. tmdb
	SourceID    int      `json:"source_id"`
	IMDbID      string   `json:"imdb_id,omitempty"` // starts with 'tt'
	MediaType   string   `json:"media_type"`        // movies or tvshows, etc.
	Season      int      `json:"season,omitempty"`
	Episode     int      `json:"episode,omitempty"`
	Query       string   `json:"search_query,omitempty"`
	Params      []string `json:"params"`
}

// return struct of the python scripts
type ProviderQueryResponse struct {
	Status    string                    `json:"go_status"`
	Provider  string                    `json:"provider"`
	IMDbID    string                    `json:"imdb_id"`    // starts with 'tt'
	MediaType string                    `json:"media_type"` // movies or tvshows, etc.
	Season    int                       `json:"season,omitempty"`
	Episode   int                       `json:"episode,omitempty"`
	Streams   *[]map[string]interface{} `json:"streams"`
}

// response containing all providers
type ProviderResponseObject struct {
	StreamMediaDetails
	Providers *[]ProviderObject `json:"providers"`
}

type StreamMediaDetails struct {
	MediaSource string `json:"media_source"`
	SourceID    int    `json:"source_id"`
	MediaType   string `json:"media_type"`       // movies or tvshows, etc.
	IMDbID      string `json:"imdb_id"`          // starts with 'tt'
	Season      *int   `json:"season,omitempty"` // shows only
	Episode     *int   `json:"episode,omitempty"`
}

type ProviderObject struct {
	Provider string          `json:"provider"` // provider name in /providers folder
	Streams  []*StreamObject `json:"streams"`
}

type ParsedData struct {
	VideoCodec    string   `json:"codec"`
	AudioCodec    []string `json:"audio"`
	Subbed        bool     `json:"subbed"`
	Dubbed        bool     `json:"dubbed"`
	AudioChannels []string `json:"channels"`
	FileContainer string   `json:"container"`
	Languages     []string `json:"languages"`
	BitDepth      string   `json:"bit_depth"` // eg. 10bit
	HDR           []string `json:"hdr"`
}

type StreamObject struct {
	Addon       string      `json:"addon"`
	Cached      string      `json:"cached"`  // whether the stream is cached
	Service     string      `json:"service"` // service such as RD, etc.
	P2P         string      `json:"p2p"`     // type of stream, such as 'debrid'
	InfoHash    string      `json:"infohash"`
	Indexer     string      `json:"indexer"`
	Filename    string      `json:"file_name"`
	FolderName  string      `json:"folder_name"` // folder name for packs
	Resolution  string      `json:"resolution"`
	FileIndex   int         `json:"file_idx"`  // file index of stream in torrent
	FileSize    int         `json:"file_size"` // file size in bytes
	Rank        int         `json:"rank"`
	Seeders     int         `json:"seeders"`
	Leechers    int         `json:"leechers"`
	URL         string      `json:"url"`
	EncodedData string      `json:"encoded_data"` // data encoded in JWT for playing streams
	ParsedData  *ParsedData `json:"data"`
}

func InitializeProviders() {

}

/*
For now, works with aiostreams, only one provider
*/
func SearchProviders(query ProviderQueryObject) (*ProviderResponseObject, error) {
	// construct cache key
	providersCacheKey := fmt.Sprintf("providers|%s|%s-%d", query.MediaType, query.MediaSource, query.SourceID)
	if query.MediaType == database.MediaTypeTVShow {
		providersCacheKey += fmt.Sprintf("|S%d|E%d", query.Season, query.Episode)
	}
	// get cache
	var cacheObject ProviderResponseObject
	cacheExists, _ := GetCache(providersCacheKey, &cacheObject)
	if cacheExists {
		return &cacheObject, nil
	}
	// Not in cache, run command
	cmd := exec.Command("python", "model/providers/aiostreams.py",
		"--connection_string", "http://localhost:3500|08f3f1c0-12ff-439d-81d1-962ac3dfe620|abcdef",
		"--imdb_id", query.IMDbID,
		"--media_type", query.MediaType,
		"--season", strconv.Itoa(query.Season),
		"--episode", strconv.Itoa(query.Episode),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, fmt.Sprintf("error: %v\n%s", err, string(out)))
		return nil, err
	}
	var obj ProviderObject
	if err := json.Unmarshal(out, &obj); err != nil {
		_ = helpers.LogErrorWithMessage(err, "Error unmarshalling json from provider")
		return nil, err
	}
	var streamInfo = StreamMediaDetails{
		MediaSource: query.MediaSource,
		SourceID:    query.SourceID,
		MediaType:   query.MediaType,
		IMDbID:      query.IMDbID,
		Season:      &query.Season,
		Episode:     &query.Episode,
	}
	providersArray := []ProviderObject{obj}
	for _, provider := range providersArray {
		for _, stream := range provider.Streams {
			var fullObject = StreamObjectFull{
				StreamMediaDetails: streamInfo,
				StreamObject:       *stream,
			}
			encodedData, err := EncodeJsonStreamJWT(fullObject)
			if err != nil {
				_ = helpers.LogErrorWithMessage(err, "Failed to encode stream into JWT string")
				continue
			}
			stream.EncodedData = encodedData
		}
	}
	result := ProviderResponseObject{
		StreamMediaDetails: streamInfo,
		Providers:          &providersArray,
	}
	// TODO revert TTL
	_, err = SetCache(providersCacheKey, result, 1*time.Second)
	if err != nil {
		// just log error, no failed return
		_ = helpers.LogErrorWithMessage(err, "Failed to set cache")
	}
	return &result, nil
}
