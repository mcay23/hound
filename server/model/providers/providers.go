package providers

import (
	"errors"
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/model/sources"
	"strconv"
	"strings"
	"time"
)

type ProvidersQueryRequest struct {
	MediaSource     string   `json:"media_source"` // eg. tmdb
	SourceID        string   `json:"source_id"`
	IMDbID          string   `json:"imdb_id,omitempty"` // starts with 'tt'
	MediaType       string   `json:"media_type"`        // movies or tvshows, etc.
	SeasonNumber    *int     `json:"season_number,omitempty"`
	EpisodeNumber   *int     `json:"episode_number,omitempty"`
	EpisodeSourceID *string  `json:"source_episode_id,omitempty"`
	EpisodeGroupID  string   `json:"episode_group_id,omitempty"`
	Query           string   `json:"search_query,omitempty"` // not used for now
	Params          []string `json:"params"`
}

// to encode into JWT string
// this struct will be encoded/decoded when playing/downloading
// a stream
type StreamObjectFull struct {
	StreamMediaDetails
	StreamObject
}

type StreamMediaDetails struct {
	MediaType       string  `json:"media_type"` // movies or tvshows, etc.
	MediaSource     string  `json:"media_source"`
	SourceID        string  `json:"source_id"`
	IMDbID          string  `json:"imdb_id"`                 // starts with 'tt'
	SeasonNumber    *int    `json:"season_number,omitempty"` // shows only
	EpisodeNumber   *int    `json:"episode_number,omitempty"`
	EpisodeSourceID *string `json:"episode_source_id,omitempty"` // tv shows only
}

type StreamObject struct {
	Provider       string      `json:"provider"`
	StreamProtocol string      `json:"stream_protocol"` // http or p2p
	URI            string      `json:"uri"`             // magnet link or http link
	InfoHash       string      `json:"info_hash"`
	Title          string      `json:"title"`
	Description    string      `json:"description"`
	Filename       *string     `json:"file_name,omitempty"` // might not be reliable
	FileIdx        *int        `json:"file_idx,omitempty"`  // file index for p2p type
	FileSize       *int        `json:"file_size,omitempty"` // file size in bytes
	Sources        *[]string   `json:"sources,omitempty"`   // trackers for p2p
	EncodedData    string      `json:"encoded_data"`        // data encoded in AES for playing streams in hound
	ParsedData     *ParsedData `json:"parsed_data,omitempty"`
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

type ProviderObject struct {
	Provider string          `json:"provider"` // provider name in /providers folder
	Streams  []*StreamObject `json:"streams"`
}

type ProviderResponseObject struct {
	StreamMediaDetails
	Providers []*ProviderObject `json:"providers"`
}

const providersCacheTTL = time.Hour * 2

func QueryProviders(query ProvidersQueryRequest) (*ProviderResponseObject, error) {
	providersCacheKey := fmt.Sprintf("providers|%s|%s-%s", query.MediaType, query.MediaSource, query.SourceID)
	if query.MediaType == database.MediaTypeTVShow {
		providersCacheKey += fmt.Sprintf("|S%d|E%d|episode_group_id:%s", *query.SeasonNumber, *query.EpisodeNumber, query.EpisodeGroupID)
	}
	// get cache
	var cacheObject ProviderResponseObject
	cacheExists, _ := database.GetCache(providersCacheKey, &cacheObject)
	if cacheExists {
		return &cacheObject, nil
	}
	streamMediaDetails := StreamMediaDetails{
		MediaType:       query.MediaType,
		MediaSource:     query.MediaSource,
		SourceID:        query.SourceID,
		IMDbID:          query.IMDbID,
		SeasonNumber:    query.SeasonNumber,
		EpisodeNumber:   query.EpisodeNumber,
		EpisodeSourceID: query.EpisodeSourceID,
	}
	// for TV shows,
	// check if the season starts with episode 1
	// some shows in tmdb don't start in episode 1
	// eg. Season 1 has 20 episodes, Season 2 starts at ep. 21
	// Sometimes happens for Japanese anime
	// This is an indication that we might want to use TVDB episode numbers
	if query.MediaType == database.MediaTypeTVShow {
		if query.SeasonNumber == nil || query.EpisodeNumber == nil {
			return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"SeasonNumber and EpisodeNumber are required for TV shows")
		}
		showID, err := strconv.Atoi(query.SourceID)
		if err != nil {
			return nil, err
		}
		seasonDetails, err := sources.GetTVSeasonTMDB(showID, *query.SeasonNumber)
		if err != nil {
			return nil, err
		}
		// if season doesn't start with 1, check if media has
		// tvdb ordering available
		if seasonDetails.Episodes[0].EpisodeNumber != 1 {
			oldEp := *query.EpisodeNumber
			firstEp := seasonDetails.Episodes[0].EpisodeNumber
			if query.EpisodeSourceID == nil {
				// find episodeID
				epItem, err := sources.GetEpisodeTMDB(showID,
					*query.SeasonNumber, *query.EpisodeNumber)
				if err != nil {
					return nil, err
				}
				epStr := strconv.Itoa(int(epItem.ID))
				query.EpisodeSourceID = &epStr
				streamMediaDetails.EpisodeSourceID = &epStr
			}
			episodeID, err := strconv.Atoi(*query.EpisodeSourceID)
			if err != nil {
				return nil, err
			}
			// if empty string is passed, automatically searches for tvdb ordering
			tvdbSeasonNumber, tvdbEpisodeNumber, err :=
				getSeasonEpisodeFromEpisodeGroup(showID, episodeID, query.EpisodeGroupID)
			if err == nil {
				query.SeasonNumber = &tvdbSeasonNumber
				query.EpisodeNumber = &tvdbEpisodeNumber
			} else {
				// otherwise, normalize episode numbers so they start from 1 anyway
				// we do this since this is more likely to align to tvdb standards,
				// which many providers use
				normalizedEp := oldEp - firstEp + 1
				query.EpisodeNumber = &normalizedEp
			}
		}
	}
	stremioStreams, err := getStremioStreams(query, streamMediaDetails)
	if err != nil {
		return nil, err
	}
	result := ProviderResponseObject{
		StreamMediaDetails: streamMediaDetails,
		Providers:          []*ProviderObject{stremioStreams},
	}
	_, err = database.SetCache(providersCacheKey, result, providersCacheTTL)
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
func getSeasonEpisodeFromEpisodeGroup(sourceID int, episodeID int, episodeGroupID string) (int, int, error) {
	if episodeGroupID == "" {
		episodeGroupID = "tvdb"
	}
	// use given episode ID or grab a "tvdb" one if it exists
	// a bit hacky, just pass in "tvdb" as episodeGroupID to search
	if episodeGroupID == "tvdb" {
		episodeGroups, err := sources.GetTVEpisodeGroupsTMDB(sourceID)
		if err != nil {
			return -1, -1, err
		}
		if len(episodeGroups.Results) == 0 {
			return -1, -1, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
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
		return -1, -1, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Could not find tvdb episode group for id tmdb-"+strconv.Itoa(sourceID))
	}
	episodeGroupDetails, err := sources.GetTVEpisodeGroupsDetailsTMDB(episodeGroupID)
	if err != nil {
		return -1, -1, helpers.LogErrorWithMessage(err, "Could not retrieve episode group details for id: tmdb-"+episodeGroupID)
	}
	if len(episodeGroupDetails.Groups) == 0 || len(episodeGroupDetails.Groups[0].Episodes) == 0 {
		return -1, -1, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Error parsing episode group details for id: tmdb-"+episodeGroupID)
	}
	var targetSeason int
	var targetEpisode int
	found := false

	// break fully if found
outerLoop:
	for _, group := range episodeGroupDetails.Groups {
		for _, episode := range group.Episodes {
			if episode.ID == int64(episodeID) {
				targetSeason = group.Order
				targetEpisode = episode.Order
				// orders are 0 indexed, seasons are already correct if "Specials" exist
				targetEpisode++
				found = true
				break outerLoop
			}
		}
	}
	if !found {
		return -1, -1, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Episode not found in episode group tmdb-"+episodeGroupID)
	}
	// If specials (season number 0) don't exist, fix order's 0-index
	if episodeGroupDetails.Groups[0].Episodes[0].SeasonNumber != 0 {
		targetSeason++
	}
	return targetSeason, targetEpisode, nil
}
