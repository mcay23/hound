package v1

import (
	"os"

	"github.com/mcay23/hound/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.Use(middlewares.CORSMiddleware)

	// public routes and login
	publicRoutes := r.Group("/api/v1")
	publicRoutes.POST("/auth/login", LoginHandler)

	// private routes, auth required, everything else
	privateRoutes := r.Group("/api/v1")
	privateRoutes.Use(middlewares.AuthMiddleware)

	// admin routes, admin only apis
	adminRoutes := r.Group("/api/v1")
	adminRoutes.Use(middlewares.AuthMiddleware)
	adminRoutes.Use(middlewares.AdminMiddleware)

	/*
		Users Routes
	*/
	adminRoutes.GET("/users", GetUsersHandler)
	adminRoutes.DELETE("/users/:id", DeleteUserHandler)
	adminRoutes.POST("/users", RegistrationHandler)

	/*
		API Keys
	*/
	privateRoutes.GET("/api_keys", GetUserAPIKeysHandler)
	privateRoutes.POST("/api_keys", CreateAPIKeyHandler)
	privateRoutes.DELETE("/api_keys/:id", RevokeAPIKeyHandler)

	/*
		General Routes
	*/
	privateRoutes.GET("/search", GeneralSearchHandler)
	privateRoutes.GET("/backdrop", GetMediaBackdrops)
	privateRoutes.GET("/continue_watching", GetContinueWatchingHandler)
	privateRoutes.GET("/watch_stats", GetWatchStatsHandler)
	privateRoutes.GET("/build_info", GetBuildInfoHandler)

	/*
		Catalog Routes
	*/
	privateRoutes.GET("/catalog/:id", GetCatalogHandler)

	/*
		Collection Routes
	*/
	privateRoutes.GET("/collection/:id", GetCollectionContentsHandler)
	privateRoutes.POST("/collection/:id", AddToCollectionHandler)
	privateRoutes.GET("/collection/recent", GetRecentCollectionContentsHandler)
	privateRoutes.GET("/collection/hound-library", GetHoundLibraryHandler)
	privateRoutes.DELETE("/collection/:id/delete", DeleteCollectionHandler) // delete whole collection
	privateRoutes.DELETE("/collection/:id", DeleteFromCollectionHandler)
	privateRoutes.GET("/collection/all", GetUserCollectionsHandler)
	privateRoutes.POST("/collection/new", CreateCollectionHandler) // add new collection

	/*
		Watch History Routes
	*/
	privateRoutes.GET("/tv/:id/history", GetWatchHistoryTVHandler)
	privateRoutes.POST("/tv/:id/history", AddWatchHistoryTVHandler)
	privateRoutes.GET("/tv/:id/season/:seasonNumber/history", GetWatchHistoryTVHandler)
	privateRoutes.POST("/tv/:id/history/rewatch", AddTVShowRewatchHandler)    // we only want multiple rewatches for tv shows
	privateRoutes.POST("/tv/:id/history/delete", DeleteWatchHistoryTVHandler) // batch deletion, we send a body so use POST which is more defined

	privateRoutes.GET("/movie/:id/history", GetWatchHistoryMovieHandler) // shared function w/ tv show history
	privateRoutes.POST("/movie/:id/history", AddWatchHistoryMovieHandler)
	privateRoutes.POST("/movie/:id/history/delete", DeleteWatchHistoryMovieHandler)
	privateRoutes.GET("/watch_activity", GetWatchActivityHandler) // returns user watch activity between two dates

	/*
		Playback Progress Routes
	*/
	privateRoutes.GET("/movie/:id/playback", GetPlaybackProgressHandler)
	privateRoutes.POST("/movie/:id/playback", SetPlaybackProgressHandler)
	privateRoutes.POST("/movie/:id/playback/delete", DeletePlaybackProgressHandler)
	privateRoutes.GET("/tv/:id/season/:seasonNumber/playback", GetPlaybackProgressHandler)
	privateRoutes.POST("/tv/:id/playback", SetPlaybackProgressHandler)
	privateRoutes.POST("/tv/:id/playback/delete", DeletePlaybackProgressHandler)

	/*
		TV Show Routes
	*/
	privateRoutes.GET("/tv/search", SearchTVShowHandler)
	privateRoutes.GET("/tv/:id", GetTVShowFromIDHandler)
	privateRoutes.GET("/tv/:id/season/:seasonNumber", GetTVSeasonHandler)
	privateRoutes.GET("/tv/:id/episode_groups", GetTVEpisodeGroupsHandler)
	privateRoutes.GET("/tv/:id/continue_watching", GetNextWatchActionHandler)

	/*
		Movies Routes
	*/
	privateRoutes.GET("/movie/search", SearchMoviesHandler)
	privateRoutes.GET("/movie/:id", GetMovieFromIDHandler)

	privateRoutes.GET("/movie/:id/continue_watching", GetNextWatchActionHandler)

	/*
		Comments
	*/
	privateRoutes.GET("/tv/:id/comments", GetCommentsTVHandler)
	privateRoutes.POST("/tv/:id/comments", PostCommentTVHandler)
	privateRoutes.GET("/movie/:id/comments", GetCommentsMovieHandler)
	privateRoutes.POST("/movie/:id/comments", PostCommentMovieHandler)
	privateRoutes.DELETE("/comments/:id", DeleteCommentHandler) // single deletion

	/*
		Video Streaming, Downloads Routes
	*/
	publicRoutes.GET("/stream/:encodedString", StreamHandler)
	privateRoutes.POST("/torrent/:encodedString", AddTorrentHandler)
	privateRoutes.POST("/download/:encodedString", DownloadHandler)                      // downloads to the server, not the client
	privateRoutes.POST("/tv/:id/season/:seasonNumber/download", DownloadTVSeasonHandler) // downloads a whole season
	privateRoutes.GET("/ingest", GetIngestTasksHandler)
	privateRoutes.POST("/ingest/:taskID/cancel", CancelIngestTaskHandler)

	/*
		Provider Profiles
	*/
	privateRoutes.GET("/provider_profiles", GetProviderProfilesHandler)
	adminRoutes.POST("/provider_profiles", CreateProviderProfileHandler)
	adminRoutes.DELETE("/provider_profiles/:id", DeleteProviderProfileHandler)
	adminRoutes.PUT("/provider_profiles/:id", UpdateProviderProfileHandler)

	/*
		Query Providers Routes
	*/
	privateRoutes.GET("/movie/:id/providers", SearchProvidersMovieHandler)
	privateRoutes.GET("/tv/:id/providers", SearchProvidersTVHandler)
	privateRoutes.GET("/movie/:id/media_files", GetMovieMediaFilesHandler)
	privateRoutes.GET("/tv/:id/media_files", GetTVShowMediaFilesHandler)

	/*
		Genres Routes
	*/
	privateRoutes.GET("/tv/genres", GetTVGenresHandler)
	privateRoutes.GET("/movie/genres", GetMovieGenresHandler)

	/*
		Media Routes
	*/
	privateRoutes.GET("/media_files", GetMediaFilesHandler) // list all downloaded media files in hound
	privateRoutes.DELETE("/media_files/:id", DeleteMediaFileHandler)

	/*
		Testing purposes only
	*/
	if os.Getenv("APP_ENV") != "production" {
		privateRoutes.GET("/decode", DecodeTestHandler)
		privateRoutes.GET("/clearcache", ClearCacheHandler)
		privateRoutes.GET("/tv/:id/episodes", GetTVEpisodesHandler)
		privateRoutes.GET("/media_files/metadata", GetMetadataHandler)
		privateRoutes.POST("/ingest", IngestFileHandler)
	}
}
