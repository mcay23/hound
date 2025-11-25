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
	RecordID      int64     `xorm:"pk autoincr 'record_id'" json:"record_id"`
	WatchType     string    `xorm:"'watch_type'" json:"watch_type"` // watch, scrobble
	WatchedAt     time.Time `xorm:"'watched_at'" json:"watched_at"`
	CreatedAt     time.Time `xorm:"created" json:"created_at"`
	UpdatedAt     time.Time `xorm:"updated" json:"updated_at"`
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
	CreatedAt     time.Time `xorm:"created" json:"created_at"`
	UpdatedAt     time.Time `xorm:"updated" json:"updated_at"`
}

type TVShowWatchEventMediaRecord struct {
	WatchEventsRecord `xorm:"extends"`
	// LEFT JOIN rewatches on show_rewatch_id
	UserID       int64 `xorm:"'user_id'" json:"user_id"`
	ShowRecordID int64 `xorm:"'show_record_id'" json:"show_record_id"` // reference to movie/show record id
	// LEFT JOIN media_records on record_id
	RecordType    string `xorm:"unique(primary) not null 'record_type'" json:"record_type"`   // movie,tvshow,season,episode
	MediaSource   string `xorm:"unique(primary) not null 'media_source'" json:"media_source"` // tmdb, openlibrary, etc. the main metadata provider
	SourceID      string `xorm:"unique(primary) not null 'source_id'" json:"source_id"`       // tmdb id, episode/season tmdb id
	SeasonNumber  *int   `xorm:"'season_number'" json:"season_number"`
	EpisodeNumber *int   `xorm:"'episode_number'" json:"episode_number"`
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

func InsertRewatchRecord(record *TVShowRewatchRecord) error {
	_, err := databaseEngine.Table(rewatchesTable).Insert(record)
	return err
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

// Create rewatch
func InsertRewatchFromSourceID(recordType string, mediaSource string, sourceID string, userID int64) (*TVShowRewatchRecord, error) {
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
		StartedAt:    time.Now().UTC(),
	}
	if err := InsertRewatchRecord(&rewatch); err != nil {
		return nil, err
	}
	return &rewatch, nil
}

func GetWatchEventsFromSourceID(recordType string, mediaSource string, sourceID string, userID int64) ([]TVShowWatchEventMediaRecord, error) {
	// cannot search by seasons for now
	if recordType != RecordTypeMovie && recordType != RecordTypeTVShow &&
		recordType != RecordTypeEpisode {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Record type not supported "+recordType)
	}
	var records []TVShowWatchEventMediaRecord

	return records, nil
}

func BatchInsertWatchEvents(records []WatchEventsRecord) error {
	if len(records) == 0 {
		return nil
	}
	session := NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return err
	}
	for start := 0; start < len(records); start += watchEventsBatchSize {
		end := start + watchEventsBatchSize
		if end > len(records) {
			end = len(records)
		}
		chunk := records[start:end]
		if _, err := session.Table(watchEventsTable).Insert(&chunk); err != nil {
			_ = session.Rollback()
			return err
		}
	}
	return session.Commit()
}
