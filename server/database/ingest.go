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
	ProtocolP2P  = "p2p"
	ProtocolHTTP = "http"
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
	DownloadPriority int       `xorm:"'download_priority'" json:"download_priority"` // priority of task, not used for now
	CopyPriority     int       `xorm:"'copy_priority'" json:"copy_priority"`
	RecordID         int64     `xorm:"'record_id'" json:"record_id"`                    // episode or movie to be ingested
	Status           string    `xorm:"'status'" json:"status"`                          // pending_insert, processing, completed
	DownloadType     string    `xorm:"'download_type'" json:"download_type"`            // p2p, http, external (not downloaded by hound)
	SourceURI        *string   `xorm:"text 'source_uri'" json:"source_uri"`             // magnet uri with trackers / http link
	FileIdx          *int      `xorm:"'file_idx'" json:"file_idx"`                      // index for  only
	LastMessage      *string   `xorm:"text 'last_message'" json:"last_message"`         // store last error message
	SourcePath       string    `xorm:"text 'source_path'" json:"source_path"`           // path to source file/download path
	DestinationPath  string    `xorm:"text 'destination_path'" json:"destination_path"` // path to final destination in hound media dir
	TotalBytes       int64     `xorm:"'total_bytes'" json:"total_bytes"`                // total bytes to be downloaded
	DownloadedBytes  int64     `xorm:"'downloaded_bytes'" json:"downloaded_bytes"`
	CopiedBytes      int64     `xorm:"'copied_bytes'" json:"copied_bytes"`      // bytes copied to final dir
	DownloadSpeed    int64     `xorm:"'download_speed'" json:"download_speed"`  // bytes per second
	CopySpeed        int64     `xorm:"'copy_speed'" json:"copy_speed"`          // bytes per second
	LastSeen         time.Time `xorm:"timestampz last_seen" json:"last_seen"`   // track stale download/copy jobs
	StartedAt        time.Time `xorm:"timestampz started_at" json:"started_at"` // time queued task was started
	FinishedAt       time.Time `xorm:"timestampz finished_at" json:"finished_at"`
	CreatedAt        time.Time `xorm:"timestampz created" json:"created_at"`
	UpdatedAt        time.Time `xorm:"timestampz updated" json:"updated_at"`
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

func FindIngestTasksForStatus(status []string) ([]IngestTask, error) {
	// if no statuses given, return all tasks
	if len(status) == 0 {
		return FindIngestTasks(IngestTask{})
	}
	var tasks []IngestTask
	err := databaseEngine.Table(IngestTasksTable).In("status", status).Desc("created_at").Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err
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
		CopyPriority:     1,
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
