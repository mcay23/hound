package v1

import (
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/sources"
	"hound/view"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AddToCollectionRequest struct {
	MediaSource string `json:"media_source" binding:"required,gt=0"`
	MediaType   string `json:"media_type"  binding:"required,gt=0"`
	SourceID    string `json:"source_id" binding:"required,gt=0"`
}

type CreateCollectionRequest struct {
	OwnerUserID     int64  `json:"owner_user_id"`
	CollectionTitle string `json:"collection_title"` // my collection, etc.
	Description     string `json:"description"`
	IsPublic        bool   `json:"is_public"`
}

// @Router /v1/collection/{id} [post]
// @Summary Add Media to Collection
// @Tags Collection
// @Accept json
// @Produce json
// @Param id path int true "Collection ID"
// @Param body body AddToCollectionRequest true "Add to Collection Request"
// @Success 200 {object} V1SuccessResponse{data=object}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func AddToCollectionHandler(c *gin.Context) {
	username := c.GetHeader("X-Username")
	body := AddToCollectionRequest{}
	if err := c.ShouldBindJSON(&body); err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to bind body: %w : %w", helpers.BadRequestError, err))
		return
	}
	idParam := c.Param("id")
	collectionID, err := strconv.Atoi(idParam)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to convert collection id to int: %w: %w", helpers.BadRequestError, err))
		return
	}
	// check valid mediaType and source
	err = validateMediaParams(body.MediaType, body.MediaSource)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	// get source ID as int, right now all sources have int ids
	sourceID, err := strconv.Atoi(body.SourceID)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to convert id to int: %w", helpers.BadRequestError))
		return
	}
	switch body.MediaType {
	case database.MediaTypeTVShow:
		err = sources.AddTVShowToCollectionTMDB(username, body.MediaSource, sourceID, int64(collectionID))
		if err != nil {
			helpers.ErrorResponse(c, fmt.Errorf("failed to add tv show to collection: %w", err))
			return
		}
	case database.MediaTypeMovie:
		err = sources.AddMovieToCollectionTMDB(username, body.MediaSource, sourceID, int64(collectionID))
		if err != nil {
			helpers.ErrorResponse(c, fmt.Errorf("failed to add movie to collection: %w", err))
			return
		}
	case database.MediaTypeGame:
		err = sources.AddGameToCollectionIGDB(username, body.MediaSource, sourceID, int64(collectionID))
		if err != nil {
			helpers.ErrorResponse(c, fmt.Errorf("failed to add game to collection: %w", err))
			return
		}
	}
	helpers.SuccessResponse(c, nil, 200)
}

// @Router /v1/collection/{id} [delete]
// @Summary Delete A Media from Collection
// @Tags Collection
// @Accept json
// @Produce json
// @Param id path int true "Collection ID"
// @Param body body AddToCollectionRequest true "Delete from Collection Request"
// @Success 200 {object} V1SuccessResponse{data=object}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func DeleteFromCollectionHandler(c *gin.Context) {
	userID, err := database.GetUserIDFromUsername(c.GetHeader("X-Username"))
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("invalid user: %w: %w", helpers.BadRequestError, err))
		return
	}
	body := AddToCollectionRequest{}
	if err := c.ShouldBindJSON(&body); err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to bind body: %w: %w", helpers.BadRequestError, err))
		return
	}
	idParam := c.Param("id")
	collectionID, err := strconv.Atoi(idParam)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to convert param id to int: %w: %w", helpers.BadRequestError, err))
		return
	}
	// check valid mediaType and source
	err = validateMediaParams(body.MediaType, body.MediaSource)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	has, record, err := database.GetMediaRecord(body.MediaType, body.MediaSource, body.SourceID)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to get media record: %w", err))
		return
	}
	if !has {
		helpers.ErrorResponse(c, fmt.Errorf("could not find media record: %w", helpers.BadRequestError))
		return
	}
	err = database.DeleteCollectionRelation(userID, record.RecordID, int64(collectionID))
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to delete collection record: %w", err))
		return
	}
	helpers.SuccessResponse(c, nil, 200)
}

// @Router /v1/collection/all [get]
// @Summary Get a User's Collections
// @Tags Collection
// @Accept json
// @Produce json
// @Success 200 {object} V1SuccessResponse{data=[]view.CollectionObject}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetUserCollectionsHandler(c *gin.Context) {
	userID, err := database.GetUserIDFromUsername(c.GetHeader("X-Username"))
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("invalid user: %w: %w", helpers.BadRequestError, err))
		return
	}
	records, _, err := database.FindCollection(database.CollectionRecord{OwnerUserID: userID}, -1, -1)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to find collection: %w: %w", helpers.InternalServerError, err))
		return
	}
	var collectionResponse = []view.CollectionObject{}
	for _, record := range records {
		temp := view.CollectionObject{
			CollectionID:    record.CollectionID,
			CollectionTitle: record.CollectionTitle,
			Description:     string(record.Description),
			OwnerUsername:   c.GetHeader("X-Username"),
			IsPublic:        record.IsPublic,
			ThumbnailURI:    record.ThumbnailURI,
			CreatedAt:       record.CreatedAt,
			UpdatedAt:       record.UpdatedAt,
		}
		collectionResponse = append(collectionResponse, temp)
	}
	helpers.SuccessResponse(c, collectionResponse, 200)
}

// @Router /v1/collection/new [post]
// @Summary Create New Collection
// @Tags Collection
// @Accept json
// @Produce json
// @Success 200 {object} V1SuccessResponse{data=object}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func CreateCollectionHandler(c *gin.Context) {
	body := CreateCollectionRequest{}
	if err := c.ShouldBindJSON(&body); err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to bind body: %w: %w", helpers.BadRequestError, err))
		return
	}
	userID, err := database.GetUserIDFromUsername(c.GetHeader("X-Username"))
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("invalid user: %w: %w", helpers.BadRequestError, err))
		return
	}
	record := database.CollectionRecord{
		CollectionTitle: body.CollectionTitle,
		Description:     body.Description,
		OwnerUserID:     userID,
		IsPublic:        body.IsPublic,
		ThumbnailURI:    "",
	}
	collectionID, err := database.CreateCollection(record)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to create collection: %w", err))
		return
	}
	helpers.SuccessResponse(c, gin.H{"collection_id": collectionID}, 200)
}

// @Router /v1/collection/{id} [get]
// @Summary Get Collection Contents
// @Tags Collection
// @Accept json
// @Produce json
// @Param id path int true "Collection ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} V1SuccessResponse{data=view.CollectionView}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetCollectionContentsHandler(c *gin.Context) {
	idParam := c.Param("id")
	limitQuery := c.Query("limit")
	offsetQuery := c.Query("offset")
	// -1 means no limit, offset
	limit, offset, err := getLimitOffset(limitQuery, offsetQuery)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to get limit/offset: %w: %w", helpers.BadRequestError, err))
		return
	}
	collectionID, err := strconv.Atoi(idParam)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to convert id to int: %w: %w", helpers.BadRequestError, err))
		return
	}
	userID, err := database.GetUserIDFromUsername(c.GetHeader("X-Username"))
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("invalid user: %w: %w", helpers.BadRequestError, err))
		return
	}
	records, collection, totalRecords, err := database.GetCollectionRecords(userID, int64(collectionID), limit, offset)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to get collection records: %w", err))
		return
	}
	var viewArray = []view.MediaRecordCatalog{}
	for _, item := range records {
		viewObject := createMediaRecordCatalogObject(item)
		viewArray = append(viewArray, viewObject)
	}
	// note collection owner can be different from calling user (public collections)
	collectionOwner, err := database.GetUsernameFromID(collection.OwnerUserID)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("invalid user: %w", err))
		return
	}
	res := view.CollectionView{
		Records: viewArray,
		Collection: &view.CollectionObject{
			CollectionID:    collection.CollectionID,
			CollectionTitle: collection.CollectionTitle,
			Description:     string(collection.Description),
			OwnerUsername:   collectionOwner,
			IsPublic:        collection.IsPublic,
			ThumbnailURI:    collection.ThumbnailURI,
			CreatedAt:       collection.CreatedAt,
			UpdatedAt:       collection.UpdatedAt,
		},
		TotalRecords: totalRecords,
		Limit:        limit,
		Offset:       offset,
	}
	helpers.SuccessResponse(c, res, 200)
}

// @Router /v1/collection/recent [get]
// @Summary Get User's Recent Collection Records
// @Description Gets 20 most recent records added to any collection
// @Tags Collection
// @Accept json
// @Produce json
// @Success 200 {object} V1SuccessResponse{data=[]view.MediaRecordCatalog}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetRecentCollectionContentsHandler(c *gin.Context) {
	userID, err := database.GetUserIDFromUsername(c.GetHeader("X-Username"))
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("invalid user: %w: %w", helpers.BadRequestError, err))
		return
	}
	// return 20 most recent
	records, err := database.GetRecentCollectionRecords(userID, 20)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to get recent collection records: %w: %w", helpers.InternalServerError, err))
		return
	}
	var viewArray []view.MediaRecordCatalog
	for _, item := range records {
		viewObject := createMediaRecordCatalogObject(item)
		viewArray = append(viewArray, viewObject)
	}
	helpers.SuccessResponse(c, viewArray, 200)
}

// @Router /v1/collection/{id}/delete [delete]
// @Summary Delete Collection
// @Tags Collection
// @Accept json
// @Produce json
// @Param id path int true "Collection ID"
// @Success 200 {object} V1SuccessResponse{data=object}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func DeleteCollectionHandler(c *gin.Context) {
	idParam := c.Param("id")
	collectionID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to convert id to int: %w: %w", helpers.BadRequestError, err))
		return
	}
	userID, err := database.GetUserIDFromUsername(c.GetHeader("X-Username"))
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("invalid user: %w: %w", helpers.BadRequestError, err))
		return
	}
	err = database.DeleteCollection(userID, collectionID)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to delete collection: %w", err))
		return
	}
	helpers.SuccessResponse(c, nil, 200)
}

func createMediaRecordCatalogObject(record database.MediaRecordGroup) view.MediaRecordCatalog {
	return view.MediaRecordCatalog{
		MediaType:        record.RecordType,
		MediaSource:      record.MediaSource,
		SourceID:         record.SourceID,
		MediaTitle:       record.MediaTitle,
		OriginalTitle:    record.OriginalTitle,
		Status:           record.Status,
		Overview:         record.Overview,
		Duration:         record.Duration,
		ReleaseDate:      record.ReleaseDate,
		LastAirDate:      record.LastAirDate,
		NextAirDate:      record.NextAirDate,
		SeasonNumber:     record.SeasonNumber,
		EpisodeNumber:    record.EpisodeNumber,
		ThumbnailURI:     record.ThumbnailURI,
		BackdropURI:      record.BackdropURI,
		Genres:           record.Genres,
		OriginalLanguage: record.OriginalLanguage,
		OriginCountry:    record.OriginCountry,
	}
}
