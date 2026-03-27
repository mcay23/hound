package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mcay23/hound/helpers"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
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
	RecordID         int64         `xorm:"pk autoincr 'record_id'" json:"record_id"`
	RecordType       string        `xorm:"unique(primary) not null 'record_type'" json:"record_type"`   // movie,tvshow,season,episode
	MediaSource      string        `xorm:"unique(primary) not null 'media_source'" json:"media_source"` // tmdb, openlibrary, etc. the main metadata provider
	SourceID         string        `xorm:"unique(primary) not null 'source_id'" json:"source_id"`       // tmdb id, episode/season tmdb id
	ParentID         *int64        `xorm:"index 'parent_id'" json:"parent_id,omitempty"`                // reference to fk record_id, null for movie, tvshow
	AncestorID       *int64        `xorm:"index 'ancestor_id'" json:"ancestor_id,omitempty"`            // reference to fk record_id of the show, for episodes
	MediaTitle       string        `xorm:"text 'media_title'" json:"media_title"`                       // movie, tvshow, season or episode title
	OriginalTitle    string        `xorm:"text 'original_title'" json:"original_title"`                 // original title in release language
	OriginalLanguage string        `xorm:"text 'original_language'" json:"original_language"`
	OriginCountry    []string      `xorm:"'origin_country'" json:"origin_country"`
	ReleaseDate      string        `xorm:"'release_date'" json:"release_date"`   // 2012-12-30, for shows/seasons - first_air_date, for episodes - air_date
	LastAirDate      string        `xorm:"'last_air_date'" json:"last_air_date"` // for shows, latest episode air date
	NextAirDate      string        `xorm:"'next_air_date'" json:"next_air_date"` // for shows, next scheduled episode air date
	SeasonNumber     *int          `xorm:"'season_number'" json:"season_number,omitempty"`
	EpisodeNumber    *int          `xorm:"'episode_number'" json:"episode_number,omitempty"`
	SortIndex        int           `xorm:"'sort_index'" json:"sort_index"`            // not in use yet, used to sort based on user preferences
	Status           string        `xorm:"'status'" json:"status"`                    // Returning Series, Released, etc.
	Overview         string        `xorm:"text 'overview'" json:"overview"`           // game of thrones is a show about ...
	Duration         int           `xorm:"'duration'" json:"duration"`                // duration/runtime in minutes
	ThumbnailURI     string        `xorm:"text 'thumbnail_uri'" json:"thumbnail_uri"` // poster image for shows/movies, still image for episode
	BackdropURI      string        `xorm:"text 'backdrop_uri'" json:"backdrop_uri"`   // backgrounds
	LogoURI          string        `xorm:"text 'logo_uri'" json:"logo_uri"`           // logo for the show/movie
	Genres           []GenreObject `xorm:"'genres'" json:"genres,omitempty"`          // to store genres, tags
	Tags             []TagObject   `xorm:"'tags'" json:"tags,omitempty"`
	CreatedAt        time.Time     `xorm:"timestampz created" json:"created_at"`
	UpdatedAt        time.Time     `xorm:"timestampz updated" json:"updated_at"`
	FullData         []byte        `xorm:"'full_data'" json:"full_data,omitempty"`            // full data from tmdb
	ContentHash      string        `xorm:"text 'content_hash'" json:"content_hash,omitempty"` // checksum to compare changes/updates
}

type GenreObject struct {
	GenreID     int64  `json:"genre_id"`
	Genre       string `json:"genre"`
	MediaType   string `json:"media_type"`
	MediaSource string `json:"media_source"`
	SourceID    int64  `json:"source_id"`
}

type MediaRecordGroup struct {
	MediaRecord  `xorm:"extends"`
	UserID       int64
	CollectionID int64
}

func instantiateMediaTables() error {
	return databaseEngine.Table(mediaRecordsTable).Sync2(new(MediaRecord))
}

func UpsertMediaRecord(mediaRecord *MediaRecord) error {
	// check if data is already in internal library
	var existingRecord MediaRecord
	has, err := databaseEngine.Table(mediaRecordsTable).Where("record_type = ?", mediaRecord.RecordType).
		Where("media_source = ?", mediaRecord.MediaSource).
		Where("source_id = ?", mediaRecord.SourceID).
		Get(&existingRecord)
	if err != nil {
		return fmt.Errorf("query %s for record_type %s, media_source %s, source_id %s: %w", mediaRecordsTable, mediaRecord.RecordType, mediaRecord.MediaSource, mediaRecord.SourceID, err)
	}

	// source_id is either movie, show, season, or episode id
	// a key on these three should be sufficiently unique (?)
	var recordID int64
	if has {
		// use existing SourceID
		recordID = existingRecord.RecordID
		if existingRecord.ContentHash != mediaRecord.ContentHash {
			// hash changed, update record in internal library
			_, err = databaseEngine.Table(mediaRecordsTable).ID(recordID).Update(mediaRecord)
			if err != nil {
				return fmt.Errorf("update %s for record_id %d (hash changed): %w", mediaRecordsTable, recordID, err)
			}
		}
		// mutate in place
		*mediaRecord = existingRecord
	} else {
		// insert media data to library table
		_, err = databaseEngine.Table(mediaRecordsTable).Insert(mediaRecord)
		if err != nil {
			// concurrent insert race, often happens when ingesting external items
			// since two episodes might attempt to upsert at the same time for new records
			if !isUniqueViolation(err) {
				return fmt.Errorf("insert %s for record_type %s, media_source %s, source_id %s: %w",
					mediaRecordsTable, mediaRecord.RecordType, mediaRecord.MediaSource, mediaRecord.SourceID, err)
			}
			has, err = databaseEngine.Table(mediaRecordsTable).Where("record_type = ?", mediaRecord.RecordType).
				Where("media_source = ?", mediaRecord.MediaSource).
				Where("source_id = ?", mediaRecord.SourceID).
				Get(&existingRecord)
			if err != nil {
				return fmt.Errorf("query %s for record_type %s, media_source %s, source_id %s (concurrent insert race): %w",
					mediaRecordsTable, mediaRecord.RecordType, mediaRecord.MediaSource, mediaRecord.SourceID, err)
			}
			if !has {
				return fmt.Errorf("query %s for record_type %s, media_source %s, source_id %s (unexpected concurrent insert race - please raise issue in github): %w: %w",
					mediaRecordsTable, mediaRecord.RecordType, mediaRecord.MediaSource, mediaRecord.SourceID, err, helpers.NotFoundError)
			}
			recordID = existingRecord.RecordID
			if existingRecord.ContentHash != mediaRecord.ContentHash {
				_, err := databaseEngine.Table(mediaRecordsTable).ID(recordID).Update(mediaRecord)
				if err != nil {
					return err
				}
			}
			*mediaRecord = existingRecord
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
		return false, fmt.Errorf("query %s for record_type %s, media_source %s, source_id %s: %w",
			mediaRecordsTable, record.RecordType, record.MediaSource, record.SourceID, err)
	}
	// it's possible that another process/worker has upserted this record between the first check and the insert,
	// thus failing the insert
	if !has {
		_, err := sess.Table(mediaRecordsTable).Insert(record)
		if err != nil {
			if !isUniqueViolation(err) {
				return false, fmt.Errorf("insert %s: %w", mediaRecordsTable, err)
			}
			// unique violation, refetch
			has, err = sess.Table(mediaRecordsTable).Where("record_type = ?", record.RecordType).
				Where("media_source = ?", record.MediaSource).
				Where("source_id = ?", record.SourceID).
				Get(&recordData)
			if err != nil {
				return false, fmt.Errorf("query %s for record_type %s, media_source %s, source_id %s (concurrent insert race): %w",
					mediaRecordsTable, record.RecordType, record.MediaSource, record.SourceID, err)
			}
			if !has {
				return false, fmt.Errorf("query %s for record_type %s, media_source %s, source_id %s (unexpected concurrent insert race - please raise issue in github): %w: %w",
					mediaRecordsTable, record.RecordType, record.MediaSource, record.SourceID, err, helpers.NotFoundError)
			}
		} else {
			return true, nil
		}
	}
	// if has, check hash, then update if not match
	if record.ContentHash == recordData.ContentHash {
		return false, nil
	}
	_, err = sess.Table(mediaRecordsTable).ID(recordData.RecordID).Update(record)
	if err != nil {
		return false, fmt.Errorf("update %s for record_id %d: %w", mediaRecordsTable, recordData.RecordID, err)
	}
	return true, nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func BatchUpsertMediaRecords(sess *xorm.Session, records []*MediaRecord) error {
	if len(records) == 0 {
		return nil
	}
	// don't want to run out of memory, batch by 500s
	const batchSize = 500
	for start := 0; start < len(records); start += batchSize {
		end := start + batchSize
		if end > len(records) {
			end = len(records)
		}
		batch := records[start:end]
		if err := batchUpsertChunk(sess, batch); err != nil {
			return fmt.Errorf("batch upsert %s of len %d, start %d, end %d: %w", mediaRecordsTable, len(batch), start, end, err)
		}
	}
	return nil
}

func batchUpsertChunk(sess *xorm.Session, records []*MediaRecord) error {
	columns := []string{
		"record_type", "media_source", "source_id", "parent_id", "ancestor_id",
		"media_title", "original_title", "original_language",
		"origin_country", "release_date", "last_air_date", "next_air_date",
		"season_number", "episode_number",
		"sort_index", "status", "overview", "duration",
		"thumbnail_uri", "backdrop_uri", "logo_uri",
		"genres", "tags", "full_data", "content_hash", "created_at", "updated_at",
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

		now := time.Now().UTC()
		valArgs = append(valArgs,
			record.RecordType,
			record.MediaSource,
			record.SourceID,
			record.ParentID,
			record.AncestorID,
			record.MediaTitle,
			record.OriginalTitle,
			record.OriginalLanguage,
			encodeJSONDB(record.OriginCountry),
			record.ReleaseDate,
			record.LastAirDate,
			record.NextAirDate,
			record.SeasonNumber,
			record.EpisodeNumber,
			record.SortIndex,
			record.Status,
			record.Overview,
			record.Duration,
			record.ThumbnailURI,
			record.BackdropURI,
			record.LogoURI,
			encodeJSONDB(record.Genres),
			encodeJSONDB(record.Tags),
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
	thumbnail_uri   = EXCLUDED.thumbnail_uri,
	backdrop_uri    = EXCLUDED.backdrop_uri,
	logo_uri        = EXCLUDED.logo_uri,
	genres          = EXCLUDED.genres,
	tags            = EXCLUDED.tags,
	full_data       = EXCLUDED.full_data,
	content_hash    = EXCLUDED.content_hash,
	ancestor_id     = EXCLUDED.ancestor_id,
	updated_at      = date_trunc('second', NOW())
WHERE media_records.content_hash IS DISTINCT FROM EXCLUDED.content_hash;
`)
	_, err := sess.DB().Exec(sb.String(), valArgs...)
	return err
}

// returns json encoding for database
// nil for empty map, slices
func encodeJSONDB(v any) any {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		return nil
	}
	switch rv.Kind() {
	case reflect.Slice, reflect.Map:
		if rv.Len() == 0 {
			return nil
		}
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

// marks a media_record to be updated on the next attempt
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
		return false, nil, fmt.Errorf("query %s nil xorm session", mediaRecordsTable)
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

func GetMediaRecordByID(recordID int64) (*MediaRecord, error) {
	var record MediaRecord
	_, err := databaseEngine.Table(mediaRecordsTable).Where("record_id = ?", recordID).Get(&record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// for an array of episode ids, check if exist in
// the database as a child of the show
// hierarchy show -> season -> episode
// returns a list of invalid episode ids
func CheckShowEpisodesIDs(mediaSource string, showSourceID string, episodeIDs []int) (map[string]string, []int, error) {
	type EpisodeIDObject struct {
		SourceID string `xorm:"source_id" json:"source_id"`
		RecordID string `xorm:"record_id" json:"record_id"`
	}
	var episodes []EpisodeIDObject
	err := databaseEngine.SQL(`
        SELECT e.source_id, e.record_id
        FROM media_records AS show
        JOIN media_records AS e
            ON  e.ancestor_id = show.record_id
            AND e.record_type = 'episode'
            AND e.media_source = show.media_source
        WHERE show.record_type = 'tvshow'
          AND show.media_source = ?
          AND show.source_id = ?;
    `, mediaSource, showSourceID).Find(&episodes)
	if err != nil {
		return nil, nil, fmt.Errorf("query %s for media_source %s, ancestor_id %s: %w", mediaRecordsTable,
			mediaSource, showSourceID, err)
	}
	if len(episodes) <= 0 {
		return nil, nil, fmt.Errorf("query %s for media_source %s, ancestor_id %s (failed to find episodes): %w", mediaRecordsTable,
			mediaSource, showSourceID, helpers.NotFoundError)
	}
	// Build index of episode IDs
	episodesMap := make(map[string]string, len(episodes))
	for _, ep := range episodes {
		episodesMap[ep.SourceID] = ep.RecordID
	}
	// Check invalidIDs
	var invalidIDs []int
	for _, id := range episodeIDs {
		if _, ok := episodesMap[strconv.Itoa(id)]; !ok {
			invalidIDs = append(invalidIDs, id)
		}
	}
	if len(invalidIDs) > 0 {
		invalidIDStr := ""
		for _, item := range invalidIDs {
			invalidIDStr += strconv.Itoa(int(item)) + ","
		}
		return nil, invalidIDs, fmt.Errorf("query %s for media_source %s, ancestor_id %s (invalid episode IDs found %s): %w", mediaRecordsTable,
			mediaSource, showSourceID, invalidIDStr, helpers.BadRequestError)
	}
	return episodesMap, nil, nil
}

// showSourceID and seasonNumber are optional
func GetEpisodeMediaRecords(mediaSource string, showSourceID string, seasonNumber *int, episodeNumber *int) ([]MediaRecord, error) {
	var episodes []MediaRecord
	sess := databaseEngine.NewSession()
	defer sess.Close()
	columns := `
		episode.record_id,
		episode.record_type,
		episode.media_source,
		episode.source_id,
		episode.parent_id,
		episode.media_title,
		episode.original_title,
		episode.thumbnail_uri,
		episode.origin_country,
		episode.release_date,
		episode.season_number,
		episode.episode_number,
		episode.sort_index,
		episode.overview,
		episode.duration,
		episode.content_hash
	`
	sess = sess.Table(mediaRecordsTable).Alias("show").
		Select(columns).
		Join("INNER", []string{"media_records", "episode"}, "episode.ancestor_id = show.record_id").
		Where("show.media_source = ?", mediaSource).
		Where("show.source_id = ?", showSourceID).
		Where("show.record_type = ?", RecordTypeTVShow).
		Where("episode.record_type = ?", "episode")
	if seasonNumber != nil {
		sess = sess.Where("episode.season_number = ?", *seasonNumber)
	}
	if episodeNumber != nil {
		sess = sess.Where("episode.episode_number = ?", *episodeNumber)
	}
	sess = sess.Asc("episode.episode_number")
	err := sess.Find(&episodes)
	if err != nil {
		return nil, err
	}
	return episodes, nil
}

// easier to accept nil then to check for nil everywhere else
func GetEpisodeMediaRecord(mediaSource string, showSourceID string,
	seasonNumber *int, episodeNumber *int) (*MediaRecord, error) {
	if seasonNumber == nil || episodeNumber == nil {
		return nil, fmt.Errorf("season number or episode number is nil: %w", helpers.BadRequestError)
	}
	episodes, err := GetEpisodeMediaRecords(mediaSource, showSourceID, seasonNumber, episodeNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get episode media records: %w: %w", helpers.InternalServerError, err)
	}
	if len(episodes) > 0 {
		return &episodes[0], nil
	}
	// fails without logging, since some optimizations grab this without knowing if it exists yet
	return nil, fmt.Errorf("query %s for media_source %s, ancestor_id %s, season_number %d, episode_number %d: %w", mediaRecordsTable,
		mediaSource, showSourceID, seasonNumber, episodeNumber, helpers.NotFoundError)
}

// This returns the movie/show-level record, not the episodes
// For shows, if you have at least 1 downloaded episode
// it will be included
func GetDownloadedParentRecords(limit int, offset int, mediaType string, genreIDs []int64) ([]MediaRecordGroup, int64, error) {
	var recordGroups []MediaRecordGroup
	// find movies with files OR shows with episodes that have files
	whereClause := `(
			mr.record_type = 'movie' AND EXISTS (
				SELECT 1 FROM %s mf WHERE mf.record_id = mr.record_id
			)
		) OR (
			mr.record_type = 'tvshow' AND EXISTS (
				SELECT 1 
				FROM %s ep 
				JOIN %s mf ON mf.record_id = ep.record_id
				WHERE ep.ancestor_id = mr.record_id AND ep.record_type = 'episode'
			)
		)`
	whereClause = fmt.Sprintf(whereClause, mediaFilesTable, mediaRecordsTable, mediaFilesTable)

	var args []interface{}
	if mediaType != "" {
		whereClause = fmt.Sprintf("(%s) AND mr.record_type = ?", whereClause)
		args = append(args, mediaType)
	}

	if len(genreIDs) > 0 {
		placeholders := make([]string, len(genreIDs))
		for i := range genreIDs {
			placeholders[i] = "?"
			args = append(args, genreIDs[i])
		}
		whereClause = fmt.Sprintf("(%s) AND EXISTS (SELECT 1 FROM %s mrg WHERE mrg.record_id = mr.record_id AND mrg.genre_id IN (%s))",
			whereClause, mediaRecordGenresTable, strings.Join(placeholders, ","))
	}

	query := fmt.Sprintf(`
		SELECT mr.*
		FROM %s mr
		WHERE %s
		ORDER BY mr.media_title ASC
	`, mediaRecordsTable, whereClause)

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s mr WHERE %s", mediaRecordsTable, whereClause)
	var totalRecords int64
	_, err := databaseEngine.SQL(countQuery, args...).Get(&totalRecords)
	if err != nil {
		return nil, 0, fmt.Errorf("count %s for media_type %s, genre_ids %s: %w", mediaRecordsTable,
			mediaType, genreIDs, err)
	}
	if limit > 0 && offset >= 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	}
	err = databaseEngine.SQL(query, args...).Find(&recordGroups)
	if err != nil {
		return nil, 0, fmt.Errorf("query %s for media_type %s, genre_ids %s: %w", mediaRecordsTable,
			mediaType, genreIDs, err)
	}
	return recordGroups, totalRecords, nil
}
