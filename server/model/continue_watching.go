package model

import (
	"errors"
	"hound/database"
	"hound/helpers"
	"hound/model/sources"
	"sort"
	"strings"
	"time"
)

/*
Next watch action determines the next media to watch is given a show,
eg. for tmdb-1234 -> play next episode, or resume episode X.
For movies, this will just be resume if progress exists

Continue Watching collects Next watch actions and collates them to
show in the home screen, ala Netflix. The principle is one tile per,
show/movie, with the next watch action as the tile action.
*/

const (
	WatchActionTypeNextEpisode = "next_episode"
	WatchActionTypeResume      = "resume"
)

type NextEpisode struct {
	SeasonNumber  *int    `json:"season_number,omitempty"`
	EpisodeNumber *int    `json:"episode_number,omitempty"`
	EpisodeID     *string `json:"episode_id,omitempty"`
	WatchActionMetadata
}

type WatchAction struct {
	MediaType       string         `json:"media_type"`
	MediaSource     string         `json:"media_source"`
	SourceID        string         `json:"source_id"`
	WatchActionType string         `json:"watch_action_type"`        // next_episode or resume
	NextEpisode     *NextEpisode   `json:"next_episode,omitempty"`   // only for next episode watch action type
	WatchProgress   *WatchProgress `json:"watch_progress,omitempty"` // only for resume watch action type
}

// info about the movie/episode
type WatchActionMetadata struct {
	MediaTitle      string  `json:"media_title"`                 // movie/show title
	EpisodeSourceID *string `json:"episode_source_id,omitempty"` // episode source id
	EpisodeTitle    *string `json:"episode_title,omitempty"`
	SeasonNumber    *int    `json:"season_number,omitempty"`  // only defined for shows
	EpisodeNumber   *int    `json:"episode_number,omitempty"` // only defined for shows
	Overview        string  `json:"overview"`
	ReleaseDate     string  `json:"release_date"`
	ThumbnailURL    string  `json:"thumbnail_url"`
}

// A nil watch action means we don't have a next watch action
func GetNextWatchAction(userID int64, mediaType string, mediaSource string, sourceID string) (*WatchAction, error) {
	if mediaSource != sources.MediaSourceTMDB {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Invalid media source, only tmdb is supported at this time")
	}
	if mediaType == database.MediaTypeMovie {
		return getNextWatchActionMovie(userID, mediaSource, sourceID)
	}
	if mediaType == database.MediaTypeTVShow {
		return getNextWatchActionTVShow(userID, mediaSource, sourceID)
	}
	return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
		"Invalid media type, only movie and tv show are supported at this time")
}

// gets the 10 most recent continue watching actions for the user
// not the most efficient performance-wise, hopefully fine for now
// evaluate if performance will be an issue for larger databases
// in the future
func GetContinueWatching(userID int64) ([]WatchAction, error) {
	// create set of tmdb ids, unix time for last watched
	tmdbIDSet := make(map[string]int64)
	// first, get all watch progress for the user
	watchProgress, err := GetWatchProgressUser(userID)
	if err != nil {
		return nil, err
	}
	for _, item := range watchProgress {
		key := item.MediaType + ":" + item.MediaSource + ":" + item.SourceID
		value, ok := tmdbIDSet[key]
		if !ok || value < item.LastWatchedAt {
			tmdbIDSet[key] = item.LastWatchedAt
		}
	}
	// get 10 most recent watch events, within 3 months
	now := time.Now()
	watchEvents, err := database.GetUniqueWatchParents(userID, 10, 0, now.AddDate(0, -3, 0))
	if err != nil {
		return nil, err
	}
	for _, item := range watchEvents {
		key := item.RecordType + ":" + item.MediaSource + ":" + item.SourceID
		value, ok := tmdbIDSet[key]
		if !ok || value < item.WatchedAt.Unix() {
			tmdbIDSet[key] = item.WatchedAt.Unix()
		}
	}
	// sort the map by time
	var sortedKeys []string
	for key := range tmdbIDSet {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Slice(sortedKeys, func(i, j int) bool {
		return tmdbIDSet[sortedKeys[i]] > tmdbIDSet[sortedKeys[j]]
	})
	// get the top 10
	var watchActions []WatchAction
	count := 0
	for _, key := range sortedKeys {
		if count >= 10 {
			break
		}
		mediaType := strings.Split(key, ":")[0]
		mediaSource := strings.Split(key, ":")[1]
		sourceID := strings.Split(key, ":")[2]
		switch mediaType {
		case database.MediaTypeMovie:
			watchAction, err := getNextWatchActionMovie(userID, mediaSource, sourceID)
			if err != nil {
				return nil, err
			}
			// completed movie, no next watch action
			if watchAction == nil {
				continue
			}
			watchActions = append(watchActions, *watchAction)
			count++
		case database.MediaTypeTVShow:
			watchAction, err := getNextWatchActionTVShow(userID, mediaSource, sourceID)
			if err != nil {
				return nil, err
			}
			// last episode, no next watch action
			if watchAction == nil {
				continue
			}
			watchActions = append(watchActions, *watchAction)
			count++
		}
	}
	return watchActions, nil
}

// simply checks the progress table for the next watch action
// if there are no incomplete watches, there is no next action
// in the future, we might want to incorporate sequels as the next action
func getNextWatchActionMovie(userID int64, mediaSource string, sourceID string) (*WatchAction, error) {
	watchProgress, err := GetWatchProgress(userID, database.MediaTypeMovie, mediaSource, sourceID, nil)
	if err != nil {
		return nil, err
	}
	if len(watchProgress) == 0 {
		return nil, nil
	}
	watchAction := WatchAction{
		MediaType:       database.MediaTypeMovie,
		MediaSource:     mediaSource,
		SourceID:        sourceID,
		WatchActionType: WatchActionTypeResume,
		WatchProgress:   watchProgress[0],
	}
	return &watchAction, nil
}

func getNextWatchActionTVShow(userID int64, mediaSource string, showID string) (*WatchAction, error) {
	// get last watched episode
	rewatch, err := database.GetActiveRewatchFromSourceID(database.MediaTypeTVShow,
		mediaSource, showID, userID)
	if err != nil {
		return nil, err
	}
	// compare last watch and last resume to see which is newer
	// if resume is newer -> suggest resume
	// if watch is newer -> suggest next episode
	lastCompleteWatch := int64(0)
	lastResume := int64(0)
	var targetWatchEvent *database.WatchEventMediaRecord
	if rewatch != nil {
		watchHistory, err := database.GetWatchEventsFromRewatchID(rewatch.RewatchID, nil)
		if err != nil {
			return nil, err
		}
		// should already be sorted by watched_at desc
		if len(watchHistory) > 0 && watchHistory[0] != nil {
			lastCompleteWatch = watchHistory[0].WatchedAt.Unix()
			targetWatchEvent = watchHistory[0]
		}
	}
	watchProgress, err := GetWatchProgress(userID, database.MediaTypeTVShow, mediaSource, showID, nil)
	if err != nil {
		return nil, err
	}
	if len(watchProgress) > 0 && watchProgress[0] != nil {
		// sort based on watched_at desc
		sort.Slice(watchProgress, func(i, j int) bool {
			return watchProgress[i].LastWatchedAt > watchProgress[j].LastWatchedAt
		})
		lastResume = watchProgress[0].LastWatchedAt
	}
	// no data at all, no action
	if lastCompleteWatch == 0 && lastResume == 0 {
		return nil, nil
	}
	watchAction := WatchAction{
		MediaType:   database.MediaTypeTVShow,
		MediaSource: mediaSource,
		SourceID:    showID,
	}
	// at least one of these will exist at this point
	// if lastcompletewatch exists, we know the show has been upserted
	// so we search there instead of making a tmdb network call
	if lastCompleteWatch > lastResume {
		// find the next episode
		currentSeason, err :=
			database.GetEpisodeMediaRecords(mediaSource, showID, targetWatchEvent.SeasonNumber, nil)
		if err != nil {
			return nil, err
		}
		var nextEpisodeRecord *database.MediaRecord
		for index, episode := range currentSeason {
			if *episode.EpisodeNumber == *targetWatchEvent.EpisodeNumber {
				if len(currentSeason) > index+1 {
					nextEpisodeRecord = &currentSeason[index+1]
				}
				break
			}
		}
		// stil not found, check next season
		if nextEpisodeRecord == nil {
			nextSeasonNumber := *targetWatchEvent.SeasonNumber + 1
			nextSeason, err :=
				database.GetEpisodeMediaRecords(mediaSource, showID, &nextSeasonNumber, nil)
			if err != nil {
				return nil, err
			}
			if len(nextSeason) > 0 {
				nextEpisodeRecord = &nextSeason[0]
			}
		}
		// no next episode, very end of show
		if nextEpisodeRecord == nil {
			return nil, nil
		}
		nextEp := NextEpisode{
			SeasonNumber:  nextEpisodeRecord.SeasonNumber,
			EpisodeNumber: nextEpisodeRecord.EpisodeNumber,
			EpisodeID:     &nextEpisodeRecord.SourceID,
		}
		// get show record
		found, showRecord, err := database.GetMediaRecord(database.RecordTypeTVShow, mediaSource, showID)
		if err != nil {
			return nil, err
		}
		// shouldn't happen since every completed watch has to have a show record
		if !found {
			return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"Show record not found in database, contact dev, this shouldn't happen")
		}
		nextEp.MediaTitle = showRecord.MediaTitle
		nextEp.Overview = nextEpisodeRecord.Overview
		nextEp.ReleaseDate = nextEpisodeRecord.ReleaseDate
		nextEp.ThumbnailURL = nextEpisodeRecord.StillURL
		nextEp.EpisodeTitle = &nextEpisodeRecord.MediaTitle
		watchAction.NextEpisode = &nextEp
		watchAction.WatchActionType = WatchActionTypeNextEpisode
	} else {
		// at this point, watch progress should exist
		watchAction.WatchProgress = watchProgress[0]
		watchAction.WatchActionType = WatchActionTypeResume
	}
	return &watchAction, nil
}
