package workers

import (
	"fmt"
	"hound/database"
	"hound/model"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

// Only p2p downloads are supported for now
func InitializeDownloadWorkers(n int) {
	// check for invalid downloads and fail them (downloading when server is just starting)
	tasks, err := database.FindIngestTasksForStatus([]string{database.IngestStatusDownloading})
	if err != nil {
		slog.Error("Failed to get pending download tasks", "error", err)
		return
	}
	for _, task := range tasks {
		failTask(&task, fmt.Errorf("invalid download task - process crashed during download"))
	}

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
	file, _, err := model.GetTorrentFile(infoHash, task.FileIdx, nil)
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
		// check if task is cancelled
		newTask, err := database.GetIngestTask(database.IngestTask{IngestTaskID: task.IngestTaskID})
		if err != nil {
			slog.Error("Failed to get ingest task", "taskID", task.IngestTaskID, "error", err)
			failTask(task, err)
			return
		}
		if newTask.Status == database.IngestStatusCanceled {
			cancelTask(newTask)
			return
		}
		// update torrent session
		session.LastUsed = time.Now().Unix()
		session, err = model.GetTorrentSession(infoHash)
		if err != nil {
			slog.Error("Failed to get torrent session", "taskID", newTask.IngestTaskID, "error", err)
			failTask(newTask, err)
			return
		}

		newTask.DownloadedBytes = file.BytesCompleted()
		newTask.DownloadSpeed = (file.BytesCompleted() - lastBytesCompleted) / 2
		lastBytesCompleted = file.BytesCompleted()
		newTask.LastSeen = time.Now().UTC()

		_, err = database.UpdateIngestTask(newTask)
		if err != nil {
			slog.Error("Failed to update task progress", "taskID", newTask.IngestTaskID, "error", err)
		}

		if file.BytesCompleted() >= file.Length() {
			slog.Info("Download finished", "workerID", workerID, "taskID", newTask.IngestTaskID)
			newTask.Status = database.IngestStatusPendingInsert
			newTask.FinishedAt = time.Now().UTC()
			// let ingest worker pick this up
			_, err := database.UpdateIngestTask(newTask)
			if err != nil {
				slog.Error("Failed to mark task done", "taskID", newTask.IngestTaskID, "error", err)
			}
			break
		}
	}
}

func cancelTask(task *database.IngestTask) {
	cancelMsg := "Task cancelled by the user"
	task.LastMessage = &cancelMsg
	task.FinishedAt = time.Now().UTC()
	_, err := database.UpdateIngestTask(task)
	if err != nil {
		slog.Error("Failed to cancel task", "taskID", task.IngestTaskID, "error", err)
	}
	slog.Info("Task cancelled by user", "taskID", task.IngestTaskID, "uri", *task.SourceURI)

	// check Torrent Session
	uri, err := metainfo.ParseMagnetUri(*task.SourceURI)
	if err != nil {
		slog.Error("Failed to parse magnet URI", "taskID", task.IngestTaskID, "error", err)
		return
	}
	session, err := model.GetTorrentSession(uri.InfoHash.HexString())
	if err != nil {
		slog.Error("Failed to get torrent session", "taskID", task.IngestTaskID, "error", err)
		return
	}

	// if no one is using, set piece priority to none
	// evaluate: this may not be required, since if the client requests the
	// stream, the piece should be newly requested again
	if session != nil && *task.FileIdx < len(session.Torrent.Files()) {
		numStreams, ok := session.ActiveStreams[*task.FileIdx]
		// active streams key doesn't exist, or no active streams
		if !ok || numStreams <= 0 {
			slog.Info("Setting piece priority to none", "uri", *task.SourceURI, "fileIdx", *task.FileIdx)
			session.Torrent.Files()[*task.FileIdx].SetPriority(torrent.PiecePriorityNone)
		}
	}
}

func failTask(task *database.IngestTask, err error) {
	slog.Error("Task failed", "taskID", task.IngestTaskID, "error", err)
	task.Status = database.IngestStatusFailed
	errorMessage := err.Error()
	task.LastMessage = &errorMessage
	task.FinishedAt = time.Now().UTC()
	database.UpdateIngestTask(task)
}
