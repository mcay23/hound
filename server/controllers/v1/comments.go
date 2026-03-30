package v1

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/sources"
	"github.com/mcay23/hound/view"

	"github.com/gin-gonic/gin"
)

type CommentRequest struct {
	CommentType   string `json:"comment_type" binding:"required,gt=0"` // review, etc.
	SeasonNumber  *int   `json:"season_number"`                        // only for tvshows, when commenting on a particular episode
	EpisodeNumber *int   `json:"episode_number"`
	CommentTitle  string `json:"title"`
	Comment       string `json:"comment"` // actual content of comment, review
	Score         int    `json:"score"`   // only required for reviews
}

// @Router /api/v1/tv/{id}/comments [get]
// @Summary Get comments for a TV show
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path string true "Media ID" example(tmdb-1234)
// @Param type query string true "Comment Type"
// @Success 200 {object} V1SuccessResponse{data=[]view.CommentObject}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetCommentsTVHandler(c *gin.Context) {
	handleGetComments(c, database.RecordTypeTVShow)
}

// @Router /api/v1/movie/{id}/comments [get]
// @Summary Get comments for a movie
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path string true "Media ID" example(tmdb-1234)
// @Param type query string true "Comment Type"
// @Success 200 {object} V1SuccessResponse{data=[]view.CommentObject}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetCommentsMovieHandler(c *gin.Context) {
	handleGetComments(c, database.RecordTypeMovie)
}

func handleGetComments(c *gin.Context, recordType string) {
	mediaSource, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	var seasonNumber, episodeNumber int
	if recordType == database.RecordTypeTVShow {
		var err error
		seasonNumber, episodeNumber, err = getSeasonEpisode(c.Query("season_number"), c.Query("episode_number"))
		if err == nil {
			recordType = database.RecordTypeEpisode
		}
	}
	commentType := c.Query("type")
	if commentType == "" {
		internal.ErrorResponse(c, fmt.Errorf("invalid type param: %w", internal.BadRequestError))
		return
	}
	var record *database.MediaRecord
	if recordType == database.RecordTypeEpisode {
		record, err = database.GetEpisodeMediaRecord(sources.MediaSourceTMDB, strconv.Itoa(sourceID), &seasonNumber, &episodeNumber)
		if err != nil {
			if errors.Is(err, internal.NotFoundError) {
				internal.SuccessResponse(c, nil, 200)
				return
			}
			internal.ErrorResponse(c, fmt.Errorf("failed to get media record: %w", err))
			return
		}
	} else {
		var has bool
		has, record, err = database.GetMediaRecord(recordType, mediaSource, strconv.Itoa(sourceID))
		if err != nil {
			internal.ErrorResponse(c, fmt.Errorf("failed to get media record: %w", err))
			return
		}
		if !has {
			internal.SuccessResponse(c, nil, 200)
			return
		}
	}
	if record == nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get media record: %w", internal.InternalServerError))
		return
	}
	comments, err := database.GetComments(record.RecordID, &commentType)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("error retrieving comments: %w", internal.InternalServerError))
		return
	}
	var commentsView []view.CommentObject
	for _, item := range *comments {
		commenter, _ := database.GetUser(item.UserID)
		if !item.IsPublic && c.GetString("username") != commenter.Username {
			continue
		}
		comment := view.CommentObject{
			CommentID:    item.CommentID,
			CommentType:  item.CommentType,
			UserID:       commenter.Username,
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
	internal.SuccessResponse(c, commentsView, 200)
}

// @Router /api/v1/tv/{id}/comments [post]
// @Summary Post a comment for a TV show
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path string true "Media ID" example(tmdb-1234)
// @Param comment body CommentRequest true "Comment"
// @Success 200 {object} V1SuccessResponse{data=object}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func PostCommentTVHandler(c *gin.Context) {
	handlePostComment(c, database.RecordTypeTVShow)
}

// @Router /api/v1/movie/{id}/comments [post]
// @Summary Post a comment for a movie
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path string true "Media ID" example(tmdb-1234)
// @Param comment body CommentRequest true "Comment"
// @Success 200 {object} V1SuccessResponse{data=object}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func PostCommentMovieHandler(c *gin.Context) {
	handlePostComment(c, database.RecordTypeMovie)
}

func handlePostComment(c *gin.Context, recordType string) {
	var body CommentRequest
	err := c.ShouldBindJSON(&body)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	mediaSource, sourceID, err := getSourceIDFromParams(c.Param("id"))
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	if body.CommentType == database.CommentTypeReview && (body.Score > 100 || body.Score < 0) {
		internal.ErrorResponse(c, fmt.Errorf("invalid score %d not (0<=score<=100): %w", body.Score, internal.BadRequestError))
		return
	}
	// get userID
	userID, err := getUserIDFromContext(c)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	// upsert top level record
	record, err := sources.UpsertMediaRecordTMDB(recordType, sourceID)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to upsert media record: %w", err))
		return
	}
	// if season and episode is specified, use episode record
	if recordType == database.RecordTypeTVShow && body.SeasonNumber != nil && body.EpisodeNumber != nil {
		record, err = database.GetEpisodeMediaRecord(mediaSource, strconv.Itoa(sourceID), body.SeasonNumber, body.EpisodeNumber)
		if err != nil {
			internal.ErrorResponse(c, fmt.Errorf("failed to get episode media record: %w", err))
			return
		}
	}
	isPublic := false
	if body.CommentType == database.CommentTypeNote {
		isPublic = true
	}
	comment := database.CommentRecord{
		UserID:       userID,
		CommentTitle: body.CommentTitle,
		IsPublic:     isPublic,
		CommentType:  body.CommentType,
		Comment:      body.Comment,
		Score:        body.Score,
		RecordID:     record.RecordID,
	}
	err = database.AddComment(&comment)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to add comment: %w", err))
		return
	}
	internal.SuccessResponse(c, gin.H{"status": "success", "comment_id": comment.CommentID}, 200)
}

// @Router /api/v1/comments/{id} [delete]
// @Summary Delete a comment
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Success 200 {object} V1SuccessResponse{data=object}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func DeleteCommentHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	// for batch deletion, split query params /comment?ids=1,2,3
	idSplit := strings.Split(c.Query("ids"), ",")
	if c.Query("ids") != "" {
		var batchIDs []int64
		for _, item := range idSplit {
			tempID, err := strconv.Atoi(item)
			if err != nil {
				internal.ErrorResponse(c, internal.LogErrorWithMessage(err, "Batch deletion: Invalid comment id in url query"))
				return
			}
			batchIDs = append(batchIDs, int64(tempID))
		}
		err = database.DeleteCommentBatch(userID, batchIDs)
		if err != nil {
			internal.ErrorResponse(c, err)
			return
		}
	} else if c.Param("id") != "" {
		// single delete case
		commentID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			internal.ErrorResponse(c, internal.LogErrorWithMessage(err, "Invalid comment id in url param"))
			return
		}
		err = database.DeleteComment(userID, int64(commentID))
		if err != nil {
			internal.ErrorResponse(c, err)
			return
		}
	} else {
		internal.ErrorResponse(c, internal.LogErrorWithMessage(err, "Invalid comment id in url param/query"))
		return
	}
	internal.SuccessResponse(c, nil, 200)
}
