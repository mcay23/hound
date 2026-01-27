package workers

import (
	"context"
	"errors"
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/model"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

type countingWriter struct {
	io.Writer
	count *int64
}

func (cw *countingWriter) Write(p []byte) (n int, err error) {
	n, err = cw.Writer.Write(p)
	atomic.AddInt64(cw.count, int64(n))
	return
}

// Only p2p downloads are supported for now
func InitializeDownloadWorkers(n int) {
	// check for invalid downloads and fail them (downloading when server is just starting)
	_, tasks, err := database.FindIngestTasksForStatus([]string{database.IngestStatusDownloading}, -1, 0)
	if err != nil {
		slog.Error("Failed to get pending download tasks", "error", err)
		return
	}
	for _, task := range tasks {
		failTask(&task.IngestTask, fmt.Errorf("invalid download task - process crashed during download"))
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
		"taskID", task.IngestTaskID, "sourceURI", *task.SourceURI)
	if strings.HasPrefix(*task.SourceURI, "http") {
		// http case
		startHTTPDownloadV2(workerID, task)
	} else {
		// p2p download case
		startP2PDownload(workerID, task)
	}
}

func startHTTPDownload(workerID int, task *database.IngestTask) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", *task.SourceURI, nil)
	if err != nil {
		slog.Error("Failed to create HTTP request", "taskID", task.IngestTaskID, "error", err)
		failTask(task, err)
		return
	}
	// mock browsers, some sites block requests without user agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	// req.Header.Set("Accept", "*/*")
	// req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	// req.Header.Set("Connection", "keep-alive")
	// req.Header.Set("Upgrade-Insecure-Requests", "1")
	// req.Header.Set("Sec-Fetch-Dest", "document")
	// req.Header.Set("Sec-Fetch-Mode", "navigate")
	// req.Header.Set("Sec-Fetch-Site", "none")
	// req.Header.Set("Sec-Fetch-User", "?1")

	client := &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2: false,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to get HTTP download", "taskID", task.IngestTaskID, "error", err)
		failTask(task, err)
		return
	}
	defer resp.Body.Close()
	slog.Info("HTTP info",
		"proto", resp.Proto,
		"contentLength", resp.ContentLength,
		"transferEncoding", resp.TransferEncoding,
	)
	if resp.StatusCode != http.StatusOK {
		slog.Info("fail headers",
			"status", resp.Status,
			"contentLength", resp.ContentLength,
			"transferEncoding", resp.TransferEncoding,
		)
		err = fmt.Errorf("bad status: %s", resp.Status)
		slog.Error("Failed to get HTTP download", "taskID", task.IngestTaskID, "error", err)
		failTask(task, err)
		return
	}
	// determine filename
	filename, err := getHTTPFilename(resp, *task.SourceURI)
	if err != nil {
		slog.Error("Failed to get HTTP filename", "taskID", task.IngestTaskID, "error", err)
		failTask(task, err)
		return
	}
	sourcePath := filepath.Join(model.HoundHttpDownloadsPath, filename)
	out, err := os.Create(sourcePath)
	if err != nil {
		slog.Error("Failed to create HTTP download file", "taskID", task.IngestTaskID, "error", err)
		failTask(task, err)
		return
	}
	defer out.Close()
	task.SourcePath = sourcePath
	task.TotalBytes = resp.ContentLength
	database.UpdateIngestTask(task)

	var downloaded int64
	pw := &countingWriter{Writer: out, count: &downloaded}
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	errChan := make(chan error, 1)
	go func() {
		_, err := io.Copy(pw, resp.Body)
		errChan <- err
	}()
	lastBytesCompleted := int64(0)
	for {
		select {
		case err := <-errChan:
			if err != nil {
				slog.Error("Failed to copy HTTP download", "taskID", task.IngestTaskID, "error", err)
				failTask(task, err)
				return
			}
			// download finished
			slog.Info("HTTP download finished", "workerID", workerID, "taskID", task.IngestTaskID)
			task.DownloadedBytes = atomic.LoadInt64(&downloaded)
			task.Status = database.IngestStatusPendingInsert
			task.FinishedAt = time.Now().UTC()
			_, err = database.UpdateIngestTask(task)
			if err != nil {
				slog.Error("Failed to mark task done", "taskID", task.IngestTaskID, "error", err)
			}
			return
		case <-ticker.C:
			// check if task is cancelled
			newTask, err := database.GetIngestTask(database.IngestTask{IngestTaskID: task.IngestTaskID})
			if err != nil {
				slog.Error("Failed to get ingest task", "taskID", task.IngestTaskID, "error", err)
				failTask(task, err)
				return
			}
			if newTask.Status == database.IngestStatusCanceled {
				cancel() // Stop the download
				cancelTask(newTask)
				return
			}
			currentDownloaded := atomic.LoadInt64(&downloaded)
			newTask.DownloadedBytes = currentDownloaded
			newTask.DownloadSpeed = (currentDownloaded - lastBytesCompleted) / 2
			lastBytesCompleted = currentDownloaded
			newTask.LastSeen = time.Now().UTC()
			_, err = database.UpdateIngestTask(newTask)
			if err != nil {
				slog.Error("Failed to update task progress", "taskID", newTask.IngestTaskID, "error", err)
			}
		}
	}
}

/*
Allows downloading with range requests, useful for CDNs which don't
allow one-shot downloads
*/
func startHTTPDownloadV2(workerID int, task *database.IngestTask) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2: false,
		},
	}

	// filename := path.Base(*task.SourceURI)
	// sourcePath := filepath.Join(model.HoundHttpDownloadsPath, filename)

	// out, err := os.OpenFile(sourcePath, os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	failTask(task, err)
	// 	return
	// }
	// defer out.Close()

	// var downloaded int64
	// stat, _ := out.Stat()
	// downloaded = stat.Size()
	// task.SourcePath = sourcePath
	// task.DownloadedBytes = downloaded
	// database.UpdateIngestTask(task)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	var downloaded int64 = 0
	var out *os.File
	lastBytesCompleted := downloaded
	for {
		req, err := http.NewRequestWithContext(ctx, "GET", *task.SourceURI, nil)
		if err != nil {
			failTask(task, err)
			return
		}
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", downloaded))
		resp, err := client.Do(req)
		if err != nil {
			failTask(task, err)
			return
		}
		// first request, open file
		if out == nil {
			filename, err := getHTTPFilename(resp, *task.SourceURI)
			if err != nil {
				failTask(task, fmt.Errorf("failed to get HTTP filename: %s", err))
				return
			}
			if !model.IsVideoFile(filepath.Ext(filename)) {
				failTask(task, fmt.Errorf("file is not a video file: %s", filename))
				return
			}
			sourcePath := filepath.Join(model.HoundHttpDownloadsPath, filename)
			out, err = os.OpenFile(sourcePath, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				failTask(task, fmt.Errorf("failed to open file: %s - %s", sourcePath, err))
				return
			}
			defer out.Close()
			stat, _ := out.Stat()
			downloaded = stat.Size()
			task.SourcePath = sourcePath
			task.TotalBytes = resp.ContentLength
			task.DownloadedBytes = downloaded
			database.UpdateIngestTask(task)
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
			resp.Body.Close()
			failTask(task, fmt.Errorf("bad status: %s", resp.Status))
			return
		}
		if _, err := out.Seek(downloaded, io.SeekStart); err != nil {
			resp.Body.Close()
			failTask(task, err)
			return
		}
		pw := &countingWriter{Writer: out, count: &downloaded}
		copyDone := make(chan error, 1)
		go func() {
			_, err := io.Copy(pw, resp.Body)
			copyDone <- err
		}()
		for {
			select {
			case err := <-copyDone:
				resp.Body.Close()
				// EOF, retry with next range
				if err == io.ErrUnexpectedEOF || err == io.EOF {
					if downloaded > lastBytesCompleted {
						lastBytesCompleted = downloaded
						goto retry
					}
					failTask(task, fmt.Errorf("stalled download at %d bytes:"+err.Error(), downloaded))
					return
				}
				if err != nil {
					failTask(task, err)
					return
				}
				// file complete
				task.DownloadedBytes = downloaded
				task.Status = database.IngestStatusPendingInsert
				task.FinishedAt = time.Now().UTC()
				database.UpdateIngestTask(task)

				slog.Info("HTTP download finished",
					"workerID", workerID,
					"taskID", task.IngestTaskID,
					"bytes", downloaded,
				)
				return

			case <-ticker.C:
				newTask, err := database.GetIngestTask(database.IngestTask{IngestTaskID: task.IngestTaskID})
				if err != nil {
					failTask(task, err)
					return
				}
				if newTask.Status == database.IngestStatusCanceled {
					cancel()
					cancelTask(newTask)
					// delete file
					slog.Info("Download cancelled, removing file", "taskID", task.IngestTaskID, "sourcePath", task.SourcePath)
					err = os.Remove(task.SourcePath)
					if err != nil {
						slog.Error("Failed to delete file", "taskID", task.IngestTaskID, "error", err)
					}
					return
				}
				current := atomic.LoadInt64(&downloaded)
				newTask.DownloadedBytes = current
				newTask.DownloadSpeed = (current - lastBytesCompleted) / 2
				lastBytesCompleted = current
				newTask.LastSeen = time.Now().UTC()
				database.UpdateIngestTask(newTask)
			}
		}
	retry:
	}
}

// get filename from http url
func getHTTPFilename(resp *http.Response, rawURL string) (string, error) {
	cd := resp.Header.Get("Content-Disposition")
	if cd != "" {
		_, params, err := mime.ParseMediaType(cd)
		if err == nil {
			if name, ok := params["filename"]; ok {
				if decoded, err := url.PathUnescape(name); err == nil {
					return decoded, nil
				}
			}
		}
	}
	// fallback to url after redirects
	if resp.Request != nil && resp.Request.URL != nil {
		if base := path.Base(resp.Request.URL.Path); base != "" && base != "/" {
			return base, nil
		}
	}
	// fallback #2
	u, err := url.Parse(rawURL)
	if err == nil && u.Path != "" {
		return path.Base(u.Path), nil
	}
	// at this point, no clue what the file extension is
	return "", errors.New(helpers.BadRequest)
}

func startP2PDownload(workerID int, task *database.IngestTask) {
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
	file, newIdx, _, err := model.GetTorrentFile(infoHash, task.FileIdx, nil)
	if err != nil {
		slog.Error("Failed to get torrent file", "taskID", task.IngestTaskID, "error", err)
		failTask(task, err)
		return
	}
	relativePath := filepath.FromSlash(file.Path())
	if task.FileIdx == nil {
		task.FileIdx = &newIdx
	}
	task.SourcePath = filepath.Join(model.HoundP2PDownloadsPath, strings.ToLower(infoHash), relativePath)
	task.TotalBytes = file.Length()
	database.UpdateIngestTask(task)

	file.Download()
	file.SetPriority(torrent.PiecePriorityNormal)
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

		var connectedSeeders int
		connectedSeeders = file.Torrent().Stats().ConnectedSeeders
		newTask.ConnectedSeeders = &connectedSeeders
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

	// protocol specific logic
	if task.DownloadType != database.ProtocolP2P {
		return
	}

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
	if session != nil && session.Torrent != nil && task.FileIdx != nil && *task.FileIdx < len(session.Torrent.Files()) {
		numStreams, ok := session.ActiveStreams[*task.FileIdx]
		// active streams key doesn't exist, or no active streams
		if !ok || numStreams <= 0 {
			// check if torrent is being downloaded by anyone else
			tasks, err := database.FindIngestTasks(database.IngestTask{SourceURI: task.SourceURI, Status: database.IngestStatusDownloading})
			if err != nil {
				slog.Error("Failed to find ingest tasks", "taskID", task.IngestTaskID, "error", err)
				return
			}
			if len(tasks) == 0 {
				slog.Info("Setting piece priority to none", "uri", *task.SourceURI, "fileIdx", *task.FileIdx)
				session.Torrent.Files()[*task.FileIdx].SetPriority(torrent.PiecePriorityNone)
			}
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
