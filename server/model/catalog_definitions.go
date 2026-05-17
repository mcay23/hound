package model

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/mcay23/hound/config"
	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/sources"
)

type CatalogDefinition struct {
	CatalogTitle  string `json:"catalog_title"`
	CatalogSource string `json:"catalog_source"`
	CatalogID     string `json:"catalog_id"`
	MediaType     string `json:"media_type,omitempty"`
	Description   string `json:"description"`
}

type CatalogDefinitionsResponse struct {
	Catalogs          []CatalogDefinition `json:"catalogs"`
	MDBListConfigured bool                `json:"mdblist_configured"`
}

type tmdbCatalogDefinition struct {
	CatalogDefinition
	DiscoverType  string
	DiscoverQuery string
}

var internalCatalogDefinitions = []CatalogDefinition{
	{
		CatalogTitle:  "Recently Added to Hound",
		CatalogSource: database.CatalogSourceInternal,
		CatalogID:     "hound-library",
		Description:   "Recently added movies and shows in Hound",
	},
	{
		CatalogTitle:  "Recently Added Shows to Hound",
		CatalogSource: database.CatalogSourceInternal,
		CatalogID:     "hound-library-shows",
		MediaType:     database.MediaTypeTVShow,
		Description:   "Recently added TV shows in Hound",
	},
	{
		CatalogTitle:  "Recently Added Movies to Hound",
		CatalogSource: database.CatalogSourceInternal,
		CatalogID:     "hound-library-movies",
		MediaType:     database.MediaTypeMovie,
		Description:   "Recently added movies in Hound",
	},
	{
		CatalogTitle:  "Recently Added to Your Collections",
		CatalogSource: database.CatalogSourceInternal,
		CatalogID:     "hound-recent-collection",
		Description:   "Media recently added to your collections",
	},
	{
		CatalogTitle:  "Movies You Might Like",
		CatalogSource: database.CatalogSourceInternal,
		CatalogID:     "hound-recommended-movies",
		MediaType:     database.MediaTypeMovie,
		Description:   "Recommended Movies for you based on your watch history",
	},
	{
		CatalogTitle:  "Shows You Might Like",
		CatalogSource: database.CatalogSourceInternal,
		CatalogID:     "hound-recommended-shows",
		MediaType:     database.MediaTypeTVShow,
		Description:   "Recommended Shows for you based on your watch history",
	},
}

var tmdbStaticCatalogDefinitions = []tmdbCatalogDefinition{
	{
		CatalogDefinition: CatalogDefinition{
			CatalogTitle:  "Trending Shows",
			CatalogSource: database.CatalogSourceTMDB,
			CatalogID:     "trending-shows",
			MediaType:     database.MediaTypeTVShow,
			Description:   "Trending TV Shows on TMDB this week.",
		},
	},
	{
		CatalogDefinition: CatalogDefinition{
			CatalogTitle:  "Trending Movies",
			CatalogSource: database.CatalogSourceTMDB,
			CatalogID:     "trending-movies",
			MediaType:     database.MediaTypeMovie,
			Description:   "Trending movies on TMDB this week.",
		},
	},
	{
		CatalogDefinition: CatalogDefinition{
			CatalogTitle:  "Movies on Netflix",
			CatalogSource: database.CatalogSourceTMDB,
			CatalogID:     "netflix-movies",
			MediaType:     database.MediaTypeMovie,
			Description:   "Movies available on Netflix",
		},
		DiscoverType:  sources.DiscoverTypeWatchProvider,
		DiscoverQuery: sources.TMDBProviderNetflix,
	},
	{
		CatalogDefinition: CatalogDefinition{
			CatalogTitle:  "Shows on Netflix",
			CatalogSource: database.CatalogSourceTMDB,
			CatalogID:     "netflix-shows",
			MediaType:     database.MediaTypeTVShow,
			Description:   "TV Shows available on Netflix",
		},
		DiscoverType:  sources.DiscoverTypeWatchProvider,
		DiscoverQuery: sources.TMDBProviderNetflix,
	},
	{
		CatalogDefinition: CatalogDefinition{
			CatalogTitle:  "Movies on Disney Plus",
			CatalogSource: database.CatalogSourceTMDB,
			CatalogID:     "disney-plus-movies",
			MediaType:     database.MediaTypeMovie,
			Description:   "Movies available on Disney Plus",
		},
		DiscoverType:  sources.DiscoverTypeWatchProvider,
		DiscoverQuery: sources.TMDBProviderDisneyPlus,
	},
	{
		CatalogDefinition: CatalogDefinition{
			CatalogTitle:  "Shows on Disney Plus",
			CatalogSource: database.CatalogSourceTMDB,
			CatalogID:     "disney-plus-shows",
			MediaType:     database.MediaTypeTVShow,
			Description:   "TV Shows available on Disney Plus",
		},
		DiscoverType:  sources.DiscoverTypeWatchProvider,
		DiscoverQuery: sources.TMDBProviderDisneyPlus,
	},
	{
		CatalogDefinition: CatalogDefinition{
			CatalogTitle:  "Movies on HBO Max",
			CatalogSource: database.CatalogSourceTMDB,
			CatalogID:     "hbo-max-movies",
			MediaType:     database.MediaTypeMovie,
			Description:   "Movies available on HBO Max",
		},
		DiscoverType:  sources.DiscoverTypeWatchProvider,
		DiscoverQuery: sources.TMDBProviderHBOMax,
	},
	{
		CatalogDefinition: CatalogDefinition{
			CatalogTitle:  "Shows on HBO Max",
			CatalogSource: database.CatalogSourceTMDB,
			CatalogID:     "hbo-max-shows",
			MediaType:     database.MediaTypeTVShow,
			Description:   "TV Shows available on HBO Max",
		},
		DiscoverType:  sources.DiscoverTypeWatchProvider,
		DiscoverQuery: sources.TMDBProviderHBOMax,
	},
	{
		CatalogDefinition: CatalogDefinition{
			CatalogTitle:  "Movies on Apple TV",
			CatalogSource: database.CatalogSourceTMDB,
			CatalogID:     "apple-tv-movies",
			MediaType:     database.MediaTypeMovie,
			Description:   "Movies available on Apple TV",
		},
		DiscoverType:  sources.DiscoverTypeWatchProvider,
		DiscoverQuery: sources.TMDBProviderAppleTV,
	},
	{
		CatalogDefinition: CatalogDefinition{
			CatalogTitle:  "Shows on Apple TV",
			CatalogSource: database.CatalogSourceTMDB,
			CatalogID:     "apple-tv-shows",
			MediaType:     database.MediaTypeTVShow,
			Description:   "TV Shows available on Apple TV",
		},
		DiscoverType:  sources.DiscoverTypeWatchProvider,
		DiscoverQuery: sources.TMDBProviderAppleTV,
	},
	{
		CatalogDefinition: CatalogDefinition{
			CatalogTitle:  "Movies on Amazon Prime",
			CatalogSource: database.CatalogSourceTMDB,
			CatalogID:     "amazon-prime-movies",
			MediaType:     database.MediaTypeMovie,
			Description:   "Movies available on Amazon Prime",
		},
		DiscoverType:  sources.DiscoverTypeWatchProvider,
		DiscoverQuery: sources.TMDBProviderAmazonVideo,
	},
	{
		CatalogDefinition: CatalogDefinition{
			CatalogTitle:  "Shows on Amazon Prime",
			CatalogSource: database.CatalogSourceTMDB,
			CatalogID:     "amazon-prime-shows",
			MediaType:     database.MediaTypeTVShow,
			Description:   "TV Shows available on Amazon Prime",
		},
		DiscoverType:  sources.DiscoverTypeWatchProvider,
		DiscoverQuery: sources.TMDBProviderAmazonVideo,
	},
	{
		CatalogDefinition: CatalogDefinition{
			CatalogTitle:  "Movies on Paramount",
			CatalogSource: database.CatalogSourceTMDB,
			CatalogID:     "paramount-movies",
			MediaType:     database.MediaTypeMovie,
			Description:   "Movies available on Paramount",
		},
		DiscoverType:  sources.DiscoverTypeWatchProvider,
		DiscoverQuery: sources.TMDBProviderParamount,
	},
	{
		CatalogDefinition: CatalogDefinition{
			CatalogTitle:  "Shows on Paramount",
			CatalogSource: database.CatalogSourceTMDB,
			CatalogID:     "paramount-shows",
			MediaType:     database.MediaTypeTVShow,
			Description:   "TV Shows available on Paramount",
		},
		DiscoverType:  sources.DiscoverTypeWatchProvider,
		DiscoverQuery: sources.TMDBProviderParamount,
	},
	{
		CatalogDefinition: CatalogDefinition{
			CatalogTitle:  "Movies on Hulu",
			CatalogSource: database.CatalogSourceTMDB,
			CatalogID:     "hulu-movies",
			MediaType:     database.MediaTypeMovie,
			Description:   "Movies available on Hulu",
		},
		DiscoverType:  sources.DiscoverTypeWatchProvider,
		DiscoverQuery: sources.TMDBProviderHulu,
	},
	{
		CatalogDefinition: CatalogDefinition{
			CatalogTitle:  "Shows on Hulu",
			CatalogSource: database.CatalogSourceTMDB,
			CatalogID:     "hulu-shows",
			MediaType:     database.MediaTypeTVShow,
			Description:   "TV Shows available on Hulu",
		},
		DiscoverType:  sources.DiscoverTypeWatchProvider,
		DiscoverQuery: sources.TMDBProviderHulu,
	},
}

func getTMDBCatalogDefinition(catalogID string) (*tmdbCatalogDefinition, bool) {
	for _, catalog := range tmdbStaticCatalogDefinitions {
		if catalog.CatalogID == catalogID {
			return &catalog, true
		}
	}
	genreCatalogs, err := getGenreCatalogDefinitions()
	if err == nil {
		for _, catalog := range genreCatalogs {
			if catalog.CatalogID == catalogID {
				genreID := -1
				if after, ok := strings.CutPrefix(catalogID, "genre-movies-"); ok {
					genreID, err = strconv.Atoi(after)
				} else if after, ok := strings.CutPrefix(catalogID, "genre-shows-"); ok {
					genreID, err = strconv.Atoi(after)
				}
				if genreID == -1 || err != nil {
					slog.Error("Error parsing genre ID from catalog ID: "+catalogID, "error", err)
					return nil, false
				}
				return &tmdbCatalogDefinition{
					CatalogDefinition: catalog,
					DiscoverType:      sources.DiscoverTypeGenre,
					DiscoverQuery:     strconv.Itoa(genreID),
				}, true
			}
		}
	}
	return nil, false
}

func GetCatalogDefinitions() ([]CatalogDefinition, error) {
	catalogs := make([]CatalogDefinition, 0, len(internalCatalogDefinitions)+len(tmdbStaticCatalogDefinitions))
	catalogs = append(catalogs, internalCatalogDefinitions...)
	for _, catalog := range tmdbStaticCatalogDefinitions {
		catalogs = append(catalogs, catalog.CatalogDefinition)
	}
	genreCatalogs, err := getGenreCatalogDefinitions()
	if err != nil {
		return nil, err
	}
	catalogs = append(catalogs, genreCatalogs...)
	return catalogs, nil
}

func GetCatalogDefinitionsResponse() (*CatalogDefinitionsResponse, error) {
	catalogs, err := GetCatalogDefinitions()
	if err != nil {
		return nil, err
	}
	return &CatalogDefinitionsResponse{
		Catalogs:          catalogs,
		MDBListConfigured: config.MDBListAPIKey != "",
	}, nil
}

func getGenreCatalogDefinitions() ([]CatalogDefinition, error) {
	movieGenres, err := database.GetGenresByType(database.MediaTypeMovie)
	if err != nil {
		return nil, fmt.Errorf("failed to get movie genre catalogs: %w", err)
	}
	tvGenres, err := database.GetGenresByType(database.MediaTypeTVShow)
	if err != nil {
		return nil, fmt.Errorf("failed to get tv genre catalogs: %w", err)
	}
	catalogs := make([]CatalogDefinition, 0, len(movieGenres)+len(tvGenres))
	for _, genre := range movieGenres {
		catalogs = append(catalogs, CatalogDefinition{
			CatalogTitle:  genre.Genre + " Movies",
			CatalogSource: database.CatalogSourceTMDB,
			CatalogID:     fmt.Sprintf("genre-movies-%d", genre.GenreID),
			MediaType:     database.MediaTypeMovie,
			Description:   fmt.Sprintf("Movies in the %s genre", genre.Genre),
		})
	}
	for _, genre := range tvGenres {
		catalogs = append(catalogs, CatalogDefinition{
			CatalogTitle:  genre.Genre + " Shows",
			CatalogSource: database.CatalogSourceTMDB,
			CatalogID:     fmt.Sprintf("genre-shows-%d", genre.GenreID),
			MediaType:     database.MediaTypeTVShow,
			Description:   fmt.Sprintf("TV Shows in the %s genre", genre.Genre),
		})
	}
	return catalogs, nil
}
