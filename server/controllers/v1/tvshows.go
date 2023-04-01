package v1

import (
	"errors"
	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/gin-gonic/gin"
	"hound/helpers"
	"hound/model/database"
	"hound/model/sources"
	"hound/view"
	"strconv"
	"strings"
)

type AddLibraryRequest struct {
	Source   string `json:"source" binding:"required,gt=0"`
	SourceID string `json:"source_id" binding:"required,gt=0"`
}

func SearchTVShowHandler(c *gin.Context) {
	queryString := c.Query("query")
	results, err := SearchTVShowCore(queryString)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to search for tv show")
		helpers.ErrorResponse(c, err)
		return
	}
	helpers.SuccessResponse(c, results, 200)
}

func GetTVShowFromIDHandler(c *gin.Context) {
	param := c.Param("id")
	split := strings.Split(param, "-")
	if len(split) != 2 {
		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
		return
	}
	id, err := strconv.ParseInt(split[1], 10, 64)
	// only accept tmdb ids for now
	if err != nil || split[0] != "tmdb" {
		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
		return
	}
	options := map[string]string{
		"append_to_response": "videos,watch/providers,credits,recommendations",
	}
	showDetails, err := sources.GetTVShowFromIDTMDB(int(id), options)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	// get profile, video urls
	for num, _ := range showDetails.Seasons {
		// this doesn't work, pointer stuff
		// item.PosterPath = tmdb.GetImageURL(item.PosterPath, tmdb.W500)
		showDetails.Seasons[num].PosterPath = GetTMDBImageURL(showDetails.Seasons[num].PosterPath, tmdb.W500)
	}
	for num, item := range showDetails.Credits.TVCredits.Cast {
		showDetails.Credits.TVCredits.Cast[num].ProfilePath = GetTMDBImageURL(item.ProfilePath, tmdb.W500)
	}
	for num, item := range showDetails.Credits.TVCredits.Crew {
		showDetails.Credits.TVCredits.Crew[num].ProfilePath = GetTMDBImageURL(item.ProfilePath, tmdb.W500)
	}

	var viewSeasons []view.SeasonObjectPartial
	for _, item := range showDetails.Seasons {
		viewSeasons = append(viewSeasons, view.SeasonObjectPartial{
			AirDate:      item.AirDate,
			EpisodeCount: item.EpisodeCount,
			ID:           item.ID,
			Name:         item.Name,
			Overview:     item.Overview,
			PosterURL:    GetTMDBImageURL(item.PosterPath, tmdb.W500),
			SeasonNumber: item.SeasonNumber,
		})
	}
	returnObject := view.TVShowFullObject{
		MediaSource:      sources.SourceTMDB,
		MediaType:        database.MediaTypeTVShow,
		OriginalName:     showDetails.OriginalName,
		SourceID:         showDetails.ID,
		MediaTitle:       showDetails.Name,
		VoteCount:        showDetails.VoteCount,
		VoteAverage:      showDetails.VoteAverage,
		PosterURL:        GetTMDBImageURL(showDetails.PosterPath, tmdb.W500),
		NumberOfEpisodes: showDetails.NumberOfEpisodes,
		NumberOfSeasons:  showDetails.NumberOfSeasons,
		Seasons:          viewSeasons,
		NextEpisodeToAir: showDetails.NextEpisodeToAir,
		Networks:         showDetails.Networks,
		EpisodeRunTime:   showDetails.EpisodeRunTime,
		CreatedBy:        showDetails.CreatedBy,
		Status:           showDetails.Status,
		FirstAirDate:     showDetails.FirstAirDate,
		Popularity:       showDetails.Popularity,
		Genres:           showDetails.Genres,
		OriginalLanguage: showDetails.OriginalLanguage,
		BackdropURL:      GetTMDBImageURL(showDetails.BackdropPath, tmdb.Original),
		Overview:         showDetails.Overview,
		OriginCountry:    showDetails.OriginCountry,
		Videos:           showDetails.Videos.TVVideos,
		WatchProviders:   showDetails.WatchProviders,
		TVCredits:        showDetails.Credits.TVCredits,
		Recommendations:  showDetails.Recommendations,
	}
	helpers.SuccessResponse(c, returnObject, 200)
}

func GetTrendingTVShowsHandler(c *gin.Context) {
	// pagination locked for now
	results, err := sources.GetTrendingTVShowsTMDB("1")
	//results2, err := sources.GetTrendingTVShowsTMDB("2")
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Error getting popular tv shows")
		helpers.ErrorResponse(c, err)
		return
	}
	//results.Results = append(results.Results, results2.Results...)
	// convert url results
	var viewArray []view.LibraryItem
	for _, item := range results.Results {
		genreArray := sources.GetGenresMap(item.GenreIDs, database.MediaTypeTVShow)
		thumbnailURL := GetTMDBImageURL(item.PosterPath, tmdb.W300)
		viewObject := view.LibraryItem{
			MediaType:    database.MediaTypeTVShow,
			MediaSource:  sources.SourceTMDB,
			SourceID:     strconv.Itoa(int(item.ID)),
			MediaTitle:   item.OriginalName,
			ReleaseDate:  item.FirstAirDate,
			Description:  item.Overview,
			ThumbnailURL: &thumbnailURL,
			Tags:         genreArray,
			UserTags:     nil,
		}
		viewArray = append(viewArray, viewObject)
	}
	helpers.SuccessResponse(c, viewArray, 200)
}

func AddTVShowToLibraryHandler(c *gin.Context) {
	username := c.GetHeader("X-Username")
	if username == "" {
		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
		return
	}
	userPayload := AddLibraryRequest{}
	if err := c.ShouldBindJSON(&userPayload); err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to bind registration body")
		helpers.ErrorResponse(c, err)
		return
	}
	if userPayload.Source != sources.SourceTMDB {
		_ = helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Only tmdb is supported at this time")
		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
		return
	}
	tmdbID, err := strconv.Atoi(userPayload.SourceID)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to convert string to tmdb id (int)")
		helpers.ErrorResponse(c, err)
		return
	}
	err = sources.AddTVShowToCollectionTMDB(username, userPayload.Source, tmdbID, nil)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to add tv show to library")
		helpers.ErrorResponse(c, err)
		return
	}
	helpers.SuccessResponse(c, gin.H{"status": "success"}, 200)
}

func GetUserTVShowLibraryHandler(c *gin.Context) {
	username := c.GetHeader("X-Username")
	limitQuery := c.Query("limit")
	offsetQuery := c.Query("offset")
	limit := 0
	offset := 0
	if limitQuery != "" && offsetQuery != "" {
		var err error
		limit, err = strconv.Atoi(limitQuery)
		if err != nil {
			_ = helpers.LogErrorWithMessage(err, "Invalid limit query param")
			helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
			return
		}
		offset, err = strconv.Atoi(offsetQuery)
		if err != nil {
			_ = helpers.LogErrorWithMessage(err, "Invalid offset query param")
			helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
			return
		}
	}
	if username == "" {
		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
		return
	}
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Error retrieving user_id from username")
		helpers.ErrorResponse(c, err)
		return
	}
	records, totalRecords, err := database.GetCollectionFromLibrary(userID, database.MediaTypeTVShow, nil, limit, offset)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Error retrieving user library")
		helpers.ErrorResponse(c, err)
		return
	}
	var viewArray []view.LibraryItem
	for _, item := range records {
		viewObject := view.LibraryItem{
			MediaType:    item.MediaType,
			MediaSource:  item.MediaSource,
			SourceID:     item.SourceID,
			MediaTitle:   item.MediaTitle,
			ReleaseDate:  item.ReleaseDate,
			Description:  string(item.Description),
			ThumbnailURL: item.ThumbnailURL,
			Tags:         item.Tags,
			UserTags:     item.UserTags,
		}
		viewArray = append(viewArray, viewObject)
	}
	returnObject := view.LibraryView{Results: &viewArray, TotalRecords: totalRecords, Limit: limit, Offset: offset}
	helpers.SuccessResponse(c, returnObject, 200)
}

func GetTVSeasonHandler(c *gin.Context) {
	seasonNumber, err := strconv.Atoi(c.Param("seasonNumber"))
	if err != nil {
		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
		return
	}
	idParam := c.Param("id")
	// idParam is in format tmdb-1234, grab id number
	split := strings.Split(idParam, "-")
	if len(split) != 2 {
		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
		return
	}
	sourceID, err := strconv.Atoi(split[1])
	// only accept tmdb ids for now
	if err != nil || split[0] != "tmdb" {
		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
		return
	}
	tvSeason, err := sources.GetTVSeasonTMDB(sourceID, seasonNumber, nil)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	// overwrite paths
	for num, item := range tvSeason.Episodes {
		tvSeason.Episodes[num].StillPath = GetTMDBImageURL(item.StillPath, tmdb.W500)
	}
	tvSeason.PosterPath = GetTMDBImageURL(tvSeason.PosterPath, tmdb.W500)
	response := view.TVSeasonResponseObject{
		MediaSource: sources.SourceTMDB,
		SourceID:    int64(sourceID),
		SeasonData:  tvSeason,
	}
	helpers.SuccessResponse(c, response, 200)
}

func GetTMDBImageURL(path string, size string) string {
	if path == "" {
		return ""
	}
	return tmdb.GetImageURL(path, size)
}

func SearchTVShowCore(queryString string) (*[]view.TMDBSearchResultObject, error){
	results, err := sources.SearchTVShowTMDB(queryString)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Error searching for tv show")
		return nil, err
	}
	// convert url results
	var convertedResults []view.TMDBSearchResultObject
	for _, item := range results.Results {
		genreArray := sources.GetGenresMap(item.GenreIDs, database.MediaTypeTVShow)
		resultObject := view.TMDBSearchResultObject{
			MediaSource:      sources.SourceTMDB,
			MediaType:        database.MediaTypeTVShow,
			OriginalName:     item.OriginalName,
			SourceID:         item.ID,
			MediaTitle:       item.Name,
			VoteCount:        item.VoteCount,
			VoteAverage:      item.VoteAverage,
			PosterURL:        GetTMDBImageURL(item.PosterPath, tmdb.W300),
			FirstAirDate:     item.FirstAirDate,
			Popularity:       item.Popularity,
			Genres:           genreArray,
			OriginalLanguage: item.OriginalLanguage,
			BackdropURL:      GetTMDBImageURL(item.BackdropPath, tmdb.Original),
			Overview:         item.Overview,
			OriginCountry:    item.OriginCountry,
		}
		convertedResults = append(convertedResults, resultObject)
	}
	return &convertedResults, nil
}