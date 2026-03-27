package model

import (
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/providers"
	"github.com/mcay23/hound/sources"

	tmdb "github.com/cyruzin/golang-tmdb"
)

const watchProgressCacheTTL = 90 * 24 * time.Hour // 90 days
const (
	PlayerWeb       = "web"
	PlayerExoplayer = "exoplayer"
	PlayerMPV       = "mpv"
)

var SupportedPlayers = []string{PlayerWeb, PlayerExoplayer, PlayerMPV}

/*
Watch progress is not stored in the db because it's not deemed as critical
Cache is preferred for now so media records don't have to be inserted for watch progress
on new movies/episodes. Media record is only inserted on completed watch history events
This is intentional to prevent downloading too much metadata for movies users peek at for
15 minutes and never watch again, for example.
*/
type WatchProgress struct {
	ClientPlatform         string          `json:"client_platform,omitempty"`                        // android-tv, etc.
	MediaType              string          `json:"media_type"`                                       // "movie" or "tvshow"
	MediaSource            string          `json:"media_source"`                                     // "tmdb"
	SourceID               string          `json:"source_id"`                                        // movie/show source id
	StreamProtocol         string          `json:"stream_protocol"`                                  // p2p, http, local, etc.
	EncodedData            string          `json:"encoded_data"`                                     // for hound-proxied sources
	SourceURI              string          `json:"source_uri"`                                       // magnet, http link, local path
	CurrentProgressSeconds int             `json:"current_progress_seconds" binding:"required,gt=0"` // how many seconds in the user is
	TotalDurationSeconds   int             `json:"total_duration_seconds" binding:"required,gt=0"`   // total duration of the media in seconds
	LastWatchedAt          int64           `json:"last_watched_at"`                                  // last unix time when the playback progress was set
	PlayerSettings         *PlayerSettings `json:"player_settings,omitempty"`                        // player, audio, subtitle selections, etc.
	WatchActionMetadata
}

// settings to help resume playback
type PlayerSettings struct {
	Player       string `json:"player,omitempty"` // mpv, exoplayer, etc.
	AudioIdx     *int   `json:"audio_idx,omitempty"`
	AudioLang    string `json:"audio_lang,omitempty"`
	SubtitleIdx  *int   `json:"subtitle_idx,omitempty"`
	SubtitleLang string `json:"subtitle_lang,omitempty"`
	ResizeMode   string `json:"resize_mode,omitempty"` // whether the video is zoomed to fill/normal, etc.
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
		return fmt.Errorf("invalid param: watchProgress is nil: %w", internal.BadRequestError)
	}
	if watchProgress.CurrentProgressSeconds > watchProgress.TotalDurationSeconds {
		return fmt.Errorf("invalid param: current progress is greater than total video duration: %w", internal.BadRequestError)
	}
	watchProgress.ClientPlatform = strings.ToLower(watchProgress.ClientPlatform)
	if watchProgress.ClientPlatform != "" && !slices.Contains(SupportedClientPlatforms, watchProgress.ClientPlatform) {
		return fmt.Errorf("invalid param: X-Client-Platform is invalid: %s: %w", watchProgress.ClientPlatform, internal.BadRequestError)
	}
	// validate player settings correct
	if watchProgress.PlayerSettings != nil {
		watchProgress.PlayerSettings.Player = strings.ToLower(watchProgress.PlayerSettings.Player)
		if watchProgress.PlayerSettings.Player != "" && !slices.Contains(SupportedPlayers, watchProgress.PlayerSettings.Player) {
			return fmt.Errorf("invalid param: player %s is not supported: %w", watchProgress.PlayerSettings.Player, internal.BadRequestError)
		}
	}
	if watchProgress.EncodedData != "" {
		data, err := providers.DecodeJsonStreamAES(watchProgress.EncodedData)
		if err != nil {
			return fmt.Errorf("failed to decode stream data: %w", err)
		}
		// sanity checks to see if tmdb ids passed in are the same as encoded data's id
		if data.SourceID != sourceID {
			return fmt.Errorf("invalid param: source id mismatch between request and encodedData: %w", internal.BadRequestError)
		}
		if data.MediaType == database.MediaTypeTVShow &&
			data.SeasonNumber != nil && data.EpisodeNumber != nil &&
			watchProgress.SeasonNumber != nil && watchProgress.EpisodeNumber != nil {
			if *data.SeasonNumber != *watchProgress.SeasonNumber {
				return fmt.Errorf("invalid param: season number mismatch between request and encodedData: %w", internal.BadRequestError)
			}
			if *data.EpisodeNumber != *watchProgress.EpisodeNumber {
				return fmt.Errorf("invalid param: episode number mismatch between request and encodedData: %w", internal.BadRequestError)
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
			return fmt.Errorf("invalid param: season/episode number is nil: %w", internal.BadRequestError)
		}
		showID, err := strconv.Atoi(sourceID)
		if err != nil {
			return fmt.Errorf("failed to parse source id: %w", err)
		}
		// get show details
		showDetails, err := sources.GetTVShowFromIDTMDB(showID)
		if err != nil {
			return fmt.Errorf("failed to get show details: %w", err)
		}
		watchProgress.MediaTitle = showDetails.Name
		targetEpisode, err := sources.GetEpisodeTMDB(showID,
			*watchProgress.SeasonNumber, *watchProgress.EpisodeNumber)
		if err != nil {
			return fmt.Errorf("failed to get episode id: %w", err)
		}
		watchProgress.EpisodeTitle = &targetEpisode.Name
		watchProgress.Overview = targetEpisode.Overview
		watchProgress.ReleaseDate = targetEpisode.AirDate
		watchProgress.ThumbnailURI = tmdb.GetImageURL(targetEpisode.StillPath, tmdb.W500)
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
			return fmt.Errorf("failed to parse source id: %w", err)
		}
		movieDetails, err := sources.GetMovieFromIDTMDB(movieID)
		if err != nil {
			return fmt.Errorf("failed to get movie details: %w", err)
		}
		watchProgress.MediaTitle = movieDetails.Title
		watchProgress.Overview = movieDetails.Overview
		watchProgress.ReleaseDate = movieDetails.ReleaseDate
		watchProgress.ThumbnailURI = tmdb.GetImageURL(movieDetails.BackdropPath, tmdb.W500)
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
			slog.Debug("failed to delete watch progress", "key", key, "error", deleteError)
		}
	}
	return deleteError
}
