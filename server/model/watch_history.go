package model

import (
	"errors"
	"hound/database"
	"hound/helpers"
	"time"
)

// only allow if previous rewatch isn't empty
func InsertRewatchFromSourceID(recordType string, mediaSource string, sourceID string,
	userID int64, startedAt time.Time) (*database.RewatchRecord, error) {
	has, record, err := database.GetMediaRecord(recordType, mediaSource, sourceID)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"No Media Record Found for "+recordType+":"+mediaSource+"-"+sourceID)
	}
	// get active rewatch
	activeRewatch, err := database.GetActiveRewatchFromSourceID(recordType, mediaSource, sourceID, userID)
	if err != nil {
		return nil, err
	}
	// don't insert if no active rewatch, or active rewatch is empty
	if activeRewatch != nil {
		watchEvents, err := database.GetWatchEventsFromRewatchID(activeRewatch.RewatchID, nil)
		if err != nil {
			return nil, err
		}
		if len(watchEvents) == 0 {
			return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"Current rewatch is empty, can't create new rewatch")
		}
	}
	rewatch := database.RewatchRecord{
		UserID:    userID,
		RecordID:  record.RecordID,
		StartedAt: startedAt,
	}
	return database.InsertRewatch(rewatch)
}
