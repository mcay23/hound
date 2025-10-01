package view

import tmdb "github.com/cyruzin/golang-tmdb"

type MovieFullObject struct {
	MediaSource string `json:"media_source"` // tmdb, openlibrary, etc
	MediaType   string `json:"media_type"`   // tmdb, openlibrary, etc
	SourceID    int64  `json:"source_id"`
	MediaTitle  string `json:"media_title"`
	BackdropURL string `json:"backdrop_url"`
	PosterURL   string `json:"poster_url"`
	Budget      int64  `json:"budget"`
	Homepage            string  `json:"homepage"`
	IMDbID              string  `json:"imdb_id"`
	OriginalLanguage    string  `json:"original_language"`
	OriginalTitle       string  `json:"original_title"`
	Overview            string  `json:"overview"`
	Popularity          float32 `json:"popularity"`
	ReleaseDate     string                      `json:"release_date"`
	Revenue         int64                       `json:"revenue"`
	Runtime         int                         `json:"runtime"`
	Status          string                      `json:"status"`
	Tagline         string                      `json:"tagline"`
	VoteAverage     float32                     `json:"vote_average"`
	VoteCount       int64                       `json:"vote_count"`
	ProductionCompanies *[]tmdb.ProductionCompany   `json:"production_companies"`
	Genres              *[]tmdb.Genre               `json:"genres"`
	MovieCredits        *tmdb.MovieCredits          `json:"credits"`
	Videos              *tmdb.VideoResults          `json:"videos"`
	Recommendations     *tmdb.MovieRecommendations  `json:"recommendations"`
	WatchProviders      *tmdb.WatchProviderResults  `json:"watch_providers"`
	Comments            *[]CommentObject            `json:"comments"`
}
