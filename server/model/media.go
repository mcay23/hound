package model

import (
	"errors"
	"fmt"
	"hound/helpers"
	"hound/model/database"
	"os"
	"path/filepath"
)

const (
	MediaPath     = "Media"
	MoviesPath    = "Movies"
	TVShowsPath   = "TV Shows"
	DownloadsPath = "Downloads"
)

/*
media deals with downloading files, creating file systems, and probing data
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
}

/*
IngestFile ingests downloaded file into the media directory
and add it to the database
*/
func IngestFile(mediaRecord *database.MediaRecord, streamDetails *StreamObjectFull, path string) error {
	if mediaRecord == nil {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Nil media record passed to IngestFile()")
	}
	targetPath := ""
	if mediaRecord.RecordType == database.RecordTypeMovie {
		targetPath = filepath.Join(MediaPath, MoviesPath)
	} else if mediaRecord.RecordType == database.RecordTypeTVShow {
		if streamDetails.SeasonNumber == nil || streamDetails.EpisodeNumber == nil {
			return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"Season number or episode number is nil")
		}
		seasonStr := "Specials"
		if *streamDetails.SeasonNumber > 0 {
			seasonStr = fmt.Sprintf("Season %d", *streamDetails.SeasonNumber)
		}
		mediaTitleStr := fmt.Sprintf("%s (%s)", mediaRecord.MediaTitle, mediaRecord.ReleaseDate[0:4])
		targetPath = filepath.Join(MediaPath, TVShowsPath, mediaTitleStr, seasonStr)
	}
	os.MkdirAll(targetPath, 0755)
	return nil
}
