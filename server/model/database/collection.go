package database

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"hound/helpers"
	"time"
)

/*
	Library - A library is a collection of all records from all users on the platform
	Collection - contains collection/list definitions
		Primary Collection - every user has a master collection which cannot be deleted
*/

const (
	libraryTable             = "library"
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

// store user saved libraries
type LibraryRecord struct {
	LibraryID    int64        `xorm:"pk autoincr 'library_id'" json:"library_id"`
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

type CollectionRelation struct {
	UserID       int64     `xorm:"unique(primary) not null 'user_id'" json:"user_id"` // refers to users table ids
	LibraryID    int64     `xorm:"unique(primary) not null 'library_id'" json:"library_id"`
	CollectionID int64     `xorm:"unique(primary) not null 'collection_id'" json:"collection_id"`
	CreatedAt    time.Time `xorm:"created" json:"created_at"`
	UpdatedAt    time.Time `xorm:"updated" json:"updated_at"`
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
	CreatedAt       time.Time    `xorm:"created" json:"created_at"`
	UpdatedAt       time.Time    `xorm:"updated" json:"updated_at"`
}

type LibraryGroup struct {
	LibraryRecord `xorm:"extends"`
	UserID        int64
	CollectionID  int64
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

func instantiateMediaTables() error {
	err := databaseEngine.Table(libraryTable).Sync2(new(LibraryRecord))
	if err != nil {
		return err
	}
	err = databaseEngine.Table(collectionsTable).Sync2(new(CollectionRecord))
	if err != nil {
		return err
	}
	err = databaseEngine.Table(collectionRelationsTable).Sync2(new(CollectionRelation))
	if err != nil {
		return err
	}
	return nil
}

func GetCollectionRecords(userID int64, collectionID int64, limit int, offset int) ([]LibraryGroup, *CollectionRecord, int64, error) {
	var libraryGroups []LibraryGroup
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
	sess := databaseEngine.Table(libraryTable)
	if limit > 0 && offset >= 0 {
		sess = sess.Limit(limit, offset)
	}
	err = sess.Where("collection_id = ?", collectionID).
		Join("INNER", collectionRelationsTable,
			fmt.Sprintf("%s.library_id = %s.library_id", libraryTable, collectionRelationsTable)).
			OrderBy(fmt.Sprintf("%s.updated_at desc", collectionRelationsTable)).
		Find(&libraryGroups)
	if err != nil {
		return nil, nil, -1, err
	}
	//TODO remove
	totalRecords, err := databaseEngine.Table(libraryTable).Where("collection_id = ?", collectionID).Join("INNER", collectionRelationsTable, fmt.Sprintf("%s.library_id = %s.library_id", libraryTable, collectionRelationsTable)).Count()
	if err != nil {
		return nil, nil, -1, err
	}
	return libraryGroups, &collection, totalRecords, nil
}

func InsertCollectionRelation(userID int64, libraryID int64, collectionID *int64) error {
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
		LibraryID:    libraryID,
		CollectionID: *collectionID,
	})
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Item already in collection")
		}
	}
	return err
}

func DeleteCollectionRelation(userID int64, libraryID int64, collectionID int64) error {
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
	affected, err := databaseEngine.Table(collectionRelationsTable).Delete(&CollectionRelation{
		UserID:       userID,
		LibraryID:    libraryID,
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
		OwnerID: userID,
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

func AddRecordToInternalLibrary(libraryRecord *LibraryRecord) (int64, error) {
	// check if data is already in internal library
	var existingRecords []LibraryRecord
	_ = databaseEngine.Table(libraryTable).Where("media_type = ?", libraryRecord.MediaType).
		Where("media_source = ?", libraryRecord.MediaSource).
		Where("source_id = ?", libraryRecord.SourceID).Find(&existingRecords)

	var libraryID int64
	if len(existingRecords) > 0 {
		// use existing SourceID
		libraryID = existingRecords[0].LibraryID
		if !bytes.Equal(existingRecords[0].FullData, libraryRecord.FullData) {
			// TODO update db with new data
		}
	} else {
		// insert media data to library table
		_, err := databaseEngine.Table(libraryTable).Insert(libraryRecord)
		if err != nil {
			return -1, err
		}
		libraryID = libraryRecord.LibraryID
	}
	return libraryID, nil
}

func GetInternalLibraryID(mediaType string, mediaSource string, sourceID string) (*int64, error) {
	var record LibraryRecord
	has, err := databaseEngine.Table(libraryTable).Where("media_type = ?", mediaType).
		Where("media_source = ?", mediaSource).
		Where("source_id = ?", sourceID).Get(&record)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"GetInternalLibraryID(): No matching record in internal library")
	}
	return &record.LibraryID, nil
}
