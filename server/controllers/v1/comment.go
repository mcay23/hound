package v1

import (
	"errors"
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/sources"
	"hound/view"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type CommentRequest struct {
	CommentType   string `json:"comment_type" binding:"required,gt=0"` // review, etc.
	IsPublic      bool   `json:"is_public"`
	SeasonNumber  *int   `json:"season_number"` // only for tvshows, when commenting on a particular episode
	EpisodeNumber *int   `json:"episode_number"`
	CommentTitle  string `json:"title"`
	Comment       string `json:"comment"` // actual content of comment, review
	Score         int    `json:"score"`   // only required for reviews
}

func GetCommentsTVHandler(c *gin.Context) {
	handleGetComments(c, database.RecordTypeTVShow)
}

func GetCommentsMovieHandler(c *gin.Context) {
	handleGetComments(c, database.RecordTypeMovie)
}

func handleGetComments(c *gin.Context, recordType string) {
	mediaSource, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	var seasonNumber, episodeNumber int
	if recordType == database.RecordTypeTVShow {
		var err error
		seasonNumber, episodeNumber, err = getSeasonEpisode(c.Query("season"), c.Query("episode"))
		if err == nil {
			recordType = database.RecordTypeEpisode
		}
	}
	commentType := c.Query("type")
	if commentType == "" {
		helpers.ErrorResponse(c, fmt.Errorf("invalid type param: %w", helpers.BadRequestError))
		return
	}
	var record *database.MediaRecord
	if recordType == database.RecordTypeEpisode {
		record, err = database.GetEpisodeMediaRecord(sources.MediaSourceTMDB, strconv.Itoa(sourceID), &seasonNumber, &episodeNumber)
		if err != nil {
			if errors.Is(err, helpers.NotFoundError) {
				helpers.SuccessResponse(c, nil, 200)
				return
			}
			helpers.ErrorResponse(c, fmt.Errorf("failed to get media record: %w", err))
			return
		}
	} else {
		var has bool
		has, record, err = database.GetMediaRecord(recordType, mediaSource, strconv.Itoa(sourceID))
		if err != nil {
			helpers.ErrorResponse(c, fmt.Errorf("failed to get media record: %w", err))
			return
		}
		if !has {
			helpers.SuccessResponse(c, nil, 200)
		}
	}
	if record == nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to get media record: %w", helpers.InternalServerError))
		return
	}
	comments, err := database.GetComments(record.RecordID, &commentType)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("error retrieving comments: %w", helpers.InternalServerError))
		return
	}
	var commentsView []view.CommentObject
	for _, item := range *comments {
		commenter, _ := database.GetUsernameFromID(item.UserID)
		if !item.IsPublic && c.GetHeader("X-Username") != commenter {
			continue
		}
		comment := view.CommentObject{
			CommentID:    item.CommentID,
			CommentType:  item.CommentType,
			UserID:       commenter,
			RecordID:     item.RecordID,
			IsPublic:     item.IsPublic,
			CommentTitle: item.CommentTitle,
			Comment:      string(item.Comment),
			Score:        item.Score,
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
		}
		commentsView = append(commentsView, comment)
	}
	helpers.SuccessResponse(c, commentsView, 200)
}

func PostCommentTVHandler(c *gin.Context) {
	handlePostComment(c, database.RecordTypeTVShow)
}

func PostCommentMovieHandler(c *gin.Context) {
	handlePostComment(c, database.RecordTypeMovie)
}

func handlePostComment(c *gin.Context, recordType string) {
	var body CommentRequest
	err := c.ShouldBindJSON(&body)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	mediaSource, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	if body.CommentType == database.CommentTypeReview && (body.Score > 100 || body.Score < 0) {
		helpers.ErrorResponse(c, fmt.Errorf("invalid score %d not (0<=score<=100): %w", body.Score, helpers.BadRequestError))
		return
	}
	// get userID
	userID, err := database.GetUserIDFromUsername(c.GetHeader("X-Username"))
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	// upsert top level record
	record, err := sources.UpsertMediaRecordTMDB(recordType, sourceID)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to upsert media record: %w", err))
		return
	}
	// if season and episode is specified, use episode record
	if body.SeasonNumber != nil && body.EpisodeNumber != nil {
		record, err = database.GetEpisodeMediaRecord(mediaSource, strconv.Itoa(sourceID), body.SeasonNumber, body.EpisodeNumber)
		if err != nil {
			helpers.ErrorResponse(c, fmt.Errorf("failed to get episode media record: %w", err))
			return
		}
	}
	comment := database.CommentRecord{
		UserID:       userID,
		CommentTitle: body.CommentTitle,
		IsPublic:     body.IsPublic,
		CommentType:  body.CommentType,
		Comment:      body.Comment,
		Score:        body.Score,
		RecordID:     record.RecordID,
	}
	err = database.AddComment(&comment)
	if err != nil {
		helpers.ErrorResponse(c, fmt.Errorf("failed to add comment: %w", err))
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
	// for batch deletion, split query params /comment?ids=1,2,3
	idSplit := strings.Split(c.Query("ids"), ",")
	if c.Query("ids") != "" {
		var batchIDs []int64
		for _, item := range idSplit {
			tempID, err := strconv.Atoi(item)
			if err != nil {
				helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Batch deletion: Invalid comment id in url query"))
				return
			}
			batchIDs = append(batchIDs, int64(tempID))
		}
		err = database.DeleteCommentBatch(userID, batchIDs)
		if err != nil {
			helpers.ErrorResponse(c, err)
			return
		}
	} else if c.Param("id") != "" {
		// single delete case
		commentID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid comment id in url param"))
			return
		}
		err = database.DeleteComment(userID, int64(commentID))
		if err != nil {
			helpers.ErrorResponse(c, err)
			return
		}
	} else {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid comment id in url param/query"))
		return
	}
	helpers.SuccessResponse(c, nil, 200)
}
