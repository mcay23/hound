package database

import (
	"fmt"
	"hound/helpers"
	"time"
)

const (
	// for comments, notes, reviews
	commentsTable      = "comments"
	CommentTypeReview  = "review"
	CommentTypeNote    = "note"
	CommentTypeComment = "comment"
)

/*
"Comments" include watch history, not just comments/reviews
*/
type CommentRecord struct {
	CommentID    int64     `xorm:"pk autoincr 'comment_id'" json:"id"`
	CommentType  string    `json:"comment_type"`
	UserID       int64     `xorm:"'user_id'" json:"user_id"`
	RecordID     int64     `xorm:"index 'record_id'" json:"record_id"`
	CommentTitle string    `xorm:"'title'" json:"title"`
	Comment      string    `xorm:"text 'comment'" json:"comment"` // actual content of comment, review
	IsPublic     bool      `xorm:"'is_public'" json:"is_public"`
	Score        int       `xorm:"'score'" json:"score"` // for review types
	CreatedAt    time.Time `xorm:"timestampz created" json:"created_at"`
	UpdatedAt    time.Time `xorm:"timestampz updated" json:"updated_at"`
}

func instantiateCommentTable() error {
	err := databaseEngine.Table(commentsTable).Sync2(new(CommentRecord))
	if err != nil {
		return err
	}
	return nil
}

func AddComment(comment *CommentRecord) error {
	if comment.CommentType != CommentTypeReview && comment.CommentType != CommentTypeComment &&
		comment.CommentType != CommentTypeNote {
		return fmt.Errorf("invalid comment type %s: %w", comment.CommentType, helpers.BadRequestError)
	}
	_, err := databaseEngine.Table(commentsTable).Insert(comment)
	return err
}

func AddCommentsBatch(comments *[]CommentRecord) error {
	_, err := databaseEngine.Table(commentsTable).Insert(comments)
	return err
}

func GetComments(recordID int64, commentType *string) (*[]CommentRecord, error) {
	var comments []CommentRecord
	sess := databaseEngine.Table(commentsTable).Where("record_id = ?", recordID).OrderBy("updated_at desc")
	if commentType != nil && *commentType != "" {
		sess.Where("comment_type = ?", commentType)
	}
	err := sess.Find(&comments)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments for recordID %d: %w", recordID, err)
	}
	return &comments, nil
}

func DeleteComment(userID int64, commentID int64) error {
	affected, err := databaseEngine.Table(commentsTable).Delete(&CommentRecord{
		UserID:    userID,
		CommentID: commentID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete commentid %d: %w", commentID, err)
	}
	if affected <= 0 {
		return fmt.Errorf("no comment found with userID %d, commentID %d: %w", userID, commentID, helpers.NotFoundError)
	}
	return nil
}

func DeleteCommentBatch(userID int64, commentIDs []int64) error {
	session := databaseEngine.NewSession()
	defer session.Close()
	_ = session.Begin()
	for _, item := range commentIDs {
		affected, err := session.Table(commentsTable).Delete(&CommentRecord{UserID: userID, CommentID: item})
		if err != nil {
			session.Rollback()
			return err
		}
		if affected <= 0 {
			session.Rollback()
			return fmt.Errorf("no comment found with userID %d, commentID %d: %w", userID, item, helpers.NotFoundError)
		}
	}
	err := session.Commit()
	return err
}
