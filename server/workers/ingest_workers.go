package workers

import (
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/model"
	"log/slog"
	"time"

	"github.com/anacrolix/torrent/metainfo"
)

func InitializeIngestWorkers(n int) {
	slog.Info("Starting ingest workers", "count", n)
	for i := range n {
		go ingestWorker(i)
	}
}

func ingestWorker(id int) {
	slog.Info("Ingest worker started", "workerID", id)
	for {
		task, err := database.GetNextPendingIngestTask()
		if err != nil {
			slog.Error("Ingest worker failed to get task", "workerID", id, "error", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if task == nil {
			time.Sleep(3 * time.Second)
			continue
		}
		processIngestTask(id, task)
	}
}

func processIngestTask(workerID int, task *database.IngestTask) {
	slog.Info("Worker picked up ingest task", "workerID", workerID, "taskID", task.IngestTaskID)
	var infoHash *string
	// p2p case, for external ingests we don't know the source
	if task.DownloadType == database.DownloadTypeP2P && task.SourceURI != nil {
		uri, err := metainfo.ParseMagnetUri(*task.SourceURI)
		if err == nil {
			hash := uri.InfoHash.HexString()
			infoHash = &hash
		}
	}
	// Fetch mediaRecord
	mediaRecord, err := database.GetMediaRecordByID(task.RecordID)
	if err != nil || mediaRecord == nil {
		err = helpers.LogErrorWithMessage(err, "failed to get media record or not found")
		failTask(task, err)
		return
	}
	var ingestRecord *database.MediaRecord
	var seasonNum, episodeNum *int
	switch mediaRecord.RecordType {
	case database.RecordTypeEpisode:
		// traverse up to show, we need this to construct destination path
		if mediaRecord.ParentID == nil {
			failTask(task, fmt.Errorf("episode record has no parent"))
			return
		}
		seasonRecord, err := database.GetMediaRecordByID(*mediaRecord.ParentID)
		if err != nil || seasonRecord == nil {
			failTask(task, fmt.Errorf("failed to get season record"))
			return
		}
		if seasonRecord.ParentID == nil {
			failTask(task, fmt.Errorf("season record has no parent"))
			return
		}
		showRecord, err := database.GetMediaRecordByID(*seasonRecord.ParentID)
		if err != nil || showRecord == nil {
			failTask(task, fmt.Errorf("failed to get show record"))
			return
		}
		ingestRecord = showRecord
		seasonNum = mediaRecord.SeasonNumber
		episodeNum = mediaRecord.EpisodeNumber
	case database.RecordTypeMovie:
		ingestRecord = mediaRecord
	default:
		failTask(task, fmt.Errorf("unsupported record type for ingestion: %s", mediaRecord.RecordType))
		return
	}
	mediaFile, err := model.IngestFile(ingestRecord, seasonNum, episodeNum, infoHash, task.FileIdx, task.SourcePath)
	if err != nil {
		slog.Error("Ingestion failed", "taskID", task.IngestTaskID, "error", err)
		failTask(task, err)
		return
	}
	// set ingest to done
	task.Status = database.IngestStatusDone
	task.DestinationPath = mediaFile.Filepath
	task.FinishedAt = time.Now()
	_, err = database.UpdateIngestTask(task)
	if err != nil {
		slog.Error("Failed to update ingest task status", "taskID", task.IngestTaskID, "error", err)
	}
	slog.Info("Ingest task completed", "taskID", task.IngestTaskID)
}
