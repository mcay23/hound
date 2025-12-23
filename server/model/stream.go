package model

import (
	"errors"
	"fmt"
	"hound/helpers"
	"log/slog"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/log"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

/*
	Handle P2P Streams, credit to https://github.com/aculix/bitplay
*/

var videoExtensions = map[string]bool{
	".mp4":  true,
	".mkv":  true,
	".avi":  true,
	".mov":  true,
	".wmv":  true,
	".flv":  true,
	".webm": true,
	".mpeg": true,
	".mpg":  true,
	".ts":   true,
	".vob":  true,
	".3gp":  true,
}

const (
	TorrentDownloadsDir = "torrent-data"
)

type TorrentSession struct {
	Torrent  *torrent.Torrent
	LastUsed time.Time
}

var (
	torrentClient  *torrent.Client
	activeSessions sync.Map // infoHash -> TorrentSession mapping
)

func InitializeP2P() {
	// client for streaming
	config := torrent.NewDefaultClientConfig()
	// uncomment for prod
	// streamingConfig.DataDir = filepath.Join(os.TempDir(), "torrent-data")
	// downloads grouped by infohash directories
	config.DefaultStorage = storage.NewFileByInfoHash(TorrentDownloadsDir)
	config.Logger.SetHandlers(log.DiscardHandler)
	var err error
	torrentClient, err = torrent.NewClient(config)
	if err != nil {
		panic(err)
	}
	go cleanupSessions()
	slog.Info("Initialized P2P Client")
}

/*
File idx is used, but if -1 (invalid) use filename
*/
func AddTorrent(infoHashStr string, sources []string) error {
	if torrentClient == nil {
		panic("Streaming torrent client is not initialized!")
	}
	var infoHash metainfo.Hash
	if err := infoHash.FromHexString(infoHashStr); err != nil {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid infoHash")
	}
	// don't return error if already exists
	if v, exists := activeSessions.Load(infoHash); exists {
		session, ok := v.(*TorrentSession)
		if !ok {
			return nil
		}
		// update last used
		session.LastUsed = time.Now()
		return nil
	}
	magnetURI := getMagnetURI(infoHash.HexString(), &sources)
	slog.Info("Retrieving Magnet...", "magnet", magnetURI)
	t, err := torrentClient.AddMagnet(*magnetURI)
	if err != nil {
		return helpers.LogErrorWithMessage(err, "Failed to add magnet")
	}
	select {
	case <-t.GotInfo():
		slog.Info("Success Retrieving Magnet Info: " + t.InfoHash().HexString())
	case <-time.After(120 * time.Second):
		return helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Timeout retrieving magnet: "+*magnetURI)
	}
	activeSessions.Store(infoHashStr, &TorrentSession{
		Torrent:  t,
		LastUsed: time.Now(),
	})
	slog.Info("Stored Magnet: " + t.InfoHash().HexString())
	return nil
}

func GetTorrentSession(infoHash string) (*TorrentSession, error) {
	v, ok := activeSessions.Load(infoHash)
	// no need to log, expected to fail sometimes
	if !ok {
		return nil, errors.New(helpers.BadRequest)
	}
	session, ok := v.(*TorrentSession)
	if !ok {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Error parsing TorrentSession")
	}
	return session, nil
}

func GetTorrentFile(infoHash string, fileIdx int, filename string, sources []string) (*torrent.File, *TorrentSession, error) {
	v, ok := activeSessions.Load(infoHash)
	if !ok {
		// force-add the torrent if it doesn't exist
		err := AddTorrent(infoHash, sources)
		if err != nil {
			return nil, nil, err
		}
		v, ok = activeSessions.Load(infoHash)
		if !ok {
			return nil, nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"Could not find torrent session")
		}
	}

	session, ok := v.(*TorrentSession)
	if !ok {
		return nil, nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Error parsing TorrentSession")
	}
	slog.Info("Starting p2p stream", "file", session.Torrent.Files()[fileIdx].DisplayPath())
	// update last used
	session.LastUsed = time.Now()
	t := session.Torrent
	if fileIdx >= len(t.Files()) {
		return nil, nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			fmt.Sprintf("Invalid fileidx: %v, total files: %v", fileIdx, len(t.Files())))
	}
	var fileToStream *torrent.File
	// fileIdx defaults to -1 for no index
	if fileIdx < 0 {
		// no idx or filename, use largest video file
		if filename == "" {
			var largestFile *torrent.File
			for _, file := range t.Files() {
				if largestFile == nil || file.Length() > largestFile.Length() {
					if IsVideoFile(largestFile.DisplayPath()) {
						largestFile = file
					}
				}
			}
			fileToStream = largestFile
			return fileToStream, session, nil
		}
		// use filename
		for _, file := range t.Files() {
			if filename == file.DisplayPath() {
				if IsVideoFile(filename) {
					fileToStream = file
				} else {
					return nil, nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), filename+" is not a valid video file")
				}
			}
		}
		// could return nil, nil
		return fileToStream, session, nil
	}
	fileToStream = t.Files()[fileIdx]
	return fileToStream, session, nil
}

// getMagnetURI returns magnet: uri from hash and trackers
func getMagnetURI(infoHash string, trackers *[]string) *string {
	if infoHash == "" {
		return nil
	}
	magnetURI := fmt.Sprintf("magnet:?xt=urn:btih:%s", strings.ToUpper(infoHash))
	if trackers == nil {
		return &magnetURI
	}
	uniqueTrackers := make(map[string]struct{})
	for _, tracker := range *trackers {
		parts := strings.SplitN(tracker, ":", 2)
		if len(parts) != 2 {
			continue
		}
		sourceType := parts[0]
		value := parts[1]
		if sourceType == "tracker" {
			if _, exists := uniqueTrackers[value]; !exists {
				uniqueTrackers[value] = struct{}{}
			}
		} else {
			slog.Info("Skipping tracker: " + sourceType)
		}
	}
	// append trackers
	var trackerParts []string
	for tracker := range uniqueTrackers {
		escapedTracker := url.QueryEscape(tracker)
		trackerParts = append(trackerParts, fmt.Sprintf("tr=%s", escapedTracker))
	}
	if len(trackerParts) > 0 {
		magnetURI += "&" + strings.Join(trackerParts, "&")
	}
	return &magnetURI
}

func cleanupSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		activeSessions.Range(func(key, value interface{}) bool {
			session := value.(*TorrentSession)
			if time.Since(session.LastUsed) > 15*time.Minute {
				slog.Info("Removed unused session: %s", key)
				session.Torrent.Drop()
				activeSessions.Delete(key)
				// clean up folder
				path := filepath.Join(TorrentDownloadsDir, session.Torrent.InfoHash().HexString())
				slog.Info("Cleaning temp folder", "path", path)
				err := os.RemoveAll(path)
				if err != nil {
					slog.Error("Failed to remove folder: "+session.Torrent.InfoHash().HexString(), "error", err)
				}
			}
			return true
		})
		// TODO evaluate
		// runtime.GC()
	}
}

func IsVideoFile(filename string) bool {
	ext := filepath.Ext(filename)
	ext = strings.ToLower(ext)
	return videoExtensions[ext]
}

func GetMimeType(filename string) string {
	ext := filepath.Ext(filename)
	ext = strings.ToLower(ext)
	return mime.TypeByExtension(ext)
}
