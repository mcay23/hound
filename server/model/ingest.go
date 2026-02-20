package model

import (
	"errors"
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/model/providers"
	"hound/sources"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/anacrolix/torrent/metainfo"
)

// whether to move, copy or preserve the source file
// to the hound directory
// preserve = keep source file where it is, typically for
// external libraries
const (
	IngestTransferMove     = "move"
	IngestTransferCopy     = "copy"
	IngestTransferPreserve = "preserve"
)

// Downloads torrent to server, not clients
func CreateIngestTaskDownload(streamDetails *providers.StreamObjectFull) error {
	if streamDetails == nil {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Nil stream details passed to DownloadTorrent()")
	}
	if streamDetails.MediaSource != sources.MediaSourceTMDB {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Invalid media source, only tmdb is supported: "+streamDetails.MediaSource)
	}
	// 1. Attempt upsert first, if failed, abort
	tmdbID, err := strconv.Atoi(streamDetails.SourceID)
	if err != nil {
		return helpers.LogErrorWithMessage(err, "Failed to convert source ID to int when downloading")
	}
	mediaRecord, err := sources.UpsertMediaRecordTMDB(streamDetails.MediaType, tmdbID)
	if err != nil {
		return helpers.LogErrorWithMessage(err, "Failed to upsert media record when downloading")
	}
	ingestRecordID := mediaRecord.RecordID // movie/episode record, not shows/seasons
	if mediaRecord.RecordType == database.RecordTypeTVShow {
		episodeRecord, err := database.GetEpisodeMediaRecord(mediaRecord.MediaSource,
			mediaRecord.SourceID, streamDetails.SeasonNumber, *streamDetails.EpisodeNumber)
		if err != nil || episodeRecord == nil {
			return helpers.LogErrorWithMessage(err, "Failed to get episode media record when downloading")
		}
		ingestRecordID = episodeRecord.RecordID
	}
	// 2. Check if a non-terminal (downloading, queued) task already exists
	tasks, err := database.FindIngestTasks(database.IngestTask{
		SourceURI: &streamDetails.URI,
		RecordID:  ingestRecordID,
	})
	if err != nil {
		return helpers.LogErrorWithMessage(err, "Failed to get ingest task when downloading")
	}
	for _, task := range tasks {
		// note that we don't check for 'done' state since the file may have been deleted afterwards
		if !slices.Contains(database.IngestTerminalStatuses, task.Status) && task.SourceURI != nil {
			uri, err := metainfo.ParseMagnetUri(*task.SourceURI)
			if err != nil {
				continue
			}
			infoHash := uri.InfoHash.HexString()
			if strings.EqualFold(infoHash, streamDetails.InfoHash) {
				return helpers.LogErrorWithMessage(errors.New(helpers.AlreadyExists),
					"Ingest task already exists - downloading/queued")
			}
		}
	}
	// 3. Check if media file already exists for this movie/episode record
	mediaFiles, err := database.GetMediaFileByRecordID(int(ingestRecordID))
	if err != nil {
		return helpers.LogErrorWithMessage(err, "Failed to get media files when downloading")
	}
	for _, mediaFile := range mediaFiles {
		if mediaFile.SourceURI != nil {
			uri, err := metainfo.ParseMagnetUri(*mediaFile.SourceURI)
			if err != nil {
				continue
			}
			infoHash := uri.InfoHash.HexString()
			// matching infohash, same file
			// there's an unhandled edge case where a single torrent may have
			// multiple versions of the same movie/episode, which is unhandled here
			if strings.EqualFold(infoHash, streamDetails.InfoHash) {
				return helpers.LogErrorWithMessage(errors.New(helpers.AlreadyExists),
					"Ingest task already exists - file already downloaded")
			}
		}
	}
	// 3. Insert ingest task
	// upsert has suceeded, if something else fails database won't be rolled back, which is fine
	_, ingestTask, err := database.InsertIngestTask(ingestRecordID, streamDetails.StreamProtocol,
		database.IngestStatusPendingDownload, streamDetails.URI, streamDetails.FileIdx)
	if err != nil {
		return helpers.LogErrorWithMessage(err, "Failed to insert ingest task when downloading")
	}
	slog.Info("Ingest task inserted successfully", "ingestTask", ingestTask)
	return nil
}

func IngestFile(mediaRecord *database.MediaRecord, seasonNumber *int, episodeNumber *int,
	infoHash *string, fileIdx *int, sourceURI *string, sourcePath string, transferMode string, fileOrigin string) (*database.MediaFile, error) {
	if mediaRecord == nil {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Nil media record passed to IngestFile()")
	}
	if transferMode != IngestTransferMove && transferMode != IngestTransferCopy && transferMode != IngestTransferPreserve {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid ingest transfer mode")
	}
	if fileOrigin != database.FileOriginHoundManaged && fileOrigin != database.FileOriginExternalLibrary {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid file origin")
	}
	if !IsVideoFile(filepath.Ext(sourcePath)) {
		return nil, helpers.LogErrorWithMessage(fmt.Errorf("File is not a video file %s", sourcePath), "File is not a video file")
	}
	// ffprobe video
	videoMetadata, err := ProbeVideoFromURI(sourcePath)
	if err != nil {
		return nil, helpers.LogErrorWithMessage(err, "Failed to probe video + "+sourcePath)
	}
	// less than 1 min is invalid, used by some providers to display
	// video not cached message, might want to explore fallback to p2p
	// in this case
	if videoMetadata.Duration < 1*time.Minute {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.VideoDurationTooShort),
			fmt.Sprintf("Video duration too short: %v (<1 minute)", videoMetadata.Duration))
	}

	var targetRecordID int64
	var targetPath string
	targetRecordID, err = getIngestTargetRecordID(mediaRecord, seasonNumber, episodeNumber)
	if err != nil {
		return nil, helpers.LogErrorWithMessage(err, "Failed to get ingest target record id")
	}
	if transferMode == IngestTransferPreserve {
		targetPath = sourcePath
	} else {
		targetDir, targetFilename, _, err := getMediaDestinationDir(mediaRecord, seasonNumber, episodeNumber,
			infoHash, fileIdx, filepath.Ext(sourcePath))
		if err != nil {
			return nil, helpers.LogErrorWithMessage(err, "Failed to get media destination dir")
		}
		err = os.MkdirAll(targetDir, 0755)
		if err != nil {
			return nil, helpers.LogErrorWithMessage(err, "Failed to create directory")
		}
		targetPath = filepath.Join(targetDir, targetFilename)
		// for external library cases, we probably just want to preserve source in original location
		// but copy may be needed one day if users want a full migration
		switch transferMode {
		case IngestTransferMove:
			// same-filesystem move is atomic
			// should be fast for hound-managed files since downloads folder and media folder are in the same
			// directory
			// TODO Edge case: this might fail/brick if file is being streamed and downloaded at the same time,
			// when download completes while file is being watched
			err = os.Rename(sourcePath, targetPath)
			if err != nil {
				// fallback to link when source is still open/locked
				linkErr := os.Link(sourcePath, targetPath)
				if linkErr != nil {
					return nil, helpers.LogErrorWithMessage(linkErr, "Failed to move file with rename+link fallback")
				}
			}
		case IngestTransferCopy:
			// try hardlink first, copy fallback
			err = os.Link(sourcePath, targetPath)
			if err != nil {
				err = copyFile(sourcePath, targetPath)
				if err != nil {
					return nil, helpers.LogErrorWithMessage(err, "Failed to copy file")
				}
			}
		}
	}
	mediaFile := database.MediaFile{
		Filepath:         targetPath,
		OriginalFilename: filepath.Base(sourcePath),
		RecordID:         targetRecordID,
		FileOrigin:       fileOrigin,
		SourceURI:        sourceURI,
		FileIdx:          fileIdx,
		VideoMetadata:    *videoMetadata,
	}
	insertedMediaFile, err := database.InsertMediaFile(&mediaFile)
	if err != nil {
		return nil, helpers.LogErrorWithMessage(err, "Failed to insert video metadata to db "+targetPath)
	}
	slog.Info("Ingestion Complete", "file", filepath.Base(sourcePath))
	return insertedMediaFile, nil
}

func getIngestTargetRecordID(mediaRecord *database.MediaRecord, seasonNumber *int, episodeNumber *int) (int64, error) {
	switch mediaRecord.RecordType {
	case database.RecordTypeMovie:
		return mediaRecord.RecordID, nil
	case database.RecordTypeTVShow:
		if seasonNumber == nil || episodeNumber == nil {
			return 0, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"Season number or episode number is nil")
		}
		episodeRecord, err := database.GetEpisodeMediaRecord(mediaRecord.MediaSource,
			mediaRecord.SourceID, seasonNumber, *episodeNumber)
		if err != nil || episodeRecord == nil {
			return 0, helpers.LogErrorWithMessage(err, "Failed to get episode media record")
		}
		return episodeRecord.RecordID, nil
	default:
		return 0, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid record type")
	}
}

func getMediaDestinationDir(mediaRecord *database.MediaRecord, seasonNumber *int, episodeNumber *int, infoHash *string,
	fileIdx *int, fileExt string) (string, string, int64, error) {
	if fileExt == "" || fileExt[0] != '.' {
		return "", "", 0, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"File extension is empty or does not include .")
	}
	targetDir := ""
	// construct title, append this later for each type
	// format eg. Big Buck Bunny (2001) {tmdb-123456}
	releaseYear := ""
	if len(mediaRecord.ReleaseDate) >= 4 {
		releaseYear = mediaRecord.ReleaseDate[0:4]
	}
	mediaTitleStr := mediaRecord.MediaTitle
	if releaseYear != "" {
		mediaTitleStr += " (" + releaseYear + ")"
	}
	// [tmdbid-1234], matches with jellyfin scheme
	mediaTitleStr += fmt.Sprintf(" [%sid-%s]", mediaRecord.MediaSource, mediaRecord.SourceID)
	mediaTitleStr = helpers.SanitizeFilename(mediaTitleStr)
	targetFilename := mediaTitleStr
	var targetRecordID int64

	switch mediaRecord.RecordType {
	case database.RecordTypeMovie:
		if infoHash != nil && *infoHash != "" {
			// append index only if it exists,
			// will often not exist for http stream sources
			targetFilename += fmt.Sprintf(" {infohash-%s", *infoHash)
			if fileIdx != nil && *fileIdx >= 0 {
				targetFilename += fmt.Sprintf("[%d]", *fileIdx)
			}
			targetFilename += "}"
		}
		targetDir = filepath.Join(HoundMoviesPath, mediaTitleStr)
		targetRecordID = mediaRecord.RecordID
	case database.RecordTypeTVShow:
		if seasonNumber == nil || episodeNumber == nil {
			return "", "", 0, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"Season number or episode number is nil")
		}
		// check if season/episode pair actually exists, and get record id of episode
		episodeRecord, err := database.GetEpisodeMediaRecord(mediaRecord.MediaSource,
			mediaRecord.SourceID, seasonNumber, *episodeNumber)
		if err != nil || episodeRecord == nil {
			return "", "", 0, helpers.LogErrorWithMessage(err, "Failed to get episode media record")
		}
		targetRecordID = episodeRecord.RecordID
		// continue to construct dir
		targetFilename = fmt.Sprintf("%s - S%02dE%02d", mediaTitleStr, *seasonNumber, *episodeNumber)
		// add infohash+fileidx, this just helps with multiple sources per episode
		// eg. Big Buck Bunny (2001) [tmdbid-123456] - S01E05 {infohash-ab23ef12[2]}.mp4
		if infoHash != nil && *infoHash != "" {
			targetFilename += fmt.Sprintf(" {infohash-%s", *infoHash)
			if fileIdx != nil && *fileIdx >= 0 {
				targetFilename += fmt.Sprintf("[%d]", *fileIdx)
			}
			targetFilename += "}"
		}
		seasonPath := fmt.Sprintf("Season %02d", *seasonNumber)
		targetDir = filepath.Join(HoundTVShowsPath, mediaTitleStr, seasonPath)
	default:
		return "", "", 0, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid record type")
	}
	name := targetFilename
	targetFilename += fileExt
	// check if the filename already exists, if so append a number
	// this is to solve the edge case -> no infohash provided, trying to
	// download the same movie/episode
	for i := 0; ; i++ {
		var candidate string
		if i == 0 {
			candidate = targetFilename
		} else {
			candidate = fmt.Sprintf("%s (%d)%s", name, i, fileExt)
		}
		path := filepath.Join(targetDir, candidate)
		_, err := os.Stat(path)
		if err == nil {
			continue
		} else if os.IsNotExist(err) {
			// file does not exist, therefore candidate is good
			targetFilename = candidate
			break
		} else {
			return "", "", 0, helpers.LogErrorWithMessage(err, "Failed to stat file: "+path)
		}
	}
	return targetDir, targetFilename, targetRecordID, nil
}

func copyFile(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return dstFile.Sync()
}

// Helper function to copy files from downloads -> media directory
// update the torrent session periodically in case copy takes time,
// so the torrent session isn't dropped and files deleted before copy is complete
// deprecate in favor of atomic move
// func copyWithUpdateTorrentSession(src, dst string, session *TorrentSession) error {
// 	srcInfo, err := os.Stat(src)
// 	if err != nil {
// 		return helpers.LogErrorWithMessage(err, "Failed to stat source file: "+src)
// 	}
// 	if !srcInfo.Mode().IsRegular() {
// 		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Source is not a regular file")
// 	}
// 	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
// 		return helpers.LogErrorWithMessage(err, "Failed to create destination directory")
// 	}
// 	// keep updating session in case copy takes time
// 	done := make(chan struct{})
// 	if session != nil {
// 		go func() {
// 			t := time.NewTicker(time.Second * 60)
// 			defer t.Stop()
// 			for {
// 				select {
// 				case <-t.C:
// 					session.LastUsed = time.Now()
// 				case <-done:
// 					return
// 				}
// 			}
// 		}()
// 	}
// 	defer close(done)
// 	_ = os.Remove(dst)
// 	// copy via hardlinks
// 	if err := os.Link(src, dst); err == nil {
// 		return nil
// 	}
// 	// fallback to regular copy
// 	in, err := os.Open(src)
// 	if err != nil {
// 		return err
// 	}
// 	defer in.Close()
// 	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
// 	if err != nil {
// 		return err
// 	}
// 	defer out.Close()
// 	if _, err := io.Copy(out, in); err != nil {
// 		return err
// 	}
// 	return out.Sync()
// }
