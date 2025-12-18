package model

import (
	"fmt"
	"hound/helpers"
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
