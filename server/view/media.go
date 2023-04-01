package view

import "hound/model/sources"

type LibraryView struct {
	Results      *[]LibraryItem `json:"results"`
	TotalRecords int64          `json:"total_records"`
	Limit        int            `json:"limit"`
	Offset       int            `json:"offset"`
}

// store user saved libraries
type LibraryItem struct {
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
	TVShowSearchResults *[]TMDBSearchResultObject         `json:"tv_results"`
	MovieSearchResults  *[]TMDBSearchResultObject         `json:"movie_results"`
	GameSearchResults   *sources.IGDBSearchResultObject `json:"game_results"`
}
