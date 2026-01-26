package view

import (
	"hound/database"
)

type MediaRecordCatalog struct {
	MediaType        string                 `json:"media_type" binding:"required"`
	MediaSource      string                 `json:"media_source" binding:"required"` // tmdb, openlibrary, etc
	SourceID         string                 `json:"source_id" binding:"required"`
	MediaTitle       string                 `json:"media_title" binding:"required"`
	OriginalTitle    string                 `json:"original_title"`
	Status           string                 `json:"status"` // Returning Series, Released, etc.
	Overview         string                 `json:"overview"`
	Duration         int                    `json:"duration"` // duration/runtime in minutes
	ReleaseDate      string                 `json:"release_date"`
	LastAirDate      string                 `json:"last_air_date"` // for shows, latest episode air date
	NextAirDate      string                 `json:"next_air_date"` // for shows, next scheduled episode air date
	SeasonNumber     *int                   `json:"season_number,omitempty"`
	EpisodeNumber    *int                   `json:"episode_number,omitempty"`
	SeasonCount      *int                   `json:"season_count,omitempty"`
	EpisodeCount     *int                   `json:"episode_count,omitempty"`
	Cast             *[]Credit              `json:"cast,omitempty"`
	Creators         *[]Credit              `json:"creators,omitempty"`
	GuestStars       *[]Credit              `json:"guest_stars,omitempty"`
	VoteCount        int64                  `json:"vote_count"`
	VoteAverage      float32                `json:"vote_average"`
	Popularity       float32                `json:"popularity"`
	ThumbnailURI     string                 `json:"thumbnail_uri"`
	BackdropURI      string                 `json:"backdrop_uri"`
	LogoURI          string                 `json:"logo_uri"`
	Genres           []database.GenreObject `json:"genres"`
	OriginalLanguage string                 `json:"original_language"`
	OriginCountry    []string               `json:"origin_country"`
	ExtraData        map[string]interface{} `json:"extra_data,omitempty"`
}

// to simplify, we typically only return top 20 cast members
// and the creators for tv show, director for movies
type Credit struct {
	MediaSource  string  `json:"meta_source" binding:"required"`
	SourceID     string  `json:"source_id" binding:"required"`
	CreditID     string  `json:"credit_id"`
	Name         string  `json:"name"`
	OriginalName string  `json:"original_name"`
	Character    *string `json:"character,omitempty"`
	ThumbnailURI string  `json:"thumbnail_uri"` // profile pic of person
	Job          string  `json:"job"`           // cast, director, etc.
}
