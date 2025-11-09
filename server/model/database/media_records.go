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

// store user saved Records
type MediaRecords struct {
	RecordID     int64        `xorm:"pk autoincr 'record_id'" json:"record_id"`
	MediaType    string       `xorm:"unique(primary) not null" json:"media_type"`            // books,tvshows, etc.
	MediaSource  string       `xorm:"unique(primary) not null" json:"media_source"`          // tmdb, openlibrary, etc
	SourceID     string       `xorm:"unique(primary) not null 'source_id'" json:"source_id"` // tmdb id, etc.
	MediaTitle   string       `json:"media_title"`                                           // game of thrones, etc.
	ReleaseDate  string       `json:"release_date"`                                          // 2012, 2013
	Description  []byte       `json:"description"`                                           // game of thrones is a show about ...
	FullData     []byte       `json:"data"`                                                  // full data
	ThumbnailURL *string      `xorm:"'thumbnail_url'" json:"thumbnail_url"`                  // url for media thumbnails
	Tags         *[]TagObject `json:"tags"`                                                  // to store genres, tags
	UserTags     *[]TagObject `json:"user_tags"`
	CreatedAt    time.Time    `xorm:"created" json:"created_at"`
	UpdatedAt    time.Time    `xorm:"updated" json:"updated_at"`
}

type MediaRecordGroup struct {
	MediaRecords `xorm:"extends"`
	UserID       int64
	CollectionID int64
}

func AddMediaRecord(mediaRecord *MediaRecords) (int64, error) {
	// check if data is already in internal library
	var existingRecords []MediaRecords
	_ = databaseEngine.Table(mediaRecordsTable).Where("media_type = ?", mediaRecord.MediaType).
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

func GetRecordID(mediaType string, mediaSource string, sourceID string) (*int64, error) {
	// using current handling, errors are usually ignored by the caller
	var record MediaRecords
	has, err := databaseEngine.Table(mediaRecordsTable).Where("media_type = ?", mediaType).
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
