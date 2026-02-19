package database

import "time"

const (
	externalLibraryItemsTable = "external_library_items"
)

const (
	ExternalLibraryItemStatusPending = "pending"
	ExternalLibraryItemStatusQueued  = "queued"
	ExternalLibraryItemStatusDone    = "done"
	ExternalLibraryItemStatusFailed  = "failed"
)

type ExternalLibraryItem struct {
	ItemID           int64      `xorm:"pk autoincr 'item_id'" json:"item_id"`
	RootPath         string     `xorm:"text 'root_path'" json:"root_path"`
	SourcePath       string     `xorm:"unique not null text 'source_path'" json:"source_path"`
	MediaType        string     `xorm:"'media_type'" json:"media_type"`
	MediaSource      string     `xorm:"'media_source'" json:"media_source"`
	SourceID         string     `xorm:"'source_id'" json:"source_id"`
	SeasonNumber     *int       `xorm:"'season_number'" json:"season_number,omitempty"`
	EpisodeNumber    *int       `xorm:"'episode_number'" json:"episode_number,omitempty"`
	FileSize         int64      `xorm:"'file_size'" json:"file_size"`
	ModifiedUnix     int64      `xorm:"'modified_unix'" json:"modified_unix"`
	Status           string     `xorm:"index 'status'" json:"status"`
	LastError        *string    `xorm:"text 'last_error'" json:"last_error"`
	LastIngestTaskID *int64     `xorm:"'last_ingest_task_id'" json:"last_ingest_task_id"`
	LastQueuedAt     *time.Time `xorm:"timestampz 'last_queued_at'" json:"last_queued_at"`
	LastCompletedAt  *time.Time `xorm:"timestampz 'last_completed_at'" json:"last_completed_at"`
	CreatedAt        time.Time  `xorm:"timestampz created" json:"created_at"`
	UpdatedAt        time.Time  `xorm:"timestampz updated" json:"updated_at"`
}

func instantiateExternalLibraryItemsTable() error {
	return databaseEngine.Table(externalLibraryItemsTable).Sync2(new(ExternalLibraryItem))
}

func GetExternalLibraryItemByPath(sourcePath string) (*ExternalLibraryItem, error) {
	var item ExternalLibraryItem
	has, err := databaseEngine.Table(externalLibraryItemsTable).
		Where("source_path = ?", sourcePath).
		Get(&item)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return &item, nil
}

func UpsertExternalLibraryItem(item *ExternalLibraryItem) error {
	if item == nil {
		return nil
	}
	var existing ExternalLibraryItem
	has, err := databaseEngine.Table(externalLibraryItemsTable).
		Where("source_path = ?", item.SourcePath).
		Get(&existing)
	if err != nil {
		return err
	}
	if !has {
		_, err = databaseEngine.Table(externalLibraryItemsTable).Insert(item)
		return err
	}
	item.ItemID = existing.ItemID
	_, err = databaseEngine.Table(externalLibraryItemsTable).
		Where("item_id = ?", existing.ItemID).
		AllCols().
		Update(item)
	return err
}
