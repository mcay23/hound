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
	Primary Collection - every user has a master collection which cannot be deleted
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
	CollectionID    int64        `xorm:"pk autoincr 'collection_id'" json:"collection_id"`
	CollectionTitle string       `xorm:"not null" json:"collection_title"` // my collection, etc.
	Description     []byte       `json:"description"`
	OwnerID         int64        `xorm:"not null 'owner_user_id'" json:"owner_user_id"`
	IsPrimary       bool         `json:"is_primary"` // is the user's primary collection, not deletable
	IsPublic        bool         `json:"is_public"`
	Tags            *[]TagObject `json:"tags"`
	ThumbnailURL    *string      `xorm:"'thumbnail_url'" json:"thumbnail_url"` // url for media thumbnails
	CreatedAt       time.Time    `xorm:"timestampz created" json:"created_at"`
	UpdatedAt       time.Time    `xorm:"timestampz updated" json:"updated_at"`
}

type CollectionRecordQuery struct {
	CollectionID *int64
	SearchQuery  *string
	OwnerID      *int64
	IsPrimary    *bool
	IsPublic     *bool
	Tags         *[]TagObject `json:"tags"`
}

type CreateCollectionRequest struct {
	OwnerID         int64  `json:"owner_id"`
	CollectionTitle string `json:"collection_title"` // my collection, etc.
	Description     string `json:"description"`
	IsPrimary       bool   `json:"is_primary"` // is the user's primary collection, not deletable
	IsPublic        bool   `json:"is_public"`
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
		return nil, nil, -1, err
	}
	if !found {
		return nil, nil, -1, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "getCollectionRecords(): No collection with this ID")
	}
	if !collection.IsPublic && collection.OwnerID != userID {
		return nil, nil, -1, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "user does not have access to collection")
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
		return nil, nil, -1, err
	}
	//TODO remove
	totalRecords, err := databaseEngine.Table(mediaRecordsTable).Where("collection_id = ?", collectionID).Join("INNER", collectionRelationsTable, fmt.Sprintf("%s.record_id = %s.record_id", mediaRecordsTable, collectionRelationsTable)).Count()
	if err != nil {
		return nil, nil, -1, err
	}
	return recordGroups, &collection, totalRecords, nil
}

func InsertCollectionRelation(userID int64, recordID int64, collectionID *int64) error {
	// if collectionID not supplied, add to user's primary collection
	if collectionID == nil {
		var collectionRecord CollectionRecord
		has, err := databaseEngine.Table(collectionsTable).Where("owner_user_id = ?", userID).Where("is_primary = ?", true).Get(&collectionRecord)
		if err != nil {
			return err
		}
		if !has {
			return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Collection not found, no primary col found for user")
		}
		collectionID = &collectionRecord.CollectionID
	} else {
		// check if collection exists in collections table
		// TODO should ideally be covered by foreign key constraint, xorm does not handle sync with fk right now
		var collectionRecord CollectionRecord
		has, err := databaseEngine.Table(collectionsTable).ID(*collectionID).Get(&collectionRecord)
		if err != nil {
			return err
		}
		if !has {
			return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Collection not found")
		}
		// check if user is authorized to add to collection
		if collectionRecord.OwnerID != userID {
			return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Collection - owner mismatch, unauthorized")
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
				return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Record already exists in collection")
			}
		}
	}
	return err
}

func DeleteCollectionRelation(userID int64, recordID int64, collectionID int64) error {
	var collectionRecord CollectionRecord
	has, err := databaseEngine.Table(collectionsTable).ID(collectionID).Get(&collectionRecord)
	if err != nil {
		return err
	}
	if !has {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Collection not found")
	}
	// check if user is authorized to add to collection
	if collectionRecord.OwnerID != userID {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Collection - owner mismatch, unauthorized")
	}
	// if user authenticated, remove
	affected, _ := databaseEngine.Table(collectionRelationsTable).Delete(&CollectionRelation{
		UserID:       userID,
		RecordID:     recordID,
		CollectionID: collectionID,
	})
	if affected == 0 {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "DeleteCollectionRelation(): No record with these parameters found")
	}
	return nil
}

func CreateCollection(record CreateCollectionRequest) (*int64, error) {
	if record.IsPrimary == true {
		_, int64, err := SearchForCollection(CollectionRecordQuery{
			OwnerID:   &record.OwnerID,
			IsPrimary: &record.IsPrimary,
		}, 1, 0)
		if err != nil {
			return nil, err
		}
		// primary record already exists
		if int64 != 0 {
			return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Primary col already exists")
		}
	}
	insert := CollectionRecord{
		CollectionTitle: record.CollectionTitle,
		Description:     []byte(record.Description),
		OwnerID:         record.OwnerID,
		IsPrimary:       record.IsPrimary,
		IsPublic:        record.IsPublic,
		Tags:            nil,
		ThumbnailURL:    nil,
	}
	_, err := databaseEngine.Table(collectionsTable).Insert(&insert)
	if err != nil {
		return nil, err
	}
	return &insert.CollectionID, nil
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
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "DeleteCollection(): Failed to delete collection IDs")
	}
	// primary collections can't be deleted
	affected, err := session.Table(collectionsTable).Where("is_primary = ?", false).Delete(&CollectionRecord{
		CollectionID: collectionID,
		OwnerID:      userID,
	})
	if err != nil {
		_ = session.Rollback()
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "DeleteCollection(): Failed to delete comments")
	}
	if affected <= 0 {
		_ = session.Rollback()
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "DeleteCollection(): No collection found with this ID or invalid user")
	}
	err = session.Commit()
	if err != nil {
		_ = session.Rollback()
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "DeleteCollection(): error committing transaction")
	}
	return nil
}

func SearchForCollection(record CollectionRecordQuery, limit int, offset int) ([]CollectionRecord, int, error) {
	var records []CollectionRecord
	sess := databaseEngine.Table(collectionsTable)
	if record.OwnerID != nil {
		sess = sess.Where("owner_user_id = ?", record.OwnerID)
	}
	if record.IsPrimary != nil {
		sess = sess.Where("is_primary = ?", record.IsPrimary)
	}
	if record.CollectionID != nil {
		sess = sess.Where("collection_id = ?", record.CollectionID)
	}
	if record.IsPublic != nil {
		sess = sess.Where("is_public = ?", record.IsPublic)
	}
	if limit > 0 && offset >= 0 {
		sess = sess.Limit(limit, offset)
	}
	err := sess.Find(&records)
	if err != nil {
		return nil, 0, err
	}
	// restart session to get total count
	sess = databaseEngine.Table(collectionsTable)
	if record.OwnerID != nil {
		sess = sess.Where("owner_user_id = ?", record.OwnerID)
	}
	if record.IsPrimary != nil {
		sess = sess.Where("is_primary = ?", record.IsPrimary)
	}
	if record.CollectionID != nil {
		sess = sess.Where("collection_id = ?", record.CollectionID)
	}
	if record.IsPublic != nil {
		sess = sess.Where("is_public = ?", record.IsPublic)
	}
	totalRecords, err := sess.Count(new(CollectionRecord))
	if err != nil {
		return nil, 0, err
	}
	return records, int(totalRecords), nil
}
