package view

import (
	"hound/model/database"
	"hound/model/sources"
	"time"
)

type CollectionView struct {
	Results      *[]LibraryObject      `json:"results"`
	Collection   *CollectionRecordView `json:"collection"`
	TotalRecords int64                 `json:"total_records"`
	Limit        int                   `json:"limit"`
	Offset       int                   `json:"offset"`
}

// store user saved libraries
type LibraryObject struct {
	MediaType    string      `json:"media_type"`    // books,tvshows, etc.
	MediaSource  string      `json:"media_source"`  // tmdb, openlibrary, etc
	SourceID     string      `json:"source_id"`     // tmdb id, etc.
	MediaTitle   string      `json:"media_title"`   // game of thrones, etc.
	ReleaseDate  string      `json:"release_date"`  //
	Description  string      `json:"description"`   // game of thrones is a show about ...
	ThumbnailURL *string     `json:"thumbnail_url"` // url for media thumbnails
	Tags         interface{} `json:"tags"`          // to store genres, tags
	UserTags     interface{} `json:"user_tags"`
}

type GeneralSearchResponse struct {
	TVShowSearchResults *[]TMDBSearchResultObject       `json:"tv_results"`
	MovieSearchResults  *[]TMDBSearchResultObject       `json:"movie_results"`
	GameSearchResults   *sources.IGDBSearchResultObject `json:"game_results"`
}

type CollectionRecordView struct {
	CollectionID    int64                 `json:"collection_id"`
	CollectionTitle string                `json:"collection_title"` // my collection, etc.
	Description     string                `json:"description"`
	Username        string                `json:"owner_user_id"`
	IsPrimary       bool                  `json:"is_primary"` // is the user's primary collection, not deletable
	IsPublic        bool                  `json:"is_public"`
	Tags            *[]database.TagObject `json:"tags"`
	ThumbnailURL    *string               `json:"thumbnail_url"` // url for media thumbnails
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

type CommentObject struct {
	CommentTitle string    `json:"title"`
	CommentID    int64     `json:"comment_id"`
	CommentType  string    `json:"comment_type"`
	UserID       string    `json:"user_id"`
	LibraryID    int64     `json:"library_id"`
	IsPrivate    bool      `json:"is_private"`
	Comment      string    `json:"comment"`  // actual content of comment, review
	TagData      string    `json:"tag_data"` // extra tag info, eg. season, episode
	Score        int       `json:"score"`
	CreatedAt    time.Time `xorm:"created" json:"created_at"`
	UpdatedAt    time.Time `xorm:"updated" json:"updated_at"`
}
