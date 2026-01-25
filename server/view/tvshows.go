package view

import (
	"hound/database"

	tmdb "github.com/cyruzin/golang-tmdb"
)

type TVSeasonResponseObject struct {
	MediaSource     string                `json:"media_source"` // tmdb, openlibrary, etc
	SourceID        int64                 `json:"source_id"`
	SeasonData      *tmdb.TVSeasonDetails `json:"season"`
	SeasonWatchInfo *[]CommentObject      `json:"watch_info"`
}

type TVShowResults struct {
	Results []MediaCatalogObject `json:"results"`
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
	MediaSource      string                     `json:"media_source"` // tmdb, openlibrary, etc
	MediaType        string                     `json:"media_type"`   // tmdb, openlibrary, etc
	SourceID         int64                      `json:"source_id"`
	MediaTitle       string                     `json:"media_title"`
	OriginalName     string                     `json:"original_name"`
	VoteCount        int64                      `json:"vote_count"`
	VoteAverage      float32                    `json:"vote_average"`
	PosterURL        string                     `json:"poster_url"`
	NumberOfEpisodes int                        `json:"number_of_episodes"`
	NumberOfSeasons  int                        `json:"number_of_seasons"`
	Seasons          []SeasonObjectPartial      `json:"seasons"`
	NextEpisodeToAir tmdb.NextEpisodeToAir      `json:"next_episode_to_air"`
	Networks         []tmdb.Network             `json:"networks"`
	EpisodeRunTime   []int                      `json:"episode_run_time"`
	CreatedBy        []tmdb.CreatedBy           `json:"created_by"`
	Status           string                     `json:"status"` // Returning Series, etc.
	FirstAirDate     string                     `json:"first_air_date"`
	Popularity       float32                    `json:"popularity"`
	Genres           []tmdb.Genre               `json:"genres"`
	OriginalLanguage string                     `json:"original_language"`
	BackdropURL      string                     `json:"backdrop_url"`
	LogoURL          string                     `json:"logo_url"`
	Overview         string                     `json:"overview"`
	OriginCountry    []string                   `json:"origin_country"`
	TVCredits        *tmdb.TVCredits            `json:"credits"`
	Videos           *tmdb.VideoResults         `json:"videos"`
	Recommendations  *tmdb.TVRecommendations    `json:"recommendations"`
	WatchProviders   *tmdb.WatchProviderResults `json:"watch_providers"`
	ExternalIDs      *tmdb.TVExternalIDs        `json:"external_ids"`
	Comments         *[]CommentObject           `json:"comments"`
}

type MediaRewatchRecordWatchEvents struct {
	database.RewatchRecord
	TargetSeason *int                              `json:"target_season,omitempty"`
	WatchEvents  []*database.WatchEventMediaRecord `json:"watch_events"`
}
