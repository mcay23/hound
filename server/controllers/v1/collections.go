package v1

import (
	"errors"
	"hound/database"
	"hound/helpers"
	"hound/sources"
	"hound/view"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AddToCollectionRequest struct {
	MediaSource  string `json:"media_source" binding:"required,gt=0"`
	MediaType    string `json:"media_type"  binding:"required,gt=0"`
	SourceID     string `json:"source_id" binding:"required,gt=0"`
	CollectionID *int64 `json:"collection_id"`
}

func AddToCollectionHandler(c *gin.Context) {
	username := c.GetHeader("X-Username")
	body := AddToCollectionRequest{}
	if err := c.ShouldBindJSON(&body); err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to bind registration body"))
		return
	}
	idParam := c.Param("id")
	collectionID, err := strconv.Atoi(idParam)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid collection id in url param"))
		return
	}
	temp := int64(collectionID)
	body.CollectionID = &temp
	// check valid mediaType and source
	err = ValidateMediaParams(body.MediaType, body.MediaSource)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	// get source ID as int, right now all sources have int ids
	sourceID, err := strconv.Atoi(body.SourceID)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to convert sourceID string to int")
		helpers.ErrorResponse(c, err)
		return
	}
	switch body.MediaType {
	case database.MediaTypeTVShow:
		err = sources.AddTVShowToCollectionTMDB(username, body.MediaSource, sourceID, body.CollectionID)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to add tv show to collection"))
			return
		}
	case database.MediaTypeMovie:
		err = sources.AddMovieToCollectionTMDB(username, body.MediaSource, sourceID, body.CollectionID)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to add movie to collection"))
			return
		}
	case database.MediaTypeGame:
		err = sources.AddGameToCollectionIGDB(username, body.MediaSource, sourceID, body.CollectionID)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to add game to collection"))
			return
		}
	}
	helpers.SuccessResponse(c, nil, 200)
}

func DeleteFromCollectionHandler(c *gin.Context) {
	userID, err := database.GetUserIDFromUsername(c.GetHeader("X-Username"))
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid user"))
		return
	}
	body := AddToCollectionRequest{}
	if err := c.ShouldBindJSON(&body); err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to bind registration body"))
		return
	}
	idParam := c.Param("id")
	collectionID, err := strconv.Atoi(idParam)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid collection id in url param"))
		return
	}
	temp := int64(collectionID)
	body.CollectionID = &temp
	// check valid mediaType and source
	err = ValidateMediaParams(body.MediaType, body.MediaSource)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	has, record, err := database.GetMediaRecord(body.MediaType, body.MediaSource, body.SourceID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error retrieving Media Record"))
		return
	}
	if !has {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Could not find Media Record"))
		return
	}
	err = database.DeleteCollectionRelation(userID, record.RecordID, *body.CollectionID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to delete collection record"))
		return
	}
	helpers.SuccessResponse(c, nil, 200)
}

func GetUserCollectionsHandler(c *gin.Context) {
	userID, err := database.GetUserIDFromUsername(c.GetHeader("X-Username"))
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid user"))
		return
	}
	query := database.CollectionRecordQuery{
		OwnerUserID: &userID,
	}
	records, _, err := database.SearchForCollection(query, -1, -1)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error searching collection"))
		return
	}
	var collectionResponse []view.CollectionObject
	for _, record := range records {
		temp := view.CollectionObject{
			CollectionID:    record.CollectionID,
			CollectionTitle: record.CollectionTitle,
			Description:     string(record.Description),
			OwnerUsername:   c.GetHeader("X-Username"),
			IsPrimary:       record.IsPrimary,
			IsPublic:        record.IsPublic,
			Tags:            record.Tags,
			ThumbnailURL:    record.ThumbnailURL,
			CreatedAt:       record.CreatedAt,
			UpdatedAt:       record.UpdatedAt,
		}
		collectionResponse = append(collectionResponse, temp)
	}
	helpers.SuccessResponse(c, collectionResponse, 200)
}

func CreateCollectionHandler(c *gin.Context) {
	body := database.CreateCollectionRequest{}
	if err := c.ShouldBindJSON(&body); err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to bind registration body"))
		return
	}
	if body.IsPrimary {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid isPrimary"))
		return
	}
	userID, err := database.GetUserIDFromUsername(c.GetHeader("X-Username"))
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid user"))
		return
	}
	body.OwnerUserID = userID
	collectionID, err := database.CreateCollection(body)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Error creating colection"+err.Error()))
		return
	}
	helpers.SuccessResponse(c, gin.H{"collection_id": collectionID}, 200)
}

func GetCollectionContentsHandler(c *gin.Context) {
	idParam := c.Param("id")
	limitQuery := c.Query("limit")
	offsetQuery := c.Query("offset")
	// -1 means no limit, offset
	limit := -1
	offset := -1
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
	userID, err := database.GetUserIDFromUsername(c.GetHeader("X-Username"))
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid user"))
		return
	}
	collectionID, err := strconv.Atoi(idParam)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid collection id in url param"))
		return
	}
	records, collection, totalRecords, err := database.GetCollectionRecords(userID, int64(collectionID), limit, offset)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to get collection records"))
		return
	}
	var viewArray []view.MediaRecordCatalog
	for _, item := range records {
		viewObject := view.MediaRecordCatalog{
			MediaType:        item.RecordType,
			MediaSource:      item.MediaSource,
			SourceID:         item.SourceID,
			MediaTitle:       item.MediaTitle,
			OriginalTitle:    item.OriginalTitle,
			ReleaseDate:      item.ReleaseDate,
			Overview:         item.Overview,
			ThumbnailURI:     item.ThumbnailURL,
			BackdropURI:      item.BackdropURL,
			Genres:           item.Genres,
			OriginalLanguage: item.OriginalLanguage,
			OriginCountry:    item.OriginCountry,
		}
		viewArray = append(viewArray, viewObject)
	}
	// note collection owner can be different from calling user (public collections)
	collectionOwner, err := database.GetUsernameFromID(collection.OwnerUserID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid user"))
		return
	}
	res := view.CollectionView{
		Results: viewArray,
		Collection: &view.CollectionObject{
			CollectionID:    collection.CollectionID,
			CollectionTitle: collection.CollectionTitle,
			Description:     string(collection.Description),
			OwnerUsername:   collectionOwner,
			IsPrimary:       collection.IsPrimary,
			IsPublic:        collection.IsPublic,
			Tags:            collection.Tags,
			ThumbnailURL:    collection.ThumbnailURL,
			CreatedAt:       collection.CreatedAt,
			UpdatedAt:       collection.UpdatedAt,
		},
		TotalRecords: totalRecords,
		Limit:        limit,
		Offset:       offset,
	}
	helpers.SuccessResponse(c, res, 200)
}

func GetRecentCollectionContentsHandler(c *gin.Context) {
	userID, err := database.GetUserIDFromUsername(c.GetHeader("X-Username"))
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid user"))
		return
	}
	// return 20 most recent
	records, err := database.GetRecentCollectionRecords(userID, 20)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to get recent collection records"))
		return
	}
	var viewArray []view.MediaRecordCatalog
	for _, item := range records {
		viewObject := view.MediaRecordCatalog{
			MediaType:        item.RecordType,
			MediaSource:      item.MediaSource,
			SourceID:         item.SourceID,
			MediaTitle:       item.MediaTitle,
			OriginalTitle:    item.OriginalTitle,
			Status:           item.Status,
			Overview:         item.Overview,
			Duration:         item.Duration,
			ReleaseDate:      item.ReleaseDate,
			LastAirDate:      item.LastAirDate,
			NextAirDate:      item.NextAirDate,
			SeasonNumber:     item.SeasonNumber,
			EpisodeNumber:    item.EpisodeNumber,
			ThumbnailURI:     item.ThumbnailURL,
			BackdropURI:      item.BackdropURL,
			StillURI:         item.StillURL,
			Genres:           item.Genres,
			OriginalLanguage: item.OriginalLanguage,
			OriginCountry:    item.OriginCountry,
		}
		viewArray = append(viewArray, viewObject)
	}
	helpers.SuccessResponse(c, viewArray, 200)
}

func DeleteCollectionHandler(c *gin.Context) {
	idParam := c.Param("id")
	collectionID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Invalid collection_id query param")
		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
		return
	}
	userID, err := database.GetUserIDFromUsername(c.GetHeader("X-Username"))
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Invalid user")
		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
		return
	}
	err = database.DeleteCollection(userID, collectionID)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to delete collection")
		helpers.ErrorResponse(c, errors.New(helpers.InternalServerError))
		return
	}
	helpers.SuccessResponse(c, nil, 200)
}
