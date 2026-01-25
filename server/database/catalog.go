package database

type MediaRecordCatalog struct {
	RecordType       string        `json:"media_type" binding:"required"`
	MediaSource      string        `json:"media_source" binding:"required"` // tmdb, openlibrary, etc
	SourceID         string        `json:"source_id" binding:"required"`
	MediaTitle       string        `json:"media_title" binding:"required"`
	OriginalTitle    string        `json:"original_title"`
	Status           string        `xorm:"'status'" json:"status"` // Returning Series, Released, etc.
	Overview         string        `json:"overview"`
	Duration         int           `xorm:"'duration'" json:"duration"` // duration/runtime in minutes
	ReleaseDate      string        `json:"release_date"`
	LastAirDate      string        `xorm:"'last_air_date'" json:"last_air_date"` // for shows, latest episode air date
	NextAirDate      string        `xorm:"'next_air_date'" json:"next_air_date"` // for shows, next scheduled episode air date
	SeasonNumber     *int          `json:"season_number,omitempty"`
	EpisodeNumber    *int          `json:"episode_number,omitempty"`
	VoteCount        int64         `json:"vote_count"`
	VoteAverage      float32       `json:"vote_average"`
	Popularity       float32       `json:"popularity"`
	ThumbnailURL     string        `json:"thumbnail_url"`
	BackdropURL      string        `json:"backdrop_url"`
	StillURL         string        `json:"still_url"`
	Genres           []GenreObject `json:"genres"`
	OriginalLanguage string        `json:"original_language"`
	OriginCountry    []string      `json:"origin_country"`
}

const (
	CatalogTypeInternal   = "catalog-internal"
	CatalogTypeCollection = "catalog-collection"
	CatalogTypeRemote     = "catalog-remote"
)
