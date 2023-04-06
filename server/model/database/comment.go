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
	if comment.CommentType != commentTypeReview && comment.CommentType != commentTypeComment &&
		comment.CommentType != commentTypeNote && comment.CommentType != commentTypeHistory {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid comment type")
	}
	_, err := databaseEngine.Table(commentsTable).Insert(comment)
	return err
}

func AddCommentsBatch(comments *[]CommentRecord) error {
	_, err := databaseEngine.Table(commentsTable).Insert(comments)
	return err
}

func GetComments(libraryID int64, commentType *string) (*[]CommentRecord, error) {
	var comments []CommentRecord
	sess := databaseEngine.Table(commentsTable).Where("library_id = ?", libraryID)
	if *commentType == commentTypeHistory {
		sess = sess.OrderBy("start_date desc")
	} else {
		sess = sess.OrderBy("updated_at desc")
	}
	if commentType != nil && *commentType != "" {
		sess.Where("comment_type = ?", commentType)
	}
	err := sess.Find(&comments)
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
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "DeleteComment(): No comment found with this ID or invalid user")
	}
	return nil
}

func DeleteCommentBatch(userID int64, commentIDs []int64) error {
	session := databaseEngine.NewSession()
	defer session.Close()
	_ = session.Begin()
	for _, item := range commentIDs {
		affected, err := session.Table(commentsTable).Delete(&CommentRecord{UserID: userID, CommentID: item})
		if err != nil || affected <= 0 {
			_ = session.Rollback()
			return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "DeleteCommentBatch(): No comment found with this ID or invalid user")
		}
	}
	_ = session.Commit()
	return nil
}
