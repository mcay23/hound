package workers

import (
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/loggers"
	"hound/model"
	"log/slog"
	"time"

	"github.com/anacrolix/torrent/metainfo"
)

func InitializeIngestWorkers() {
	slog.Debug("Starting ingest workers", "count", model.MaxConcurrentIngests)
	for i := range model.MaxConcurrentIngests {
		go ingestWorker(i)
	}
}

func ingestWorker(id int) {
	slog.Debug("Ingest worker started", "workerID", id)
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
	loggers.IngestLogger().Info("Worker picked up ingest task", "workerID", workerID, "taskID", task.IngestTaskID)
	var infoHashStr string
	var infoHash *string
	// p2p case, for external ingests we don't know the source
	if task.DownloadProtocol == database.ProtocolP2P && task.SourceURI != nil {
		uri, err := metainfo.ParseMagnetUri(*task.SourceURI)
		if err == nil {
			infoHashStr = uri.InfoHash.HexString()
			infoHash = &infoHashStr
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
	var mediaFile *database.MediaFile
	if task.DownloadProtocol == database.ProtocolExternal {
		mediaFile, err = model.IngestFile(ingestRecord, seasonNum, episodeNum, infoHash, task.FileIdx, task.SourceURI, task.SourcePath, model.IngestTransferPreserve, database.FileOriginExternalLibrary)
	} else {
		mediaFile, err = model.IngestFile(ingestRecord, seasonNum, episodeNum, infoHash, task.FileIdx, task.SourceURI, task.SourcePath, model.IngestTransferMove, database.FileOriginHoundManaged)
	}
	if err != nil {
		slog.Error("Ingestion failed", "taskID", task.IngestTaskID, "error", err)
		failTask(task, err)
		return
	}
	// set ingest to done
	task.Status = database.IngestStatusDone
	task.DestinationPath = mediaFile.Filepath
	task.FinishedAt = time.Now().UTC()
	_, err = database.UpdateIngestTask(task)
	if err != nil {
		slog.Error("Failed to update ingest task status", "taskID", task.IngestTaskID, "error", err)
	}
	if task.DownloadProtocol == database.ProtocolExternal {
		item, getErr := database.GetExternalLibraryItemByPath(task.SourcePath)
		if getErr == nil && item != nil {
			now := time.Now().UTC()
			item.Status = database.ExternalLibraryItemStatusDone
			item.LastError = nil
			item.LastCompletedAt = &now
			item.LastIngestTaskID = &task.IngestTaskID
			err = database.UpsertExternalLibraryItem(item)
			if err != nil {
				helpers.LogErrorWithMessage(err, "Failed to upsert external library item")
			}
		}
	}
	slog.Info("Ingest task completed", "taskID", task.IngestTaskID)
}
