package v1

import (
	"fmt"
	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/helpers"
	"github.com/mcay23/hound/model"
	"github.com/mcay23/hound/sources"
	"github.com/mcay23/hound/view"
	"strconv"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/gin-gonic/gin"
)

// @Router /v1/tv/search [get]
// @Summary Search TV Shows
// @Tags TV Show, Search
// @Accept json
// @Produce json
// @Param query query string true "Search Query"
// @Success 200 {object} V1SuccessResponse{data=[]view.MediaRecordCatalog}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func SearchTVShowHandler(c *gin.Context) {
	queryString := c.Query("query")
	results, err := model.SearchTVShows(queryString)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to search tv show for query %s: %w", queryString, err))
		return
	}
	helpers.SuccessResponse(c, results, 200)
}

// @Router /v1/tv/{id} [get]
// @Summary Get TV Show Details
// @Tags TV Show
// @Accept json
// @Produce json
// @Param id path string true "Media ID" example(tmdb-1234)
// @Success 200 {object} V1SuccessResponse{data=view.TVShowCatalogObject}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetTVShowFromIDHandler(c *gin.Context) {
	mediaSource, showID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, fmt.Errorf("request id param invalid: %w: %w", helpers.BadRequestError, err))
		return
	}
	showDetails, err := sources.GetTVShowFromIDTMDB(showID)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to get tv show from id %d: %w", showID, err))
		return
	}
	// create top level show
	duration := 0
	if len(showDetails.EpisodeRunTime) > 0 {
		duration = showDetails.EpisodeRunTime[0]
	}
	genreArray := database.ConvertGenres(sources.MediaSourceTMDB, database.MediaTypeTVShow, showDetails.Genres)
	logoURI := ""
	if len(showDetails.Images.Logos) > 0 {
		logoURI = helpers.GetTMDBImageURL(showDetails.Images.Logos[0].FilePath, tmdb.W500)
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
			LogoURI:          logoURI,
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
			ThumbnailURI: helpers.GetTMDBImageURL(cast.ProfilePath, tmdb.W500),
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
			ThumbnailURI: helpers.GetTMDBImageURL(creator.ProfilePath, tmdb.W500),
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
			SeasonNumber: &season.SeasonNumber,
			EpisodeCount: &season.EpisodeCount,
			ThumbnailURI: helpers.GetTMDBImageURL(season.PosterPath, tmdb.W500),
			ReleaseDate:  season.AirDate,
		})
	}
	showObject.Seasons = seasonArray
	helpers.SuccessResponse(c, showObject, 200)
}

// @Router /v1/tv/{id}/season/{seasonNumber} [get]
// @Summary Get TV Season Details
// @Tags TV Show
// @Accept json
// @Produce json
// @Param id path string true "Media ID" example(tmdb-1234)
// @Param seasonNumber path int true "Season Number"
// @Success 200 {object} V1SuccessResponse{data=view.TVSeasonCatalogObject}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetTVSeasonHandler(c *gin.Context) {
	mediaSource, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, fmt.Errorf("request id param invalid: %w: %w", helpers.BadRequestError, err))
		return
	}
	seasonNumber, err := strconv.Atoi(c.Param("seasonNumber"))
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("invalid season number: %w: %w", helpers.BadRequestError, err))
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
				MediaSource:  sources.MediaSourceTMDB,
				SourceID:     strconv.Itoa(int(item.ID)),
				CreditID:     item.CreditID,
				Name:         item.Name,
				Character:    &item.Character,
				ThumbnailURI: helpers.GetTMDBImageURL(item.ProfilePath, tmdb.W500),
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

// @Router /v1/tv/{id}/episode_groups [get]
// @Summary Get TV Episode Groups
// @Tags TV Show
// @Accept json
// @Produce json
// @Param id path string true "Media ID" example(tmdb-1234)
// @Success 200 {object} V1SuccessResponse{data=object}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetTVEpisodeGroupsHandler(c *gin.Context) {
	mediaSource, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, fmt.Errorf("request id param invalid: %w: %w", helpers.BadRequestError, err))
		return
	}
	episodeGroups, err := sources.GetTVEpisodeGroupsTMDB(sourceID)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to get tv episode groups for id %d: %w", sourceID, err))
		return
	}
	helpers.SuccessResponse(c, episodeGroups.Results, 200)
}
