package model

import (
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/loggers"
	"hound/providers"
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
func CreateIngestTaskDownload(streamDetails *providers.StreamObjectFull, prefs *database.IngestDownloadPreferences, skipDownloaded bool) error {
	if streamDetails == nil {
		return fmt.Errorf("nil stream details passed to DownloadTorrent(): %w", helpers.BadRequestError)
	}
	if streamDetails.MediaSource != sources.MediaSourceTMDB {
		return fmt.Errorf("invalid media source, only tmdb is supported: %s: %w", streamDetails.MediaSource, helpers.BadRequestError)
	}
	if streamDetails.MediaType != database.RecordTypeMovie && streamDetails.MediaType != database.RecordTypeTVShow {
		return fmt.Errorf("invalid media type, only movie and tvshow are supported: %s: %w", streamDetails.MediaType, helpers.BadRequestError)
	}
	// 1. Attempt upsert first, if failed, abort
	tmdbID, err := strconv.Atoi(streamDetails.SourceID)
	if err != nil {
		return fmt.Errorf("failed to convert source id to int: %s: %w", streamDetails.SourceID, err)
	}
	mediaRecord, err := sources.UpsertMediaRecordTMDB(streamDetails.MediaType, tmdbID)
	if err != nil {
		return fmt.Errorf("failed to upsert media record: %s-%d: %w", streamDetails.MediaType, tmdbID, err)
	}
	// get movie/episode record, not shows/seasons
	childRecord := mediaRecord
	if mediaRecord.RecordType == database.RecordTypeTVShow {
		episodeRecord, err := database.GetEpisodeMediaRecord(mediaRecord.MediaSource,
			mediaRecord.SourceID, streamDetails.SeasonNumber, streamDetails.EpisodeNumber)
		if err != nil || episodeRecord == nil {
			return fmt.Errorf("failed to get episode media record for tvshow %s-%d s%d-e%d: %w",
				streamDetails.MediaType, tmdbID, streamDetails.SeasonNumber, *streamDetails.EpisodeNumber, err)
		}
		childRecord = episodeRecord
	}
	// 2. Check if a non-terminal task or media file already exists
	if streamDetails.StreamProtocol != "" && streamDetails.URI != "" {
		err = CheckDuplicateDownloadTask(childRecord, -1, streamDetails.StreamProtocol,
			streamDetails.URI, streamDetails.InfoHash, streamDetails.FileIdx, skipDownloaded)
		if err != nil {
			return err
		}
	} else {
		// URI/stream protocol empty, this is resolved by the worker
		// Do simple checking to see if it's already downloading/being downloaded
		// for simpler logic, for now, don't allow concurrent season downloads even if they have
		// different preferences
		tasks, _ := database.FindIngestTasks(database.IngestTask{RecordID: childRecord.RecordID})
		for _, task := range tasks {
			if !slices.Contains(database.IngestTerminalStatuses, task.Status) {
				return fmt.Errorf("file already queued/downloading: %s-%d: %w",
					childRecord.MediaSource, childRecord.SourceID, helpers.AlreadyExistsError)
			}
		}
		if skipDownloaded {
			// check already existing files
			mediaFiles, err := database.GetMediaFileByRecordID(int(childRecord.RecordID))
			if err != nil {
				return err
			}
			for _, file := range mediaFiles {
				_, err := os.Stat(file.Filepath)
				// files that don't exist don't count
				if os.IsNotExist(err) {
					continue
				}
				return fmt.Errorf("file already downloaded: %s-%s: %w",
					childRecord.MediaSource, childRecord.SourceID, helpers.AlreadyExistsError)
			}
		}
	}
	// 3. Insert ingest task
	taskToInsert := &database.IngestTask{
		RecordID:            childRecord.RecordID,
		DownloadProtocol:    streamDetails.StreamProtocol,
		Status:              database.IngestStatusPendingDownload,
		FileIdx:             streamDetails.FileIdx,
		DownloadPreferences: prefs,
	}
	if streamDetails.URI != "" {
		taskToInsert.SourceURI = &streamDetails.URI
	}
	_, _, err = database.InsertIngestTask(taskToInsert)
	if err != nil {
		return err
	}
	slog.Info("Ingest task inserted successfully", "ingestTask", taskToInsert)
	return nil
}

/*
CheckDuplicateDownloadTask checks if a download task is a duplicate (already downloaded or being downloaded)
This is not a fool-proof check, false negatives may occur
*/
func CheckDuplicateDownloadTask(mediaRecord *database.MediaRecord, currentTaskID int64, protocol string,
	sourceURI string, currInfoHash string, currFileIdx *int, skipDownloaded bool) error {
	if mediaRecord == nil {
		return fmt.Errorf("nil mediaRecord: %w", helpers.BadRequestError)
	}
	if sourceURI == "" {
		return fmt.Errorf("empty sourceURI: %w", helpers.BadRequestError)
	}
	tasks, err := database.FindIngestTasks(database.IngestTask{
		RecordID: mediaRecord.RecordID,
	})
	if err != nil {
		return fmt.Errorf("failed to get ingest tasks during duplicate check for record %d: %w", mediaRecord.RecordID, err)
	}
	// 1. Check ingest tasks for queued/downloaded tasks for this particular record
	for _, task := range tasks {
		if task.IngestTaskID == currentTaskID || task.DownloadProtocol != protocol {
			continue
		}
		// note that we skip 'done' tasks since there's no guarantee the file
		// still exists even if it was downloaded before
		if !slices.Contains(database.IngestTerminalStatuses, task.Status) && task.SourceURI != nil {
			// for http case, same sourceURI will be the same file
			switch protocol {
			case database.ProtocolProxyHTTP:
				if *task.SourceURI == sourceURI {
					return fmt.Errorf("http file already queued/downloading for %s %s-%d: %w", mediaRecord.RecordType,
						mediaRecord.MediaSource, mediaRecord.SourceID, helpers.AlreadyExistsError)
				}
			case database.ProtocolP2P:
				// for p2p case, sourceURI is the magnetURI w/ trackers, depending on the file index
				// it might be a different file. Here, we know that the episode/movie record is the
				// same, but technically you might be downloading a different version of a movie
				// from the same magnet torrent, so we check for fileidx equality
				magnet, err := metainfo.ParseMagnetUri(*task.SourceURI)
				// check if task's sourceURI has the same torrent infohash
				if err == nil && strings.EqualFold(magnet.InfoHash.HexString(), currInfoHash) {
					// if it does, we still need to know if the file idx is the same
					if currFileIdx != nil && task.FileIdx != nil && *currFileIdx == *task.FileIdx {
						return fmt.Errorf("p2p file already queued/downloading for %s %s-%d: %w", mediaRecord.RecordType,
							mediaRecord.MediaSource, mediaRecord.SourceID, helpers.AlreadyExistsError)
					} else if currFileIdx == nil && task.FileIdx == nil {
						// when both is nil, some providers implicitly expect largest file
						// here, we assume it refers to the same file
						return fmt.Errorf("p2p file already queued/downloading for %s %s-%d: %w", mediaRecord.RecordType,
							mediaRecord.MediaSource, mediaRecord.SourceID, helpers.AlreadyExistsError)
					}
				}
			}
		}
	}
	// 2. Check against media_files table
	mediaFiles, err := database.GetMediaFileByRecordID(int(mediaRecord.RecordID))
	if err != nil {
		return fmt.Errorf("failed to get media files for record %d: %w", mediaRecord.RecordID, err)
	}

	if skipDownloaded && len(mediaFiles) > 0 {
		return fmt.Errorf("(skipDownloaded=true) file already downloaded for %s %s-%d: %w", mediaRecord.RecordType,
			mediaRecord.MediaSource, mediaRecord.SourceID, helpers.AlreadyExistsError)
	}

	for _, mediaFile := range mediaFiles {
		// files may have been loaded in different ways, sourceURI not guaranteed to always
		// exist for ingested files (?). If sourceURI doesn't exist, we have no reliable way to
		// check for duplicates unless we use a video hashing algorithm
		if mediaFile.SourceURI != nil {
			// if unable to find file (manually deleted?), skip
			_, err := os.Stat(mediaFile.Filepath)
			if os.IsNotExist(err) {
				continue
			}
			if protocol == database.ProtocolProxyHTTP && strings.HasPrefix(*mediaFile.SourceURI, "http") &&
				*mediaFile.SourceURI == sourceURI {
				return fmt.Errorf("http file already downloaded for %s %s-%d: %w", mediaRecord.RecordType,
					mediaRecord.MediaSource, mediaRecord.SourceID, helpers.AlreadyExistsError)
			} else if protocol == database.ProtocolP2P {
				magnet, err := metainfo.ParseMagnetUri(*mediaFile.SourceURI)
				if err == nil && strings.EqualFold(magnet.InfoHash.HexString(), currInfoHash) {
					if currFileIdx != nil && mediaFile.FileIdx != nil && *currFileIdx == *mediaFile.FileIdx {
						return fmt.Errorf("p2p file already downloaded for %s %s-%d: %w", mediaRecord.RecordType,
							mediaRecord.MediaSource, mediaRecord.SourceID, helpers.AlreadyExistsError)
					} else if currFileIdx == nil && mediaFile.FileIdx == nil {
						return fmt.Errorf("p2p file already downloaded for %s %s-%d: %w", mediaRecord.RecordType,
							mediaRecord.MediaSource, mediaRecord.SourceID, helpers.AlreadyExistsError)
					}
				}
			}
		}
	}
	return nil
}

func IngestFile(mediaRecord *database.MediaRecord, seasonNumber *int, episodeNumber *int,
	infoHash *string, fileIdx *int, sourceURI *string, sourcePath string, transferMode string, fileOrigin string) (*database.MediaFile, error) {
	if mediaRecord == nil {
		return nil, fmt.Errorf("nil mediaRecord: %w", helpers.BadRequestError)
	}
	if transferMode != IngestTransferMove && transferMode != IngestTransferCopy && transferMode != IngestTransferPreserve {
		return nil, fmt.Errorf("invalid ingest transfer mode: %s: %w", transferMode, helpers.BadRequestError)
	}
	if fileOrigin != database.FileOriginHoundManaged && fileOrigin != database.FileOriginExternalLibrary {
		return nil, fmt.Errorf("invalid file origin: %s: %w", fileOrigin, helpers.BadRequestError)
	}
	if !IsVideoFile(filepath.Ext(sourcePath)) {
		return nil, fmt.Errorf("file is not a video file: %s: %w", sourcePath, helpers.BadRequestError)
	}
	// ffprobe video
	videoMetadata, err := ProbeVideoFromURI(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to probe video: %s: %w", sourcePath, err)
	}
	// less than 1 min is invalid, used by some providers to display
	// video not cached message, might want to explore fallback to p2p
	// in this case
	if videoMetadata.Duration < 1*time.Minute {
		return nil, fmt.Errorf("video duration too short: %v (<1 minute): %w", videoMetadata.Duration, helpers.BadRequestError)
	}

	var targetRecordID int64
	var targetPath string
	// for tv shows, get episode's record id
	targetRecordID, err = getIngestTargetRecordID(mediaRecord, seasonNumber, episodeNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get ingest target for %s %s-%d: %w", mediaRecord.RecordType,
			mediaRecord.MediaSource, mediaRecord.SourceID, err)
	}
	if transferMode == IngestTransferPreserve {
		targetPath = sourcePath
	} else {
		targetDir, targetFilename, _, err := getMediaDestinationDir(mediaRecord, seasonNumber, episodeNumber,
			infoHash, fileIdx, filepath.Ext(sourcePath))
		if err != nil {
			return nil, fmt.Errorf("failed to get media destination dir: %w", err)
		}
		err = os.MkdirAll(targetDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create directory: %s: %w", targetDir, err)
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
					return nil, fmt.Errorf("failed to move file with rename+link fallback: %w", linkErr)
				}
			}
		case IngestTransferCopy:
			// try hardlink first, copy fallback
			err = os.Link(sourcePath, targetPath)
			if err != nil {
				err = copyFile(sourcePath, targetPath)
				if err != nil {
					return nil, fmt.Errorf("failed to copy file: %w", err)
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
		return nil, err
	}
	slog.Info("Ingestion Complete", "file", filepath.Base(sourcePath))
	loggers.IngestLogger().Info("[External Library: Ingestion Complete]", "path", sourcePath, "SourceID", mediaRecord.SourceID,
		"Title", mediaRecord.MediaTitle, "Release", mediaRecord.ReleaseDate, "Season", seasonNumber, "Episode", episodeNumber)
	return insertedMediaFile, nil
}

func getIngestTargetRecordID(mediaRecord *database.MediaRecord, seasonNumber *int, episodeNumber *int) (int64, error) {
	switch mediaRecord.RecordType {
	case database.RecordTypeMovie:
		return mediaRecord.RecordID, nil
	case database.RecordTypeTVShow:
		if seasonNumber == nil || episodeNumber == nil {
			return 0, fmt.Errorf("season number or episode number is nil: %w", helpers.BadRequestError)
		}
		episodeRecord, err := database.GetEpisodeMediaRecord(mediaRecord.MediaSource,
			mediaRecord.SourceID, seasonNumber, episodeNumber)
		if err != nil || episodeRecord == nil {
			return 0, fmt.Errorf("failed to get episode media record for %s %s-%d: %w", mediaRecord.RecordType,
				mediaRecord.MediaSource, mediaRecord.SourceID, err)
		}
		return episodeRecord.RecordID, nil
	default:
		return 0, fmt.Errorf("invalid record type: %s: %w", mediaRecord.RecordType, helpers.BadRequestError)
	}
}

func getMediaDestinationDir(mediaRecord *database.MediaRecord, seasonNumber *int, episodeNumber *int, infoHash *string,
	fileIdx *int, fileExt string) (string, string, int64, error) {
	if fileExt == "" || fileExt[0] != '.' {
		return "", "", 0, fmt.Errorf("file extension is empty or does not include . for %s: %w", fileExt, helpers.BadRequestError)
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
			return "", "", 0, fmt.Errorf("season number or episode number is nil for %s %s-%d: %w", mediaRecord.RecordType,
				mediaRecord.MediaSource, mediaRecord.SourceID, helpers.BadRequestError)
		}
		// check if season/episode pair actually exists, and get record id of episode
		episodeRecord, err := database.GetEpisodeMediaRecord(mediaRecord.MediaSource,
			mediaRecord.SourceID, seasonNumber, episodeNumber)
		if err != nil || episodeRecord == nil {
			return "", "", 0, fmt.Errorf("failed to get episode media record: %w", err)
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
		return "", "", 0, fmt.Errorf("invalid record type: %s: %w", mediaRecord.RecordType, helpers.BadRequestError)
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
			return "", "", 0, fmt.Errorf("failed to stat file: %s: %w", path, err)
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
