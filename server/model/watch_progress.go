package model

import (
	"errors"
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/model/sources"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

const watchProgressCacheTTL = 24 * 180 * time.Hour // 180 days

/*
Watch progress is not stored in the db because it's not deemed as critical
Cache is preferred for now so media records don't have to be inserted for watch progress
on new movies/episodes. Media record is only inserted on completed watch history events
This is intentional to prevent downloading too much metadata for movies users peek at for
15 minutes and never watch again, for example.
*/
type WatchProgress struct {
	MediaSource            string  `json:"media_source"`                                     // "tmdb"
	ParentSourceID         string  `json:"parent_source_id"`                                 // movie/show source id
	StreamType             string  `json:"stream_type"`                                      // p2p, http, local, etc.
	EncodedData            string  `json:"encoded_data"`                                     // for hound-proxied sources
	SourceURI              string  `json:"source_uri"`                                       // magnet, http link, local path
	SeasonNumber           *int    `json:"season_number,omitempty"`                          // only defined for shows
	EpisodeNumber          *int    `json:"episode_number,omitempty"`                         // only defined for shows
	EpisodeID              *string `json:"episode_id,omitempty"`                             // episode source id
	CurrentProgressSeconds int     `json:"current_progress_seconds" binding:"required,gt=0"` // how many seconds in the user is
	TotalDurationSeconds   int     `json:"total_duration_seconds" binding:"required,gt=0"`   // total duration of the media in seconds
	LastWatchedAt          int64   `json:"last_watched_at"`                                  // last unix time when the playback progress was set
}

// eg. watch_progress|userid:123|mediaType:movie|source:tmdb-123|season:nil|episode:nil
// eg. watch_progress|userid:123|mediaType:show|source:tmdb-123|season:1|episode:2
// each user should only have one watch_progress of a movie/episode at one time
// subsequent writes are updates to the existing watch_progress
const WATCH_PROGRESS_CACHE_KEY = "watch_progress|userid:%d|mediaType:%s|source:%s-%s|season:%v|episode:%v"

func GetWatchProgress(userID int64, mediaType string, mediaSource string,
	sourceID string, seasonNumber *int) ([]*WatchProgress, error) {
	prefixFormat := strings.Split(WATCH_PROGRESS_CACHE_KEY, "|season")[0]
	keyPrefix := fmt.Sprintf(prefixFormat, userID, mediaType, mediaSource, sourceID)
	if mediaType == database.MediaTypeTVShow && seasonNumber != nil {
		keyPrefix += fmt.Sprintf("|season:%v", *seasonNumber)
	}
	keys, err := database.GetKeysWithPrefix(keyPrefix)
	if err != nil {
		return nil, err
	} else if len(keys) == 0 {
		return nil, nil
	}
	var watchProgressArray []*WatchProgress
	for _, key := range keys {
		item := WatchProgress{}
		exists, err := database.GetCache(key, &item)
		if err != nil {
			return nil, err
		}
		if exists {
			watchProgressArray = append(watchProgressArray, &item)
		}
	}
	return watchProgressArray, nil
}

func SetWatchProgress(userID int64, mediaType string, mediaSource string,
	sourceID string, watchProgress *WatchProgress) error {

	if watchProgress == nil {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"invalid param: watchProgress is nil")
	}
	if watchProgress.CurrentProgressSeconds > watchProgress.TotalDurationSeconds {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"invalid param: current progress is greater than total duration")
	}
	watchProgress.MediaSource = mediaSource
	watchProgress.ParentSourceID = sourceID
	watchProgress.LastWatchedAt = time.Now().Unix()
	// dyamically fill episodeID
	if mediaType == database.MediaTypeTVShow {
		if watchProgress.SeasonNumber == nil || watchProgress.EpisodeNumber == nil {
			return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"invalid param: season/episode number is nil")
		}
		tmdbID, err := strconv.Atoi(sourceID)
		if err != nil {
			return helpers.LogErrorWithMessage(err, "failed to parse source id")
		}
		targetEpisode, err := sources.GetEpisodeTMDB(tmdbID,
			*watchProgress.SeasonNumber, *watchProgress.EpisodeNumber)
		if err != nil {
			return helpers.LogErrorWithMessage(err, "failed to get episode id")
		}
		episodeIDStr := strconv.Itoa(int(targetEpisode.ID))
		watchProgress.EpisodeID = &episodeIDStr
		cacheKey := fmt.Sprintf(WATCH_PROGRESS_CACHE_KEY, userID, mediaType, mediaSource, sourceID,
			*watchProgress.SeasonNumber, *watchProgress.EpisodeNumber)
		_, err = database.SetCache(cacheKey, watchProgress, watchProgressCacheTTL)
		if err != nil {
			return err
		}
		slog.Info("Watch Progress Set", "key", cacheKey)
		return nil
	}
	// for movies, don't send in season/episode number
	cacheKey := fmt.Sprintf(WATCH_PROGRESS_CACHE_KEY, userID, mediaType, mediaSource, sourceID,
		nil, nil)
	_, err := database.SetCache(cacheKey, watchProgress, watchProgressCacheTTL)
	slog.Info("Watch Progress Set", "key", cacheKey)
	return err
}

// Delete all watch progress before deleteBefore
// If nil, delete all watch progress
func DeleteWatchProgress(userID int64, mediaType string, mediaSource string,
	sourceID string, seasonNumber *int, episodeNumber *int, deleteBefore *time.Time) error {
	prefixFormat := strings.Split(WATCH_PROGRESS_CACHE_KEY, "|season")[0]
	keyPrefix := fmt.Sprintf(prefixFormat, userID, mediaType, mediaSource, sourceID)
	if mediaType == database.MediaTypeTVShow {
		if seasonNumber != nil {
			keyPrefix += fmt.Sprintf("|season:%v", *seasonNumber)
			if episodeNumber != nil {
				keyPrefix += fmt.Sprintf("|episode:%v", *episodeNumber)
			}
		}
	}
	keys, err := database.GetKeysWithPrefix(keyPrefix)
	if err != nil {
		return err
	}
	var deleteError error
	for _, key := range keys {
		// skip checks if deleteBefore == nil, minor optimization
		// over setting deleteBefore to time.Now()
		if deleteBefore != nil {
			var watchProgress WatchProgress
			exists, err := database.GetCache(key, &watchProgress)
			if err != nil {
				return err
			}
			if !exists {
				continue
			}
			// skip if setting a watch before the current scrobble activity
			// eg. you mark movie as watched at 3 months ago,
			// don't delete current progress
			if watchProgress.LastWatchedAt > deleteBefore.Unix() {
				continue
			}
		}
		deleteError = database.DeleteCache(key)
		if deleteError != nil {
			// don't return
			helpers.LogErrorWithMessage(err, "failed to delete watch progress")
		}
	}
	return deleteError
}
