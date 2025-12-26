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
	privateRoutes.POST("/collection/:id", AddToCollectionHandler)
	privateRoutes.GET("/collection/:id", GetCollectionContentsHandler)
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

	privateRoutes.GET("/movie/:id/history", GetWatchHistoryHandler) // shared function w/ tv show history
	privateRoutes.POST("/movie/:id/history", AddWatchHistoryMovieHandler)
	privateRoutes.POST("/movie/:id/history/delete", DeleteWatchHistoryHandler)

	/*
		TV Show Routes
	*/
	privateRoutes.GET("/tv/search", SearchTVShowHandler)
	privateRoutes.GET("/tv/trending", GetTrendingTVShowsHandler)
	privateRoutes.GET("/tv/:id", GetTVShowFromIDHandler)
	privateRoutes.GET("/tv/:id/season/:seasonNumber", GetTVSeasonHandler)
	privateRoutes.GET("/tv/:id/episode_groups", GetTVEpisodeGroupsHandler)
	privateRoutes.GET("/tv/:id/episodes", GetTVEpisodesHandler)
	privateRoutes.GET("/tv/:id/comments", GetCommentsHandler)
	privateRoutes.POST("/tv/:id/comments", PostCommentHandler)

	/*
		Movies Routes
	*/
	privateRoutes.GET("/movie/search", SearchMoviesHandler)
	privateRoutes.GET("/movie/trending", GetTrendingMoviesHandler)
	privateRoutes.GET("/movie/:id", GetMovieFromIDHandler)
	privateRoutes.POST("/movie/:id/comments", PostCommentHandler)
	privateRoutes.GET("/movie/:id/comments", GetCommentsHandler)

	/*
		Games Routes
	*/
	privateRoutes.GET("/game/search", SearchGamesHandler)
	privateRoutes.GET("/game/:id", GetGameFromIDHandler)
	privateRoutes.POST("/game/:id/comments", PostCommentHandler)
	privateRoutes.GET("/game/:id/comments", GetCommentsHandler)

	/*
		Video Streaming, Downloads Routes
	*/
	publicRoutes.GET("/stream/:encodedString", StreamHandler)
	privateRoutes.POST("/torrent/:encodedString", AddTorrentHandler)
	privateRoutes.POST("/torrent/:encodedString/download", DownloadTorrentHandler) // downloads to the server, not the client
	privateRoutes.GET("/media/downloads", GetDownloadsHandler)

	/*
		Query Providers Routes
	*/
	privateRoutes.GET("/movie/:id/providers", SearchProvidersHandler)
	privateRoutes.GET("/tv/:id/providers", SearchProvidersHandler)

	/*
		Media Routes
	*/
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
}
