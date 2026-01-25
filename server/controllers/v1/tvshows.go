package v1

import (
	"errors"
	"hound/database"
	"hound/helpers"
	"hound/sources"
	"hound/view"
	"strconv"
	"strings"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/gin-gonic/gin"
)

func SearchTVShowHandler(c *gin.Context) {
	queryString := c.Query("query")
	results, err := SearchTVShowCore(queryString)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to search for tv show"))
		return
	}
	helpers.SuccessResponse(c, results, 200)
}

func GetTVShowFromIDHandler(c *gin.Context) {
	mediaSource, sourceID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "request id param invalid"+err.Error()))
		return
	}
	showDetails, err := sources.GetTVShowFromIDTMDB(sourceID)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	// get profile, video urls
	for num := range showDetails.Seasons {
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
	logoURL := ""
	if len(showDetails.Images.Logos) > 0 {
		logoURL = GetTMDBImageURL(showDetails.Images.Logos[0].FilePath, tmdb.W500)
	}
	returnObject := view.TVShowFullObject{
		MediaSource:      sources.MediaSourceTMDB,
		MediaType:        database.MediaTypeTVShow,
		OriginalName:     showDetails.OriginalName,
		SourceID:         showDetails.ID,
		MediaTitle:       showDetails.Name,
		VoteCount:        showDetails.VoteCount,
		VoteAverage:      showDetails.VoteAverage,
		PosterURL:        GetTMDBImageURL(showDetails.PosterPath, tmdb.W500),
		LogoURL:          logoURL,
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
		Videos:           showDetails.Videos,
		WatchProviders:   showDetails.WatchProviders,
		TVCredits:        showDetails.Credits.TVCredits,
		Recommendations:  showDetails.Recommendations,
		ExternalIDs:      showDetails.TVExternalIDs,
	}
	_, record, err := database.GetMediaRecord(database.MediaTypeTVShow, sources.MediaSourceTMDB, strconv.Itoa(int(showDetails.ID)))
	if err == nil {
		commentType := c.Query("type")
		comments, err := GetCommentsCore(c.GetHeader("X-Username"), record.RecordID, &commentType)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Error retrieving comments"))
			return
		}
		returnObject.Comments = comments
	}
	helpers.SuccessResponse(c, returnObject, 200)
}

//func GetUserTVShowLibraryHandler(c *gin.Context) {
//	username := c.GetHeader("X-Username")
//	limitQuery := c.Query("limit")
//	offsetQuery := c.Query("offset")
//	limit := 0
//	offset := 0
//	if limitQuery != "" && offsetQuery != "" {
//		var err error
//		limit, err = strconv.Atoi(limitQuery)
//		if err != nil {
//			_ = helpers.LogErrorWithMessage(err, "Invalid limit query param")
//			helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
//			return
//		}
//		offset, err = strconv.Atoi(offsetQuery)
//		if err != nil {
//			_ = helpers.LogErrorWithMessage(err, "Invalid offset query param")
//			helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
//			return
//		}
//	}
//	if username == "" {
//		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
//		return
//	}
//	userID, err := database.GetUserIDFromUsername(username)
//	if err != nil {
//		helpers.ErrorResponse(c, err)
//		return
//	}
//	records, totalRecords, err := database.GetCollectionRecords(userID, 3, limit, offset)
//	if err != nil {
//		_ = helpers.LogErrorWithMessage(err, "Error retrieving user library")
//		helpers.ErrorResponse(c, err)
//		return
//	}
//	var viewArray []view.LibraryObject
//	for _, item := range records {
//		viewObject := view.LibraryObject{
//			MediaType:    item.MediaType,
//			MediaSource:  item.MediaSource,
//			SourceID:     item.SourceID,
//			MediaTitle:   item.MediaTitle,
//			ReleaseDate:  item.ReleaseDate,
//			Description:  string(item.Description),
//			ThumbnailURL: item.ThumbnailURL,
//			Tags:         item.Tags,
//			UserTags:     item.UserTags,
//		}
//		viewArray = append(viewArray, viewObject)
//	}
//	returnObject := view.CollectionContentsView{Results: &viewArray, TotalRecords: totalRecords, Limit: limit, Offset: offset}
//	helpers.SuccessResponse(c, returnObject, 200)
//}

func GetTVSeasonHandler(c *gin.Context) {
	seasonNumber, err := strconv.Atoi(c.Param("seasonNumber"))
	if err != nil {
		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
		return
	}
	mediaSource, sourceID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "request id param invalid"+err.Error()))
		return
	}
	tvSeason, err := sources.GetTVSeasonTMDB(sourceID, seasonNumber)
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
		MediaSource: sources.MediaSourceTMDB,
		SourceID:    int64(sourceID),
		SeasonData:  tvSeason,
	}
	has, record, err := database.GetMediaRecord(database.MediaTypeTVShow, sources.MediaSourceTMDB, strconv.Itoa(sourceID))
	// if record id exists, retrieve watch history
	if err == nil && has {
		commentType := "history"
		comments, err := GetCommentsCore(c.GetHeader("X-Username"), record.RecordID, &commentType)
		if err != nil {
			helpers.ErrorResponse(c, err)
			return
		}
		var filteredComments []view.CommentObject
		for _, item := range *comments {
			if strings.HasPrefix(item.TagData, "S"+strconv.Itoa(seasonNumber)) {
				filteredComments = append(filteredComments, item)
			}
		}
		response.SeasonWatchInfo = &filteredComments
	}
	helpers.SuccessResponse(c, response, 200)
}

func GetTVEpisodeGroupsHandler(c *gin.Context) {
	mediaSource, sourceID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"request id param invalid"+err.Error()))
		return
	}
	episodeGroups, err := sources.GetTVEpisodeGroupsTMDB(sourceID)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	helpers.SuccessResponse(c, episodeGroups.Results, 200)
}

func GetTMDBImageURL(path string, size string) string {
	if path == "" {
		return ""
	}
	return tmdb.GetImageURL(path, size)
}

func SearchTVShowCore(queryString string) (*[]view.MediaCatalogObject, error) {
	results, err := sources.SearchTVShowTMDB(queryString)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Error searching for tv show")
		return nil, err
	}
	// convert url results
	var convertedResults []view.MediaCatalogObject
	for _, item := range results.Results {
		genreArray := sources.GetGenresMap(item.GenreIDs, database.MediaTypeTVShow)
		resultObject := view.MediaCatalogObject{
			MediaSource:      sources.MediaSourceTMDB,
			MediaType:        database.MediaTypeTVShow,
			OriginalName:     item.OriginalName,
			SourceID:         strconv.Itoa(int(item.ID)),
			MediaTitle:       item.Name,
			VoteCount:        item.VoteCount,
			VoteAverage:      item.VoteAverage,
			ThumbnailURL:     GetTMDBImageURL(item.PosterPath, tmdb.W300),
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
