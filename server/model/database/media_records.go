package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"hound/helpers"
	"strings"
	"time"

	"xorm.io/xorm"
)

/*
	MediaRecords - A all records from all users on the hound server
		           mostly used for archival purposes
*/

const mediaRecordsTable = "media_records"

const (
	RecordTypeTVShow  = MediaTypeTVShow
	RecordTypeMovie   = MediaTypeMovie
	RecordTypeSeason  = "season"
	RecordTypeEpisode = "episode"
)

// store user saved Records
type MediaRecord struct {
	RecordID         int64        `xorm:"pk autoincr 'record_id'" json:"record_id"`
	RecordType       string       `xorm:"unique(primary) not null 'record_type'" json:"record_type"`   // movie,tvshow,season,episode
	MediaSource      string       `xorm:"unique(primary) not null 'media_source'" json:"media_source"` // tmdb, openlibrary, etc. the main metadata provider
	SourceID         string       `xorm:"unique(primary) not null 'source_id'" json:"source_id"`       // tmdb id, episode/season tmdb id
	ParentID         *int64       `xorm:"'parent_id'" json:"parent_id"`                                // reference to fk record_id, null for movie, tvshow
	MediaTitle       string       `xorm:"text 'media_title'" json:"media_title"`                       // movie, tvshow, season or episode title
	OriginalTitle    string       `xorm:"text 'original_title'" json:"original_title"`                 // original title in release language
	OriginalLanguage string       `xorm:"text 'original_language'" json:"original_language"`
	OriginCountry    []string     `xorm:"'origin_country'" json:"origin_country"`
	ReleaseDate      string       `xorm:"'release_date'" json:"release_date"`   // 2012-12-30, for shows/seasons - first_air_date, for episodes - air_date
	LastAirDate      string       `xorm:"'last_air_date'" json:"last_air_date"` // for shows, latest episode air date
	NextAirDate      string       `xorm:"'next_air_date'" json:"next_air_date"` // for shows, next scheduled episode air date
	SeasonNumber     *int         `xorm:"'season_number'" json:"season_number"`
	EpisodeNumber    *int         `xorm:"'episode_number'" json:"episode_number"`
	SortIndex        int          `xorm:"'sort_index'" json:"sort_index"`       // not in use yet, used to sort based on user preferences
	Status           string       `xorm:"'status'" json:"status"`               // Returning Series, Released, etc.
	Overview         string       `xorm:"text 'overview'" json:"overview"`      // game of thrones is a show about ...
	Duration         int          `xorm:"'duration'" json:"duration"`           // duration/runtime in minutes
	ThumbnailURL     string       `xorm:"'thumbnail_url'" json:"thumbnail_url"` // poster image for tmdb
	BackdropURL      string       `xorm:"'backdrop_url'" json:"backdrop_url"`   // backgrounds
	StillURL         string       `xorm:"'still_url'" json:"still_url"`         // episodes, still frame for thumbnail
	Tags             *[]TagObject `json:"tags"`                                 // to store genres, tags
	UserTags         *[]TagObject `xorm:"'user_tags'" json:"user_tags"`
	FullData         []byte       `xorm:"'full_data'" json:"full_data"`       // full data from tmdb
	ContentHash      string       `xorm:"'content_hash'" json:"content_hash"` // checksum to compare changes/updates
	CreatedAt        time.Time    `xorm:"created" json:"created_at"`
	UpdatedAt        time.Time    `xorm:"updated" json:"updated_at"`
}

type MediaRecordGroup struct {
	MediaRecord  `xorm:"extends"`
	UserID       int64
	CollectionID int64
}

// For recursive entry
type MediaRecordNode struct {
	Root     *MediaRecord
	Children []*MediaRecordNode
}

func UpsertMediaRecord(mediaRecord *MediaRecord) error {
	// check if data is already in internal library
	var existingRecords []MediaRecord
	_ = databaseEngine.Table(mediaRecordsTable).Where("record_type = ?", mediaRecord.RecordType).
		Where("media_source = ?", mediaRecord.MediaSource).
		Where("source_id = ?", mediaRecord.SourceID).
		Where("season_number = ?", mediaRecord.SeasonNumber).
		Where("episode_number = ?", mediaRecord.EpisodeNumber).Find(&existingRecords)

	// source_id is either movie, show, season, or episode id
	// a key on these three should be sufficiently unique (?)
	var recordID int64
	if len(existingRecords) > 0 {
		// use existing SourceID
		recordID = existingRecords[0].RecordID
		if existingRecords[0].ContentHash != mediaRecord.ContentHash {
			// hash changed, update record in internal library
			_, err := databaseEngine.Table(mediaRecordsTable).ID(recordID).Update(&mediaRecord)
			if err != nil {
				return err
			}
		}
		// mutate in place
		*mediaRecord = existingRecords[0]
	} else {
		// insert media data to library table
		_, err := databaseEngine.Table(mediaRecordsTable).Insert(&mediaRecord)
		if err != nil {
			return err
		}
	}
	return nil
}

// Caller responsible for session, can rollback
// returns true if upserted
func UpsertMediaRecordsTrx(sess *xorm.Session, record *MediaRecord) (bool, error) {
	var recordData MediaRecord
	has, err := sess.Table(mediaRecordsTable).Where("record_type = ?", record.RecordType).
		Where("media_source = ?", record.MediaSource).
		Where("source_id = ?", record.SourceID).
		Get(&recordData)
	if err != nil {
		return false, err
	}
	if !has {
		_, err := sess.Table(mediaRecordsTable).Insert(record)
		if err != nil {
			return false, err
		}
	}
	// if has, check hash, then update if not match
	if record.ContentHash == recordData.ContentHash {
		return false, nil
	}
	_, err = sess.Table(mediaRecordsTable).ID(recordData.RecordID).Update(record)
	if err != nil {
		return false, err
	}
	return true, nil
}

func BatchUpsertMediaRecords(sess *xorm.Session, records []*MediaRecord) error {
	if len(records) == 0 {
		return nil
	}
	// don't want to run out of memory
	// we need to batch
	const batchSize = 100
	for start := 0; start < len(records); start += batchSize {
		end := start + batchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[start:end]

		if err := batchUpsertChunk(sess, batch); err != nil {
			return err
		}
	}

	return nil
}

func batchUpsertChunk(sess *xorm.Session, records []*MediaRecord) error {
	columns := []string{
		"record_type", "media_source", "source_id", "parent_id",
		"media_title", "original_title", "original_language",
		"origin_country", "release_date", "last_air_date", "next_air_date",
		"season_number", "episode_number",
		"sort_index", "status", "overview", "duration",
		"thumbnail_url", "backdrop_url", "still_url",
		"tags", "user_tags", "full_data", "content_hash", "created_at", "updated_at",
	}

	var sb strings.Builder
	sb.Grow(len(records) * 1024)

	sb.WriteString("INSERT INTO media_records (")
	sb.WriteString(strings.Join(columns, ","))
	sb.WriteString(") VALUES ")

	valArgs := make([]any, 0, len(records)*len(columns))
	argIndex := 1

	for idx, record := range records {
		if idx > 0 {
			sb.WriteString(",")
		}

		sb.WriteString("(")
		for c := range columns {
			if c > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(fmt.Sprintf("$%d", argIndex))
			argIndex++
		}
		sb.WriteString(")")

		// Truncate time to seconds to remove microseconds
		now := time.Now().UTC().Truncate(time.Second)
		valArgs = append(valArgs,
			record.RecordType,
			record.MediaSource,
			record.SourceID,
			record.ParentID,
			record.MediaTitle,
			record.OriginalTitle,
			record.OriginalLanguage,
			encodeJSON(record.OriginCountry),
			record.ReleaseDate,
			record.LastAirDate,
			record.NextAirDate,
			record.SeasonNumber,
			record.EpisodeNumber,
			record.SortIndex,
			record.Status,
			record.Overview,
			record.Duration,
			record.ThumbnailURL,
			record.BackdropURL,
			record.StillURL,
			encodeJSON(record.Tags),
			encodeJSON(record.UserTags),
			record.FullData,
			record.ContentHash,
			now, // created_at
			now, // updated_at
		)
	}
	sb.WriteString(`
ON CONFLICT (record_type, media_source, source_id)
DO UPDATE SET
	parent_id       = EXCLUDED.parent_id,
	media_title     = EXCLUDED.media_title,
	original_title  = EXCLUDED.original_title,
	original_language = EXCLUDED.original_language,
	origin_country  = EXCLUDED.origin_country,
	release_date    = EXCLUDED.release_date,
	last_air_date   = EXCLUDED.last_air_date,
	next_air_date   = EXCLUDED.next_air_date,
	sort_index      = EXCLUDED.sort_index,
	status          = EXCLUDED.status,
	overview        = EXCLUDED.overview,
	duration        = EXCLUDED.duration,
	thumbnail_url   = EXCLUDED.thumbnail_url,
	backdrop_url    = EXCLUDED.backdrop_url,
	still_url       = EXCLUDED.still_url,
	tags            = EXCLUDED.tags,
	user_tags       = EXCLUDED.user_tags,
	full_data       = EXCLUDED.full_data,
	content_hash    = EXCLUDED.content_hash,
	updated_at      = date_trunc('second', NOW())
WHERE media_records.content_hash IS DISTINCT FROM EXCLUDED.content_hash;
`)
	_, err := sess.DB().Exec(sb.String(), valArgs...)
	return err
}

func encodeJSON(v any) []byte {
	if v == nil {
		return nil
	}
	b, _ := json.Marshal(v)
	return b
}

func MarkForUpdate(recordType string, mediaSource string, sourceID string) error {
	_, err := databaseEngine.Table(mediaRecordsTable).Where("record_type = ?", recordType).
		Where("media_source = ?", mediaSource).
		Where("source_id = ?", sourceID).Update(map[string]interface{}{
		"content_hash": "xxx",
	})
	return err
}

func GetMediaRecord(recordType string, mediaSource string, sourceID string) (bool, *MediaRecord, error) {
	session := databaseEngine.NewSession()
	defer session.Close()
	return GetMediaRecordTrx(session, recordType, mediaSource, sourceID)
}

// each mediaSource, sourceID combination should be unique
// for shows, episodes, etc.
func GetMediaRecordTrx(session *xorm.Session, recordType string, mediaSource string, sourceID string) (bool, *MediaRecord, error) {
	var record MediaRecord
	if session == nil {
		return false, nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "GetMediaRecordTrx(): Session is nil")
	}
	query := session.Table(mediaRecordsTable).
		Where("record_type = ?", recordType).
		Where("media_source = ?", mediaSource).
		Where("source_id = ?", sourceID)
	has, err := query.Get(&record)
	if err != nil {
		return has, nil, err
	}
	return has, &record, nil
}
