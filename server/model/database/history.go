package database

import "time"

const (
	// for comments, notes, reviews
	historyTable = "comment"
)

type HistoryRecord struct {
	HistoryID int64     `xorm:"pk autoincr 'history_id'" json:"id"`
	UserID    int64     `xorm:"'user_id'" json:"user_id"`
	RecordID  int64     `xorm:"'record_id'" json:"record_id"`
	TagData   string    `json:"tag_data"` // extra tag info, eg. season, episode
	CreatedAt time.Time `xorm:"created" json:"created_at"`
}
