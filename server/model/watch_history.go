package model

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

const scrobbleCacheTTL = 48 * time.Hour

type WatchHistoryTVShowPayload struct {
	SeasonNumber  *int    `json:"season_number"`
	EpisodeNumber *int    `json:"episode_number"`
	EpisodeIDs    *[]int  `json:"episode_ids"` // tmdb unique id for episode
	ActionType    string  `json:"action_type" binding:"required,oneof=watch scrobble"`
	RewatchID     *int64  `json:"rewatch_id"`
	WatchedAt     *string `json:"watched_at"`
}

type WatchHistoryMoviePayload struct {
	ActionType string  `json:"action_type" binding:"required,oneof=watch scrobble"`
	WatchedAt  *string `json:"watched_at"`
}

func CreateTVShowWatchHistory(userID int64, mediaSource string, showID int, watchHistoryPayload WatchHistoryTVShowPayload) (*[]int, *[]int, error) {
	actionType := strings.ToLower(watchHistoryPayload.ActionType)
	if actionType != database.ActionWatch && actionType != database.ActionScrobble {
		return nil, nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid action type")
	}
	watchTime := time.Now().UTC()
	if watchHistoryPayload.WatchedAt != nil && *watchHistoryPayload.WatchedAt != "" {
		parsed, err := time.Parse(time.RFC3339, *watchHistoryPayload.WatchedAt)
		if err != nil {
			return nil, nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid watched_at timestamp")
		}
		watchTime = parsed.UTC()
	}
	if watchHistoryPayload.EpisodeIDs == nil || len(*watchHistoryPayload.EpisodeIDs) == 0 {
		if watchHistoryPayload.SeasonNumber == nil || watchHistoryPayload.EpisodeNumber == nil {
			return nil, nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"No valid episode ids / season episode pair provided")
		}
	}
	// 1. Upsert show
	if _, err := sources.UpsertTVShowRecordTMDB(showID); err != nil {
		return nil, nil, helpers.LogErrorWithMessage(err, "Error upserting tv show: "+mediaSource+"-"+strconv.Itoa(showID))
	}
	// 2. if using season/episode number, get episode id
	// if episodeIDs are not nil, season/episode number is ignored
	// might want to refactor to allow batch entry by season/episode number
	if watchHistoryPayload.EpisodeIDs == nil {
		showIDstr := strconv.Itoa(showID)
		episodeRecord, err := database.GetEpisodeMediaRecord(mediaSource, showIDstr, watchHistoryPayload.SeasonNumber, *watchHistoryPayload.EpisodeNumber)
		if err != nil {
			return nil, nil, helpers.LogErrorWithMessage(err, "Error getting episode record for this show, check if it exists")
		}
		targetEpisodeID, err := strconv.Atoi(episodeRecord.SourceID)
		if err != nil {
			return nil, nil, helpers.LogErrorWithMessage(err, "Error converting episode id to string")
		}
		watchHistoryPayload.EpisodeIDs = &[]int{targetEpisodeID}
	}
	// 3. Parse episode ids, check for duplicated episode ids
	episodeSet := make(map[int]struct{})
	var episodeIDs []int
	for _, id := range *watchHistoryPayload.EpisodeIDs {
		if _, exists := episodeSet[id]; exists {
			continue
		}
		episodeSet[id] = struct{}{}
		episodeIDs = append(episodeIDs, id)
	}
	// check if episode ids are in the database, and belong to the correct show
	episodeMap, invalidIDs, err := database.CheckShowEpisodesIDs(mediaSource, strconv.Itoa(showID), episodeIDs)
	if len(invalidIDs) > 0 {
		errorStr := ""
		for _, item := range invalidIDs {
			errorStr += strconv.Itoa(item) + ","
		}
		return nil, nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid Episode IDs found:"+errorStr)
	}
	if err != nil {
		return nil, nil, helpers.LogErrorWithMessage(err, "Error checking episode ids for tv show:"+mediaSource+"-"+strconv.Itoa(showID))
	}
	targetRewatchID := int64(-1)
	if watchHistoryPayload.RewatchID != nil {
		targetRewatchID = *watchHistoryPayload.RewatchID
		// check if rewatch_id exists for this user
		rewatchRecords, err := database.GetRewatchesFromSourceID(database.RecordTypeTVShow, mediaSource, strconv.Itoa(showID), userID)
		if err != nil {
			return nil, nil, helpers.LogErrorWithMessage(err, "Error getting show rewatch records")
		}
		found := false
		for _, item := range rewatchRecords {
			if item.RewatchID == targetRewatchID {
				found = true
				break
			}
		}
		if !found {
			return nil, nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Could not find this rewatch ID in the database")
		}
	}
	// 4. Get most current rewatch or create new rewatch if none if rewatch payload is empty
	if targetRewatchID == -1 {
		rewatchRecord, err := database.GetActiveRewatchFromSourceID(database.MediaTypeTVShow, mediaSource, strconv.Itoa(showID), userID)
		if err != nil {
			return nil, nil, helpers.LogErrorWithMessage(err, "Error getting active rewatch: "+mediaSource+"-"+strconv.Itoa(showID))
		}
		// add rewatch record if none exists
		if rewatchRecord == nil {
			rewatchRecord, err = InsertRewatchFromSourceID(database.MediaTypeTVShow, mediaSource,
				strconv.Itoa(showID), userID, time.Now().UTC())
			if err != nil {
				return nil, nil, helpers.LogErrorWithMessage(err, "Error creating rewatch record")
			}
		}
		targetRewatchID = rewatchRecord.RewatchID
	}
	// 5. Filter cached scrobbles, since we don't want to accidentally double insert scrobbles
	// if they are inserted within X hours of each other. Watches are fine since it's manual
	type pendingInsert struct {
		EpisodeID    int
		EpisodeIDStr string
		CacheKey     string
	}
	pendingRecords := make([]database.WatchEventsRecord, 0, len(episodeIDs))
	pendingMetadata := make([]pendingInsert, 0, len(episodeIDs))
	skippedEpisodeIDs := []int{}
	// create cache keys for scrobbles to prevent accidental duplicate inserts
	for _, episodeID := range episodeIDs {
		episodeIDStr := strconv.Itoa(episodeID)
		cacheKey := ""
		if actionType == database.ActionScrobble {
			// realistically, scrobbles should only operate on the latest rewatch session
			cacheKey = fmt.Sprintf("watch_history:scrobble:userid-%d:rewatchid-%d:%s:%s-%s", userID, targetRewatchID,
				database.RecordTypeEpisode, mediaSource, episodeIDStr)
			var cached bool
			cacheHit, err := database.GetCache(cacheKey, &cached)
			if err != nil {
				return nil, nil, helpers.LogErrorWithMessage(err, "Error checking scrobble cache")
			}
			// if cache hit and scrobble, skip insert
			if cacheHit {
				skippedEpisodeIDs = append(skippedEpisodeIDs, episodeID)
				continue
			}
		}
		int64Val, err := strconv.ParseInt(episodeMap[episodeIDStr], 10, 64)
		if err != nil {
			return nil, nil, helpers.LogErrorWithMessage(err, "Error parsing episode id")
		}
		pendingRecords = append(pendingRecords, database.WatchEventsRecord{
			RewatchID: targetRewatchID,
			RecordID:  int64Val,
			WatchType: actionType,
			WatchedAt: watchTime,
		})
		pendingMetadata = append(pendingMetadata, pendingInsert{
			EpisodeID:    episodeID,
			EpisodeIDStr: episodeIDStr,
			CacheKey:     cacheKey,
		})
	}
	if len(pendingRecords) == 0 {
		return nil, &skippedEpisodeIDs, nil
	}
	if err := database.BatchInsertWatchEvents(pendingRecords); err != nil {
		return nil, nil, helpers.LogErrorWithMessage(err, "Error inserting watch events records")
	}
	// set idempotence cache for scrobbles, 48 hours
	// only set once inserts are successful
	insertedEpisodeIDs := make([]int, len(pendingMetadata))
	for idx, meta := range pendingMetadata {
		insertedEpisodeIDs[idx] = meta.EpisodeID
		if meta.CacheKey != "" {
			if _, err := database.SetCache(meta.CacheKey, true, scrobbleCacheTTL); err != nil {
				return nil, nil, helpers.LogErrorWithMessage(err, "Error caching scrobble entry")
			}
		}
	}
	// delete all watch progress before watchTime
	// only for season/episode pair case,
	// not for batch insertion
	if watchHistoryPayload.SeasonNumber != nil && watchHistoryPayload.EpisodeNumber != nil {
		err = DeleteWatchProgress(userID, database.MediaTypeTVShow, mediaSource,
			strconv.Itoa(showID), watchHistoryPayload.SeasonNumber, watchHistoryPayload.EpisodeNumber, &watchTime)
		if err != nil {
			_ = helpers.LogErrorWithMessage(err, "Error deleting watch progress")
		}
	}
	return &insertedEpisodeIDs, &skippedEpisodeIDs, nil
}

func CreateMovieWatchHistory(userID int64, mediaSource string, sourceID int, watchHistoryPayload WatchHistoryMoviePayload) (*int, error) {
	actionType := strings.ToLower(watchHistoryPayload.ActionType)
	if actionType != database.ActionWatch && actionType != database.ActionScrobble {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid action type")
	}
	watchTime := time.Now().UTC()
	if watchHistoryPayload.WatchedAt != nil && *watchHistoryPayload.WatchedAt != "" {
		parsed, err := time.Parse(time.RFC3339, *watchHistoryPayload.WatchedAt)
		if err != nil {
			return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid watched_at timestamp")
		}
		watchTime = parsed.UTC()
	}
	// 1. Upsert movie record
	movieRecord, err := sources.UpsertMovieRecordTMDB(sourceID)
	if err != nil {
		return nil, helpers.LogErrorWithMessage(err, "Error upserting media record for "+mediaSource+"-"+strconv.Itoa(sourceID))
	}
	// 2. Get most current rewatch or create new rewatch if none exists
	rewatchRecord, err := database.GetActiveRewatchFromSourceID(database.MediaTypeMovie, mediaSource, strconv.Itoa(sourceID), userID)
	if err != nil {
		return nil, helpers.LogErrorWithMessage(err, "Error getting active rewatch: "+mediaSource+"-"+strconv.Itoa(sourceID))
	}
	// add rewatch record if none exists
	if rewatchRecord == nil {
		rewatchRecord, err = InsertRewatchFromSourceID(database.MediaTypeMovie, mediaSource,
			strconv.Itoa(sourceID), userID, time.Now().UTC())
		if err != nil {
			return nil, helpers.LogErrorWithMessage(err, "Error creating rewatch record")
		}
	}
	watchEvent := database.WatchEventsRecord{
		RewatchID: rewatchRecord.RewatchID,
		RecordID:  movieRecord.RecordID,
		WatchType: actionType,
		WatchedAt: watchTime,
	}
	if actionType == database.ActionScrobble {
		// check cache for recent scrobble
		cacheKey := fmt.Sprintf("watch_history:scrobble:userid-%d:rewatchid-%d:%s:%s-%d",
			userID, rewatchRecord.RewatchID, database.RecordTypeMovie, mediaSource, sourceID)
		var cached bool
		cacheHit, err := database.GetCache(cacheKey, &cached)
		if err != nil {
			return nil, helpers.LogErrorWithMessage(err, "Error checking scrobble cache")
		}
		// if cache hit, return without inserting
		if cacheHit {
			return nil, nil
		}
		// set cache for scrobbles to prevent accident duplicate inserts
		if _, err := database.SetCache(cacheKey, true, scrobbleCacheTTL); err != nil {
			_ = helpers.LogErrorWithMessage(err, "Error caching scrobble entry")
		}
	}
	err = database.BatchInsertWatchEvents([]database.WatchEventsRecord{watchEvent})
	if err != nil {
		return nil, helpers.LogErrorWithMessage(err, "Error inserting watch event to db: "+mediaSource+"-"+strconv.Itoa(sourceID))
	}
	// delete all watch progress before watchTime
	err = DeleteWatchProgress(userID, database.MediaTypeMovie, mediaSource,
		strconv.Itoa(sourceID), nil, nil, &watchTime)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Error deleting watch progress")
	}
	return &sourceID, nil
}

// only allow if previous rewatch isn't empty
func InsertRewatchFromSourceID(recordType string, mediaSource string, sourceID string,
	userID int64, startedAt time.Time) (*database.RewatchRecord, error) {
	has, record, err := database.GetMediaRecord(recordType, mediaSource, sourceID)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"No Media Record Found for "+recordType+":"+mediaSource+"-"+sourceID)
	}
	// get active rewatch
	activeRewatch, err := database.GetActiveRewatchFromSourceID(recordType, mediaSource, sourceID, userID)
	if err != nil {
		return nil, err
	}
	// don't insert if no active rewatch, or active rewatch is empty
	if activeRewatch != nil {
		watchEvents, err := database.GetWatchEventsFromRewatchID(activeRewatch.RewatchID, nil)
		if err != nil {
			return nil, err
		}
		if len(watchEvents) == 0 {
			return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"Current active rewatch is empty, can't create new rewatch")
		}
	}
	rewatch := database.RewatchRecord{
		UserID:    userID,
		RecordID:  record.RecordID,
		StartedAt: startedAt,
	}
	rewatchRecord, err := database.InsertRewatch(rewatch)
	if err != nil {
		return nil, err
	}
	// close previous rewatch, delete watch progress before startedAt
	if activeRewatch != nil {
		_ = database.FinishRewatch(activeRewatch.RewatchID, time.Now().UTC())
	}
	_ = DeleteWatchProgress(userID, recordType, mediaSource, sourceID, nil, nil, &startedAt)
	return rewatchRecord, nil
}
