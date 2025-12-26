package workers

import (
	"hound/database"
	"hound/model"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/anacrolix/torrent/metainfo"
)

// Only p2p downloads are supported for now
func InitializeDownloadWorkers(n int) {
	slog.Info("Starting download workers", "count", n)
	for i := range n {
		go downloadWorker(i)
	}
}

func downloadWorker(id int) {
	slog.Info("Download worker started", "workerID", id)
	for {
		task, err := database.GetNextPendingDownloadTask()
		if err != nil {
			slog.Error("Worker failed to get task", "workerID", id, "error", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if task == nil {
			time.Sleep(3 * time.Second)
			continue
		}
		processTask(id, task)
	}
}

func processTask(workerID int, task *database.IngestTask) {
	slog.Info("Worker picked up download task", "workerID", workerID,
		"taskID", task.IngestTaskID, "infoHash", task.SourceURI)
	uri, err := metainfo.ParseMagnetUri(*task.SourceURI)
	if err != nil {
		slog.Error("Failed to parse magnet URI", "taskID", task.IngestTaskID, "error", err)
		failTask(task, err)
		return
	}
	infoHash := uri.InfoHash.HexString()
	err = model.AddTorrent(infoHash, nil)
	if err != nil {
		slog.Error("Failed to add torrent", "taskID", task.IngestTaskID, "error", err)
		failTask(task, err)
		return
	}
	session, err := model.GetTorrentSession(infoHash)
	if err != nil {
		slog.Error("Failed to get torrent session", "taskID", task.IngestTaskID, "error", err)
		failTask(task, err)
		return
	}
	file, _, err := model.GetTorrentFile(infoHash, task.FileIdx, "", nil)
	if err != nil {
		slog.Error("Failed to get torrent file", "taskID", task.IngestTaskID, "error", err)
		failTask(task, err)
		return
	}
	relativePath := filepath.FromSlash(file.Path())
	task.SourcePath = filepath.Join(model.HoundP2PDownloadsPath, strings.ToLower(infoHash), relativePath)
	task.TotalBytes = file.Length()
	database.UpdateIngestTask(task)

	file.Download()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	lastBytesCompleted := int64(0)
	for range ticker.C {
		session.LastUsed = time.Now() // keep torrent session alive
		task.DownloadedBytes = file.BytesCompleted()
		task.DownloadSpeed = (file.BytesCompleted() - lastBytesCompleted) / 2
		lastBytesCompleted = file.BytesCompleted()
		task.LastSeen = time.Now()

		_, err := database.UpdateIngestTask(task)
		if err != nil {
			slog.Error("Failed to update task progress", "taskID", task.IngestTaskID, "error", err)
		}

		if file.BytesCompleted() >= file.Length() {
			slog.Info("Download finished", "workerID", workerID, "taskID", task.IngestTaskID)
			task.Status = database.IngestStatusPendingInsert
			task.FinishedAt = time.Now()
			// let ingest worker pick this up
			_, err := database.UpdateIngestTask(task)
			if err != nil {
				slog.Error("Failed to mark task done", "taskID", task.IngestTaskID, "error", err)
			}
			break
		}
	}
}

func failTask(task *database.IngestTask, err error) {
	slog.Error("Task failed", "taskID", task.IngestTaskID, "error", err)
	task.Status = database.IngestStatusFailed
	errorMessage := err.Error()
	task.LastMessage = &errorMessage
	task.FinishedAt = time.Now()
	database.UpdateIngestTask(task)
}
