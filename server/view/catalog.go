package view

import "hound/sources"

type MediaCatalogObject struct {
	MediaType        string                 `json:"media_type" binding:"required"`
	MediaSource      string                 `json:"media_source" binding:"required"` // tmdb, openlibrary, etc
	SourceID         string                 `json:"source_id" binding:"required"`
	MediaTitle       string                 `json:"media_title" binding:"required"`
	OriginalName     string                 `json:"original_name"`
	Overview         string                 `json:"overview"`
	VoteCount        int64                  `json:"vote_count"`
	VoteAverage      float32                `json:"vote_average"`
	Popularity       float32                `json:"popularity"`
	ThumbnailURL     string                 `json:"thumbnail_url"`
	BackdropURL      string                 `json:"backdrop_url"`
	ReleaseDate      string                 `json:"release_date"`
	FirstAirDate     string                 `json:"first_air_date"` // for shows
	Genres           *[]sources.GenreObject `json:"genres"`
	OriginalLanguage string                 `json:"original_language"`
	OriginCountry    []string               `json:"origin_country"`
}
