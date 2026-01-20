package database

import (
	"time"
)

/*
	Queries to get watch stats for users,
	such as no. of movies watched, duration, etc.
*/

// movies, shows watched are unique, so watching the same movie twice will
// only count once
// However total duration counts all watches
type WatchStats struct {
	MoviesWatched         int64     `json:"movies_watched"`
	EpisodesWatched       int64     `json:"episodes_watched"` // a show is included even though it's not completed
	TotalMoviesDuration   int64     `json:"total_movies_duration"`
	TotalEpisodesDuration int64     `json:"total_episodes_duration"`
	StartTime             time.Time `json:"start_time"` // query start, finish time
	FinishTime            time.Time `json:"finish_time"`
}

func GetWatchStats(userID int64, startTime *time.Time, finishTime *time.Time) (*WatchStats, error) {
	// if time not provided, return lifetime stats
	if startTime == nil {
		zeroTime := time.Time{}
		startTime = &zeroTime
	}
	if finishTime == nil {
		now := time.Now()
		finishTime = &now
	}
	stats := &WatchStats{
		StartTime:  *startTime,
		FinishTime: *finishTime,
	}
	// 1. Unique movies watched
	moviesCount, err := databaseEngine.Table(watchEventsTable).Alias("we").
		Join("INNER", "rewatches r", "r.rewatch_id = we.rewatch_id").
		Join("INNER", "media_records mr", "mr.record_id = we.record_id").
		Where("r.user_id = ?", userID).
		And("mr.record_type = ?", RecordTypeMovie).
		And("we.watched_at BETWEEN ? AND ?", startTime, finishTime).
		Distinct("we.record_id").
		Count()
	if err != nil {
		return nil, err
	}
	stats.MoviesWatched = moviesCount
	// 2. Unique episodes watched
	episodesCount, err := databaseEngine.Table(watchEventsTable).Alias("we").
		Join("INNER", "rewatches r", "r.rewatch_id = we.rewatch_id").
		Join("INNER", "media_records mr", "mr.record_id = we.record_id").
		Where("r.user_id = ?", userID).
		And("mr.record_type = ?", RecordTypeEpisode).
		And("we.watched_at BETWEEN ? AND ?", startTime, finishTime).
		Distinct("we.record_id").
		Count()
	if err != nil {
		return nil, err
	}
	stats.EpisodesWatched = episodesCount
	// 3. Total movies duration
	type SumResult struct {
		Total int64
	}
	var movieDuration SumResult
	_, err = databaseEngine.Table(watchEventsTable).Alias("we").
		Join("INNER", "rewatches r", "r.rewatch_id = we.rewatch_id").
		Join("INNER", "media_records mr", "mr.record_id = we.record_id").
		Where("r.user_id = ?", userID).
		And("mr.record_type = ?", RecordTypeMovie).
		And("we.watched_at BETWEEN ? AND ?", startTime, finishTime).
		Select("SUM(mr.duration) as total").
		Get(&movieDuration)
	if err != nil {
		return nil, err
	}
	stats.TotalMoviesDuration = movieDuration.Total
	// 4. Total shows duration
	var showDuration SumResult
	_, err = databaseEngine.Table(watchEventsTable).Alias("we").
		Join("INNER", "rewatches r", "r.rewatch_id = we.rewatch_id").
		Join("INNER", "media_records mr", "mr.record_id = we.record_id").
		Where("r.user_id = ?", userID).
		And("mr.record_type = ?", RecordTypeEpisode).
		And("we.watched_at BETWEEN ? AND ?", startTime, finishTime).
		Select("SUM(mr.duration) as total").
		Get(&showDuration)
	if err != nil {
		return nil, err
	}
	stats.TotalEpisodesDuration = showDuration.Total

	return stats, nil
}
