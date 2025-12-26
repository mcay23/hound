package workers

import (
	"hound/database"
	"hound/model"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"time"
)

func InitializeWorkers(downloadWorkers int, ingestWorkers int) {
	InitializeDownloadWorkers(downloadWorkers)
	InitializeIngestWorkers(ingestWorkers)
	go cleanUpDownloads()
}

func cleanUpDownloads() {
	// clean up downloads if not in use
	// p2p case
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		entries, err := os.ReadDir(model.HoundP2PDownloadsPath)
		if err != nil {
			slog.Error("Failed to read p2p downloads directory", "error", err)
			return
		}
		// for p2p, directories are infohash names
		for _, folder := range entries {
			if folder.IsDir() {
				infoHash := folder.Name()
				// check if still being ingested
				task, err := database.GetIngestTask(database.IngestTask{SourceURI: model.GetMagnetURI(infoHash, nil)})
				if err != nil {
					slog.Error("Failed to get ingest task", "infohash", infoHash, "error", err)
					continue
				}
				terminalStates := []string{database.IngestStatusDone, database.IngestStatusFailed, database.IngestStatusCanceled}
				// if torrent session exists, never remove, wait for it to be cleaned up
				if !model.CheckTorrentSession(infoHash) {
					if task == nil {
						removeP2PFiles(infoHash)
					} else if slices.Contains(terminalStates, task.Status) {
						removeP2PFiles(infoHash)
					}
				}
			}
		}
	}
}

func removeP2PFiles(infoHash string) {
	slog.Info("Removing unused torrent files", "infohash", infoHash)
	targetDir := filepath.Join(model.HoundP2PDownloadsPath, infoHash)
	err := os.Chmod(targetDir, 0666)
	if err != nil {
		slog.Error("Failed to chmod dir", "dir", targetDir, "error", err)
	}
	err = os.RemoveAll(targetDir)
	if err != nil {
		slog.Error("Failed to remove dir", "dir", targetDir, "error", err)
	}
}
