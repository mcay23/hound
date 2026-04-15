package database

import (
	"fmt"
	"time"
)

const (
	IngestTasksTable = "ingest_tasks"
)

// Status values. Ingestions roughly follow this order
const (
	IngestStatusPendingDownload  = "pending_download"
	IngestStatusMetadataFetching = "metadata_fetching" // for torrents
	IngestStatusDownloading      = "downloading"
	IngestStatusPendingInsert    = "pending_insert" // downloading finished, for external ingestion, status starts here
	IngestStatusCopying          = "copying"
	IngestStatusDone             = "done"
	// failure states. these are considered terminal states
	IngestStatusFailed   = "failed"
	IngestStatusCanceled = "canceled"
)

const (
	ProtocolP2P       = "p2p"
	ProtocolProxyHTTP = "proxy-http"
	ProtocolFileHTTP  = "file-http"
	ProtocolExternal  = "external"
)

const (
	MatchTypeInfoHash = "match_info_hash"
	MatchTypeString   = "match_string"
)

// tasks in terminal statuses won't change, retries must be made as a new task
var (
	IngestTerminalStatuses = []string{
		IngestStatusDone,
		IngestStatusFailed,
		IngestStatusCanceled,
	}
	IngestActiveStatuses = []string{
		IngestStatusPendingDownload,
		IngestStatusMetadataFetching,
		IngestStatusDownloading,
		IngestStatusPendingInsert,
		IngestStatusCopying,
	}
)

type IngestTask struct {
	IngestTaskID        int64                      `xorm:"pk autoincr 'ingest_task_id'" json:"ingest_task_id"`
	DownloadPriority    int                        `xorm:"'download_priority'" json:"download_priority"`             // priority of task, not used for now
	RecordID            int64                      `xorm:"index 'record_id'" json:"record_id"`                       // episode or movie to be ingested
	Status              string                     `xorm:"index 'status'" json:"status"`                             // pending_insert, processing, completed
	DownloadProtocol    string                     `xorm:"'download_protocol'" json:"download_protocol"`             // p2p, http, external (not downloaded by hound)
	SourceURI           *string                    `xorm:"text 'source_uri'" json:"source_uri"`                      // magnet uri with trackers / http link
	FileIdx             *int                       `xorm:"'file_idx'" json:"file_idx"`                               // index for p2p only
	DownloadPreferences *IngestDownloadPreferences `xorm:"jsonb 'download_preferences'" json:"download_preferences"` // for season/auto downloads
	LastMessage         *string                    `xorm:"text 'last_message'" json:"last_message"`                  // store last error message
	SourcePath          string                     `xorm:"text 'source_path'" json:"source_path"`                    // path to source file/download path
	DestinationPath     string                     `xorm:"text 'destination_path'" json:"destination_path"`          // path to final destination in hound media dir
	TotalBytes          int64                      `xorm:"'total_bytes'" json:"total_bytes"`                         // total bytes to be downloaded
	DownloadedBytes     int64                      `xorm:"'downloaded_bytes'" json:"downloaded_bytes"`
	DownloadSpeed       int64                      `xorm:"'download_speed'" json:"download_speed"`       // bytes per second
	ConnectedSeeders    *int                       `xorm:"'connected_seeders'" json:"connected_seeders"` // number of seeders (p2p only)
	LastSeen            time.Time                  `xorm:"timestampz last_seen" json:"last_seen"`        // track stale download/copy jobs
	StartedAt           time.Time                  `xorm:"timestampz started_at" json:"started_at"`      // time queued task was started
	FinishedAt          time.Time                  `xorm:"timestampz finished_at" json:"finished_at"`
	CreatedAt           time.Time                  `xorm:"timestampz index created" json:"created_at"`
	UpdatedAt           time.Time                  `xorm:"timestampz updated" json:"updated_at"`
}

type IngestTaskFullRecord struct {
	IngestTask         `xorm:"extends"`
	MediaType          string       `json:"media_type"`
	MovieMediaRecord   *MediaRecord `json:"movie_media_record,omitempty"`
	EpisodeMediaRecord *MediaRecord `json:"episode_media_record,omitempty"`
	ShowMediaRecord    *MediaRecord `json:"show_media_record,omitempty"`
}

/*
This defines download preferences when the user downloads downloads automatically
eg. Downloading a whole season. The download worker search current providers, and gets
the best match based on the preferences.

Flow:
1. Providers return a list of sources
2. The first preference from the PreferenceList is evaluated
3. The first source from the providers response is evaluated
4. If a match is found, this becomes the task's sourceURI
5. Else, repeat the process with the second, third, ... preference until
a match is found
6. If no match is found, choose the first result or fail the task depending
on StrictMatch

Note that AIOStreams is recommended for basically all users, and this has even
more robust sort/filter systems, it's recommended to use that to set quality,
resolution, bitrate, language preferences.

This is mostly to help downloading seasons, so we can more consistently download
episodes from the same torrent based on string matching or infohash matching
*/
type IngestDownloadPreferences struct {
	StrictMatch       bool                 `json:"strict_match"` // whether if no match is found, to fail the download
	PreferenceList    []DownloadPreference `json:"preference_list"`
	ProviderProfileID int64                `json:"provider_profile_id"`
}

type DownloadPreference struct {
	MatchType             string                      `json:"match_type"`
	InfoHashPreference    *DownloadPreferenceInfoHash `json:"info_hash_preference,omitempty"`
	StringMatchPreference *DownloadPreferenceString   `json:"string_match_preference,omitempty"`
}

type DownloadPreferenceInfoHash struct {
	InfoHash string `json:"info_hash"`
}

type DownloadPreferenceString struct {
	MatchString   string `json:"match_string"`
	CaseSensitive bool   `json:"case_sensitive"`
}

// ingest_jobs track ingestion of files from download -> inserted into hound
// external files are inserted at the pending_insert stage
func instantiateIngestTasksTable() error {
	return databaseEngine.Table(IngestTasksTable).Sync2(new(IngestTask))
}

func FindIngestTasks(task IngestTask) ([]IngestTask, error) {
	var tasks []IngestTask
	err := databaseEngine.Table(IngestTasksTable).Desc("created_at").Find(&tasks, &task)
	if err != nil {
		return nil, err
	}
	return tasks, err
}

func FindIngestTasksForStatus(status []string, limit int, offset int) (int, []IngestTaskFullRecord, error) {
	// if no statuses given, return all tasks
	if len(status) == 0 {
		ct, err := databaseEngine.Table(IngestTasksTable).Count()
		if err != nil {
			return 0, nil, fmt.Errorf("count %s: %w", IngestTasksTable, err)
		}
		var tasks []IngestTask
		sess := databaseEngine.Table(IngestTasksTable).Desc("created_at")
		if limit > 0 && offset >= 0 {
			sess = sess.Limit(limit, offset)
		}
		err = sess.Omit("full_data").Find(&tasks)
		if err != nil {
			return 0, nil, fmt.Errorf("query %s: %w", IngestTasksTable, err)
		}
		return int(ct), enrichIngestTasks(tasks), nil
	}
	// status given, find tasks with status
	ct := databaseEngine.Table(IngestTasksTable).In("status", status)
	totalRecords, err := ct.Count()
	if err != nil {
		return 0, nil, fmt.Errorf("count %s: %w", IngestTasksTable, err)
	}
	var tasks []IngestTask
	sess := databaseEngine.Table(IngestTasksTable).
		In("status", status).
		Desc("created_at")
	if limit > 0 && offset >= 0 {
		sess = sess.Limit(limit, offset)
	}
	err = sess.Omit("full_data").Find(&tasks)
	if err != nil {
		return 0, nil, fmt.Errorf("query %s: %w", IngestTasksTable, err)
	}
	return int(totalRecords), enrichIngestTasks(tasks), nil
}

// this gets the movie record, or both the episode and show record for tv shows
// a bit computationally expensive, might need a better solution
func enrichIngestTasks(tasks []IngestTask) []IngestTaskFullRecord {
	const reducedFields = "record_id, record_type, media_source, source_id, parent_id, ancestor_id, media_title, original_title, original_language, origin_country, release_date, last_air_date, next_air_date, season_number, episode_number, sort_index, status, overview, duration, thumbnail_uri, backdrop_uri, logo_uri, genres, tags, created_at, updated_at"
	if len(tasks) == 0 {
		return []IngestTaskFullRecord{}
	}
	// collect record ids
	recordIDs := make([]int64, len(tasks))
	for i, t := range tasks {
		recordIDs[i] = t.RecordID
	}
	// fetch records
	var allRecords []MediaRecord
	databaseEngine.Table(mediaRecordsTable).Select(reducedFields).In("record_id", recordIDs).Find(&allRecords)
	// map records by id and collect ancestor (show) ids for episodes
	recordMap := make(map[int64]MediaRecord)
	showIDs := make([]int64, 0)
	for _, r := range allRecords {
		recordMap[r.RecordID] = r
		if r.RecordType == RecordTypeEpisode && r.AncestorID != nil {
			showIDs = append(showIDs, *r.AncestorID)
		}
	}
	// fetch show records (for episodes)
	var showRecords []MediaRecord
	if len(showIDs) > 0 {
		databaseEngine.Table(mediaRecordsTable).In("record_id", showIDs).Select(reducedFields).Find(&showRecords)
	}
	showMap := make(map[int64]MediaRecord)
	for _, s := range showRecords {
		showMap[s.RecordID] = s
	}

	enriched := make([]IngestTaskFullRecord, len(tasks))
	for i, t := range tasks {
		er := IngestTaskFullRecord{IngestTask: t}
		if r, ok := recordMap[t.RecordID]; ok {
			switch r.RecordType {
			case RecordTypeMovie:
				er.MovieMediaRecord = &r
				er.MediaType = MediaTypeMovie
			case RecordTypeEpisode:
				er.EpisodeMediaRecord = &r
				er.MediaType = MediaTypeTVShow
				// check for show record (ancestor)
				if r.AncestorID != nil {
					if show, ok := showMap[*r.AncestorID]; ok {
						er.ShowMediaRecord = &show
					}
				}
			}
		}
		enriched[i] = er
	}
	return enriched
}

func GetIngestTask(task IngestTask) (*IngestTask, error) {
	has, err := databaseEngine.Table(IngestTasksTable).Get(&task)
	if !has {
		return nil, nil
	}
	return &task, err
}

func InsertIngestTask(task *IngestTask) (bool, *IngestTask, error) {
	if task.DownloadPriority == 0 {
		task.DownloadPriority = 1
	}
	_, err := databaseEngine.Table(IngestTasksTable).Insert(task)
	return true, task, err
}

func UpdateIngestTask(task *IngestTask) (int, error) {
	affected, err := databaseEngine.Table(IngestTasksTable).
		Where("ingest_task_id = ?", task.IngestTaskID).
		Update(task)
	return int(affected), err
}

func UpdateStatus(ingestTaskID int64, status string) (bool, error) {
	_, err := databaseEngine.Table(IngestTasksTable).
		Where("ingest_task_id = ?", ingestTaskID).
		Update(IngestTask{Status: status})
	return true, err
}

// GetNextPendingDownloadTask atomically gets the next pending download task for workers and marks as downloading
// use ForUpdate() lock to prevent multiple workers from picking up task
func GetNextPendingDownloadTask() (*IngestTask, error) {
	var task IngestTask
	sess := databaseEngine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, err
	}
	// postgres for update to prevent race conditions
	has, err := sess.SQL("SELECT * FROM "+IngestTasksTable+" WHERE status = ? ORDER BY ingest_task_id ASC LIMIT 1 FOR UPDATE",
		IngestStatusPendingDownload).Get(&task)
	if err != nil {
		sess.Rollback()
		return nil, fmt.Errorf("query %s for status %s: %w", IngestTasksTable, IngestStatusPendingDownload, err)
	}
	if !has {
		sess.Rollback()
		return nil, nil
	}
	task.Status = IngestStatusDownloading
	task.StartedAt = time.Now().UTC()
	task.LastSeen = time.Now().UTC()
	if _, err := sess.Table(IngestTasksTable).ID(task.IngestTaskID).
		Cols("status", "started_at", "last_seen").Update(&task); err != nil {
		sess.Rollback()
		return nil, fmt.Errorf("update %s for ingest_task_id %d: %w", IngestTasksTable, task.IngestTaskID, err)
	}
	sess.Commit()
	return &task, nil
}

// GetNextPendingIngestTask atomically gets the next pending ingest task for workers and marks as copying
func GetNextPendingIngestTask() (*IngestTask, error) {
	var task IngestTask
	sess := databaseEngine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, err
	}
	has, err := sess.SQL("SELECT * FROM "+IngestTasksTable+" WHERE status = ? ORDER BY ingest_task_id ASC LIMIT 1 FOR UPDATE",
		IngestStatusPendingInsert).Get(&task)
	if err != nil {
		sess.Rollback()
		return nil, fmt.Errorf("query %s for status %s: %w", IngestTasksTable, IngestStatusPendingInsert, err)
	}
	if !has {
		sess.Rollback()
		return nil, nil
	}
	task.Status = IngestStatusCopying
	task.StartedAt = time.Now().UTC()
	task.LastSeen = time.Now().UTC()
	if _, err := sess.Table(IngestTasksTable).ID(task.IngestTaskID).
		Cols("status", "started_at", "last_seen").Update(&task); err != nil {
		sess.Rollback()
		return nil, fmt.Errorf("update %s for ingest_task_id %d: %w", IngestTasksTable, task.IngestTaskID, err)
	}
	sess.Commit()
	return &task, nil
}
