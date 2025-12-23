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
	DownloadTypeP2P      = "p2p"
	DownloadTypeHTTP     = "http"
	DownloadTypeExternal = "external"
)

type IngestTask struct {
	IngestTaskID    int64     `xorm:"pk autoincr 'ingest_task_id'" json:"ingest_task_id"`
	RecordID        int64     `xorm:"'record_id'" json:"record_id"`               // episode or movie to be ingested
	Status          string    `xorm:"'status'" json:"status"`                     // pending_insert, processing, completed
	DownloadType    string    `xorm:"'download_type'" json:"download_type"`       // p2p, http, external (not downloaded by hound)
	SourceURI       string    `xorm:"'source_uri'" json:"source_uri"`             // magnet uri with trackers / http link
	LastMessage     string    `xorm:"'last_message'" json:"last_message"`         // store last error message
	SourcePath      string    `xorm:"'source_path'" json:"source_path"`           // path to source file/download path
	DestinationPath string    `xorm:"'destination_path'" json:"destination_path"` // path to final destination in hound media dir
	TotalBytes      int64     `xorm:"'total_bytes'" json:"total_bytes"`           // total bytes to be downloaded
	DownloadedBytes int64     `xorm:"'downloaded_bytes'" json:"downloaded_bytes"`
	CopiedBytes     int64     `xorm:"'copied_bytes'" json:"copied_bytes"`      // bytes copied to final dir
	DownloadSpeed   int64     `xorm:"'download_speed'" json:"download_speed"`  // bytes per second
	CopySpeed       int64     `xorm:"'copy_speed'" json:"copy_speed"`          // bytes per second
	LastSeen        time.Time `xorm:"timestampz last_seen" json:"last_seen"`   // track stale download/copy jobs
	StartedAt       time.Time `xorm:"timestampz started_at" json:"started_at"` // time queued task was started
	FinishedAt      time.Time `xorm:"timestampz finished_at" json:"finished_at"`
	CreatedAt       time.Time `xorm:"timestampz created" json:"created_at"`
	UpdatedAt       time.Time `xorm:"timestampz updated" json:"updated_at"`
}

// ingest_jobs track ingestion of files from download -> inserted into hound
// external files are inserted at the pending_insert stage
func instantiateIngestTasksTable() error {
	databaseEngine.Sync2(new(IngestTask))
	return nil
}

func GetIngestTasks(mediaSource string, sourceID string) ([]*IngestTask, error) {
	var tasks []*IngestTask
	err := databaseEngine.Table(IngestTasksTable).
		Join("INNER", "media_records", "media_records.record_id = ingest_tasks.record_id").
		Where("media_records.media_source = ? AND media_records.source_id = ?", mediaSource, sourceID).Find(&tasks)
	return tasks, err
}

func InsertIngestTask(recordID int64, downloadType string, status string, uri string) (bool, *IngestTask, error) {
	task := IngestTask{
		RecordID:     recordID,
		DownloadType: downloadType,
		Status:       status,
		SourceURI:    uri,
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
