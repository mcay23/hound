package internal

import "path/filepath"

const (
	// all hound data should live in this folder
	// downloads and media are subdirectories of this folder, so move
	// between downloads and media should be fast
	dataDir      = "Hound Data"
	mediaDir     = "Media"
	downloadsDir = "Downloads"
)

var (
	HoundMoviesPath          = filepath.Join(dataDir, mediaDir, "Movies")
	HoundTVShowsPath         = filepath.Join(dataDir, mediaDir, "TV Shows")
	HoundP2PDownloadsPath    = filepath.Join(dataDir, downloadsDir, "p2p")
	HoundHttpDownloadsPath   = filepath.Join(dataDir, downloadsDir, "http")
	HoundIPTVDownloadsPath   = filepath.Join(dataDir, downloadsDir, "iptv")
	HoundExternalLibraryPath = filepath.Join("External Library")
)
