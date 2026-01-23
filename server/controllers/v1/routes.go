package v1

import (
	"hound/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.Use(middlewares.CORSMiddleware)

	// public routes, registration and login
	publicRoutes := r.Group("/api/v1")
	publicRoutes.POST("/auth/register", RegistrationHandler)
	publicRoutes.POST("/auth/login", LoginHandler)

	// private routes, auth required, everything else
	privateRoutes := r.Group("/api/v1")
	privateRoutes.Use(middlewares.JWTMiddleware)

	/*
		General Routes
	*/
	privateRoutes.GET("/search", GeneralSearchHandler)
	privateRoutes.GET("/backdrops", GetMediaBackdrops)
	privateRoutes.GET("/continue_watching", GetContinueWatchingHandler)
	privateRoutes.GET("/watch_stats", GetWatchStatsHandler)
	privateRoutes.POST("/collection/:id", AddToCollectionHandler)
	privateRoutes.GET("/collection/:id", GetCollectionContentsHandler)
	privateRoutes.GET("/collection/recent", GetRecentCollectionContentsHandler)
	privateRoutes.DELETE("/collection/:id", DeleteFromCollectionHandler)
	privateRoutes.GET("/collection/all", GetUserCollectionsHandler)
	privateRoutes.POST("/collection/new", CreateCollectionHandler)          // add new collection
	privateRoutes.DELETE("/collection/delete/:id", DeleteCollectionHandler) // delete whole collection
	privateRoutes.DELETE("/comments", DeleteCommentHandler)                 // ?ids=23,52,43 (batch deletion)
	privateRoutes.DELETE("/comments/:id", DeleteCommentHandler)             // single deletion

	/*
		Watch History Routes
	*/
	privateRoutes.GET("/tv/:id/history", GetWatchHistoryHandler)
	privateRoutes.POST("/tv/:id/history", AddWatchHistoryTVShowHandler)
	privateRoutes.POST("/tv/:id/history/delete", DeleteWatchHistoryHandler) // batch deletion, we send a body so use POST which is more defined
	privateRoutes.GET("/tv/:id/season/:seasonNumber/history", GetWatchHistoryHandler)
	privateRoutes.POST("/tv/:id/history/rewatch", AddTVShowRewatchHandler) // we only want multiple rewatches for tv shows
	privateRoutes.GET("/movie/:id/history", GetWatchHistoryHandler)        // shared function w/ tv show history
	privateRoutes.POST("/movie/:id/history", AddWatchHistoryMovieHandler)
	privateRoutes.POST("/movie/:id/history/delete", DeleteWatchHistoryHandler)
	privateRoutes.GET("/watch_history/activity", GetWatchActivityHandler) // returns user watch activity between two dates

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
	privateRoutes.GET("/tv/trending", GetTrendingTVShowsHandler)
	privateRoutes.GET("/tv/:id", GetTVShowFromIDHandler)
	privateRoutes.GET("/tv/:id/season/:seasonNumber", GetTVSeasonHandler)
	privateRoutes.GET("/tv/:id/episode_groups", GetTVEpisodeGroupsHandler)
	privateRoutes.GET("/tv/:id/comments", GetCommentsHandler)
	privateRoutes.POST("/tv/:id/comments", PostCommentHandler)
	privateRoutes.GET("/tv/:id/continue_watching", GetNextWatchActionHandler)

	/*
		Movies Routes
	*/
	privateRoutes.GET("/movie/search", SearchMoviesHandler)
	privateRoutes.GET("/movie/trending", GetTrendingMoviesHandler)
	privateRoutes.GET("/movie/:id", GetMovieFromIDHandler)
	privateRoutes.POST("/movie/:id/comments", PostCommentHandler)
	privateRoutes.GET("/movie/:id/comments", GetCommentsHandler)
	privateRoutes.GET("/movie/:id/continue_watching", GetNextWatchActionHandler)

	/*
		Games Routes - games are being deprecated
	*/
	// privateRoutes.GET("/game/search", SearchGamesHandler)
	// privateRoutes.GET("/game/:id", GetGameFromIDHandler)
	// privateRoutes.POST("/game/:id/comments", PostCommentHandler)
	// privateRoutes.GET("/game/:id/comments", GetCommentsHandler)

	/*
		Video Streaming, Downloads Routes
	*/
	publicRoutes.GET("/stream/:encodedString", StreamHandler)
	privateRoutes.POST("/torrent/:encodedString", AddTorrentHandler)
	privateRoutes.POST("/torrent/:encodedString/download", DownloadHandler) // downloads to the server, not the client
	privateRoutes.GET("/media/downloads", GetDownloadsHandler)
	privateRoutes.POST("/media/downloads/:taskID/cancel", CancelDownloadHandler)

	/*
		Query Providers Routes
	*/
	privateRoutes.GET("/movie/:id/providers", SearchProvidersMovieHandler)
	privateRoutes.GET("/tv/:id/providers", SearchProvidersTVShowsHandler)
	privateRoutes.GET("/movie/:id/media_files", SearchMovieMediaFilesHandler)
	privateRoutes.GET("/tv/:id/media_files", SearchTVShowMediaFilesHandler)

	/*
		Media Routes
	*/
	privateRoutes.GET("/media/files", GetMediaFilesHandler) // list all downloaded media files in hound
	privateRoutes.POST("/media/ingest", IngestFileHandler)

	/*
		Metadata Routes
	*/
	privateRoutes.GET("/media/metadata", GetMetadataHandler)

	/*
		Testing purposes only
	*/
	privateRoutes.GET("/decode", DecodeTestHandler)
	privateRoutes.GET("/clearcache", ClearCacheHandler)
	privateRoutes.GET("/tv/:id/episodes", GetTVEpisodesHandler)
}
