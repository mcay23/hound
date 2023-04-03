package database

import (
	"errors"
	"hound/helpers"
	"time"
)

const (
	// for comments, notes, reviews
	commentTable       = "comment"
	commentTypeReview  = "review"
	commentTypeNote    = "note"
	commentTypeComment = "comment"
)

type CommentRecord struct {
	CommentID   int64     `xorm:"pk autoincr 'comment_id'" json:"id"`
	CommentType string    `json:"comment_type"`
	UserID      int64     `xorm:"'user_id'" json:"user_id"`
	LibraryID   int64     `xorm:"'library_id'" json:"library_id"`
	IsPrivate   bool      `json:"is_private"`
	Comment     []byte    `json:"comment"`  // actual content of comment, review
	TagData     string    `json:"tag_data"` // extra tag info, eg. season, episode
	Score       int       `json:"score"`
	CreatedAt   time.Time `xorm:"created" json:"created_at"`
	UpdatedAt   time.Time `xorm:"updated" json:"updated_at"`
}

func instantiateCommentTable() error {
	err := databaseEngine.Table(commentTable).Sync2(new(CommentRecord))
	if err != nil {
		return err
	}
	return nil
}

func AddComment(comment *CommentRecord) error {
	if comment.CommentType != commentTypeReview && comment.CommentType != commentTypeComment && comment.CommentType != commentTypeNote {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid comment type")
	}
	_, err := databaseEngine.Table(commentTable).Insert(comment)
	return err
}
