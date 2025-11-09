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
	privateRoutes.POST("/collection/new", CreateCollectionHandler)
	privateRoutes.DELETE("/collection/delete/:id", DeleteCollectionHandler)
	privateRoutes.DELETE("/comments", DeleteCommentHandler)     // ?ids=23,52,43 (batch deletion)
	privateRoutes.DELETE("/comments/:id", DeleteCommentHandler) // single deletion

	/*
		TV Show Routes
	*/
	privateRoutes.GET("/tv/search", SearchTVShowHandler)
	privateRoutes.GET("/tv/trending", GetTrendingTVShowsHandler)
	privateRoutes.GET("/tv/:id", GetTVShowFromIDHandler)
	privateRoutes.GET("/tv/:id/season/:seasonNumber", GetTVSeasonHandler)
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
		Video Streaming Routes
	*/
	publicRoutes.GET("/stream/:encodedString", StreamHandler)
	privateRoutes.POST("/torrent/:encodedString", AddTorrentHandler)
	//privateRoutes.GET("/tv/:id/stream/:encodedString", StreamHandler)

	/*
		Query Providers Routes
	*/
	privateRoutes.GET("/movie/:id/providers", SearchProvidersHandler)
	privateRoutes.GET("/tv/:id/providers", SearchProvidersHandler)

	/*
		Testing purposes only
	*/
	privateRoutes.GET("/decode", DecodeTestHandler)
	privateRoutes.GET("/clearcache", ClearCacheHandler)
}
