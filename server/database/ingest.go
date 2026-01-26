package database

import "time"

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
)

// tasks in terminal statuses won't change, retries must be made as a new task
var (
	IngestTerminalStatuses = []string{
		IngestStatusDone,
		IngestStatusFailed,
		IngestStatusCanceled,
	}
)

type IngestTask struct {
	IngestTaskID     int64     `xorm:"pk autoincr 'ingest_task_id'" json:"ingest_task_id"`
	DownloadPriority int       `xorm:"'download_priority'" json:"download_priority"`    // priority of task, not used for now
	RecordID         int64     `xorm:"index 'record_id'" json:"record_id"`              // episode or movie to be ingested
	Status           string    `xorm:"index 'status'" json:"status"`                    // pending_insert, processing, completed
	DownloadType     string    `xorm:"'download_type'" json:"download_type"`            // p2p, http, external (not downloaded by hound)
	SourceURI        *string   `xorm:"text 'source_uri'" json:"source_uri"`             // magnet uri with trackers / http link
	FileIdx          *int      `xorm:"'file_idx'" json:"file_idx"`                      // index for  only
	LastMessage      *string   `xorm:"text 'last_message'" json:"last_message"`         // store last error message
	SourcePath       string    `xorm:"text 'source_path'" json:"source_path"`           // path to source file/download path
	DestinationPath  string    `xorm:"text 'destination_path'" json:"destination_path"` // path to final destination in hound media dir
	TotalBytes       int64     `xorm:"'total_bytes'" json:"total_bytes"`                // total bytes to be downloaded
	DownloadedBytes  int64     `xorm:"'downloaded_bytes'" json:"downloaded_bytes"`
	DownloadSpeed    int64     `xorm:"'download_speed'" json:"download_speed"`       // bytes per second
	ConnectedSeeders *int      `xorm:"'connected_seeders'" json:"connected_seeders"` // number of seeders (p2p only)
	LastSeen         time.Time `xorm:"timestampz last_seen" json:"last_seen"`        // track stale download/copy jobs
	StartedAt        time.Time `xorm:"timestampz started_at" json:"started_at"`      // time queued task was started
	FinishedAt       time.Time `xorm:"timestampz finished_at" json:"finished_at"`
	CreatedAt        time.Time `xorm:"timestampz index created" json:"created_at"`
	UpdatedAt        time.Time `xorm:"timestampz updated" json:"updated_at"`
}

type IngestTaskFullRecord struct {
	IngestTask         `xorm:"extends"`
	MediaType          string       `json:"media_type"`
	MovieMediaRecord   *MediaRecord `json:"movie_media_record,omitempty"`
	EpisodeMediaRecord *MediaRecord `json:"episode_media_record,omitempty"`
	ShowMediaRecord    *MediaRecord `json:"show_media_record,omitempty"`
}

// ingest_jobs track ingestion of files from download -> inserted into hound
// external files are inserted at the pending_insert stage
func instantiateIngestTasksTable() error {
	databaseEngine.Table(IngestTasksTable).Sync2(new(IngestTask))
	return nil
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
			return 0, nil, err
		}
		var tasks []IngestTask
		sess := databaseEngine.Table(IngestTasksTable).Desc("created_at")
		if limit > 0 && offset >= 0 {
			sess = sess.Limit(limit, offset)
		}
		err = sess.Omit("full_data").Find(&tasks)
		if err != nil {
			return 0, nil, err
		}
		return int(ct), enrichIngestTasks(tasks), err
	}
	// status given, find tasks with status
	ct := databaseEngine.Table(IngestTasksTable).In("status", status)
	totalRecords, err := ct.Count()
	if err != nil {
		return 0, nil, err
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
		return 0, nil, err
	}
	return int(totalRecords), enrichIngestTasks(tasks), err
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

func InsertIngestTask(recordID int64, downloadType string, status string,
	sourceURI string, fileIdx *int) (bool, *IngestTask, error) {
	task := IngestTask{
		RecordID:         recordID,
		DownloadPriority: 1,
		DownloadType:     downloadType,
		Status:           status,
		SourceURI:        &sourceURI,
		FileIdx:          fileIdx,
	}
	_, err := databaseEngine.Table(IngestTasksTable).Insert(&task)
	return true, &task, err
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
	has, err := sess.SQL("SELECT * FROM "+IngestTasksTable+" WHERE status = ? ORDER BY ingest_task_id ASC LIMIT 1 FOR UPDATE",
		IngestStatusPendingDownload).Get(&task)
	if err != nil {
		sess.Rollback()
		return nil, err
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
		return nil, err
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
		return nil, err
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
		return nil, err
	}
	sess.Commit()
	return &task, nil
}
