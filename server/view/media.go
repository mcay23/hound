package view

import (
	"hound/database"
	"hound/sources"
	"time"
)

type CollectionView struct {
	Records      []MediaRecordCatalog `json:"records"`
	Collection   *CollectionObject    `json:"collection"`
	TotalRecords int64                `json:"total_records"`
	Limit        int                  `json:"limit"`
	Offset       int                  `json:"offset"`
}

type WatchActivityResponse struct {
	WatchActivity []*database.WatchEventMediaRecord `json:"watch_activity"`
	TotalRecords  int64                             `json:"total_records"`
	Limit         int                               `json:"limit"`
	Offset        int                               `json:"offset"`
}

type GeneralSearchResponse struct {
	TVShowSearchResults *[]MediaRecordCatalog           `json:"tv_results"`
	MovieSearchResults  *[]MediaRecordCatalog           `json:"movie_results"`
	GameSearchResults   *sources.IGDBSearchResultObject `json:"game_results"`
}

type CollectionObject struct {
	CollectionID    int64                 `json:"collection_id"`
	CollectionTitle string                `json:"collection_title"` // my collection, etc.
	Description     string                `json:"description"`
	OwnerUsername   string                `json:"owner_username"`
	IsPublic        bool                  `json:"is_public"`
	Tags            *[]database.TagObject `json:"tags"`
	ThumbnailURI    string                `json:"thumbnail_uri,omitempty"` // url for media thumbnails
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

type CommentObject struct {
	CommentTitle string    `json:"title"`
	CommentID    int64     `json:"comment_id"`
	CommentType  string    `json:"comment_type"`
	UserID       string    `json:"user_id"`
	RecordID     int64     `json:"record_id"`
	IsPrivate    bool      `json:"is_private"`
	Comment      string    `json:"comment"`  // actual content of comment, review
	TagData      string    `json:"tag_data"` // extra tag info, eg. season, episode
	Score        int       `json:"score"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	CreatedAt    time.Time `xorm:"timestampz created" json:"created_at"`
	UpdatedAt    time.Time `xorm:"timestampz updated" json:"updated_at"`
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
