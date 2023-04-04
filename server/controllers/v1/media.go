package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"hound/helpers"
	"hound/model/database"
	"hound/model/sources"
	"hound/view"
	"regexp"
	"strconv"
)

type AddToCollectionRequest struct {
	MediaSource  string `json:"media_source" binding:"required,gt=0"`
	MediaType    string `json:"media_type"  binding:"required,gt=0"`
	SourceID     string `json:"source_id" binding:"required,gt=0"`
	CollectionID *int64 `json:"collection_id"`
}

type CommentRequest struct {
	MediaSource string `json:"media_source" binding:"required,gt=0"`
	MediaType   string `json:"media_type"  binding:"required,gt=0"`
	SourceID    string `json:"source_id" binding:"required,gt=0"`
	CommentType string `json:"comment_type" binding:"required,gt=0"` // review, etc.
	IsPrivate   bool   `json:"is_private" binding:"required"`
	Comment     []byte `json:"comment" binding:"required,gt=0"` // actual content of comment, review
	TagData     string `json:"tag_data"`                        // extra tag info, eg. season, episode
	Score       int    `json:"score"`                           // only for reviews
}

func GeneralSearchHandler(c *gin.Context) {
	queryString := c.Query("q")
	// search tmdb
	tvResults, _ := SearchTVShowCore(queryString)
	movieResults, _ := SearchMoviesCore(queryString)
	// search igdb
	gameResults, _ := sources.SearchGameIGDB(queryString)

	helpers.SuccessResponse(c, view.GeneralSearchResponse{
		TVShowSearchResults: tvResults,
		MovieSearchResults:  movieResults,
		GameSearchResults:   &gameResults,
	}, 200)
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
	if body.MediaType == database.MediaTypeTVShow {
		err = sources.AddTVShowToCollectionTMDB(username, body.MediaSource, sourceID, body.CollectionID)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to add tv show to collection"))
			return
		}
	} else if body.MediaType == database.MediaTypeMovie {
		err = sources.AddMovieToCollectionTMDB(username, body.MediaSource, sourceID, body.CollectionID)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to add movie to collection"))
			return
		}
	} else if body.MediaType == database.MediaTypeGame {
		err = sources.AddGameToCollectionIGDB(username, body.MediaSource, sourceID, body.CollectionID)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to add game to collection"))
			return
		}
	}
	helpers.SuccessResponse(c, gin.H{"status": "success"}, 200)
}

func DeleteFromCollectionHandler(c *gin.Context)  {
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
	libraryID, err := database.GetInternalLibraryID(body.MediaType, body.MediaSource, body.SourceID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid user"))
		return
	}
	err = database.DeleteCollectionRelation(userID, *libraryID, *body.CollectionID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to delete collection record"))
		return
	}
	helpers.SuccessResponse(c, gin.H{"status": "success"}, 200)
}

func GetUserCollectionsHandler(c *gin.Context) {
	userID, err := database.GetUserIDFromUsername(c.GetHeader("X-Username"))
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid user"))
		return
	}
	query := database.CollectionRecordQuery{
		OwnerID: &userID,
	}
	records, _, err := database.SearchForCollection(query, -1, -1)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error searching collection"))
		return
	}
	var collectionResponse []view.CollectionRecordView
	for _, record := range records {
		temp := view.CollectionRecordView{
			CollectionID:    record.CollectionID,
			CollectionTitle: record.CollectionTitle,
			Description:     string(record.Description),
			Username:        c.GetHeader("X-Username"),
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
	body.OwnerID = userID
	collectionID, err := database.CreateCollection(body)
	helpers.SuccessResponse(c, gin.H{"status": "success", "collection_id": collectionID}, 200)
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
	var viewArray []view.LibraryObject
	for _, item := range records {
		viewObject := view.LibraryObject{
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
	// note collection owner can be different from calling user (public collections)
	collectionOwner, err := database.GetUsernameFromID(collection.OwnerID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid user"))
		return
	}
	helpers.SuccessResponse(c, view.CollectionView{
		Results: &viewArray,
		Collection: &view.CollectionRecordView{
			CollectionID:    collection.CollectionID,
			CollectionTitle: collection.CollectionTitle,
			Description:     string(collection.Description),
			Username:        collectionOwner,
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
	}, 200)
}

func PostComment(c *gin.Context) {
	var body CommentRequest
	err := c.ShouldBindJSON(&body)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	// TODO sourceID is int only for all sources (tmdb, igdb) but might be different in other sources
	sourceID, err := strconv.Atoi(body.SourceID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "cannot cast sourceid to string"))
		return
	}
	// get userID
	userID, err := database.GetUserIDFromUsername(c.GetHeader("X-Username"))
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	// check valid mediaType and source
	err = ValidateMediaParams(body.MediaType, body.MediaSource)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	var comment *database.CommentRecord
	if body.MediaType == database.MediaTypeTVShow {
		// no match, return error - format is S2E12, S2, show etc.
		if match, _ := regexp.MatchString(`S\d+$|S\d+E\d+$|show$`, body.TagData); !match {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid TagData format, regex failed"))
			return
		}
		record, err := sources.GetLibraryObjectTMDB(database.MediaTypeTVShow, sourceID)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Failed to get library object tmdb"))
		}
		// add item to internal library if not there
		libraryID, err := database.AddRecordToInternalLibrary(record)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Failed to insert record to library"))
		}
		comment = &database.CommentRecord{
			UserID:      userID,
			LibraryID:   libraryID,
			IsPrivate:   body.IsPrivate,
			CommentType: body.CommentType,
			Comment:     body.Comment,
			TagData:     body.TagData,
			Score:       body.Score,
		}
	} else if body.MediaType == database.MediaTypeMovie {
		record, err := sources.GetLibraryObjectTMDB(database.MediaTypeMovie, sourceID)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Failed to get library object tmdb"))
		}
		// add item to internal library if not there
		libraryID, err := database.AddRecordToInternalLibrary(record)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Failed to insert record to library"))
		}
		// omit tag data, not used in movie comments/review
		comment = &database.CommentRecord{
			UserID:      userID,
			LibraryID:   libraryID,
			IsPrivate:   body.IsPrivate,
			CommentType: body.CommentType,
			Comment:     body.Comment,
			Score:       body.Score,
		}
	} else if body.MediaType == database.MediaTypeGame {
		record, err := sources.GetLibraryObjectIGDB(sourceID)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Failed to get library object tmdb"))
		}
		// add item to internal library if not there
		libraryID, err := database.AddRecordToInternalLibrary(record)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Failed to insert record to library"))
		}
		// omit tag data, not used in game comments/review
		comment = &database.CommentRecord{
			UserID:      userID,
			LibraryID:   libraryID,
			IsPrivate:   body.IsPrivate,
			CommentType: body.CommentType,
			Comment:     body.Comment,
			Score:       body.Score,
		}
	} else {
		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
		return
	}
	err = database.AddComment(comment)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Failed to add comment"))
	}
}
