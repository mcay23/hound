package database

import (
	"fmt"
	"hound/helpers"
	"log/slog"
	"sort"
	"strconv"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
	"xorm.io/xorm"
)

const (
	genresTable            = "genres"
	mediaRecordGenresTable = "media_record_genres"
)

type GenreRecord struct {
	GenreID     int64     `xorm:"pk autoincr 'genre_id'" json:"genre_id"`
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

/*
In Hound, genres are persisted to DB but also loaded to the cache for fast retrieval
Might refactor to use in memory vars instead of cache to simplify code
*/

func PopulateGenresCache() error {
	var genres []GenreRecord
	err := databaseEngine.Table(genresTable).Find(&genres)
	if err != nil {
		return fmt.Errorf("query %s: %w", genresTable, err)
	}
	for _, genre := range genres {
		key := getGenreCacheKey(genre.MediaSource, genre.MediaType, genre.SourceID)
		_, err := SetCache(key, genre, -1) // No expiration for genres
		if err != nil {
			return err
		}
	}
	return nil
}

func getGenreCacheKey(mediaSource, mediaType string, sourceID int64) string {
	return "genre:" + mediaSource + ":" + mediaType + ":" + strconv.FormatInt(sourceID, 10)
}

func GetGenreFromCache(mediaSource, mediaType string, sourceID int64) *GenreRecord {
	key := getGenreCacheKey(mediaSource, mediaType, sourceID)
	var genre GenreRecord
	exists, err := GetCache(key, &genre)
	if err != nil {
		slog.Debug("Error retrieving genre from cache, returning nil", "error", err)
		return nil
	}
	if exists {
		return &genre
	}
	return nil
}

func GetGenresByType(mediaType string) ([]GenreRecord, error) {
	// get all genre keys from the cache
	// only tmdb is supported for now
	prefix := "genre:tmdb:" + mediaType + ":"
	keys, err := GetKeysWithPrefix(prefix)
	if err != nil {
		return nil, err
	}
	genres := make([]GenreRecord, 0, len(keys))
	for _, key := range keys {
		var genre GenreRecord
		exists, err := GetCache(key, &genre)
		if err != nil {
			slog.Debug("error retrieving genre from cache", "error", err, "key", key)
			continue
		}
		if exists {
			genres = append(genres, genre)
		}
	}
	sort.Slice(genres, func(i, j int) bool {
		return genres[i].Genre < genres[j].Genre
	})
	return genres, nil
}

func instantiateGenresTables() error {
	if err := databaseEngine.Table(genresTable).Sync2(new(GenreRecord)); err != nil {
		return err
	}
	if err := databaseEngine.Table(mediaRecordGenresTable).Sync2(new(MediaRecordGenre)); err != nil {
		return err
	}
	// delete genres cache on startup
	genreKeys, err := GetKeysWithPrefix("genre:")
	if err != nil {
		return err
	}
	for _, key := range genreKeys {
		err := DeleteCache(key)
		if err != nil {
			return err
		}
	}
	return PopulateGenresCache()
}

func ConvertGenres(mediaSource string, mediaType string, genres []tmdb.Genre) []GenreObject {
	converted := make([]GenreObject, 0, len(genres))
	for _, genre := range genres {
		genreRecord := GetGenreFromCache(mediaSource, mediaType, int64(genre.ID))
		if genreRecord != nil {
			converted = append(converted, GenreObject{
				GenreID:     genreRecord.GenreID,
				Genre:       genreRecord.Genre,
				MediaSource: mediaSource,
				SourceID:    genreRecord.SourceID,
				MediaType:   mediaType,
			})
		}
	}
	return converted
}

// UpsertGenres inserts new genres, or update db records if genre names were to change
// This shouldn't be run after startup, race conditions possible?
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
	// refresh cache after upsert
	_ = PopulateGenresCache()
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
			Where("source_id = ?", genre.SourceID).
			Where("media_type = ?", mediaType).
			Get(&existing)
		if err != nil {
			return nil, fmt.Errorf("query %s for media_source %s, media_type %s, source_id %d: %w",
				genresTable, mediaSource, mediaType, genre.SourceID, err)
		}
		if !has {
			insert := GenreRecord{
				Genre:       genre.Genre,
				MediaSource: mediaSource,
				SourceID:    genre.SourceID,
				MediaType:   mediaType,
			}
			_, err = sess.Table(genresTable).Insert(&insert)
			if err != nil {
				if !isUniqueViolation(err) {
					return nil, fmt.Errorf("query %s for media_source %s, media_type %s, source_id %d: %w: %w",
						genresTable, mediaSource, mediaType, genre.SourceID, helpers.AlreadyExistsError, err)
				}
				// concurrent upsert race, refetch
				has, err = sess.Table(genresTable).
					Where("media_source = ?", mediaSource).
					Where("source_id = ?", genre.SourceID).
					Where("media_type = ?", mediaType).
					Get(&existing)
				if err != nil {
					return nil, fmt.Errorf("query %s for media_source %s, media_type %s, source_id %d: %w",
						genresTable, mediaSource, mediaType, genre.SourceID, err)
				}
				if !has {
					return nil, fmt.Errorf("query %s for media_source %s, media_type %s, source_id %d: %w",
						genresTable, mediaSource, mediaType, genre.SourceID, err)
				}
			} else {
				existing = insert
			}
		}
		if existing.Genre != genre.Genre {
			existing.Genre = genre.Genre
			_, err = sess.Table(genresTable).ID(existing.GenreID).Cols("genre").Update(&existing)
			if err != nil {
				return nil, fmt.Errorf("query %s for genre_id %d: %w",
					genresTable, existing.GenreID, err)
			}
		}
		sourceToInternal[genre.SourceID] = existing.GenreID
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
		return fmt.Errorf("query %s for media_record_id %d: %w",
			mediaRecordGenresTable, mediaRecordID, err)
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
	if err != nil {
		return fmt.Errorf("insert %s for media_record_id %d: %w",
			mediaRecordGenresTable, mediaRecordID, err)
	}
	return nil
}
