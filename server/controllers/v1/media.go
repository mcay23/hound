package v1

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"hound/helpers"
	"hound/model/database"
	"hound/model/sources"
	"hound/view"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type AddToCollectionRequest struct {
	MediaSource  string `json:"media_source" binding:"required,gt=0"`
	MediaType    string `json:"media_type"  binding:"required,gt=0"`
	SourceID     string `json:"source_id" binding:"required,gt=0"`
	CollectionID *int64 `json:"collection_id"`
}

type CommentRequest struct {
	CommentType  string    `json:"comment_type" binding:"required,gt=0"` // review, etc.
	IsPrivate    bool      `json:"is_private"`
	CommentTitle string    `json:"title"`
	Comment      string    `json:"comment"`    // actual content of comment, review
	StartDate    time.Time `json:"start_date"` // for watch history
	EndDate      time.Time `json:"end_date"`
	TagData      string    `json:"tag_data"` // extra tag info, eg. season, episode
	Score        int       `json:"score"`    // only for reviews
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
	helpers.SuccessResponse(c, gin.H{"status": "success"}, 200)
}

func GetCommentsHandler(c *gin.Context) {
	idParam := c.Param("id")
	mediaSource, sourceID, err := ParseID(idParam)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	requestURL := strings.Split(c.Request.URL.Path, "/")
	if len(requestURL) <= 0 {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "request url invalid (should not happen)"))
		return
	}
	mediaType := requestURL[3]
	if mediaType == "tv" {
		mediaType = database.MediaTypeTVShow
	}
	libraryID, err := database.GetInternalLibraryID(mediaType, mediaSource, strconv.Itoa(sourceID))
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "No internal library ID found"))
		return
	}
	commentType := c.Query("type")
	comments, err := GetCommentsCore(c.GetHeader("X-Username"), *libraryID, &commentType)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Error retrieving comments"))
		return
	}
	helpers.SuccessResponse(c, comments, 200)
}

func PostCommentHandler(c *gin.Context) {
	var body CommentRequest
	err := c.ShouldBindJSON(&body)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	if body.Score > 100 || body.Score < 0 {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "request url invalid (should not happen)"))
		return
	}
	idParam := c.Param("id")
	mediaSource, sourceID, err := ParseID(idParam)
	requestURL := strings.Split(c.Request.URL.Path, "/")
	if len(requestURL) <= 0 {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "request url invalid (should not happen)"))
		return
	}
	mediaType := requestURL[3]
	if mediaType == "tv" {
		mediaType = database.MediaTypeTVShow
	}
	// get userID
	userID, err := database.GetUserIDFromUsername(c.GetHeader("X-Username"))
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	// check valid mediaType and source
	err = ValidateMediaParams(mediaType, mediaSource)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	comment := database.CommentRecord{
		UserID:       userID,
		CommentTitle: body.CommentTitle,
		IsPrivate:    body.IsPrivate,
		CommentType:  body.CommentType,
		Comment:      []byte(body.Comment),
		StartDate:    body.StartDate,
		EndDate:      body.EndDate,
		TagData:      body.TagData,
		Score:        body.Score,
	}
	if mediaType == database.MediaTypeTVShow || mediaType == database.MediaTypeMovie {
		record, err := sources.GetLibraryObjectTMDB(mediaType, sourceID)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Failed to get library object tmdb"))
			return
		}
		// TODO bound checking for tag data (season and episode)
		// add item to internal library if not there
		libraryID, err := database.AddRecordToInternalLibrary(record)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Failed to insert record to library"))
			return
		}
		if mediaType == database.MediaTypeTVShow {
			if match, _ := regexp.MatchString(`S\d+$|S\d+E\d+$`, body.TagData); !match {
				fmt.Println(body.TagData)
				helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid TagData format, regex failed"))
				return
			}
			// mark seasons as watch case, no episode data
			if !strings.Contains(body.TagData, "E") {
				seasonNumber, err := strconv.Atoi(strings.Split(body.TagData, "E")[0][1:])
				if err != nil {
					helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid TagData format"))
					return
				}
				season, err := sources.GetTVSeasonTMDB(sourceID, seasonNumber, nil)
				if err != nil {
					helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Error retrieving season"))
					return
				}
				minEpisode := season.Episodes[0].EpisodeNumber
				maxEpisode := 0
				for _, ep := range season.Episodes {
					if ep.EpisodeNumber < minEpisode {
						minEpisode = ep.EpisodeNumber
					}
					if ep.EpisodeNumber > maxEpisode {
						maxEpisode = ep.EpisodeNumber
					}
				}
				err = sources.MarkTVSeasonAsWatchedTMDB(userID, libraryID, seasonNumber, minEpisode, maxEpisode, body.StartDate)
				if err != nil {
					helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Error during batch insertion"))
					return
				}
				helpers.SuccessResponse(c, gin.H{"status": "success"}, 200)
				return
			}
		}
		comment.LibraryID = libraryID
	} else if mediaType == database.MediaTypeGame {
		record, err := sources.GetLibraryObjectIGDB(sourceID)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Failed to get library object tmdb"))
			return
		}
		// add item to internal library if not there
		libraryID, err := database.AddRecordToInternalLibrary(record)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Failed to insert record to library"))
			return
		}
		comment.LibraryID = libraryID
	} else {
		helpers.ErrorResponse(c, errors.New(helpers.BadRequest))
		return
	}
	err = database.AddComment(&comment)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Failed to add comment"))
		return
	}
	helpers.SuccessResponse(c, gin.H{"status": "success", "comment_id": comment.CommentID}, 200)
}

func DeleteCommentHandler(c *gin.Context) {
	username := c.GetHeader("X-Username")
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid user"))
		return
	}
	idParam := c.Query("ids")
	idSplit := strings.Split(idParam, ",")
	if len(idSplit) > 1 {
		// batch delete case
		var batchIDs []int64
		for _, item := range idSplit {
			tempID, err := strconv.Atoi(item)
			if err != nil {
				helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid comment id in url param"))
				return
			}
			batchIDs = append(batchIDs, int64(tempID))
		}
		err = database.DeleteCommentBatch(userID, batchIDs)
		if err != nil {
			helpers.ErrorResponse(c, err)
			return
		}
	} else {
		// single delete case
		commentID, err := strconv.Atoi(idParam)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid comment id in url param"))
			return
		}
		err = database.DeleteComment(userID, int64(commentID))
		if err != nil {
			helpers.ErrorResponse(c, err)
			return
		}
	}
	helpers.SuccessResponse(c, gin.H{"status": "success"}, 200)
}

func ParseID(idParam string) (string, int, error) {
	split := strings.Split(idParam, "-")
	if len(split) != 2 {
		return "", -1, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid id in param")
	}
	// TODO sourceID is int only for all sources (tmdb, igdb) but might be different in other sources
	sourceID, err := strconv.Atoi(split[1])
	if err != nil {
		return "", -1, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "cannot cast sourceid to string")
	}
	mediaSource := split[0]
	return mediaSource, sourceID, nil
}

func GetCommentsCore(username string, libraryID int64, commentType *string) (*[]view.CommentObject, error) {
	comments, err := database.GetComments(libraryID, commentType)
	if err != nil {
		return nil, err
	}
	var commentsView []view.CommentObject
	for _, item := range *comments {
		commenter, _ := database.GetUsernameFromID(item.UserID)
		if item.IsPrivate && username != commenter {
			continue
		}
		comment := view.CommentObject{
			CommentTitle: item.CommentTitle,
			CommentID:    item.CommentID,
			CommentType:  item.CommentType,
			UserID:       commenter,
			LibraryID:    item.LibraryID,
			IsPrivate:    item.IsPrivate,
			Comment:      string(item.Comment),
			TagData:      item.TagData,
			Score:        item.Score,
			StartDate:    item.StartDate,
			EndDate:      item.EndDate,
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
		}
		commentsView = append(commentsView, comment)
	}
	return &commentsView, nil
}
