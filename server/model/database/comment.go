package database

import (
	"errors"
	"hound/helpers"
	"time"
)

const (
	// for comments, notes, reviews
	commentsTable      = "comments"
	commentTypeReview  = "review"
	commentTypeNote    = "note"
	commentTypeComment = "comment"
	// watch history, play history, etc.
	commentTypeHistory = "history"
)

type CommentRecord struct {
	CommentID    int64     `xorm:"pk autoincr 'comment_id'" json:"id"`
	CommentType  string    `json:"comment_type"`
	UserID       int64     `xorm:"'user_id'" json:"user_id"`
	LibraryID    int64     `xorm:"'library_id'" json:"library_id"`
	IsPrivate    bool      `json:"is_private"`
	CommentTitle string    `json:"title"`
	Comment      []byte    `json:"comment"`  // actual content of comment, review
	TagData      string    `json:"tag_data"` // extra tag info, eg. season, episode
	Score        int       `json:"score"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	CreatedAt    time.Time `xorm:"created" json:"created_at"`
	UpdatedAt    time.Time `xorm:"updated" json:"updated_at"`
}

func instantiateCommentTable() error {
	err := databaseEngine.Table(commentsTable).Sync2(new(CommentRecord))
	if err != nil {
		return err
	}
	return nil
}

func AddComment(comment *CommentRecord) error {
	if comment.CommentType != commentTypeReview && comment.CommentType != commentTypeComment && comment.CommentType != commentTypeNote {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid comment type")
	}
	_, err := databaseEngine.Table(commentsTable).Insert(comment)
	return err
}

func GetComments(libraryID int64) (*[]CommentRecord, error) {
	var comments []CommentRecord
	err := databaseEngine.Table(commentsTable).Where("library_id = ?", libraryID).OrderBy("updated_at desc").Find(&comments)
	if err != nil {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "GetComments(): Failed to get comments")
	}
	return &comments, nil
}

func DeleteComment(userID int64, commentID int64) error {
	affected, err := databaseEngine.Table(commentsTable).Delete(&CommentRecord{
		UserID:    userID,
		CommentID: commentID,
	})
	if err != nil {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "DeleteComment(): Failed to delete comments")
	}
	if affected <= 0 {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "GetComments(): No comment found with this ID or invalid user")
	}
	return nil
}
