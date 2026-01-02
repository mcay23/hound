package model

import (
	"errors"
	"fmt"
	"hound/helpers"
	"log/slog"
	"mime"
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

type TorrentSession struct {
	Torrent       *torrent.Torrent
	ActiveStreams map[int]int // file idx -> num streams
	LastUsed      int64
	Mu            sync.RWMutex
}

var (
	torrentClient  *torrent.Client
	activeSessions sync.Map // infoHash -> TorrentSession mapping
)

func InitializeP2P() {
	config := torrent.NewDefaultClientConfig()
	// downloads grouped by infohash directories
	config.DefaultStorage = storage.NewFileByInfoHash(HoundP2PDownloadsPath)
	config.Logger.SetHandlers(log.DiscardHandler)
	var err error
	torrentClient, err = torrent.NewClient(config)
	if err != nil {
		panic(err)
	}
	go cleanupSessions()
	slog.Info("Initialized P2P Client")
}

func AddActiveTorrentStream(infoHash string, fileIdx int) error {
	v, ok := activeSessions.Load(infoHash)
	if !ok {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Error getting TorrentSession")
	}
	session, ok := v.(*TorrentSession)
	if !ok {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Error parsing TorrentSession")
	}
	session.Mu.Lock()
	defer session.Mu.Unlock()
	session.LastUsed = time.Now().Unix()
	if _, ok := session.ActiveStreams[fileIdx]; !ok {
		session.ActiveStreams[fileIdx] = 1
	} else {
		session.ActiveStreams[fileIdx]++
	}
	slog.Info("Active stream opened", "infoHash", infoHash, "fileIdx", fileIdx, "activeStreams", session.ActiveStreams)
	return nil
}

func RemoveActiveTorrentStream(infoHash string, fileIdx int) error {
	v, ok := activeSessions.Load(infoHash)
	if !ok {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Error getting TorrentSession")
	}
	session, ok := v.(*TorrentSession)
	if !ok {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Error parsing TorrentSession")
	}
	session.Mu.Lock()
	defer session.Mu.Unlock()
	// update last used so it's not removed immediately
	session.LastUsed = time.Now().Unix()
	if _, ok := session.ActiveStreams[fileIdx]; !ok || session.ActiveStreams[fileIdx] <= 0 {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Trying to remove non-existent stream")
	}
	session.ActiveStreams[fileIdx]--
	slog.Info("Active stream closed", "infoHash", infoHash, "fileIdx", fileIdx, "activeStreams", session.ActiveStreams)
	return nil
}

/*
File idx is used, but if -1 (invalid) use filename
*/
func AddTorrent(infoHashStr string, sources *[]string) error {
	if torrentClient == nil {
		panic("Streaming torrent client is not initialized!")
	}
	var hashCheck metainfo.Hash
	if err := hashCheck.FromHexString(infoHashStr); err != nil {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid infoHash")
	}
	// don't return error if already exists
	if v, exists := activeSessions.Load(infoHashStr); exists {
		session, ok := v.(*TorrentSession)
		if !ok {
			return nil
		}
		// update last used
		session.Mu.Lock()
		session.LastUsed = time.Now().Unix()
		session.Mu.Unlock()
		return nil
	}
	magnetURI := helpers.GetMagnetURI(infoHashStr, sources)
	slog.Info("Retrieving Magnet...", "magnet", magnetURI)
	t, err := torrentClient.AddMagnet(magnetURI)
	if err != nil {
		return helpers.LogErrorWithMessage(err, "Failed to add magnet")
	}
	select {
	case <-t.GotInfo():
		slog.Info("Success Retrieving Magnet Info: " + t.InfoHash().HexString())
	case <-time.After(120 * time.Second):
		return helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Timeout retrieving magnet: "+magnetURI)
	}
	activeSessions.Store(infoHashStr, &TorrentSession{
		Torrent:       t,
		LastUsed:      time.Now().Unix(),
		ActiveStreams: make(map[int]int),
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

// check if a torrent session exists
func CheckTorrentSession(infoHash string) bool {
	_, ok := activeSessions.Load(infoHash)
	return ok
}

func GetTorrentFile(infoHash string, fileIdx *int, sources *[]string) (*torrent.File, int, *TorrentSession, error) {
	v, ok := activeSessions.Load(infoHash)
	if !ok {
		// Add the torrent if it doesn't exist
		err := AddTorrent(infoHash, sources)
		if err != nil {
			return nil, -1, nil, err
		}
		v, ok = activeSessions.Load(infoHash)
		if !ok {
			return nil, -1, nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"Could not find torrent session")
		}
	}
	session, ok := v.(*TorrentSession)
	if !ok {
		return nil, -1, nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Error parsing TorrentSession")
	}
	// update last used
	session.Mu.Lock()
	session.LastUsed = time.Now().Unix()
	session.Mu.Unlock()
	t := session.Torrent

	// use largest fileidx if not specified
	if fileIdx == nil {
		largestIdx := -1
		for idx, file := range t.Files() {
			if !IsVideoFile(file.DisplayPath()) {
				continue
			}
			if largestIdx == -1 || file.Length() > t.Files()[largestIdx].Length() {
				largestIdx = idx
			}
		}
		if largestIdx == -1 {
			return nil, -1, nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"Could not find video file")
		}
		fileIdx = &largestIdx
	}
	if *fileIdx >= len(t.Files()) {
		return nil, -1, nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			fmt.Sprintf("Invalid fileidx: %v, total files: %v", *fileIdx, len(t.Files())))
	}
	slog.Info("grabbing p2p file", "file", t.Files()[*fileIdx].DisplayPath())
	return t.Files()[*fileIdx], *fileIdx, session, nil
}

func cleanupSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		activeSessions.Range(func(key, value interface{}) bool {
			session := value.(*TorrentSession)
			// check if active streams exist
			session.Mu.RLock()
			totalStreams := 0
			for _, val := range session.ActiveStreams {
				totalStreams += val
			}
			lastUsed := session.LastUsed
			session.Mu.RUnlock()

			if totalStreams != 0 {
				return true
			}
			// 10 minute grace period
			if time.Now().Unix()-lastUsed > 600 {
				session.Torrent.Drop()
				activeSessions.Delete(key)
				slog.Info("Removed unused session: %s", key)
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
