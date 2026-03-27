package workers

import (
	"errors"
	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/helpers"
	"github.com/mcay23/hound/loggers"
	"github.com/mcay23/hound/model"
	"github.com/mcay23/hound/sources"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const externalQueueBuffer = 4096

type externalLibraryRoot struct {
	RootPath  string
	MediaType string
}

type externalLibraryQueueItem struct {
	Path          string
	RootPath      string
	MediaType     string
	IsInitialScan bool
}

var (
	externalLibraryQueue       chan externalLibraryQueueItem
	externalLibraryItemInQueue sync.Map
)

/*
Quick rundown:
Three producers:
1. initialExternalLibraryScan
2. periodicExternalLibraryRescan
3. watchExternalLibrary

Initial scan is run at every startup,
watch runs on file changes, renames

Producers create a queue of paths to process, if item is already
in queue, it's skipped

External Library Queue workers pick up from the queue and checks
if the file is already in the external_library_items table, and exits
early if it's processed and unchanged (size, modified). Next, it attempts
to match files to tmdb ids, season, episodes, if this is successful,
an ingest task is created and ingest workers pick up the task and
run ffprobe, etc. before inserting to media_files

Performance is mostly bottlenecked in tmdb network call, in the future
hardcoding popular shows/movies -> tmdb-id might be a fast trade-off, but
the slowest part is still upserting tv show records which have many
seasons and episodes, since each season is a separate network call

Right now, tmdb doesn't have rate-limits, but in the future, throttling
workers might be a good idea

Note that this will cause mismatches, especially for tv shows if the season/episode
ordering of the source directory is not identical to tmdb's ordering
*/
func InitializeExternalLibraryWorkers() {
	if !model.ExternalLibraryEnabled {
		return
	}
	roots := getExternalLibraryRoots()
	if len(roots) == 0 {
		slog.Warn("External library enabled but no valid root paths configured")
		return
	}
	externalLibraryQueue = make(chan externalLibraryQueueItem, externalQueueBuffer)
	consumerCount := model.MaxExternalLibraryWorkers
	for i := 0; i < consumerCount; i++ {
		go externalLibraryQueueWorker()
	}
	for _, root := range roots {
		go initialExternalLibraryScan(root)
		go periodicExternalLibraryRescan(root, model.ExternalScanInterval)
		go watchExternalLibrary(root)
		slog.Info("External library root started", "root", root.RootPath, "mediaType", root.MediaType)
	}
}

func getExternalLibraryRoots() []externalLibraryRoot {
	roots := make([]externalLibraryRoot, 0, 2)
	if strings.TrimSpace(model.ExternalLibraryMovies) != "" {
		roots = append(roots, externalLibraryRoot{
			RootPath:  filepath.Clean(model.ExternalLibraryMovies),
			MediaType: database.MediaTypeMovie,
		})
	}
	if strings.TrimSpace(model.ExternalLibraryTV) != "" {
		roots = append(roots, externalLibraryRoot{
			RootPath:  filepath.Clean(model.ExternalLibraryTV),
			MediaType: database.MediaTypeTVShow,
		})
	}
	valid := make([]externalLibraryRoot, 0, len(roots))
	for _, root := range roots {
		stat, err := os.Stat(root.RootPath)
		if err != nil || !stat.IsDir() {
			slog.Error("External library path is invalid", "path", root.RootPath, "mediaType", root.MediaType, "error", err)
			continue
		}
		valid = append(valid, root)
	}
	return valid
}

func initialExternalLibraryScan(root externalLibraryRoot) {
	slog.Info("Starting initial external library scan", "root", root.RootPath, "mediaType", root.MediaType)
	scanExternalLibrary(root, true)
}

func scanExternalLibrary(root externalLibraryRoot, isInitialScan bool) {
	err := filepath.WalkDir(root.RootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			slog.Warn("External library scan encountered error on path", "path", path, "error", err)
			return nil
		}
		if d.IsDir() {
			return nil
		}
		enqueueExternalPath(path, root, isInitialScan)
		return nil
	})
	if err != nil {
		slog.Error("External library scan failed (walkdir error)", "root", root.RootPath, "mediaType", root.MediaType, "isInitial", isInitialScan, "error", err)
	}
}

func periodicExternalLibraryRescan(root externalLibraryRoot, intervalMinutes int) {
	if intervalMinutes <= 0 {
		intervalMinutes = 360
	}
	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		slog.Info("Running periodic external library rescan", "root", root.RootPath, "mediaType", root.MediaType)
		scanExternalLibrary(root, false)
	}
}

func watchExternalLibrary(root externalLibraryRoot) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("Failed to initialize external library watcher", "error", err)
		return
	}
	defer watcher.Close()

	err = filepath.WalkDir(root.RootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			slog.Warn("External library watcher failed to access path", "path", path, "error", err)
			return nil
		}
		if d.IsDir() {
			_ = watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		slog.Error("Failed to register watcher directories", "error", err)
		return
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Rename == fsnotify.Rename {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					_ = watcher.Add(event.Name)
				}
				enqueueExternalPath(event.Name, root, false)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			slog.Error("External library watcher error", "error", err)
		}
	}
}

func enqueueExternalPath(path string, root externalLibraryRoot, isInitialScan bool) {
	cleanPath := filepath.Clean(path)
	if !model.IsVideoFile(cleanPath) {
		return
	}
	key := root.RootPath + "|" + root.MediaType + "|" + cleanPath
	if _, loaded := externalLibraryItemInQueue.LoadOrStore(key, struct{}{}); loaded {
		return
	}
	externalLibraryQueue <- externalLibraryQueueItem{
		Path:          cleanPath,
		RootPath:      root.RootPath,
		MediaType:     root.MediaType,
		IsInitialScan: isInitialScan,
	}
}

func externalLibraryQueueWorker() {
	for item := range externalLibraryQueue {
		func(item externalLibraryQueueItem) {
			key := item.RootPath + "|" + item.MediaType + "|" + item.Path
			defer externalLibraryItemInQueue.Delete(key)
			defer func() {
				if r := recover(); r != nil {
					slog.Error("External library worker panic recovered",
						"path", item.Path,
						"root", item.RootPath,
						"mediaType", item.MediaType,
						"panic", r)
				}
			}()
			processExternalPath(item)
		}(item)
	}
}

func processExternalPath(item externalLibraryQueueItem) {
	stat, err := os.Stat(item.Path)
	if err != nil || stat.IsDir() {
		return
	}
	// isFileCopying adds 1 second to every process, we disable it for initial
	// scans for faster processing
	// trade-off -> if files are being copied and then hound is started,
	// this may cause files being processed prematurely, which may be an issue
	if !item.IsInitialScan && isFileCopying(item.Path, stat) {
		return
	}
	dbItem, err := database.GetExternalLibraryItemByPath(item.Path)
	if err != nil {
		slog.Error("Failed to read external library item", "path", item.Path, "error", err)
		return
	}
	// already exists, unchanged in db, no need to re-evaluate
	if dbItem != nil &&
		dbItem.FileSize == stat.Size() &&
		dbItem.ModifiedUnix == stat.ModTime().Unix() &&
		(dbItem.Status == database.ExternalLibraryItemStatusDone || dbItem.Status == database.ExternalLibraryItemStatusQueued) {
		return
	}
	upsert := &database.ExternalLibraryItem{
		RootPath:     item.RootPath,
		SourcePath:   item.Path,
		FileSize:     stat.Size(),
		ModifiedUnix: stat.ModTime().Unix(),
		Status:       database.ExternalLibraryItemStatusPending,
	}
	if dbItem != nil {
		upsert.ItemID = dbItem.ItemID
	}
	loggers.IngestLogger().Info("[External Library: File detected]", "path", item.Path)
	err = database.UpsertExternalLibraryItem(upsert)
	if err != nil {
		slog.Error("Failed to upsert external library item", "error", err)
	}

	ingestTask, parsed, err := model.QueueExternalLibraryFile(item.RootPath, item.Path, item.MediaType)
	if err != nil {
		status := database.ExternalLibraryItemStatusFailed
		lastError := err.Error()
		if errors.Is(err, helpers.AlreadyExistsError) {
			status = database.ExternalLibraryItemStatusDone
			lastError = ""
		}
		upsert.Status = status
		if lastError != "" {
			upsert.LastError = &lastError
		} else {
			upsert.LastError = nil
		}
		err = database.UpsertExternalLibraryItem(upsert)
		if err != nil {
			slog.Error("Failed to upsert external library item", "error", err)
		}
		return
	}
	now := time.Now().UTC()
	upsert.MediaType = parsed.MediaType
	upsert.MediaSource = sources.MediaSourceTMDB
	upsert.SourceID = parsed.SourceID
	upsert.SeasonNumber = parsed.SeasonNumber
	upsert.EpisodeNumber = parsed.EpisodeNumber
	upsert.Status = database.ExternalLibraryItemStatusQueued
	upsert.LastError = nil
	upsert.LastIngestTaskID = &ingestTask.IngestTaskID
	upsert.LastQueuedAt = &now
	err = database.UpsertExternalLibraryItem(upsert)
	if err != nil {
		slog.Error("Failed to upsert external library item", "error", err)
	}
}

// Minimal guard against processing a file while it's still being copied.
// If size or mtime changes within a short window, defer processing.
func isFileCopying(path string, first os.FileInfo) bool {
	time.Sleep(1 * time.Second)
	second, err := os.Stat(path)
	if err != nil || second.IsDir() {
		return true
	}
	return first.Size() != second.Size() || first.ModTime().UnixNano() != second.ModTime().UnixNano()
}
