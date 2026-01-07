package model

import (
	"errors"
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/model/providers"
	"hound/model/sources"
	"log/slog"
	"strconv"
	"strings"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
)

const watchProgressCacheTTL = 90 * 24 * time.Hour // 90 days

/*
Watch progress is not stored in the db because it's not deemed as critical
Cache is preferred for now so media records don't have to be inserted for watch progress
on new movies/episodes. Media record is only inserted on completed watch history events
This is intentional to prevent downloading too much metadata for movies users peek at for
15 minutes and never watch again, for example.
*/
type WatchProgress struct {
	MediaType              string `json:"media_type"`                                       // "movie" or "tvshow"
	MediaSource            string `json:"media_source"`                                     // "tmdb"
	SourceID               string `json:"source_id"`                                        // movie/show source id
	StreamProtocol         string `json:"stream_protocol"`                                  // p2p, http, local, etc.
	EncodedData            string `json:"encoded_data"`                                     // for hound-proxied sources
	SourceURI              string `json:"source_uri"`                                       // magnet, http link, local path
	CurrentProgressSeconds int    `json:"current_progress_seconds" binding:"required,gt=0"` // how many seconds in the user is
	TotalDurationSeconds   int    `json:"total_duration_seconds" binding:"required,gt=0"`   // total duration of the media in seconds
	LastWatchedAt          int64  `json:"last_watched_at"`                                  // last unix time when the playback progress was set
	WatchActionMetadata
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

// Gets all the user's watch progress, this is potentially expensive in some cases (?)
// but under normal flows, a user shouldn't have many continue watches since they are flushed
// every three months. Otherwise, detecting complete watches needs to be more accurate if
// there are too many half-watched movies/episodes
func GetWatchProgressUser(userID int64) ([]*WatchProgress, error) {
	keys, err := database.GetKeysWithPrefix(fmt.Sprintf("watch_progress|userid:%d", userID))
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
	if watchProgress.EncodedData != "" {
		data, err := providers.DecodeJsonStreamAES(watchProgress.EncodedData)
		if err != nil {
			return helpers.LogErrorWithMessage(err, "failed to decode stream data")
		}
		// sanity checks to see if tmdb ids passed in are the same as encoded data's id
		if data.SourceID != sourceID {
			return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"invalid param: source id mismatch between request and encodedData")
		}
		if data.MediaType == database.MediaTypeTVShow &&
			data.SeasonNumber != nil && data.EpisodeNumber != nil &&
			watchProgress.SeasonNumber != nil && watchProgress.EpisodeNumber != nil {
			if *data.SeasonNumber != *watchProgress.SeasonNumber {
				return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
					"invalid param: season number mismatch between request and encodedData")
			}
			if *data.EpisodeNumber != *watchProgress.EpisodeNumber {
				return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
					"invalid param: episode number mismatch between request and encodedData")
			}
		}
	}
	watchProgress.MediaType = mediaType
	watchProgress.MediaSource = mediaSource
	watchProgress.SourceID = sourceID
	watchProgress.LastWatchedAt = time.Now().Unix()
	// dyamically fill episodeID
	if mediaType == database.MediaTypeTVShow {
		if watchProgress.SeasonNumber == nil || watchProgress.EpisodeNumber == nil {
			return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"invalid param: season/episode number is nil")
		}
		showID, err := strconv.Atoi(sourceID)
		if err != nil {
			return helpers.LogErrorWithMessage(err, "failed to parse source id")
		}
		// get show details
		showDetails, err := sources.GetTVShowFromIDTMDB(showID)
		if err != nil {
			return helpers.LogErrorWithMessage(err, "failed to get show details")
		}
		watchProgress.MediaTitle = showDetails.Name
		targetEpisode, err := sources.GetEpisodeTMDB(showID,
			*watchProgress.SeasonNumber, *watchProgress.EpisodeNumber)
		if err != nil {
			return helpers.LogErrorWithMessage(err, "failed to get episode id")
		}
		watchProgress.EpisodeTitle = &targetEpisode.Name
		watchProgress.Overview = targetEpisode.Overview
		watchProgress.ReleaseDate = targetEpisode.AirDate
		watchProgress.ThumbnailURL = tmdb.GetImageURL(targetEpisode.StillPath, tmdb.W500)
		episodeIDStr := strconv.Itoa(int(targetEpisode.ID))
		watchProgress.EpisodeSourceID = &episodeIDStr
		cacheKey := fmt.Sprintf(WATCH_PROGRESS_CACHE_KEY, userID, mediaType, mediaSource, sourceID,
			*watchProgress.SeasonNumber, *watchProgress.EpisodeNumber)
		_, err = database.SetCache(cacheKey, watchProgress, watchProgressCacheTTL)
		if err != nil {
			return err
		}
		slog.Info("Watch Progress Set", "key", cacheKey)
		return nil
	} else {
		// get details and store as well, we need this for continue watching page
		// in home screen when fetched
		movieID, err := strconv.Atoi(sourceID)
		if err != nil {
			return helpers.LogErrorWithMessage(err, "failed to parse source id")
		}
		movieDetails, err := sources.GetMovieFromIDTMDB(movieID)
		if err != nil {
			return helpers.LogErrorWithMessage(err, "failed to get movie details")
		}
		watchProgress.MediaTitle = movieDetails.Title
		watchProgress.Overview = movieDetails.Overview
		watchProgress.ReleaseDate = movieDetails.ReleaseDate
		watchProgress.ThumbnailURL = tmdb.GetImageURL(movieDetails.BackdropPath, tmdb.W500)
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
