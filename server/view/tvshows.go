package view

import (
	tmdb "github.com/cyruzin/golang-tmdb"
	"hound/model/sources"
)

type TVGenre struct {
}

type TMDBSearchResultObject struct {
	MediaSource      string                 `json:"media_source" binding:"required"` // tmdb, openlibrary, etc
	MediaType        string                 `json:"media_type" binding:"required"`
	SourceID         int64                  `json:"source_id" binding:"required"`
	MediaTitle       string                 `json:"media_title" binding:"required"`
	OriginalName     string                 `json:"original_name"`
	VoteCount        int64                  `json:"vote_count"`
	VoteAverage      float32                `json:"vote_average"`
	PosterURL        string                 `json:"poster_url"`
	FirstAirDate     string                 `json:"first_air_date"`
	ReleaseDate      string                 `json:"release_date"`
	Popularity       float32                `json:"popularity"`
	Genres           *[]sources.GenreObject `json:"genres"`
	OriginalLanguage string                 `json:"original_language"`
	BackdropURL      string                 `json:"backdrop_url"`
	Overview         string                 `json:"overview"`
	OriginCountry    []string               `json:"origin_country"`
}

type TVSeasonResponseObject struct {
	MediaSource     string                `json:"media_source"` // tmdb, openlibrary, etc
	SourceID        int64                 `json:"source_id"`
	SeasonData      *tmdb.TVSeasonDetails `json:"season"`
	SeasonWatchInfo *[]CommentObject        `json:"watch_info"`
}

type TVShowResults struct {
	Results []TMDBSearchResultObject `json:"results"`
}

type TVShowDetails struct {
	*tmdb.TVDetails
	BackdropURL string `json:"backdrop_url"`
	PosterURL   string `json:"poster_url"`
}

type SeasonObjectPartial struct {
	AirDate      string `json:"air_date"`
	EpisodeCount int    `json:"episode_count"`
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	PosterURL    string `json:"poster_url"`
	SeasonNumber int    `json:"season_number"`
}

type TVShowFullObject struct {
	MediaSource      string                `json:"media_source"` // tmdb, openlibrary, etc
	MediaType        string                `json:"media_type"`   // tmdb, openlibrary, etc
	SourceID         int64                 `json:"source_id"`
	MediaTitle       string                `json:"media_title"`
	OriginalName     string                `json:"original_name"`
	VoteCount        int64                 `json:"vote_count"`
	VoteAverage      float32               `json:"vote_average"`
	PosterURL        string                `json:"poster_url"`
	NumberOfEpisodes int                   `json:"number_of_episodes"`
	NumberOfSeasons  int                   `json:"number_of_seasons"`
	Seasons          []SeasonObjectPartial `json:"seasons"`
	NextEpisodeToAir struct {
		AirDate        string  `json:"air_date"`
		EpisodeNumber  int     `json:"episode_number"`
		ID             int64   `json:"id"`
		Name           string  `json:"name"`
		Overview       string  `json:"overview"`
		ProductionCode string  `json:"production_code"`
		SeasonNumber   int     `json:"season_number"`
		ShowID         int64   `json:"show_id"`
		StillPath      string  `json:"still_path"`
		VoteAverage    float32 `json:"vote_average"`
		VoteCount      int64   `json:"vote_count"`
	} `json:"next_episode_to_air"`
	Networks []struct {
		Name          string `json:"name"`
		ID            int64  `json:"id"`
		LogoPath      string `json:"logo_path"`
		OriginCountry string `json:"origin_country"`
	} `json:"networks"`
	EpisodeRunTime []int `json:"episode_run_time"`
	CreatedBy      []struct {
		ID          int64  `json:"id"`
		CreditID    string `json:"credit_id"`
		Name        string `json:"name"`
		Gender      int    `json:"gender"`
		ProfilePath string `json:"profile_path"`
	} `json:"created_by"`
	Status       string  `json:"status"` // Returning Series, etc.
	FirstAirDate string  `json:"first_air_date"`
	Popularity   float32 `json:"popularity"`
	Genres       []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	OriginalLanguage string                  `json:"original_language"`
	BackdropURL      string                  `json:"backdrop_url"`
	Overview         string                  `json:"overview"`
	OriginCountry    []string                `json:"origin_country"`
	TVCredits        *tmdb.TVCredits         `json:"credits"`
	Videos           *tmdb.TVVideos          `json:"videos"`
	Recommendations  *tmdb.TVRecommendations `json:"recommendations"`
	WatchProviders   *tmdb.TVWatchProviders  `json:"watch_providers"`
	Comments         *[]CommentObject        `json:"comments"`
}
