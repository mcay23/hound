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

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
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

type TorrentSession struct {
	Torrent  *torrent.Torrent
	LastUsed time.Time
}

var (
	client         *torrent.Client
	activeTorrents sync.Map // infoHash -> TorrentSession
)

func InitializeP2P() {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = filepath.Join(os.TempDir(), "torrent-data")
	var err error
	client, err = torrent.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	go cleanupSessions()
	slog.Info("Initialized P2P Client")
}

/*
File idx is used, but if -1 (invalid) use filename
*/
func AddTorrent(infoHashStr string, sources []string, fileIdx int, filename string) error {
	if client == nil {
		panic("Torrent client is not initialized!")
	}
	var infoHash metainfo.Hash
	if err := infoHash.FromHexString(infoHashStr); err != nil {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid infoHash")
	}
	if _, exists := activeTorrents.Load(infoHash); exists {
		return nil
	}
	magnetURI := getMagnetLink(infoHash.HexString(), sources)
	slog.Info("Retrieving Magnet...", "magnet", magnetURI)
	t, err := client.AddMagnet(magnetURI)

	if err != nil {
		return helpers.LogErrorWithMessage(err, "Failed to add magnet")
	}
	select {
	case <-t.GotInfo():
		slog.Info("Success Retrieving Magnet Info: " + t.InfoHash().HexString())
	case <-time.After(120 * time.Second):
		return helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Timeout retrieving magnet: "+magnetURI)
	}
	activeTorrents.Store(infoHashStr, &TorrentSession{
		Torrent:  t,
		LastUsed: time.Now(),
	})
	slog.Info("Stored Magnet: " + t.InfoHash().HexString())
	return nil
}

func GetTorrentFile(infoHash string, fileIdx int, filename string) (*torrent.File, error) {
	v, ok := activeTorrents.Load(infoHash)
	if !ok {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "No session with this hash found, add the torrent first")
	}
	session, ok := v.(*TorrentSession)
	if !ok {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Error parsing TorrentSession")
	}
	// update last used
	session.LastUsed = time.Now()
	t := session.Torrent
	if fileIdx >= len(t.Files()) {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
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
			return fileToStream, nil
		}
		// use filename
		for _, file := range t.Files() {
			if filename == file.DisplayPath() {
				if IsVideoFile(filename) {
					fileToStream = file
				} else {
					return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), filename+" is not a valid video file")
				}
			}
		}
		// could return nil, nil
		return fileToStream, nil
	}
	fileToStream = t.Files()[fileIdx]
	return fileToStream, nil
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

// getMagnetLink returns magnet: link from hash and trackers
func getMagnetLink(infoHash string, sources []string) string {
	if infoHash == "" {
		return ""
	}
	uniqueTrackers := make(map[string]struct{})
	for _, source := range sources {
		parts := strings.SplitN(source, ":", 2)
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
	magnetURI := fmt.Sprintf("magnet:?xt=urn:btih:%s", strings.ToUpper(infoHash))

	// append trackers
	var trackerParts []string
	for tracker := range uniqueTrackers {
		escapedTracker := url.QueryEscape(tracker)
		trackerParts = append(trackerParts, fmt.Sprintf("tr=%s", escapedTracker))
	}
	if len(trackerParts) > 0 {
		magnetURI += "&" + strings.Join(trackerParts, "&")
	}
	return magnetURI
}

func cleanupSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		activeTorrents.Range(func(key, value interface{}) bool {
			session := value.(*TorrentSession)

			if time.Since(session.LastUsed) > 15*time.Minute {
				session.Torrent.Drop()
				activeTorrents.Delete(key)
				slog.Info("Removed unused session: %s", key)
			}
			return true
		})
		// TODO evaluate
		// runtime.GC()
	}
}
