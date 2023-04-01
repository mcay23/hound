package database

import (
	"errors"
	"fmt"
	"hound/helpers"
	"time"
	"xorm.io/xorm"
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
	historyTable             = "history"
	commentTable             = "comments"
	MediaTypeTVShow          = "tvshow"
	MediaTypeMovie           = "movie"
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
	UserID       int64 `xorm:"unique(primary) not null 'user_id'" json:"user_id"` // refers to users table ids
	LibraryID    int64 `xorm:"unique(primary) not null 'library_id'" json:"library_id"`
	CollectionID int64 `xorm:"unique(primary) not null 'collection_id'" json:"collection_id"`
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

func AddItemToCollection(userID int64, collectionID *int64, libraryRecord *LibraryRecord) error {
	// if no collection SourceID provided, add to user's primary collection (every user has one)
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
	}
	session := databaseEngine.NewSession()
	defer session.Close()

	var existingRecords []LibraryRecord

	// check if data is already in internal library
	_ = session.Table(libraryTable).Where("media_type = ?", libraryRecord.MediaType).
		Where("media_source = ?", libraryRecord.MediaSource).
		Where("source_id = ?", libraryRecord.SourceID).Find(&existingRecords)

	var libraryID int64
	if len(existingRecords) > 0 {
		// use existing SourceID
		libraryID = existingRecords[0].LibraryID
		// TODO update db with new data
	} else {
		// insert media data to library table
		err := session.Begin()
		if err != nil {
			return err
		}
		_, err = session.Table(libraryTable).Insert(libraryRecord)
		if err != nil {
			_ = session.Rollback()
			return err
		}
		libraryID = libraryRecord.LibraryID
	}

	// create relation between data, user, and collection
	err := insertCollectionRelation(session, userID, libraryID, *collectionID)
	if err != nil {
		_ = session.Rollback()
		return err
	}
	err = session.Commit()
	if err != nil {
		return err
	}
	return nil
}

func GetCollectionFromLibrary(userID int64, mediaType string, collectionID *int64, limit int, offset int) ([]LibraryGroup, int64, error) {
	//var records []LibraryRecord
	if collectionID == nil {
		var collectionRecord CollectionRecord
		has, err := databaseEngine.Table(collectionsTable).Where("owner_user_id = ?", userID).
			Where("is_primary = ?", true).Get(&collectionRecord)
		if err != nil {
			return nil, 0, err
		}
		if !has {
			return nil, 0, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Collection not found, no primary col found for user")
		}
		collectionID = &collectionRecord.CollectionID
	}
	//totalRecords, err := databaseEngine.Table(libraryTable).Where("user_id = ?", userID).Where("media_type = ?", mediaType).
	//	Where("collection_id = ?", collectionID).Count(new(LibraryRecord))
	var libraryGroups []LibraryGroup
	err := databaseEngine.Table(libraryTable).Limit(limit, offset).Where("collection_id = ?", collectionID).Join("INNER", collectionRelationsTable, fmt.Sprintf("%s.library_id = %s.library_id", libraryTable, collectionRelationsTable)).Find(&libraryGroups)
	if err != nil {
		return nil, -1, err
	}
	//TODO remove
	totalRecords, err := databaseEngine.Table(libraryTable).Where("collection_id = ?", collectionID).Join("INNER", collectionRelationsTable, fmt.Sprintf("%s.library_id = %s.library_id", libraryTable, collectionRelationsTable)).Count()
	if err != nil {
		return nil, -1, err
	}
	if err != nil {
		return nil, -1, err
	}
	return libraryGroups, totalRecords, nil
}

func insertCollectionRelation(session *xorm.Session, userID int64, libraryID int64, collectionID int64) error {
	// check if collection exists in collections table
	// TODO should ideally be covered by foreign key constraint, xorm does not handle sync with fk right now
	var collectionRecord CollectionRecord
	has, err := session.Table(collectionsTable).ID(collectionID).Get(&collectionRecord)
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
	// insert record to db
	_, err = databaseEngine.Table(collectionRelationsTable).Insert(CollectionRelation{
		UserID:       userID,
		LibraryID:    libraryID,
		CollectionID: collectionID,
	})
	return err
}

func CreateCollection(record CollectionRecord) error {
	if record.IsPrimary == true {
		_, int64, err := SearchCollection(CollectionRecordQuery{
			OwnerID:   &record.OwnerID,
			IsPrimary: &record.IsPrimary,
		}, 1, 0)
		if err != nil {
			return err
		}
		// primary record already exists
		if int64 != 0 {
			return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Primary col already exists")
		}
	}
	_, err := databaseEngine.Table(collectionsTable).Insert(record)
	if err != nil {
		return err
	}
	return nil
}

func SearchCollection(record CollectionRecordQuery, limit int, offset int) ([]CollectionRecord, int64, error) {
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
	err := sess.Limit(limit, offset).Find(&records)
	if err != nil {
		return nil, -1, err
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
		return nil, -1, err
	}
	return records, totalRecords, nil
}
