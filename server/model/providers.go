package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/model/sources"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// for calling the python scripts
type ProviderQueryObject struct {
	MediaSource     string   `json:"media_source"` // eg. tmdb
	SourceID        int      `json:"source_id"`
	IMDbID          string   `json:"imdb_id,omitempty"` // starts with 'tt'
	MediaType       string   `json:"media_type"`        // movies or tvshows, etc.
	Season          int      `json:"season,omitempty"`
	Episode         int      `json:"episode,omitempty"`
	SourceEpisodeID int      `json:"source_episode_id,omitempty"`
	EpisodeGroupID  string   `json:"episode_group_id,omitempty"`
	Query           string   `json:"search_query,omitempty"`
	Params          []string `json:"params"`
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
	MediaSource     string `json:"media_source"`
	SourceID        int    `json:"source_id"`
	MediaType       string `json:"media_type"`              // movies or tvshows, etc.
	IMDbID          string `json:"imdb_id"`                 // starts with 'tt'
	SeasonNumber    *int   `json:"season_number,omitempty"` // shows only
	EpisodeNumber   *int   `json:"episode_number,omitempty"`
	SourceEpisodeID *int   `json:"source_episode_id,omitempty"` // tv shows only
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
	FileIdx     *int        `json:"file_idx"`  // file index of stream in torrent
	FileSize    int         `json:"file_size"` // file size in bytes
	Rank        int         `json:"rank"`
	Seeders     int         `json:"seeders"`
	Leechers    int         `json:"leechers"`
	Sources     []string    `json:"sources"` // trackers for p2p
	URI         string      `json:"uri"`
	EncodedData string      `json:"encoded_data"` // data encoded in JWT for playing streams
	ParsedData  *ParsedData `json:"data"`
}

var providersCacheDuration time.Duration = 2 * time.Hour

func InitializeProviders() {

}

/*
For now, works with aiostreams, only one provider
*/
func SearchProviders(query ProviderQueryObject) (*ProviderResponseObject, error) {
	// construct cache key
	providersCacheKey := fmt.Sprintf("providers|%s|%s-%d", query.MediaType, query.MediaSource, query.SourceID)
	if query.MediaType == database.MediaTypeTVShow {
		providersCacheKey += fmt.Sprintf("|S%d|E%d|episode_group_id:%s", query.Season, query.Episode, query.EpisodeGroupID)
	}
	// get cache
	var cacheObject ProviderResponseObject
	cacheExists, _ := database.GetCache(providersCacheKey, &cacheObject)
	if cacheExists {
		return &cacheObject, nil
	}
	searchSeason := query.Season
	searchEpisode := query.Episode
	// Not in cache fetch new season/episode number if requested
	if query.EpisodeGroupID != "" {
		targetSeason, targetEpisode, err := getTVDBSeasonEpisodeFromTMDBID(query.SourceID, query.SourceEpisodeID, query.EpisodeGroupID)
		// if we try querying "tvdb" but no results are found (eg. no episode groups for this show),
		// just use the original season/episode without handling the error
		// else if a episode_group_id is provided but invalid, return an error
		if err == nil {
			searchSeason = *targetSeason
			searchEpisode = *targetEpisode
		} else if query.EpisodeGroupID != "tvdb" {
			return nil, err
		}
	}
	cmd := exec.Command("python", "model/providers/aiostreams.py",
		"--connection_string", "http://localhost:3500|08f3f1c0-12ff-439d-81d1-962ac3dfe620|abcdef",
		"--imdb_id", query.IMDbID,
		"--media_type", query.MediaType,
		"--season", strconv.Itoa(searchSeason),
		"--episode", strconv.Itoa(searchEpisode),
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
		MediaSource:     query.MediaSource,
		SourceID:        query.SourceID,
		MediaType:       query.MediaType,
		IMDbID:          query.IMDbID,
		SeasonNumber:    &query.Season,
		EpisodeNumber:   &query.Episode,
		SourceEpisodeID: &query.SourceEpisodeID,
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
			if stream.P2P == database.ProtocolP2P {
				stream.URI = *GetMagnetURI(stream.InfoHash, nil)
			}
		}
	}
	result := ProviderResponseObject{
		StreamMediaDetails: streamInfo,
		Providers:          &providersArray,
	}
	_, err = database.SetCache(providersCacheKey, result, providersCacheDuration)
	if err != nil {
		// just log error, no failed return
		_ = helpers.LogErrorWithMessage(err, "Failed to set cache")
	}
	return &result, nil
}

/*
Returns season-episode number in an episode group for given episodeID
This is useful to convert from tmdb to tvdb orderings
*/
func getTVDBSeasonEpisodeFromTMDBID(sourceID int, episodeID int, episodeGroupID string) (*int, *int, error) {
	// use given episode ID or grab a "tvdb" one if it exists
	// a bit hacky, just pass in "tvdb" as episodeGroupID to search
	if episodeGroupID == "tvdb" {
		episodeGroups, err := sources.GetTVEpisodeGroupsTMDB(sourceID)
		if err != nil {
			return nil, nil, err
		}
		if len(episodeGroups.Results) == 0 {
			return nil, nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"No episode groups for id tmdb-"+strconv.Itoa(sourceID))
		}
		for _, item := range episodeGroups.Results {
			if strings.Contains(strings.ToLower(item.Name), "tvdb") ||
				strings.Contains(strings.ToLower(item.Description), "tvdb") {
				episodeGroupID = item.ID
				break
			}
		}
	}
	// not found case, episodeGroupID isn't updated yet
	if episodeGroupID == "tvdb" {
		return nil, nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Could not find tvdb episode group for id tmdb-"+strconv.Itoa(sourceID))
	}
	episodeGroupDetails, err := sources.GetTVEpisodeGroupsDetailsTMDB(episodeGroupID)
	if err != nil {
		return nil, nil, helpers.LogErrorWithMessage(err, "Could not retrieve episode group details for id: tmdb-"+episodeGroupID)
	}
	if len(episodeGroupDetails.Groups) == 0 || len(episodeGroupDetails.Groups[0].Episodes) == 0 {
		return nil, nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Error parsing episode group details for id: tmdb-"+episodeGroupID)
	}
	var targetSeason *int
	var targetEpisode *int

	// break fully if found
outerLoop:
	for _, group := range episodeGroupDetails.Groups {
		for _, episode := range group.Episodes {
			if episode.ID == int64(episodeID) {
				targetSeason = &group.Order
				targetEpisode = &episode.Order
				// orders are 0 indexed, seasons are already correct if "Specials" exist
				(*targetEpisode)++
				break outerLoop
			}
		}
	}
	// If specials (season number 0) don't exist, fix order's 0-index
	if episodeGroupDetails.Groups[0].Episodes[0].SeasonNumber != 0 {
		(*targetSeason)++
	}
	return targetSeason, targetEpisode, nil
}
