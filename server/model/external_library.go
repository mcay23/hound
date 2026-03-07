package model

import (
	"errors"
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/loggers"
	"hound/sources"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	seasonEpisodePattern = regexp.MustCompile(`(?i)[Ss](\d{1,2})[Ee](\d{1,4})`)
	altEpisodePattern    = regexp.MustCompile(`(?i)(\d{1,2})[xX](\d{1,4})`)
	seasonFolderPattern  = regexp.MustCompile(`(?i)season[ ._-]*(\d{1,2})`)
	yearPattern          = regexp.MustCompile(`\(\s*((?:19|20)\d{2})\s*\)`)
	tmdbIDPattern        = regexp.MustCompile(`(?i)\[(?:tmdbid|tmdb)-(\d+)\]`)
	junkTokenPattern     = regexp.MustCompile(`(?i)\b(480p|720p|1080p|2160p|x264|x265|h264|h265|hevc|av1|vp9|bluray|brrip|webrip|web-dl|remux|dvdrip|hdr|aac|dts)\b`)
	showUpsertLocks      sync.Map
)

const (
	externalTMDBMatchCachePrefix = "external_match|tmdb"
	externalTMDBMatchCacheTTL    = 12 * time.Hour
	maxSearchChecks              = 5 // top n results to search for matching
)

type ParsedExternalMedia struct {
	MediaType     string
	Title         string
	SourceID      string
	Year          int
	SeasonNumber  *int
	EpisodeNumber *int
}

func QueueExternalLibraryFile(rootPath string, filePath string, mediaType string) (*database.IngestTask, *ParsedExternalMedia, error) {
	cleanRoot := filepath.Clean(rootPath)
	cleanPath := filepath.Clean(filePath)
	stat, err := os.Stat(cleanPath)
	if err != nil {
		return nil, nil, err
	}
	if stat.IsDir() || !IsVideoFile(cleanPath) {
		return nil, nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Not a video file")
	}
	parsed, err := parseExternalMediaPath(cleanRoot, cleanPath, mediaType)
	if err != nil {
		return nil, nil, err
	}
	var ingestRecordID int64
	switch parsed.MediaType {
	case database.MediaTypeMovie:
		sourceID := -1
		if parsed.SourceID != "" {
			sourceID, err = strconv.Atoi(parsed.SourceID)
			if err != nil || sourceID <= 0 {
				return nil, parsed, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
					"Invalid tmdb id in movie folder name")
			}
		} else {
			sourceID, err = findBestMovieTMDBID(parsed.Title, parsed.Year)
			if err != nil {
				return nil, parsed, err
			}
			parsed.SourceID = strconv.Itoa(sourceID)
		}
		has, record, err := database.GetMediaRecord(database.RecordTypeMovie, sources.MediaSourceTMDB, parsed.SourceID)
		if err == nil && has && record != nil {
			ingestRecordID = record.RecordID
		} else {
			record, err = sources.UpsertMediaRecordTMDB(database.MediaTypeMovie, sourceID)
			if err != nil {
				// Retry fetching just in case another worker succeeded in upserting concurrently
				has, retryRecord, retryErr := database.GetMediaRecord(database.RecordTypeMovie,
					sources.MediaSourceTMDB, parsed.SourceID)
				if retryErr == nil && has && retryRecord != nil {
					record = retryRecord
					ingestRecordID = record.RecordID
				} else {
					return nil, parsed, err
				}
			} else {
				ingestRecordID = record.RecordID
			}
		}
		loggers.IngestLogger().Info("[Matched Movie]", "path", filePath, "source_id", sourceID,
			"title", record.MediaTitle, "release_date", record.ReleaseDate)
	case database.MediaTypeTVShow:
		sourceID := -1
		if parsed.SourceID != "" {
			sourceID, err = strconv.Atoi(parsed.SourceID)
			if err != nil || sourceID <= 0 {
				return nil, parsed, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
					"Invalid tmdb id in tv show folder name")
			}
		} else {
			sourceID, err = findBestTVTMDBID(parsed.Title, parsed.Year)
			if err != nil {
				return nil, parsed, err
			}
			parsed.SourceID = strconv.Itoa(sourceID)
		}
		// if episode record already exists, we don't want to make an extra call
		// just to get show title, just debug from source id
		logTitle := "<see source id>"
		logReleaseDate := "<see source id>"
		// Fast path for repeat episodes - if already in DB, skip show upsert
		// however, with concurrent workers, good possibility multiple upserts are attempted.
		// use a lock to prevent multiple upserts for the same show.
		epRecord, err := database.GetEpisodeMediaRecord(sources.MediaSourceTMDB, parsed.SourceID, parsed.SeasonNumber, *parsed.EpisodeNumber)
		if err == nil && epRecord != nil {
			ingestRecordID = epRecord.RecordID
		} else {
			lock, _ := showUpsertLocks.LoadOrStore(sources.MediaSourceTMDB+"-"+parsed.SourceID, &sync.Mutex{})
			mu := lock.(*sync.Mutex)
			mu.Lock()

			epRecord, err = database.GetEpisodeMediaRecord(sources.MediaSourceTMDB, parsed.SourceID, parsed.SeasonNumber, *parsed.EpisodeNumber)
			if err == nil && epRecord != nil {
				ingestRecordID = epRecord.RecordID
				mu.Unlock()
			} else {
				record, err := sources.UpsertMediaRecordTMDB(database.MediaTypeTVShow, sourceID)
				if err != nil {
					loggers.IngestLogger().Info("[Match TV Show Failed]", "error", "Failed to upsert media record",
						"sourceID", sourceID, "season", parsed.SeasonNumber, "episode", parsed.EpisodeNumber)
					mu.Unlock()
					return nil, parsed, err
				}
				logTitle = record.MediaTitle
				logReleaseDate = record.ReleaseDate
				epRecord, err = database.GetEpisodeMediaRecord(record.MediaSource, record.SourceID,
					parsed.SeasonNumber, *parsed.EpisodeNumber)
				mu.Unlock()
				if err != nil || epRecord == nil {
					loggers.IngestLogger().Info("[Match TV Show Failed]", "error", "Failed to get episode record",
						"sourceID", record.SourceID, "season", parsed.SeasonNumber, "episode", parsed.EpisodeNumber)
					return nil, parsed, helpers.LogErrorWithMessage(err, "Failed to resolve episode record")
				}
				ingestRecordID = epRecord.RecordID
			}
		}
		loggers.IngestLogger().Info("[Matched TV Show]", "path", filePath, "source_id",
			sourceID, "title", logTitle, "release_date", logReleaseDate, "season",
			epRecord.SeasonNumber, "episode", epRecord.EpisodeNumber)
	default:
		return nil, parsed, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Unsupported media type")
	}
	tasks, err := database.FindIngestTasks(database.IngestTask{
		RecordID:         ingestRecordID,
		SourcePath:       cleanPath,
		DownloadProtocol: database.ProtocolExternal,
	})
	if err != nil {
		return nil, parsed, err
	}
	for _, task := range tasks {
		if !slices.Contains(database.IngestTerminalStatuses, task.Status) {
			return nil, parsed, helpers.LogErrorWithMessage(errors.New(helpers.AlreadyExists), "Ingest task already queued")
		}
	}

	sourceURI := "file://" + filepath.ToSlash(cleanPath)
	taskToInsert := &database.IngestTask{
		RecordID:         ingestRecordID,
		DownloadProtocol: database.ProtocolExternal,
		Status:           database.IngestStatusPendingInsert,
		SourceURI:        &sourceURI,
	}
	_, ingestTask, err := database.InsertIngestTask(taskToInsert)
	if err != nil {
		return nil, parsed, err
	}
	ingestTask.SourcePath = cleanPath
	ingestTask.TotalBytes = stat.Size()
	ingestTask.DownloadedBytes = stat.Size()
	ingestTask.LastSeen = time.Now().UTC()
	_, err = database.UpdateIngestTask(ingestTask)
	if err != nil {
		return nil, parsed, err
	}
	return ingestTask, parsed, nil
}

// finds tmdb id for title, year, media type
func parseExternalMediaPath(rootPath string, filePath string, mediaType string) (*ParsedExternalMedia, error) {
	rel, err := filepath.Rel(rootPath, filePath)
	if err != nil {
		return nil, err
	}
	rel = filepath.Clean(rel)
	if strings.HasPrefix(rel, "..") {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "File is outside external library root")
	}
	parts := splitPath(rel)
	filename := filepath.Base(filePath)
	parentDir := filepath.Base(filepath.Dir(filePath))

	switch mediaType {
	case database.MediaTypeMovie:
		sourceID := extractTMDBID(parentDir)
		title := cleanTitle(parentDir)
		if title == "" {
			title = cleanTitle(strings.TrimSuffix(filename, filepath.Ext(filename)))
		}
		year := extractYear(parentDir)
		if year == 0 {
			year = extractYear(filename)
		}
		return &ParsedExternalMedia{
			MediaType: database.MediaTypeMovie,
			Title:     title,
			SourceID:  sourceID,
			Year:      year,
		}, nil
	case database.MediaTypeTVShow:
		if len(parts) < 1 {
			return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Unsupported tv show path structure")
		}
		showDir := parts[0]
		sourceID := extractTMDBID(showDir)
		showTitle := cleanTitle(showDir)
		if showTitle == "" {
			showTitle = cleanTitle(parentDir)
		}
		season, episode := extractSeasonEpisode(filename)
		if season == nil || episode == nil {
			season = extractSeasonFromParts(parts)
		}
		if season == nil {
			return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "No season number found in path")
		}
		if episode == nil {
			return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "No episode number found in filename")
		}
		return &ParsedExternalMedia{
			MediaType:     database.MediaTypeTVShow,
			Title:         showTitle,
			SourceID:      sourceID,
			Year:          extractYear(showDir),
			SeasonNumber:  season,
			EpisodeNumber: episode,
		}, nil
	default:
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid media type hint")
	}
}

/*
There's quite a bit of reliance on tmdb results here,
typically tmdb has its own algorithm to find fuzzy matches,
even though the search query can be quite different than the result
since a show/movie may have many aliases

eg. tmdb successfully returns the result:
query: Sangatsu no Lion -> result: March comes in like a lion

This is a bit difficult to query in hound without a second network call
to check for alternative titles, so we rely on tmdb results for now.
if hound fails to match title, it still can succeed if year is the same
and tmdb deems it a top n result

This fails if a tv/movie title doesn't really have a good match in tmdb
and returns a junk result that happens to have the same year,
but it should be a relatively small edge case

Whether this approach causes too many false positives is something that
needs to be evaluated
*/
func findBestMovieTMDBID(title string, year int) (int, error) {
	cacheKey := getExternalTMDBMatchCacheKey(database.MediaTypeMovie, title, year)
	var cachedID int
	cacheExists, _ := database.GetCache(cacheKey, &cachedID)
	if cacheExists && cachedID > 0 {
		return cachedID, nil
	}
	results, err := sources.SearchMoviesTMDB(title)
	if err != nil {
		return -1, err
	}
	bestScore := -1
	bestID := -1
	target := normalizeTitle(title)
	for idx, candidate := range results.Results {
		if idx > maxSearchChecks {
			break
		}
		candidateTitle := normalizeTitle(candidate.Title)
		score := 0
		if candidateTitle == target {
			score += 5
		} else if strings.Contains(candidateTitle, target) || strings.Contains(target, candidateTitle) {
			score += 3
		}
		// for top results, also search original title if initial matching fails (?)
		// if score <= 0 && idx <= 3 {
		// 	tvDetails, err := sources.GetTVShowFromIDTMDB(int(candidate.ID))
		// 	if err != nil {
		// 		_ = helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError),
		// 			"findBestMovieTMDBID(): Error getting tmdb show"+err.Error())
		// 	}
		// 	if len(tvDetails.OriginCountry) > 0 {
		// 		originCountry := strings.ToLower(tvDetails.OriginCountry[0])
		// 		if tvDetails != nil {
		// 			for _, result := range tvDetails.AlternativeTitles.Results {
		// 				if strings.ToLower(result.Iso3166_1) == originCountry {
		// 					candidateTitle = normalizeTitle(result.Title)
		// 					if candidateTitle == target {
		// 						score += 5
		// 					}
		// 					if strings.Contains(candidateTitle, target) || strings.Contains(target, candidateTitle) {
		// 						score += 3
		// 					}
		// 				}
		// 			}
		// 		}
		// 	}
		// }
		// off by 1 years are accepted
		if year > 0 && len(candidate.ReleaseDate) >= 4 {
			cYear, _ := strconv.Atoi(candidate.ReleaseDate[:4])
			switch cYear {
			case year:
				score += 5
			case year - 1, year + 1:
				score += 1
			}
		}
		if score > bestScore {
			bestScore = score
			bestID = int(candidate.ID)
		}
	}
	_, _ = database.SetCache(cacheKey, bestID, externalTMDBMatchCacheTTL)
	return bestID, nil
}

// see findBestMovieTMDBID comments for explanation
func findBestTVTMDBID(title string, year int) (int, error) {
	cacheKey := getExternalTMDBMatchCacheKey(database.MediaTypeTVShow, title, year)
	var cachedID int
	cacheExists, _ := database.GetCache(cacheKey, &cachedID)
	if cacheExists && cachedID > 0 {
		return cachedID, nil
	}
	results, err := sources.SearchTVShowTMDB(title)
	if err != nil {
		return -1, err
	}
	bestScore := -1
	bestID := -1
	target := normalizeTitle(title)
	for idx, candidate := range results.Results {
		if idx > maxSearchChecks {
			break
		}
		candidateTitle := normalizeTitle(candidate.Name)
		score := 0
		if candidateTitle == target {
			score += 5
		} else if strings.Contains(candidateTitle, target) || strings.Contains(target, candidateTitle) {
			score += 3
		}
		if year > 0 && len(candidate.FirstAirDate) >= 4 {
			cYear, _ := strconv.Atoi(candidate.FirstAirDate[:4])
			switch cYear {
			case year:
				score += 5
			case year - 1, year + 1:
				score += 1
			}
		}
		if score > bestScore {
			bestScore = score
			bestID = int(candidate.ID)
		}
	}
	if bestID <= 0 {
		return -1, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), fmt.Sprintf("No TMDB match for tv show: %s", title))
	}
	_, _ = database.SetCache(cacheKey, bestID, externalTMDBMatchCacheTTL)
	return bestID, nil
}

func getExternalTMDBMatchCacheKey(mediaType string, title string, year int) string {
	return fmt.Sprintf("%s|%s|%s|year:%d", externalTMDBMatchCachePrefix, mediaType, normalizeTitle(title), year)
}

func splitPath(path string) []string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		filtered = append(filtered, part)
	}
	return filtered
}

func cleanTitle(raw string) string {
	title := strings.TrimSpace(raw)
	title = strings.TrimSuffix(title, filepath.Ext(title))
	title = tmdbIDPattern.ReplaceAllString(title, "")
	title = strings.ReplaceAll(title, ".", " ")
	title = strings.ReplaceAll(title, "_", " ")
	title = strings.ReplaceAll(title, "-", " ")
	title = yearPattern.ReplaceAllString(title, "")
	title = junkTokenPattern.ReplaceAllString(title, "")
	title = strings.Join(strings.Fields(title), " ")
	return strings.TrimSpace(title)
}

func normalizeTitle(title string) string {
	s := strings.ToLower(cleanTitle(title))
	return strings.Join(strings.Fields(s), " ")
}

func extractYear(input string) int {
	matches := yearPattern.FindStringSubmatch(input)
	if len(matches) < 2 {
		return 0
	}
	year, _ := strconv.Atoi(matches[1])
	return year
}

func extractSeasonEpisode(filename string) (*int, *int) {
	if matches := seasonEpisodePattern.FindStringSubmatch(filename); len(matches) == 3 {
		season, _ := strconv.Atoi(matches[1])
		episode, _ := strconv.Atoi(matches[2])
		return &season, &episode
	}
	if matches := altEpisodePattern.FindStringSubmatch(filename); len(matches) == 3 {
		season, _ := strconv.Atoi(matches[1])
		episode, _ := strconv.Atoi(matches[2])
		return &season, &episode
	}
	return nil, nil
}

func extractSeasonFromParts(parts []string) *int {
	for _, part := range parts {
		if matches := seasonFolderPattern.FindStringSubmatch(part); len(matches) == 2 {
			season, _ := strconv.Atoi(matches[1])
			return &season
		}
	}
	return nil
}

func extractTMDBID(input string) string {
	matches := tmdbIDPattern.FindStringSubmatch(input)
	if len(matches) != 2 {
		return ""
	}
	return matches[1]
}
