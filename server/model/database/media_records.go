package database

import (
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
	SourceID         string       `xorm:"unique(primary) not null 'source_id'" json:"source_id"` // tmdb id, episode/season tmdb id
	ParentID         int64        `xorm:"'parent_id'" json:"parent_id"`                          // reference to fk record_id, null for movie, tvshow
	MediaTitle       string       `xorm:"'media_title'" json:"media_title"`                      // movie, tvshow, season or episode title
	OriginalTitle    string       `xorm:"'original_title'" json:"original_title"`                // original title in release language
	OriginalLanguage string       `xorm:"'original_language'" json:"original_language"`
	OriginCountry    []string     `xorm:"'origin_country'" json:"origin_country"`
	ReleaseDate      string       `xorm:"'release_date'" json:"release_date"`   // 2012-12-30, for shows/seasons - first_air_date, for episodes - air_date
	LastAirDate      string       `xorm:"'last_air_date'" json:"last_air_date"` // for shows, latest episode air date
	NextAirDate      string       `xorm:"'next_air_date'" json:"next_air_date"` // for shows, next scheduled episode air date
	SeasonNumber     int          `xorm:"'season_number'" json:"season_number"`
	EpisodeNumber    int          `xorm:"'episode_number'" json:"episode_number"`
	SortIndex        int          `xorm:"'sort_index'" json:"sort_index"`       // not in use yet, used to sort based on user preferences
	Status           string       `xorm:"'status'" json:"status"`               // Returning Series, Released, etc.
	Overview         string       `xorm:"'overview'" json:"overview"`           // game of thrones is a show about ...
	Duration         int          `xorm:"'duration'" json:"duration"`           // duration/runtime in minutes
	ThumbnailURL     string       `xorm:"'thumbnail_url'" json:"thumbnail_url"` // poster image for tmdb
	BackdropURL      string       `xorm:"'backdrop_url'" json:"backdrop_url"`   // backgrounds
	StillURL         string       `xorm:"'still_url'" json:"still_url"`         // episodes, still frame for thumbnail
	Tags             *[]TagObject `json:"tags"`                                 // to store genres, tags
	UserTags         *[]TagObject `xorm:"'user_tags'" json:"user_tags"`
	FullData         []byte       `json:"data"`                                // full data from tmdb
	ContentHash      string       `xorm:"'coontent_hash'" json:"content_hash"` // checksum to compare changes/updates
	CreatedAt        time.Time    `xorm:"created" json:"created_at"`
	UpdatedAt        time.Time    `xorm:"updated" json:"updated_at"`
}

type MediaRecordGroup struct {
	MediaRecord  `xorm:"extends"`
	UserID       int64
	CollectionID int64
}

func UpsertMediaRecord(mediaRecord *MediaRecord) (int64, error) {
	// check if data is already in internal library
	var existingRecords []MediaRecord
	_ = databaseEngine.Table(mediaRecordsTable).Where("record_type = ?", mediaRecord.RecordType).
		Where("media_source = ?", mediaRecord.MediaSource).
		Where("source_id = ?", mediaRecord.SourceID).Find(&existingRecords)

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
				return -1, err
			}
			recordID = mediaRecord.RecordID
		}
	} else {
		// insert media data to library table
		_, err := databaseEngine.Table(mediaRecordsTable).Insert(&mediaRecord)
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
