package database

import (
	"bytes"
	"errors"
	"hound/helpers"
	"time"
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
	RecordType       string       `xorm:"unique(primary) not null" json:"record_type"`           // movie,tvshow,season,episode
	MediaSource      string       `xorm:"unique(primary) not null" json:"media_source"`          // tmdb, openlibrary, etc. the main metadata provider
	SourceID         string       `xorm:"unique(primary) not null 'source_id'" json:"source_id"` // tmdb id, etc.
	ParentID         int64        `xorm:"'parent_id'" json:"parent_id"`                          // reference to fk record_id, null for movie, tvshow
	MediaTitle       string       `json:"media_title"`                                           // game of thrones, etc.
	OriginalTitle    string       `json:"original_title"`                                        // original title in release language
	OriginalLanguage string       `json:"original_language"`
	OriginCountry    []string     `json:"origin_country"`
	ReleaseDate      string       `json:"release_date"` // 2012_12_30
	SeasonNumber     int          `json:"season_number"`
	EpisodeNumber    int          `json:"episode_number"`
	SortIndex        int          `json:"sort_index"`                           // not in use yet, used to sort based on user preferences
	Status           string       `json:"status"`                               // Returning Series, Released, etc.
	Overview         string       `json:"overview"`                             // game of thrones is a show about ...
	Duration         int          `json:"duration"`                             // duration in minutes
	ThumbnailURL     string       `xorm:"'thumbnail_url'" json:"thumbnail_url"` // poster image for tmdb
	BackdropURL      string       `xorm:"'backdrop_url'" json:"backdrop_url"`   // backgrounds
	StillURL         string       `xorm:"'still_url'" json:"still_url"`         // episodes, still frame for thumbnail
	Tags             *[]TagObject `json:"tags"`                                 // to store genres, tags
	UserTags         *[]TagObject `json:"user_tags"`
	FullData         []byte       `json:"data"` // full data from tmdb
	CreatedAt        time.Time    `xorm:"created" json:"created_at"`
	UpdatedAt        time.Time    `xorm:"updated" json:"updated_at"`
}

type MediaRecordGroup struct {
	MediaRecord  `xorm:"extends"`
	UserID       int64
	CollectionID int64
}

func AddMediaRecord(mediaRecord *MediaRecord) (int64, error) {
	// check if data is already in internal library
	var existingRecords []MediaRecord
	_ = databaseEngine.Table(mediaRecordsTable).Where("record_type = ?", mediaRecord.RecordType).
		Where("media_source = ?", mediaRecord.MediaSource).
		Where("source_id = ?", mediaRecord.SourceID).Find(&existingRecords)

	var recordID int64
	if len(existingRecords) > 0 {
		// use existing SourceID
		recordID = existingRecords[0].RecordID
		if !bytes.Equal(existingRecords[0].FullData, mediaRecord.FullData) {
			// TODO update db with new data
		}
	} else {
		// insert media data to library table
		_, err := databaseEngine.Table(mediaRecordsTable).Insert(mediaRecord)
		if err != nil {
			return -1, err
		}
		recordID = mediaRecord.RecordID
	}
	return recordID, nil
}

func GetRecordID(recordType string, mediaSource string, sourceID string) (*int64, error) {
	// using current handling, errors are usually ignored by the caller
	var record MediaRecord
	has, err := databaseEngine.Table(mediaRecordsTable).Where("record_type = ?", recordType).
		Where("media_source = ?", mediaSource).
		Where("source_id = ?", sourceID).Get(&record)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"GetRecordID(): No matching record in internal library")
	}
	return &record.RecordID, nil
}
