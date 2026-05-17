package model

import (
	"fmt"
	"log/slog"
	"math/rand"
	"strconv"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/sources"
	"github.com/mcay23/hound/view"
)

const (
	userRecommendationsCacheTTL    = 24 * time.Hour
	userRecommendationsStaleExpiry = 14 * 24 * time.Hour
	userRecommendationsCacheKey    = "user_recommendations|%d|%s|%d"
)

var (
	recsLock           catalogLock
	recsRefreshingLock catalogLock
)

/*
Gets recommendations based on user watch history
Rough algorithm:
1. Get watched movies/shows from last 6 months
2. Shuffle and select a random subset
3. For each item in the subset, get 3 recommendations from TMDB
4. Deduplicate, shuffle, choose 'limit' items, and return
Since this is an expensive call, it has two cache keys, an expiry key,
and the main cache key. If cache is expired but still in cache,
return the cached object first and update in background.
*/
func GetUserRecommendations(userID int64, mediaType string, limit int) ([]view.MediaRecordCatalog, error) {
	cacheKey := fmt.Sprintf(userRecommendationsCacheKey, userID, mediaType, limit)
	// Prevent concurrent requests since performance is slow
	userLock := recsLock.getLock(cacheKey)
	userLock.Lock()
	defer userLock.Unlock()

	var recs []view.MediaRecordCatalog
	found, _ := database.GetCache(cacheKey, &recs)
	if found {
		// check if expired. If expired, run in background while returning current cache early
		expiryKey := cacheKey + "|expiry"
		notExpired, _ := database.GetCache(expiryKey, nil)
		if notExpired {
			return recs, nil
		}
		// only allow one refresh to run at a time
		refreshLock := recsRefreshingLock.getLock(cacheKey)
		if refreshLock.TryLock() {
			go func() {
				defer refreshLock.Unlock()
				_, err := getUserRecommendationsInternal(userID, mediaType, limit)
				if err != nil {
					slog.Error("failed refreshing recommendations", "userID", userID, "mediaType", mediaType, "limit", limit, "err", err)
				}
			}()
		}
		return recs, nil
	}
	return getUserRecommendationsInternal(userID, mediaType, limit)
}

func getUserRecommendationsInternal(userID int64, mediaType string, limit int) ([]view.MediaRecordCatalog, error) {
	cacheKey := fmt.Sprintf(userRecommendationsCacheKey, userID, mediaType, limit)
	after := time.Now().AddDate(-6, 0, 0)
	parents, err := database.GetUniqueWatchedParents(userID, limit*2, 0, after)
	if err != nil {
		return nil, fmt.Errorf("error getting watched parents: %w", err)
	}
	sourceItem := []database.WatchEventMediaRecord{}
	for _, item := range parents {
		if item != nil && item.RecordType == mediaType {
			sourceItem = append(sourceItem, *item)
		}
	}
	rand.Shuffle(len(sourceItem), func(i, j int) {
		sourceItem[i], sourceItem[j] = sourceItem[j], sourceItem[i]
	})
	if len(sourceItem) > limit {
		sourceItem = sourceItem[:limit]
	}
	rawRecs := []view.MediaRecordCatalog{}
	recsPerSource := 3
	for _, item := range sourceItem {
		sourceID, err := strconv.Atoi(item.SourceID)
		if err != nil {
			slog.Error("Invalid tmdb source id: " + item.SourceID)
			continue
		}
		if mediaType == database.MediaTypeMovie {
			results, err := sources.GetMovieFromIDTMDB(sourceID)
			if err != nil {
				continue
			}
			recs := createCatalogObject(results.Recommendations.MovieRecommendationsResults, nil, recsPerSource)
			rawRecs = append(rawRecs, recs...)
		} else if mediaType == database.MediaTypeTVShow {
			results, err := sources.GetTVShowFromIDTMDB(sourceID)
			if err != nil {
				continue
			}
			recs := createCatalogObject(nil, results.Recommendations.TVRecommendationsResults, recsPerSource)
			rawRecs = append(rawRecs, recs...)
		}
	}
	// deduplicate results
	seen := make(map[string]struct{})
	dedupedRecs := make([]view.MediaRecordCatalog, 0, len(rawRecs))
	for _, item := range rawRecs {
		key := item.MediaSource + "-" + item.MediaType + "-" + item.SourceID
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		dedupedRecs = append(dedupedRecs, item)
	}
	rand.Shuffle(len(dedupedRecs), func(i, j int) {
		dedupedRecs[i], dedupedRecs[j] = dedupedRecs[j], dedupedRecs[i]
	})
	// if recommendations are sparse, cache shorter
	if len(dedupedRecs) >= limit {
		dedupedRecs = dedupedRecs[:limit]
		database.SetCache(cacheKey, dedupedRecs, userRecommendationsCacheTTL)
		database.SetCache(cacheKey+"|expiry", 1, userRecommendationsStaleExpiry)
	} else {
		database.SetCache(cacheKey, dedupedRecs, 1*time.Hour)
		database.SetCache(cacheKey+"|expiry", 1, userRecommendationsStaleExpiry)
	}
	return dedupedRecs, nil
}

// limit clamps the number of recommendations per source
// some things such as duration and logo image are missing
func createCatalogObject(movieRecs *tmdb.MovieRecommendationsResults,
	showRecs *tmdb.TVRecommendationsResults, limit int) []view.MediaRecordCatalog {
	viewArray := []view.MediaRecordCatalog{}
	if movieRecs != nil {
		shuffledResults := movieRecs.Results
		rand.Shuffle(len(shuffledResults), func(i, j int) {
			shuffledResults[i], shuffledResults[j] = shuffledResults[j], shuffledResults[i]
		})
		if len(shuffledResults) > limit {
			shuffledResults = shuffledResults[:limit]
		}
		for _, item := range shuffledResults {
			catalogObject := view.MediaRecordCatalog{
				MediaType:        database.MediaTypeMovie,
				MediaSource:      sources.MediaSourceTMDB,
				SourceID:         strconv.Itoa(int(item.ID)),
				MediaTitle:       item.Title,
				OriginalTitle:    item.OriginalTitle,
				Overview:         item.Overview,
				VoteCount:        item.VoteCount,
				VoteAverage:      item.VoteAverage,
				Popularity:       item.Popularity,
				ThumbnailURI:     internal.GetTMDBImageURL(item.PosterPath, "w300"),
				BackdropURI:      internal.GetTMDBImageURL(item.BackdropPath, "original"),
				ReleaseDate:      item.ReleaseDate,
				Genres:           sources.GetGenresMapFromTMDBIDs(item.GenreIDs, database.MediaTypeMovie),
				OriginalLanguage: item.OriginalLanguage,
			}
			viewArray = append(viewArray, catalogObject)
		}
	} else if showRecs != nil {
		shuffledResults := showRecs.Results
		rand.Shuffle(len(shuffledResults), func(i, j int) {
			shuffledResults[i], shuffledResults[j] = shuffledResults[j], shuffledResults[i]
		})
		if len(shuffledResults) > limit {
			shuffledResults = shuffledResults[:limit]
		}
		for _, item := range shuffledResults {
			catalogObject := view.MediaRecordCatalog{
				MediaType:        database.MediaTypeTVShow,
				MediaSource:      sources.MediaSourceTMDB,
				SourceID:         strconv.Itoa(int(item.ID)),
				MediaTitle:       item.Name,
				OriginalTitle:    item.OriginalName,
				Overview:         item.Overview,
				VoteCount:        item.VoteCount,
				VoteAverage:      item.VoteAverage,
				Popularity:       item.Popularity,
				ThumbnailURI:     internal.GetTMDBImageURL(item.PosterPath, "w300"),
				BackdropURI:      internal.GetTMDBImageURL(item.BackdropPath, "original"),
				ReleaseDate:      item.FirstAirDate,
				Genres:           sources.GetGenresMapFromTMDBIDs(item.GenreIDs, database.MediaTypeTVShow),
				OriginalLanguage: item.OriginalLanguage,
			}
			viewArray = append(viewArray, catalogObject)
		}
	}
	return viewArray
}
