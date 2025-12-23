package model

import (
	"errors"
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/services"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	MediaPath     = "Media"
	MoviesPath    = "Movies"
	TVShowsPath   = "TV Shows"
	DownloadsPath = "Downloads"
)

/*
media deals with file ingestion pipeline download->create files->process metadata...etc.
*/
func InitializeMedia() {
	// create media directories
	err := os.MkdirAll(filepath.Join(MediaPath, MoviesPath), 0755)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to create media directory")
		panic(fmt.Errorf("fatal error creating media directory %w", err))
	}
	err = os.MkdirAll(filepath.Join(MediaPath, TVShowsPath), 0755)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to create media directory")
		panic(fmt.Errorf("fatal error creating media directory %w", err))
	}
	err = os.MkdirAll(filepath.Join(DownloadsPath), 0755)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to create downloads directory")
		panic(fmt.Errorf("fatal error creating downloads directory %w", err))
	}
}

/*
IngestFile copies the downloaded file into the media directory
and adds its metadata to the database
*/
func IngestFile(mediaRecord *database.MediaRecord, seasonNumber *int, episodeNumber *int,
	infoHash *string, fileIdx *int, sourcePath string) (*database.MediaFile, error) {
	if mediaRecord == nil {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Nil media record passed to IngestFile()")
	}
	if !IsVideoFile(filepath.Ext(sourcePath)) {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "File is not a video file")
	}
	targetDir := ""
	targetFilename := ""
	var targetRecordID int64
	switch mediaRecord.RecordType {
	case database.RecordTypeMovie:
		targetDir = filepath.Join(MediaPath, MoviesPath)
		targetRecordID = mediaRecord.RecordID
	case database.RecordTypeTVShow:
		if seasonNumber == nil || episodeNumber == nil {
			return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"Season number or episode number is nil")
		}
		// check if season/episode pair actually exists
		episodeRecord, err := database.GetEpisodeMediaRecord(mediaRecord.MediaSource,
			&mediaRecord.SourceID, seasonNumber, *episodeNumber)
		if err != nil || episodeRecord == nil {
			return nil, helpers.LogErrorWithMessage(err, "Failed to get episode media record")
		}
		targetRecordID = episodeRecord.RecordID
		// continue to construct dir
		// eg. Big Buck Bunny (2001) {tmdb-123456}
		mediaTitleStr := fmt.Sprintf("%s (%s) {%s-%s}", mediaRecord.MediaTitle, mediaRecord.ReleaseDate[0:4],
			mediaRecord.MediaSource, mediaRecord.SourceID)
		mediaTitleStr = helpers.SanitizeFilename(mediaTitleStr)
		targetFilename = fmt.Sprintf("%s - S%02dE%02d", mediaTitleStr, *seasonNumber, *episodeNumber)
		// add infohash, this just helps with multiple sources per episode
		// of course, it's possible to have multiple qualities per infohash
		// eg. Big Buck Bunny (2001) {tmdb-123456} - S1E5 {tmdb-5123} {infohash-ab23ef12[2]}.mp4
		if infoHash != nil && *infoHash != "" && fileIdx != nil && *fileIdx >= 0 {
			targetFilename += fmt.Sprintf(" {infohash-%s[%d]}", *infoHash, *fileIdx)
		}
		seasonStr := fmt.Sprintf("Season %02d", *seasonNumber)
		targetFilename += filepath.Ext(sourcePath)
		targetDir = filepath.Join(MediaPath, TVShowsPath, mediaTitleStr, seasonStr)
	default:
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid record type")
	}
	var session *TorrentSession
	if infoHash != nil && *infoHash != "" {
		session, _ = GetTorrentSession(*infoHash)
	}
	if err := copyWithUpdateTorrentSession(sourcePath, filepath.Join(targetDir, targetFilename), session); err != nil {
		return nil, helpers.LogErrorWithMessage(err, "Failed to copy file")
	}
	videoMetadata, err := ProbeVideoFromURI(filepath.Join(targetDir, targetFilename))
	if err != nil {
		return nil, helpers.LogErrorWithMessage(err, "Failed to probe video + "+filepath.Join(targetDir, targetFilename))
	}
	mediaFile := database.MediaFile{
		Filepath:         filepath.Join(targetDir, targetFilename),
		OriginalFilename: filepath.Base(sourcePath),
		RecordID:         targetRecordID,
		SourceURI:        getMagnetURI(*infoHash, nil),
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

// Helper function to copy files from downloads -> media directory
// update the torrent session periodically in case copy takes time,
// so the torrent session isn't dropped and files deleted before copy is complete
func copyWithUpdateTorrentSession(src, dst string, session *TorrentSession) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return helpers.LogErrorWithMessage(err, "Failed to stat source file: "+src)
	}
	if !srcInfo.Mode().IsRegular() {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Source is not a regular file")
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return helpers.LogErrorWithMessage(err, "Failed to create destination directory")
	}
	// keep updating session in case copy takes time
	done := make(chan struct{})
	if session != nil {
		go func() {
			t := time.NewTicker(time.Second * 60)
			defer t.Stop()
			for {
				select {
				case <-t.C:
					session.LastUsed = time.Now()
				case <-done:
					return
				}
			}
		}()
	}
	defer close(done)
	_ = os.Remove(dst)
	// copy via hardlinks
	if err := os.Link(src, dst); err == nil {
		return nil
	}
	// fallback to regular copy
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func ProbeVideoFromURI(uri string) (*database.VideoMetadata, error) {
	rawOutput, err := services.FFProbe(uri)
	if err != nil {
		return nil, err
	}
	return simplifyMetadata(uri, rawOutput)
}

// helper convert ffprobe output to a metadata struct to store in db
func simplifyMetadata(uri string, raw *services.FfprobeOutput) (*database.VideoMetadata, error) {
	size, _ := strconv.ParseInt(raw.Format.Size, 10, 64)
	durationSeconds, _ := strconv.ParseFloat(raw.Format.Duration, 64)
	metadata := &database.VideoMetadata{
		Filename:           filepath.Base(uri),
		Filesize:           size,
		FileFormat:         raw.Format.FileFormat,
		FileFormatLongName: raw.Format.FileFormatLongName,
		Duration:           time.Duration(durationSeconds * float64(time.Second)),
		Bitrate:            raw.Format.Bitrate,
	}
	for _, rawStream := range raw.Streams {
		stream := database.Stream{
			CodecType:      rawStream.CodecType,
			CodecName:      rawStream.CodecName,
			CodecLongName:  rawStream.CodecLongName,
			Profile:        rawStream.Profile,
			Channels:       rawStream.Channels,
			ChannelLayout:  rawStream.ChannelLayout,
			PixelFormat:    rawStream.PixelFormat,
			ColorPrimaries: rawStream.ColorPrimaries,
			ColorTransfer:  rawStream.ColorTransfer,
			ColorSpace:     rawStream.ColorSpace,
			ColorRange:     rawStream.ColorRange,
		}
		// should be ISO-639-2 3 letter codes
		if lang, ok := rawStream.Tags["language"]; ok {
			stream.Language = lang
		}
		if title, ok := rawStream.Tags["title"]; ok {
			stream.Title = title
		}
		metadata.Streams = append(metadata.Streams, stream)
		if rawStream.CodecType == "video" {
			metadata.Width = rawStream.Width
			metadata.Height = rawStream.Height
			metadata.Framerate = rawStream.AvgFrameRate
		}
	}
	return metadata, nil
}
