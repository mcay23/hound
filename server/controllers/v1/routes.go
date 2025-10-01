package v1

import (
	"github.com/gin-gonic/gin"
	"hound/middlewares"
)

func SetupRoutes(r *gin.Engine) {
	r.Use(middlewares.CORSMiddleware)
	// public routes, registration and login
	publicRoutes := r.Group("/api/v1/auth")
	publicRoutes.POST("/register", RegistrationHandler)
	publicRoutes.POST("/login", LoginHandler)

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
	privateRoutes.DELETE("/comments", DeleteCommentHandler)

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
	privateRoutes.GET("movie/:id/stream", StreamHandler)
	privateRoutes.GET("tv/:id/stream", StreamHandler)

	/*
		Query Providers Routes
	 */
	privateRoutes.GET("/providers", SearchProvidersHandler)
}
