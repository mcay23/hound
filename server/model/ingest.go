package model

import (
	"errors"
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/model/providers"
	"hound/model/sources"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"
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
	// check if task already exists
	if streamDetails.FileIdx == nil {
		fmt.Println("nil file idx")
	} else {
		fmt.Println("in", *streamDetails.FileIdx)
	}
	tasks, err := database.FindIngestTasks(database.IngestTask{
		SourceURI: &streamDetails.URI,
		FileIdx:   streamDetails.FileIdx,
	})
	if err != nil {
		return helpers.LogErrorWithMessage(err, "Failed to get ingest task when downloading")
	}
	// if a non-terminal task exists for this file, abort
	// eg. downloading/queued
	for _, task := range tasks {
		if task.FileIdx == nil {
			fmt.Println("out nil")
		} else {
			fmt.Println("out", *task.FileIdx)
		}
		if streamDetails.FileIdx == nil || task.FileIdx == nil {
			fmt.Println("in nil")
			return helpers.LogErrorWithMessage(errors.New(helpers.AlreadyExists),
				"Ingest task already exists - nil file idx")
		}
		if !slices.Contains(database.IngestTerminalStatuses, task.Status) && *task.FileIdx == *streamDetails.FileIdx {
			return helpers.LogErrorWithMessage(errors.New(helpers.AlreadyExists),
				"Ingest task already exists")
		}
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
	// 2. Insert ingest task
	// upsert has suceeded, if something else fails database won't be rolled back, which is fine
	_, ingestTask, err := database.InsertIngestTask(ingestRecordID, streamDetails.StreamProtocol,
		database.IngestStatusPendingDownload, streamDetails.URI, streamDetails.FileIdx)
	if err != nil {
		return helpers.LogErrorWithMessage(err, "Failed to insert ingest task when downloading")
	}
	slog.Info("Ingest task inserted successfully", "ingestTask", ingestTask)
	return nil
}

/*
IngestFile copies the downloaded file into the media directory
and adds its metadata to the database
*/
func IngestFile(mediaRecord *database.MediaRecord, seasonNumber *int, episodeNumber *int,
	infoHash *string, fileIdx *int, sourceURI *string, sourcePath string) (*database.MediaFile, error) {
	if mediaRecord == nil {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Nil media record passed to IngestFile()")
	}
	if !IsVideoFile(filepath.Ext(sourcePath)) {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "File is not a video file")
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

	targetDir, targetFilename, targetRecordID, err := GetMediaDestinationDir(mediaRecord, seasonNumber, episodeNumber,
		infoHash, fileIdx, filepath.Ext(sourcePath))
	if err != nil {
		return nil, helpers.LogErrorWithMessage(err, "Failed to get media destination dir")
	}

	// rename, should be atomic since same filesystem
	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		return nil, helpers.LogErrorWithMessage(err, "Failed to create directory")
	}
	// need stricter evaluation for p2p file being streamed
	err = os.Rename(sourcePath, filepath.Join(targetDir, targetFilename))
	if err != nil {
		slog.Error("Failed to rename file, trying link", "sourcePath", sourcePath, "targetPath",
			filepath.Join(targetDir, targetFilename), "error", err)
		// fallback to link, if file is being used
		linkErr := os.Link(sourcePath, filepath.Join(targetDir, targetFilename))
		if linkErr != nil {
			return nil, helpers.LogErrorWithMessage(linkErr, "Failed to rename+link file")
		}
	}

	mediaFile := database.MediaFile{
		Filepath:         filepath.Join(targetDir, targetFilename),
		OriginalFilename: filepath.Base(sourcePath),
		RecordID:         targetRecordID,
		SourceURI:        sourceURI,
		FileIdx:          fileIdx,
		VideoMetadata:    *videoMetadata,
	}
	insertedMediaFile, err := database.InsertMediaFile(&mediaFile)
	if err != nil {
		return nil, helpers.LogErrorWithMessage(err, "Failed to insert video metadata to db"+
			filepath.Join(targetDir, targetFilename))
	}
	slog.Info("Ingestion Complete", "file", filepath.Base(sourcePath))
	return insertedMediaFile, nil
}

func GetMediaDestinationDir(mediaRecord *database.MediaRecord, seasonNumber *int, episodeNumber *int, infoHash *string,
	fileIdx *int, fileExt string) (string, string, int64, error) {
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
	mediaTitleStr += fmt.Sprintf(" {%s-%s}", mediaRecord.MediaSource, mediaRecord.SourceID)
	mediaTitleStr = helpers.SanitizeFilename(mediaTitleStr)
	targetFilename := mediaTitleStr
	var targetRecordID int64

	switch mediaRecord.RecordType {
	case database.RecordTypeMovie:
		if infoHash != nil && *infoHash != "" {
			// append index only if it exists,
			// will probably not exist for http stream sources
			targetFilename += fmt.Sprintf(" {infohash-%s", *infoHash)
			if fileIdx != nil && *fileIdx >= 0 {
				targetFilename += fmt.Sprintf("[%d]", *fileIdx)
			}
			targetFilename += "}"
		}
		targetDir = filepath.Join(HoundMoviesPath, mediaTitleStr)
		targetFilename += fileExt
		targetRecordID = mediaRecord.RecordID
	case database.RecordTypeTVShow:
		if seasonNumber == nil || episodeNumber == nil {
			return "", "", 0, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"Season number or episode number is nil")
		}
		// check if season/episode pair actually exists, and get record id if episode
		episodeRecord, err := database.GetEpisodeMediaRecord(mediaRecord.MediaSource,
			mediaRecord.SourceID, seasonNumber, *episodeNumber)
		if err != nil || episodeRecord == nil {
			return "", "", 0, helpers.LogErrorWithMessage(err, "Failed to get episode media record")
		}
		targetRecordID = episodeRecord.RecordID
		// continue to construct dir
		targetFilename = fmt.Sprintf("%s - S%02dE%02d", mediaTitleStr, *seasonNumber, *episodeNumber)
		// add infohash+fileidx, this just helps with multiple sources per episode
		// eg. Big Buck Bunny (2001) {tmdb-123456} - S1E5 {tmdb-5123} {infohash-ab23ef12[2]}.mp4
		if infoHash != nil && *infoHash != "" {
			targetFilename += fmt.Sprintf(" {infohash-%s", *infoHash)
			if fileIdx != nil && *fileIdx >= 0 {
				targetFilename += fmt.Sprintf("[%d]", *fileIdx)
			}
			targetFilename += "}"
		}
		seasonPath := fmt.Sprintf("Season %02d", *seasonNumber)
		targetDir = filepath.Join(HoundTVShowsPath, mediaTitleStr, seasonPath)
		targetFilename += fileExt
	default:
		return "", "", 0, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid record type")
	}
	return targetDir, targetFilename, targetRecordID, nil
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
