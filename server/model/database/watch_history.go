package database

import "time"

const (
	// for watch history
	watchEventsTable = "watch_events"
	// current watch progress
	watchProgressTable = "watch_progress"
	rewatchesTable     = "rewatches"
)

// stores completed watches of movies or episodes
type WatchEventsRecord struct {
	WatchID   int64     `xorm:"pk autoincr 'watch_id'" json:"id"`
	UserID    int64     `xorm:"'user_id'" json:"user_id"`
	RecordID  int64     `xorm:"'record_id'" json:"record_id"` // reference to movie/episode record id
	RewatchID int64     `xorm:"'rewatch_id'" json:"rewatch_id"`
	WatchedAt time.Time `xorm:"'watched_at'" json:"watched_at"`
	CreatedAt time.Time `xorm:"created" json:"created_at"`
	UpdatedAt time.Time `xorm:"updated" json:"updated_at"`
}

// Every show/movie has one active 'rewatch' session at a time
// starting a new rewatch will unmark the watch history in the client's UI
type RewatchRecord struct {
	RewatchID  int64     `xorm:"pk autoincr 'rewatch_id'" json:"id"`
	UserID     int64     `xorm:"'user_id'" json:"user_id"`
	RecordID   int64     `xorm:"'record_id'" json:"record_id"` // reference to movie/show record id
	StartedAt  time.Time `xorm:"'started_at'" json:"started_at"`
	FinishedAt time.Time `xorm:"'finished_at'" json:"finished_at"`
	CreatedAt  time.Time `xorm:"created" json:"created_at"`
	UpdatedAt  time.Time `xorm:"updated" json:"updated_at"`
}

func instantiateWatchTables() error {
	return databaseEngine.Sync2(new(WatchEventsRecord), new(RewatchRecord))
}

func InsertWatchEventsRecord(record *WatchEventsRecord) error {
	_, err := databaseEngine.Table(watchEventsTable).Insert(record)
	return err
}

func InsertRewatchRecord(record *RewatchRecord) error {
	_, err := databaseEngine.Table(rewatchesTable).Insert(record)
	return err
}

// given media_type, media_source, source_id, user_id
// get record id and rewatches
// func GetRewatchesFromSourceID(media_type string, media_source string, source_id string, user_id string) (*[]RewatchRecord, error) {

// }

// Gets the current active rewatch
// this is the rewatch with the latest start time
func GetActiveRewatchFromSourceID(media_type string, media_source string, source_id string, user_id string) (*RewatchRecord, error) {
	var record RewatchRecord
	has, err := databaseEngine.Table(rewatchesTable).Where("user_id = ?", user_id).
		Where("record_id = ?", source_id).Get(&record)
	if !has {
		return nil, nil
	}
	return &record, err
}

func UpdateWatchEventsRecord(record *WatchEventsRecord) error {
	_, err := databaseEngine.ID(record.WatchID).Update(record)
	return err
}
