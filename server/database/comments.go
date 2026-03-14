package database

import (
	"fmt"
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

/*
"Comments" include watch history, not just comments/reviews
*/
type CommentRecord struct {
	CommentID    int64     `xorm:"pk autoincr 'comment_id'" json:"id"`
	CommentType  string    `json:"comment_type"`
	UserID       int64     `xorm:"'user_id'" json:"user_id"`
	RecordID     int64     `xorm:"index 'record_id'" json:"record_id"`
	IsPrivate    bool      `xorm:"'is_private'" json:"is_private"`
	CommentTitle string    `xorm:"'title'" json:"title"`
	Comment      string    `xorm:"text 'comment'" json:"comment"` // actual content of comment, review
	TagData      string    `xorm:"'tag_data'" json:"tag_data"`    // extra tag info, eg. season, episode
	Score        int       `xorm:"'score'" json:"score"`
	StartDate    time.Time `xorm:"timestampz 'start_date'" json:"start_date"`
	EndDate      time.Time `xorm:"timestampz 'end_date'" json:"end_date"`
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
	if comment.CommentType != commentTypeReview && comment.CommentType != commentTypeComment &&
		comment.CommentType != commentTypeNote && comment.CommentType != commentTypeHistory {
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
	sess := databaseEngine.Table(commentsTable).Where("record_id = ?", recordID)
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
