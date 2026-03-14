package database

import (
	"errors"
	"fmt"
	"hound/helpers"
	"time"

	"github.com/lib/pq"
)

/*
	Collection - contains collection/list definitions
*/

const (
	collectionsTable         = "collections"
	collectionRelationsTable = "collection_relations"
)

// stores watch/read history for media types by user
type History struct {
	UserID         int64 `xorm:"not null"`
	ConsumeHistory []time.Time
}

type TagObject struct {
	TagID   int64
	TagName string
}

type CollectionRelation struct {
	UserID       int64     `xorm:"unique(primary) not null 'user_id'" json:"user_id"` // refers to users table ids
	RecordID     int64     `xorm:"unique(primary) not null 'record_id'" json:"record_id"`
	CollectionID int64     `xorm:"unique(primary) not null 'collection_id'" json:"collection_id"`
	CreatedAt    time.Time `xorm:"timestampz created" json:"created_at"`
	UpdatedAt    time.Time `xorm:"timestampz updated" json:"updated_at"`
}

type CollectionRecord struct {
	CollectionID    int64     `xorm:"pk autoincr 'collection_id'" json:"collection_id"`
	CollectionTitle string    `xorm:"not null" json:"collection_title"` // my collection, etc.
	Description     string    `xorm:"text 'description'" json:"description"`
	OwnerUserID     int64     `xorm:"index 'owner_user_id'" json:"owner_user_id"`
	IsPublic        bool      `json:"is_public"`
	ThumbnailURI    string    `xorm:"'thumbnail_uri'" json:"thumbnail_uri"` // url for media thumbnails
	CreatedAt       time.Time `xorm:"timestampz created" json:"created_at"`
	UpdatedAt       time.Time `xorm:"timestampz updated" json:"updated_at"`
}

func instantiateCollectionTables() error {
	err := databaseEngine.Table(collectionsTable).Sync2(new(CollectionRecord))
	if err != nil {
		return err
	}
	return databaseEngine.Table(collectionRelationsTable).Sync2(new(CollectionRelation))
}

func GetCollectionRecords(userID int64, collectionID int64, limit int, offset int) ([]MediaRecordGroup, *CollectionRecord, int64, error) {
	var recordGroups []MediaRecordGroup
	var collection CollectionRecord
	found, err := databaseEngine.Table(collectionsTable).ID(collectionID).Get(&collection)
	if err != nil {
		return nil, nil, -1, fmt.Errorf("query %s for collection_id %d: %w: %w", collectionsTable,
			collectionID, helpers.InternalServerError, err)
	}
	if !found {
		return nil, nil, -1, fmt.Errorf("query %s for collection_id %d: %w", collectionsTable,
			collectionID, helpers.NotFoundError)
	}
	if !collection.IsPublic && collection.OwnerUserID != userID {
		return nil, nil, -1, fmt.Errorf("query %s for collection_id %d, owner_user_id %d does not have access: %w",
			collectionsTable, collectionID, userID, helpers.UnauthorizedError)
	}
	sess := databaseEngine.Table(mediaRecordsTable)
	if limit > 0 && offset >= 0 {
		sess = sess.Limit(limit, offset)
	}
	err = sess.Where("collection_id = ?", collectionID).
		Join("INNER", collectionRelationsTable,
			fmt.Sprintf("%s.record_id = %s.record_id", mediaRecordsTable, collectionRelationsTable)).
		OrderBy(fmt.Sprintf("%s.updated_at desc", collectionRelationsTable)).
		Find(&recordGroups)
	if err != nil {
		return nil, nil, -1, fmt.Errorf("query %s, %s for collection_id %d: %w", mediaRecordsTable,
			collectionRelationsTable, collectionID, err)
	}
	//TODO remove
	totalRecords, err := databaseEngine.Table(mediaRecordsTable).Where("collection_id = ?", collectionID).
		Join("INNER", collectionRelationsTable, fmt.Sprintf("%s.record_id = %s.record_id", mediaRecordsTable, collectionRelationsTable)).Count()
	if err != nil {
		return nil, nil, -1, fmt.Errorf("query %s, %s for collection_id %d: %w",
			mediaRecordsTable, collectionRelationsTable, collectionID, err)
	}
	return recordGroups, &collection, totalRecords, nil
}

func GetRecentCollectionRecords(userID int64, limit int) ([]MediaRecordGroup, error) {
	var recordGroups []MediaRecordGroup
	// distinct on to deduplicate, grab most recent
	query := fmt.Sprintf(`
		SELECT * FROM (
			SELECT DISTINCT ON (mr.record_id) mr.*, cr.user_id, cr.collection_id, cr.created_at as added_at
			FROM %s mr
			INNER JOIN %s cr ON mr.record_id = cr.record_id
			WHERE cr.user_id = ?
			ORDER BY mr.record_id, cr.created_at DESC
		) sub
		ORDER BY added_at DESC
		LIMIT ?
	`, mediaRecordsTable, collectionRelationsTable)
	err := databaseEngine.SQL(query, userID, limit).Find(&recordGroups)
	if err != nil {
		return recordGroups, fmt.Errorf("query %s: %w", collectionRelationsTable, err)
	}
	return recordGroups, nil
}

func InsertCollectionRelation(userID int64, recordID int64, collectionID *int64) error {
	// if collectionID not supplied, add to user's primary collection
	if collectionID == nil {
		var collectionRecord CollectionRecord
		has, err := databaseEngine.Table(collectionsTable).Where("owner_user_id = ?", userID).Where("is_primary = ?", true).Get(&collectionRecord)
		if err != nil {
			return fmt.Errorf("query %s for owner_user_id %d, is_primary true: %w", collectionsTable, userID, err)
		}
		if !has {
			return fmt.Errorf("query %s owner_user_id %d, is_primary true: %w", collectionsTable, userID, helpers.NotFoundError)
		}
		collectionID = &collectionRecord.CollectionID
	} else {
		// check if collection exists in collections table
		// TODO should ideally be covered by foreign key constraint, xorm does not handle sync with fk right now
		var collectionRecord CollectionRecord
		has, err := databaseEngine.Table(collectionsTable).ID(*collectionID).Get(&collectionRecord)
		if err != nil {
			return fmt.Errorf("query %s for collection_id %d: %w", collectionsTable, *collectionID, err)
		}
		if !has {
			return fmt.Errorf("query %s for collection_id %d: %w", collectionsTable, *collectionID, helpers.NotFoundError)
		}
		// check if user is authorized to add to collection
		if collectionRecord.OwnerUserID != userID {
			return fmt.Errorf("insert %s for collection_id %d, owner_user_id %d: %w",
				collectionsTable, *collectionID, userID, helpers.UnauthorizedError)
		}
	}
	// insert record to db
	_, err := databaseEngine.Table(collectionRelationsTable).Insert(CollectionRelation{
		UserID:       userID,
		RecordID:     recordID,
		CollectionID: *collectionID,
	})
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// unique key failed
			if pqErr.Code == "23505" {
				return fmt.Errorf("insert %s for record_id %d, collection_id %d: %w",
					collectionRelationsTable, recordID, *collectionID, helpers.AlreadyExistsError)
			}
		}
	}
	return err
}

func DeleteCollectionRelation(userID int64, recordID int64, collectionID int64) error {
	var collectionRecord CollectionRecord
	has, err := databaseEngine.Table(collectionsTable).ID(collectionID).Get(&collectionRecord)
	if err != nil {
		return fmt.Errorf("query %s for collection_id %d: %w", collectionsTable, collectionID, err)
	}
	if !has {
		return fmt.Errorf("query %s for collection_id %d: %w", collectionsTable, collectionID, helpers.NotFoundError)
	}
	// check if user is authorized to this collection
	if collectionRecord.OwnerUserID != userID {
		return fmt.Errorf("delete %s for collection_id %d, record_id %d, user_id %d: %w",
			collectionRelationsTable, collectionID, recordID, userID, helpers.UnauthorizedError)
	}
	// if user authenticated, remove
	affected, _ := databaseEngine.Table(collectionRelationsTable).Delete(&CollectionRelation{
		UserID:       userID,
		RecordID:     recordID,
		CollectionID: collectionID,
	})
	if affected == 0 {
		return fmt.Errorf("delete %s for userID %d, record_id %d, collection_id %d: %w",
			collectionRelationsTable, userID, recordID, collectionID, helpers.NotFoundError)
	}
	return nil
}

func CreateCollection(record CollectionRecord) (*int64, error) {
	if record.OwnerUserID <= 0 {
		return nil, fmt.Errorf("insert %s for owner_user_id %d invalid owner_user_id: %w", collectionsTable, record.OwnerUserID, helpers.BadRequestError)
	}
	_, err := databaseEngine.Table(collectionsTable).Insert(&record)
	if err != nil {
		return nil, fmt.Errorf("insert %s for owner_user_id %d: %w", collectionsTable, record.OwnerUserID, err)
	}
	return &record.CollectionID, nil
}

func DeleteCollection(userID int64, collectionID int64) error {
	session := databaseEngine.NewSession()
	defer session.Close()
	_ = session.Begin()
	_, err := session.Table(collectionRelationsTable).Delete(&CollectionRelation{
		CollectionID: collectionID,
	})
	if err != nil {
		_ = session.Rollback()
		return fmt.Errorf("delete %s for owner_user_id %d, collection_id %d: %w", collectionRelationsTable, userID, collectionID, err)
	}
	// primary collections can't be deleted
	affected, err := session.Table(collectionsTable).Where("is_primary = ?", false).Delete(&CollectionRecord{
		CollectionID: collectionID,
		OwnerUserID:  userID,
	})
	if err != nil {
		_ = session.Rollback()
		return fmt.Errorf("delete %s for owner_user_id %d, collection_id %d: %w", collectionsTable, userID, collectionID, err)
	}
	if affected <= 0 {
		_ = session.Rollback()
		return fmt.Errorf("query %s for owner_user_id %d, collection_id %d: %w", collectionsTable, userID, collectionID, helpers.NotFoundError)
	}
	err = session.Commit()
	return err
}

func FindCollection(query CollectionRecord, limit int, offset int) ([]CollectionRecord, int, error) {
	var records []CollectionRecord
	sess := databaseEngine.Table(collectionsTable)
	if limit > 0 && offset >= 0 {
		sess = sess.Limit(limit, offset)
	}
	err := sess.Find(&records, &query)
	if err != nil {
		return nil, 0, fmt.Errorf("query %s: %w", collectionsTable, err)
	}
	// restart session to get total count
	sess = databaseEngine.Table(collectionsTable)
	totalRecords, err := sess.Count(&query)
	if err != nil {
		return nil, 0, fmt.Errorf("count %s: %w", collectionsTable, err)
	}
	return records, int(totalRecords), nil
}
