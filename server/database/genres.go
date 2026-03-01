package database

import (
	"errors"
	"hound/helpers"
	"sort"
	"time"

	"xorm.io/xorm"
)

const (
	genresTable            = "genres"
	mediaRecordGenresTable = "media_record_genres"
)

type GenreRecord struct {
	ID          int64     `xorm:"pk autoincr 'id'" json:"id"`
	Genre       string    `xorm:"varchar(128) not null 'genre'" json:"genre"`
	MediaType   string    `xorm:"unique(uk_genres_source) not null 'media_type'" json:"media_type"`
	MediaSource string    `xorm:"unique(uk_genres_source) not null 'media_source'" json:"media_source"`
	SourceID    int64     `xorm:"unique(uk_genres_source) not null 'source_id'" json:"source_id"`
	CreatedAt   time.Time `xorm:"timestampz created" json:"created_at"`
	UpdatedAt   time.Time `xorm:"timestampz updated" json:"updated_at"`
}

type MediaRecordGenre struct {
	RecordID  int64     `xorm:"unique(uk_media_record_genres) index not null 'record_id'" json:"record_id"`
	GenreID   int64     `xorm:"unique(uk_media_record_genres) index not null 'genre_id'" json:"genre_id"`
	CreatedAt time.Time `xorm:"timestampz created" json:"created_at"`
	UpdatedAt time.Time `xorm:"timestampz updated" json:"updated_at"`
}

func instantiateGenresTables() error {
	if err := databaseEngine.Table(genresTable).Sync2(new(GenreRecord)); err != nil {
		return err
	}
	return databaseEngine.Table(mediaRecordGenresTable).Sync2(new(MediaRecordGenre))
}

func UpsertGenres(mediaSource string, mediaType string, genres []GenreObject) (map[int64]int64, error) {
	session := databaseEngine.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return nil, err
	}
	sourceToInternal, err := upsertGenresTrx(session, mediaSource, mediaType, genres)
	if err != nil {
		_ = session.Rollback()
		return nil, err
	}
	if err := session.Commit(); err != nil {
		return nil, err
	}
	return sourceToInternal, nil
}

// Inserts new genres, or update db records if genre names were to change
// Not sure if genre name changes has ever happened with tmdb
func upsertGenresTrx(sess *xorm.Session, mediaSource string, mediaType string, genres []GenreObject) (map[int64]int64, error) {
	sourceToInternal := make(map[int64]int64, len(genres))
	for _, genre := range genres {
		var existing GenreRecord
		has, err := sess.Table(genresTable).
			Where("media_source = ?", mediaSource).
			Where("source_id = ?", genre.ID).
			Where("media_type = ?", mediaType).
			Get(&existing)
		if err != nil {
			return nil, err
		}
		if !has {
			insert := GenreRecord{
				Genre:       genre.Name,
				MediaSource: mediaSource,
				SourceID:    genre.ID,
				MediaType:   mediaType,
			}
			_, err = sess.Table(genresTable).Insert(&insert)
			if err != nil {
				if !isUniqueViolation(err) {
					return nil, err
				}
				// concurrent upsert race, refetch
				has, err = sess.Table(genresTable).
					Where("media_source = ?", mediaSource).
					Where("source_id = ?", genre.ID).
					Where("media_type = ?", mediaType).
					Get(&existing)
				if err != nil {
					return nil, err
				}
				if !has {
					return nil, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError),
						"Failed to find genre after unique violation")
				}
			} else {
				existing = insert
			}
		}
		if existing.Genre != genre.Name {
			existing.Genre = genre.Name
			_, err = sess.Table(genresTable).ID(existing.ID).Cols("genre").Update(&existing)
			if err != nil {
				return nil, err
			}
		}
		sourceToInternal[genre.ID] = existing.ID
	}
	return sourceToInternal, nil
}

// atomic update media record's genres by removing all genres for a media record, then repopulating
func ReplaceMediaRecordGenresByIDs(mediaRecordID int64, genreIDs []int64) error {
	session := databaseEngine.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return err
	}
	if err := ReplaceMediaRecordGenresByIDsTrx(session, mediaRecordID, genreIDs); err != nil {
		_ = session.Rollback()
		return err
	}
	return session.Commit()
}

func ReplaceMediaRecordGenresByIDsTrx(sess *xorm.Session, mediaRecordID int64, genreIDs []int64) error {
	if _, err := sess.Table(mediaRecordGenresTable).Where("record_id = ?", mediaRecordID).Delete(&MediaRecordGenre{}); err != nil {
		return err
	}
	if len(genreIDs) == 0 {
		return nil
	}
	// keep insert set deterministic and avoid duplicate relation rows
	cleaned := make([]int64, 0, len(genreIDs))
	seen := make(map[int64]struct{}, len(genreIDs))
	for _, id := range genreIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		cleaned = append(cleaned, id)
	}
	sort.Slice(cleaned, func(i, j int) bool { return cleaned[i] < cleaned[j] })
	records := make([]MediaRecordGenre, 0, len(cleaned))
	for _, genreID := range cleaned {
		records = append(records, MediaRecordGenre{
			RecordID: mediaRecordID,
			GenreID:  genreID,
		})
	}
	_, err := sess.Table(mediaRecordGenresTable).Insert(&records)
	return err
}
