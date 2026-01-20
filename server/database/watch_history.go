package database

import (
	"errors"
	"hound/helpers"
	"time"

	"xorm.io/xorm"
)

const (
	// for watch history
	watchEventsTable = "watch_events"
	// current watch progress
	watchProgressTable   = "watch_progress"
	rewatchesTable       = "rewatches"
	ActionScrobble       = "scrobble"
	ActionWatch          = "watch"
	watchEventsBatchSize = 100
)

// stores completed watches of movies or episodes
type WatchEventsRecord struct {
	WatchEventID int64     `xorm:"pk autoincr 'watch_event_id'" json:"watch_event_id"`
	RewatchID    int64     `xorm:"index 'rewatch_id'" json:"rewatch_id"`
	RecordID     int64     `xorm:"index 'record_id'" json:"record_id"`
	WatchType    string    `xorm:"'watch_type'" json:"watch_type"` // watch, scrobble
	WatchedAt    time.Time `xorm:"'watched_at'" json:"watched_at"`
	CreatedAt    time.Time `xorm:"timestampz created" json:"created_at"`
	UpdatedAt    time.Time `xorm:"timestampz updated" json:"updated_at"`
}

// Every show has one active 'rewatch' session at a time
// starting a new rewatch will unmark the watch history in the client's UI
// the active/current rewatch record is the one with the newest started_at
// this is only used for tv shows
type RewatchRecord struct {
	RewatchID  int64     `xorm:"pk autoincr 'rewatch_id'" json:"rewatch_id"`
	UserID     int64     `xorm:"index 'user_id'" json:"user_id"`
	RecordID   int64     `xorm:"index 'record_id'" json:"record_id"` // record id for movie/show
	StartedAt  time.Time `xorm:"'started_at'" json:"rewatch_started_at"`
	FinishedAt time.Time `xorm:"'finished_at'" json:"rewatch_finished_at"`
	CreatedAt  time.Time `xorm:"timestampz created" json:"created_at"`
	UpdatedAt  time.Time `xorm:"timestampz updated" json:"updated_at"`
}

// combination fo a watch event and a media record
type WatchEventMediaRecord struct {
	WatchEventsRecord `xorm:"extends"`
	RecordType        string `xorm:"'record_type'" json:"record_type"`      // movie,episode
	MediaSource       string `xorm:"'media_source'" json:"media_source"`    // tmdb, openlibrary, etc. the main metadata provider
	SourceID          string `xorm:"'source_id'" json:"source_id"`          // tmdb id, episode/season tmdb id
	MediaTitle        string `xorm:"text 'media_title'" json:"media_title"` // movie, tvshow, season or episode title
	SeasonNumber      *int   `xorm:"'season_number'" json:"season_number,omitempty"`
	EpisodeNumber     *int   `xorm:"'episode_number'" json:"episode_number,omitempty"`
	ReleaseDate       string `xorm:"'release_date'" json:"release_date"` // 2012-12-30, for shows/seasons - first_air_date, for episodes - air_date
	Overview          string `xorm:"text 'overview'" json:"overview"`    // game of thrones is a show about ...
	Duration          int    `xorm:"'duration'" json:"duration"`         // duration/runtime in minutes
}

func instantiateWatchTables() error {
	err := databaseEngine.Table(watchEventsTable).Sync2(new(WatchEventsRecord))
	if err != nil {
		return err
	}
	err = databaseEngine.Table(rewatchesTable).Sync2(new(RewatchRecord))
	if err != nil {
		return err
	}
	return nil
}

// gets a users watch activity (list of events) between start and end time
// if nil, return all activity
func GetWatchActivity(userID int64, startTime *time.Time, endTime *time.Time, limit int, offset int) ([]*WatchEventMediaRecord, int64, error) {
	var records []*WatchEventMediaRecord
	query := func() *xorm.Session {
		sess := databaseEngine.Table(watchEventsTable).Alias("we").
			Join("INNER", rewatchesTable+" r", "r.rewatch_id = we.rewatch_id").
			Join("INNER", mediaRecordsTable+" mr", "mr.record_id = we.record_id").
			Where("r.user_id = ?", userID)
		if startTime != nil {
			sess = sess.Where("we.watched_at >= ?", *startTime)
		}
		if endTime != nil {
			sess = sess.Where("we.watched_at <= ?", *endTime)
		}
		return sess
	}
	totalRecords, err := query().Count()
	if err != nil {
		return nil, 0, err
	}
	sess := query()
	if limit > 0 {
		sess = sess.Limit(limit, offset)
	}
	sess = sess.OrderBy("we.watched_at DESC, we.watch_event_id DESC")
	err = sess.Find(&records)
	return records, totalRecords, err
}

// Gets the current active rewatch
// this is the rewatch with the latest start time
func GetActiveRewatchFromSourceID(recordType string, mediaSource string, sourceID string, userID int64) (*RewatchRecord, error) {
	// find record id
	has, record, err := GetMediaRecord(recordType, mediaSource, sourceID)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	var rewatchRecord RewatchRecord
	has, err = databaseEngine.Table(rewatchesTable).
		Where("user_id = ?", userID).
		Where("record_id = ?", record.RecordID).
		Desc("started_at").
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
func GetRewatchesFromSourceID(recordType string, mediaSource string, sourceID string, userID int64) ([]*RewatchRecord, error) {
	var records []*RewatchRecord
	err := databaseEngine.Table(rewatchesTable).
		Where("user_id = ?", userID).
		Where("record_id in (select record_id from media_records where record_type = ? and media_source = ? and source_id = ?)",
			recordType, mediaSource, sourceID).
		Desc("started_at").
		Find(&records)
	return records, err
}

// useful to answer, what's the 10 most recent unique movies/shows the user has watched
// we want to get the parent of the media record, so that we can group by show or movie
func GetUniqueWatchParents(userID int64, limit int, offset int, after time.Time) ([]*WatchEventMediaRecord, error) {
	var records []*WatchEventMediaRecord
	err := databaseEngine.
		Table("watch_events we").
		Join("INNER", "rewatches r", "r.rewatch_id = we.rewatch_id").
		Join("INNER", "media_records mr", "mr.record_id = we.record_id").
		Join("LEFT", "media_records season",
			"season.record_id = mr.parent_id AND mr.record_type = 'episode'").
		Join("LEFT", "media_records show",
			"show.record_id = season.parent_id AND season.record_type = 'season'").
		Where("r.user_id = ?", userID).
		Where("we.watched_at > ?", after).
		Omit("mr.full_data").
		Select(`
			DISTINCT ON (
				COALESCE(show.record_type,  mr.record_type),
				COALESCE(show.media_source, mr.media_source),
				COALESCE(show.source_id,    mr.source_id)
			)
			we.*,
			COALESCE(show.record_id,    mr.record_id)      AS record_id,
			COALESCE(show.record_type,  mr.record_type)    AS record_type,
			COALESCE(show.media_source, mr.media_source)   AS media_source,
			COALESCE(show.source_id,    mr.source_id)      AS source_id,
			COALESCE(show.media_title,  mr.media_title)    AS media_title,
			COALESCE(show.overview,     mr.overview)       AS overview,
			COALESCE(show.thumbnail_url, mr.thumbnail_url) AS thumbnail_url,
			COALESCE(show.backdrop_url,  mr.backdrop_url)  AS backdrop_url,
			COALESCE(show.release_date,  mr.release_date)  AS release_date,
			mr.season_number,
			mr.episode_number,
			mr.duration
		`).
		OrderBy(`
			COALESCE(show.record_type,  mr.record_type),
			COALESCE(show.media_source, mr.media_source),
			COALESCE(show.source_id,    mr.source_id),
			we.watched_at DESC,
			we.watch_event_id DESC
		`).Limit(limit, offset).
		Find(&records)
	return records, err
}

func InsertRewatch(rewatch RewatchRecord) (*RewatchRecord, error) {
	_, err := databaseEngine.Table(rewatchesTable).Insert(&rewatch)
	if err != nil {
		return nil, err
	}
	return &rewatch, nil
}

func FinishRewatch(rewatchID int64, finishedAt time.Time) error {
	_, err := databaseEngine.Table(rewatchesTable).
		Where("rewatch_id = ?", rewatchID).
		Update(RewatchRecord{FinishedAt: finishedAt})
	return err
}

// get rewatches joined with media_records by record_id for a certain rewatch id
func GetWatchEventsFromRewatchID(rewatchID int64, seasonNumber *int) ([]*WatchEventMediaRecord, error) {
	var records []*WatchEventMediaRecord
	sess := NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, err
	}
	sess = sess.Table(watchEventsTable).
		Where("rewatch_id = ?", rewatchID).
		Join("LEFT", "media_records", "media_records.record_id = watch_events.record_id").
		Omit("media_records.full_data")
	if seasonNumber != nil {
		sess = sess.Where("media_records.season_number = ?", *seasonNumber)
	}
	sess = sess.OrderBy("watch_events.watched_at DESC, watch_events.watch_event_id DESC")
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

func BatchDeleteWatchEvents(watchEventIDs []int64, userID int64, recordID int) error {
	if len(watchEventIDs) == 0 {
		return nil
	}
	sess := NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}
	// make sure all events belong to the user, record
	count, err := sess.Table("watch_events").
		Join("INNER", "rewatches", "rewatches.rewatch_id = watch_events.rewatch_id").
		In("watch_events.watch_event_id", watchEventIDs).
		And("rewatches.user_id = ?", userID).
		And("rewatches.record_id = ?", recordID).
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
