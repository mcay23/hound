package model

import (
	"errors"
	"hound/database"
	"hound/helpers"
	"hound/model/sources"
	"log/slog"
)

// Downloads torrent to server, not clients
func CreateIngestTaskDownload(streamDetails *StreamObjectFull) error {
	if streamDetails == nil {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Nil stream details passed to DownloadTorrent()")
	}
	if streamDetails.MediaSource != sources.MediaSourceTMDB {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Invalid media source, only tmdb is supported: "+streamDetails.MediaSource)
	}
	// 1. Attempt upsert first, if failed, abort
	mediaRecord, err := sources.UpsertMediaRecordTMDB(streamDetails.MediaType, streamDetails.SourceID)
	if err != nil {
		return helpers.LogErrorWithMessage(err, "Failed to upsert media record when downloading")
	}
	ingestRecordID := mediaRecord.RecordID // movie/episode record, not shows/seasons
	if mediaRecord.RecordType == database.RecordTypeTVShow {
		episodeRecord, err := database.GetEpisodeMediaRecord(mediaRecord.MediaSource,
			&mediaRecord.SourceID, streamDetails.SeasonNumber, *streamDetails.EpisodeNumber)
		if err != nil || episodeRecord == nil {
			return helpers.LogErrorWithMessage(err, "Failed to get episode media record when downloading")
		}
		ingestRecordID = episodeRecord.RecordID
	}
	// 2. Insert ingest task
	// upsert has suceeded, if something else fails database won't be rolled back, which is fine
	// don't store trackers in uri
	_, ingestTask, err := database.InsertIngestTask(ingestRecordID, database.DownloadTypeP2P,
		database.IngestStatusPendingDownload, *getMagnetURI(streamDetails.InfoHash, nil),
		streamDetails.FileIndex)
	if err != nil {
		return helpers.LogErrorWithMessage(err, "Failed to insert ingest task when downloading")
	}
	slog.Info("Ingest task inserted successfully", "ingestTask", ingestTask)
	return nil
	// file, session, err := GetTorrentFile(streamDetails.InfoHash, streamDetails.FileIndex,
	// 	streamDetails.Filename, streamDetails.Sources)
	// relativePath := filepath.FromSlash(file.Path())
	// currentDir, _ := os.Getwd()
	// absolutePath := filepath.Join(currentDir, TorrentDownloadsDir, strings.ToLower(streamDetails.InfoHash), relativePath)
	// fmt.Println("SeasonNumber", *streamDetails.SeasonNumber, "EpisodeNumber", *streamDetails.EpisodeNumber)
	// fmt.Println("path", absolutePath)
	// if err != nil {
	// 	return err
	// }

	// go func() {
	// 	file.Download()
	// 	for {
	// 		// keep session alive while still downloading
	// 		session.LastUsed = time.Now()
	// 		completed := file.BytesCompleted()
	// 		total := file.Length()
	// 		fmt.Printf("\rProgress: %.2f%% (%d/%d bytes)",
	// 			float64(completed)/float64(total)*100, completed, total)

	// 		if completed >= total {
	// 			slog.Info("Download finished, starting ingestion", "infohash", streamDetails.InfoHash,
	// 				"filename", streamDetails.Filename)
	// 			_, err := IngestFile(mediaRecord, streamDetails.SeasonNumber, streamDetails.EpisodeNumber,
	// 				&streamDetails.InfoHash, &streamDetails.FileIndex, absolutePath)
	// 			if err != nil {
	// 				slog.Error("Failed to ingest file!", "error", err)
	// 			}
	// 			break
	// 		}
	// 		time.Sleep(5 * time.Second)
	// 	}
	// }()
	// return nil
}
