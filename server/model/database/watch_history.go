package database

import (
	"errors"
	"hound/helpers"
	"time"
)

const (
	// for watch history
	watchEventsTable = "watch_events"
	// current watch progress
	watchProgressTable   = "watch_progress"
	rewatchesTable       = "show_rewatches"
	ActionScrobble       = "scrobble"
	ActionWatch          = "watch"
	watchEventsBatchSize = 100
)

// stores completed watches of movies or episodes
type WatchEventsRecord struct {
	WatchEventID  int64     `xorm:"pk autoincr 'watch_event_id'" json:"watch_event_id"`
	ShowRewatchID *int64    `xorm:"'show_rewatch_id'" json:"show_rewatch_id"`
	RecordID      int64     `xorm:"'record_id'" json:"record_id"`
	WatchType     string    `xorm:"'watch_type'" json:"watch_type"` // watch, scrobble
	WatchedAt     time.Time `xorm:"'watched_at'" json:"watched_at"`
	CreatedAt     time.Time `xorm:"timestampz created" json:"created_at"`
	UpdatedAt     time.Time `xorm:"timestampz updated" json:"updated_at"`
}

// Every show has one active 'rewatch' session at a time
// starting a new rewatch will unmark the watch history in the client's UI
// the active/current rewatch record is the one with the newest started_at
// this is only used for tv shows
type TVShowRewatchRecord struct {
	ShowRewatchID int64     `xorm:"pk autoincr 'show_rewatch_id'" json:"show_rewatch_id"`
	UserID        int64     `xorm:"'user_id'" json:"user_id"`
	ShowRecordID  int64     `xorm:"'show_record_id'" json:"show_record_id"` // show record id
	StartedAt     time.Time `xorm:"'started_at'" json:"rewatch_started_at"`
	FinishedAt    time.Time `xorm:"'finished_at'" json:"rewatch_finished_at"`
	CreatedAt     time.Time `xorm:"timestampz created" json:"created_at"`
	UpdatedAt     time.Time `xorm:"timestampz updated" json:"updated_at"`
}

// combination fo a watch event and a media record
type TVShowWatchEventMediaRecord struct {
	WatchEventsRecord `xorm:"extends"`
	RecordType        string `xorm:"'record_type'" json:"record_type"`      // movie,episode
	MediaSource       string `xorm:"'media_source'" json:"media_source"`    // tmdb, openlibrary, etc. the main metadata provider
	SourceID          string `xorm:"'source_id'" json:"source_id"`          // tmdb id, episode/season tmdb id
	MediaTitle        string `xorm:"text 'media_title'" json:"media_title"` // movie, tvshow, season or episode title
	SeasonNumber      *int   `xorm:"'season_number'" json:"season_number"`
	EpisodeNumber     *int   `xorm:"'episode_number'" json:"episode_number"`
	ReleaseDate       string `xorm:"'release_date'" json:"release_date"` // 2012-12-30, for shows/seasons - first_air_date, for episodes - air_date
	Overview          string `xorm:"text 'overview'" json:"overview"`    // game of thrones is a show about ...
	Duration          int    `xorm:"'duration'" json:"duration"`         // duration/runtime in minutes
}

func instantiateWatchTables() error {
	err := databaseEngine.Table(watchEventsTable).Sync2(new(WatchEventsRecord))
	if err != nil {
		return err
	}
	err = databaseEngine.Table(rewatchesTable).Sync2(new(TVShowRewatchRecord))
	if err != nil {
		return err
	}
	return nil
}

// Gets the current active rewatch
// this is the rewatch with the latest start time
// show_record_id should only be show records
func GetActiveRewatchFromSourceID(recordType string, mediaSource string, sourceID string, userID int64) (*TVShowRewatchRecord, error) {
	if recordType != RecordTypeTVShow {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Rewatches are only supported for tvshows")
	}
	// find record id
	has, record, err := GetMediaRecord(recordType, mediaSource, sourceID)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"No Media Record Found for "+recordType+":"+mediaSource+"-"+sourceID)
	}
	var rewatchRecord TVShowRewatchRecord
	has, err = databaseEngine.Table(rewatchesTable).
		Where("user_id = ?", userID).
		Where("show_record_id = ?", record.RecordID).
		Desc("started_at").
		Limit(1).
		Get(&rewatchRecord)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return &rewatchRecord, err
}

// get rewatches for a certain show, given user id and show source id
func GetShowRewatchesFromSourceID(mediaSource string, sourceID string, userID int64) ([]*TVShowRewatchRecord, error) {
	var records []*TVShowRewatchRecord
	err := databaseEngine.Table(rewatchesTable).
		Where("user_id = ?", userID).
		Where("show_record_id in (select record_id from media_records where record_type = ? and media_source = ? and source_id = ?)",
			RecordTypeTVShow, mediaSource, sourceID).
		Desc("started_at").
		Find(&records)
	return records, err
}

func InsertShowRewatchFromSourceID(recordType string, mediaSource string, sourceID string,
	userID int64, startedAt time.Time) (*TVShowRewatchRecord, error) {
	if recordType != RecordTypeTVShow {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "InsertRewatchFromSourceID(): Only tvshows are allowed")
	}
	has, record, err := GetMediaRecord(recordType, mediaSource, sourceID)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"No Media Record Found for "+recordType+":"+mediaSource+"-"+sourceID)
	}
	rewatch := TVShowRewatchRecord{
		UserID:       userID,
		ShowRecordID: record.RecordID,
		StartedAt:    startedAt,
	}
	_, err = databaseEngine.Table(rewatchesTable).Insert(&rewatch)
	if err != nil {
		return nil, err
	}
	return &rewatch, nil
}

// get rewatches joined with media_records by record_id for a certain rewatch id
func GetWatchEventsFromRewatchID(rewatchID int64, seasonNumber *int) ([]*TVShowWatchEventMediaRecord, error) {
	var records []*TVShowWatchEventMediaRecord
	sess := NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, err
	}
	sess = sess.Table(watchEventsTable).
		Where("show_rewatch_id = ?", rewatchID).
		Join("LEFT", "media_records", "media_records.record_id = watch_events.record_id").
		Omit("media_records.full_data")
	if seasonNumber != nil {
		sess = sess.Where("media_records.season_number = ?", *seasonNumber)
	}
	err := sess.Find(&records)
	return records, err
}

// batch the inserts since we also insert the full json data
// might take some memory
func BatchInsertWatchEvents(records []WatchEventsRecord) error {
	if len(records) == 0 {
		return nil
	}
	sess := NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}
	for start := 0; start < len(records); start += watchEventsBatchSize {
		end := start + watchEventsBatchSize
		if end > len(records) {
			end = len(records)
		}
		chunk := records[start:end]
		if _, err := sess.Table(watchEventsTable).Insert(&chunk); err != nil {
			_ = sess.Rollback()
			return err
		}
	}
	return sess.Commit()
}

func BatchDeleteWatchEvents(watchEventIDs []int64, userID int64, showRecordID int) error {
	if len(watchEventIDs) == 0 {
		return nil
	}
	sess := NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}
	// make sure all events belong to the user, show record
	count, err := sess.Table("watch_events").
		Join("INNER", "show_rewatches", "show_rewatches.show_rewatch_id = watch_events.show_rewatch_id").
		In("watch_events.watch_event_id", watchEventIDs).
		And("show_rewatches.user_id = ?", userID).
		And("show_rewatches.show_record_id = ?", showRecordID).
		Count()
	if err != nil {
		_ = sess.Rollback()
		return err
	}
	if count != int64(len(watchEventIDs)) {
		_ = sess.Rollback()
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Watch history does not belong to the user")
	}
	// count correct, delete
	_, err = sess.Table(watchEventsTable).In("watch_event_id", watchEventIDs).Delete(&WatchEventsRecord{})
	if err != nil {
		_ = sess.Rollback()
		return err
	}
	return sess.Commit()
}
