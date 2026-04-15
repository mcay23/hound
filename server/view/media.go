package view

import (
	"time"

	"github.com/mcay23/hound/database"
)

type CollectionView struct {
	Records      []MediaRecordCatalog `json:"records"`
	Collection   *CollectionObject    `json:"collection"`
	TotalRecords int64                `json:"total_records"`
	Limit        int                  `json:"limit"`
	Offset       int                  `json:"offset"`
}

type WatchActivityResponse struct {
	WatchActivity []*database.WatchActivity `json:"watch_activity"`
	TotalRecords  int64                     `json:"total_records"`
	Limit         int                       `json:"limit"`
	Offset        int                       `json:"offset"`
}

type GeneralSearchResponse struct {
	TVShowSearchResults *[]MediaRecordCatalog `json:"tv_results"`
	MovieSearchResults  *[]MediaRecordCatalog `json:"movie_results"`
}

type CollectionObject struct {
	CollectionID     int64                 `json:"collection_id"`
	CollectionTitle  string                `json:"collection_title"` // my collection, etc.
	Description      string                `json:"description"`
	OwnerUsername    string                `json:"owner_username"`
	OwnerDisplayName string                `json:"owner_display_name"`
	IsPublic         bool                  `json:"is_public"`
	Tags             *[]database.TagObject `json:"tags"`
	ThumbnailURI     string                `json:"thumbnail_uri,omitempty"` // url for media thumbnails
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
}

type CommentObject struct {
	CommentID        int64     `json:"comment_id"`
	CommentType      string    `json:"comment_type"`
	OwnerUsername    string    `json:"owner_username"`
	OwnerDisplayName string    `json:"owner_display_name"`
	RecordID         int64     `json:"record_id"`
	IsPublic         bool      `json:"is_public"`
	CommentTitle     string    `json:"title"`
	Comment          string    `json:"comment"` // actual content of comment, review
	Score            int       `json:"score"`
	CreatedAt        time.Time `xorm:"timestampz created" json:"created_at"`
	UpdatedAt        time.Time `xorm:"timestampz updated" json:"updated_at"`
}

type MediaFilesResponse struct {
	TotalRecords int                   `json:"total_records"`
	Limit        int                   `json:"limit"`
	Offset       int                   `json:"offset"`
	Files        []*database.MediaFile `json:"files"`
}

type IngestTaskResponse struct {
	TotalRecords int                             `json:"total_records"`
	Limit        int                             `json:"limit"`
	Offset       int                             `json:"offset"`
	Tasks        []database.IngestTaskFullRecord `json:"tasks"`
}
