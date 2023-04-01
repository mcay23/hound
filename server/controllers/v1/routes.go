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
	/*
		TV Show Routes
	 */
	privateRoutes.GET("/tv/search", SearchTVShowHandler)
	privateRoutes.GET("/tv/trending", GetTrendingTVShowsHandler)
	privateRoutes.GET("/tv/:id", GetTVShowHandlerFromID)
	privateRoutes.POST("/tv/lib", AddTVShowToLibraryHandler)
	privateRoutes.GET("/tv/lib", GetUserTVShowLibraryHandler)
	privateRoutes.GET("/tv/:id/season/:seasonNumber", GetTVSeasonHandler)
	/*
		Movies Routes
	 */
	privateRoutes.GET("/movie/search", SearchMoviesHandler)
	privateRoutes.GET("/movie/trending", GetTrendingMoviesHandler)
	privateRoutes.GET("/movie/:id", GetMovieFromIDHandler)

	//err := database.CreateCollection(database.CollectionRecord{
	//	CollectionTitle: "my new collection",
	//	Description:     []byte("this is a description"),
	//	OwnerID:         1,
	//	IsPrimary:       true,
	//	IsPublic:        true,
	//	Tags:            nil,
	//	ThumbnailURL:    nil,
	//})
	//if err != nil {
	//	helpers.LogErrorWithMessage(err, "err")
	//}
	//id := int64(1)
	//rec, num, err := database.SearchCollection(database.CollectionRecordQuery{
	//	CollectionID: &id,
	//	OwnerID:      nil,
	//	IsPrimary:    nil,
	//	IsPublic:     nil,
	//	Tags:         nil,
	//}, 10, 0)
	//if err != nil {
	//	helpers.LogErrorWithMessage(err, "err")
	//} else {
	//	fmt.Println(rec, num)
	//}
}
