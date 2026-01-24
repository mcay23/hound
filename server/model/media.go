package model

import (
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/services"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	// all hound data should live in this folder
	// downloads and media are subdirectories of this folder, so move
	// between downloads and media should be fast
	dataDir      = "Hound Data"
	mediaDir     = "Media"
	downloadsDir = "Downloads"
)

var (
	HoundMoviesPath        = filepath.Join(dataDir, mediaDir, "Movies")
	HoundTVShowsPath       = filepath.Join(dataDir, mediaDir, "TV Shows")
	HoundP2PDownloadsPath  = filepath.Join(dataDir, downloadsDir, "p2p")
	HoundHttpDownloadsPath = filepath.Join(dataDir, downloadsDir, "http")
)

/*
media deals with file ingestion pipeline download->create files->process metadata...etc.
*/
func InitializeMedia() {
	// create media directories
	err := os.MkdirAll(HoundMoviesPath, 0755)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to create media directory")
		panic(fmt.Errorf("fatal error creating media directory %w", err))
	}
	err = os.MkdirAll(HoundTVShowsPath, 0755)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to create media directory")
		panic(fmt.Errorf("fatal error creating media directory %w", err))
	}
	err = os.MkdirAll(HoundP2PDownloadsPath, 0755)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to create p2p downloads directory")
		panic(fmt.Errorf("fatal error creating p2p downloads directory %w", err))
	}
	err = os.MkdirAll(HoundHttpDownloadsPath, 0755)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to create http downloads directory")
		panic(fmt.Errorf("fatal error creating http downloads directory %w", err))
	}
}

func DeleteMediaFile(fileID int) error {
	// delete file first
	file, err := database.GetMediaFile(fileID)
	if err != nil {
		return err
	}
	err = os.Remove(file.Filepath)
	if err != nil {
		// if file doesn't exist, continue to delete mediafile record
		if !os.IsNotExist(err) {
			return err
		} else {
			slog.Info("File doesn't exist in dir, deleting media_file db record", "filepath", file.Filepath)
		}
	}
	err = database.DeleteMediaFileRecord(fileID)
	if err != nil {
		return helpers.LogErrorWithMessage(err, "Failed to delete media_file db record")
	}
	slog.Info("Deleted media_file record", "file_id", fileID, "filepath", file.Filepath)
	return nil
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
