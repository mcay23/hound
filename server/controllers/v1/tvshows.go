package v1

import (
	"errors"
	"hound/database"
	"hound/helpers"
	"hound/model"
	"hound/sources"
	"hound/view"
	"strconv"
	"strings"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/gin-gonic/gin"
)

func SearchTVShowHandler(c *gin.Context) {
	queryString := c.Query("query")
	results, err := model.SearchTVShows(queryString)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to search for tv show"))
		return
	}
	helpers.SuccessResponse(c, results, 200)
}

// returns a tmdb-like response but with media record catalog structure
func GetTVShowFromIDHandlerV2(c *gin.Context) {
	mediaSource, showID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "request id param invalid"+err.Error()))
		return
	}
	showDetails, err := sources.GetTVShowFromIDTMDB(showID)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	// create top level show
	duration := 0
	if len(showDetails.EpisodeRunTime) > 0 {
		duration = showDetails.EpisodeRunTime[0]
	}
	genreArray := []database.GenreObject{}
	for _, genre := range showDetails.Genres {
		genreArray = append(genreArray, database.GenreObject{
			ID:   genre.ID,
			Name: genre.Name,
		})
	}
	logoURL := ""
	if len(showDetails.Images.Logos) > 0 {
		logoURL = helpers.GetTMDBImageURL(showDetails.Images.Logos[0].FilePath, tmdb.W500)
	}
	showObject := view.TVShowCatalogObject{
		MediaRecordCatalog: view.MediaRecordCatalog{
			MediaSource:      sources.MediaSourceTMDB,
			MediaType:        database.RecordTypeTVShow,
			SourceID:         strconv.Itoa(int(showID)),
			MediaTitle:       showDetails.Name,
			OriginalTitle:    showDetails.OriginalName,
			VoteCount:        showDetails.VoteCount,
			VoteAverage:      showDetails.VoteAverage,
			Popularity:       showDetails.Popularity,
			ThumbnailURI:     helpers.GetTMDBImageURL(showDetails.PosterPath, tmdb.W500),
			SeasonCount:      &showDetails.NumberOfSeasons,
			EpisodeCount:     &showDetails.NumberOfEpisodes,
			LastAirDate:      showDetails.LastAirDate,
			NextAirDate:      showDetails.NextEpisodeToAir.AirDate,
			ReleaseDate:      showDetails.FirstAirDate,
			Duration:         duration,
			Status:           showDetails.Status,
			Genres:           genreArray,
			OriginalLanguage: showDetails.OriginalLanguage,
			BackdropURI:      helpers.GetTMDBImageURL(showDetails.BackdropPath, tmdb.Original),
			LogoURI:          logoURL,
			Overview:         showDetails.Overview,
			OriginCountry:    showDetails.OriginCountry,
		},
	}
	// append top 20 cast members
	castArray := []view.Credit{}
	for idx, cast := range showDetails.Credits.TVCredits.Cast {
		castArray = append(castArray, view.Credit{
			MediaSource:  sources.MediaSourceTMDB,
			SourceID:     strconv.Itoa(int(cast.ID)),
			CreditID:     cast.CreditID,
			Name:         cast.Name,
			OriginalName: cast.OriginalName,
			Character:    &cast.Character,
			ProfileURI:   helpers.GetTMDBImageURL(cast.ProfilePath, tmdb.W500),
			Job:          "Cast",
		})
		if idx == 20 {
			break
		}
	}
	showObject.Cast = &castArray
	creatorsArray := []view.Credit{}
	for _, creator := range showDetails.CreatedBy {
		creatorsArray = append(creatorsArray, view.Credit{
			MediaSource:  sources.MediaSourceTMDB,
			SourceID:     strconv.Itoa(int(creator.ID)),
			CreditID:     creator.CreditID,
			Name:         creator.Name,
			OriginalName: creator.Name,
			ProfileURI:   helpers.GetTMDBImageURL(creator.ProfilePath, tmdb.W500),
			Job:          "Creator",
		})
	}
	showObject.Creators = &creatorsArray
	// continue to append seasons
	seasonArray := []view.MediaRecordCatalog{}
	for _, season := range showDetails.Seasons {
		seasonArray = append(seasonArray, view.MediaRecordCatalog{
			MediaSource:  sources.MediaSourceTMDB,
			MediaType:    database.RecordTypeSeason,
			SourceID:     strconv.Itoa(int(season.ID)),
			Overview:     season.Overview,
			MediaTitle:   season.Name,
			EpisodeCount: &season.EpisodeCount,
			ThumbnailURI: helpers.GetTMDBImageURL(season.PosterPath, tmdb.W500),
			ReleaseDate:  season.AirDate,
		})
	}
	showObject.Seasons = seasonArray
	helpers.SuccessResponse(c, showObject, 200)
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
		showDetails.Seasons[num].PosterPath = helpers.GetTMDBImageURL(showDetails.Seasons[num].PosterPath, tmdb.W500)
	}
	for num, item := range showDetails.Credits.TVCredits.Cast {
		showDetails.Credits.TVCredits.Cast[num].ProfilePath = helpers.GetTMDBImageURL(item.ProfilePath, tmdb.W500)
	}
	for num, item := range showDetails.Credits.TVCredits.Crew {
		showDetails.Credits.TVCredits.Crew[num].ProfilePath = helpers.GetTMDBImageURL(item.ProfilePath, tmdb.W500)
	}
	var viewSeasons []view.SeasonObjectPartial
	for _, item := range showDetails.Seasons {
		viewSeasons = append(viewSeasons, view.SeasonObjectPartial{
			AirDate:      item.AirDate,
			EpisodeCount: item.EpisodeCount,
			ID:           item.ID,
			Name:         item.Name,
			Overview:     item.Overview,
			PosterURL:    helpers.GetTMDBImageURL(item.PosterPath, tmdb.W500),
			SeasonNumber: item.SeasonNumber,
		})
	}
	logoURL := ""
	if len(showDetails.Images.Logos) > 0 {
		logoURL = helpers.GetTMDBImageURL(showDetails.Images.Logos[0].FilePath, tmdb.W500)
	}
	returnObject := view.TVShowFullObject{
		MediaSource:      sources.MediaSourceTMDB,
		MediaType:        database.MediaTypeTVShow,
		OriginalName:     showDetails.OriginalName,
		SourceID:         showDetails.ID,
		MediaTitle:       showDetails.Name,
		VoteCount:        showDetails.VoteCount,
		VoteAverage:      showDetails.VoteAverage,
		PosterURL:        helpers.GetTMDBImageURL(showDetails.PosterPath, tmdb.W500),
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
		BackdropURL:      helpers.GetTMDBImageURL(showDetails.BackdropPath, tmdb.Original),
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

func GetTVSeasonHandlerV2(c *gin.Context) {
	mediaSource, sourceID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "request id param invalid"+err.Error()))
		return
	}
	seasonNumber, err := strconv.Atoi(c.Param("seasonNumber"))
	if err != nil {
		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
		return
	}
	seasonDetails, err := sources.GetTVSeasonTMDB(sourceID, seasonNumber)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	seasonObject := view.TVSeasonCatalogObject{
		MediaRecordCatalog: view.MediaRecordCatalog{
			MediaType:    database.RecordTypeSeason,
			MediaSource:  sources.MediaSourceTMDB,
			SourceID:     strconv.Itoa(int(seasonDetails.ID)),
			SeasonNumber: &seasonDetails.SeasonNumber,
			ReleaseDate:  seasonDetails.AirDate,
			MediaTitle:   seasonDetails.Name,
			Overview:     seasonDetails.Overview,
			ThumbnailURI: helpers.GetTMDBImageURL(seasonDetails.PosterPath, tmdb.W500),
		},
	}
	episodesArray := []view.MediaRecordCatalog{}
	for _, item := range seasonDetails.Episodes {
		epRecord := view.MediaRecordCatalog{
			MediaSource:   sources.MediaSourceTMDB,
			MediaType:     database.RecordTypeEpisode,
			SourceID:      strconv.Itoa(int(item.ID)),
			SeasonNumber:  &item.SeasonNumber,
			EpisodeNumber: &item.EpisodeNumber,
			MediaTitle:    item.Name,
			Overview:      item.Overview,
			Duration:      item.Runtime,
			ReleaseDate:   item.AirDate,
			ThumbnailURI:  helpers.GetTMDBImageURL(item.StillPath, tmdb.W500),
		}
		guestStarsArr := []view.Credit{}
		for idx, item := range item.GuestStars {
			guestStarsArr = append(guestStarsArr, view.Credit{
				MediaSource: sources.MediaSourceTMDB,
				SourceID:    strconv.Itoa(int(item.ID)),
				CreditID:    item.CreditID,
				Name:        item.Name,
				Character:   &item.Character,
				ProfileURI:  helpers.GetTMDBImageURL(item.ProfilePath, tmdb.W500),
			})
			if idx == 20 {
				break
			}
		}
		epRecord.GuestStars = &guestStarsArr
		episodesArray = append(episodesArray, epRecord)
	}
	seasonObject.Episodes = episodesArray
	helpers.SuccessResponse(c, seasonObject, 200)
}

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
		tvSeason.Episodes[num].StillPath = helpers.GetTMDBImageURL(item.StillPath, tmdb.W500)
	}
	tvSeason.PosterPath = helpers.GetTMDBImageURL(tvSeason.PosterPath, tmdb.W500)

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
