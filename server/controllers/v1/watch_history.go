package v1

import (
	"fmt"
	"hound/helpers"
	"hound/model/sources"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Only episode ids that belong to the same show should be inserted at the same time
type AddWatchHistoryPayload struct {
	EpisodeIDs []string `json:"episode_id" binding:"required,gt=0"` // tmdb-120390, use tmdb unique id for episode
}

func AddWatchHistoryTVShowHandler(c *gin.Context) {
	username := c.GetHeader("X-Username")
	watchHistoryPayload := AddWatchHistoryPayload{}
	if err := c.ShouldBindJSON(&watchHistoryPayload); err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to bind watch history body: "+c.Param("id")))
		return
	}
	// 1. Parse episode ids
	mediaSource, showID, err := GetSourceIDFromParams(c.Param("id"))
	if err != nil || mediaSource != sources.SourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error parsing source_id: "+c.Param("id")))
		return
	}
	var episodeIDs []string
	for _, item := range watchHistoryPayload.EpisodeIDs {
		_, id, err := GetSourceIDFromParams(item)
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error parsing episode IDs: "+c.Param("id")))
			return
		}
		episodeIDs = append(episodeIDs, strconv.Itoa(id))
	}
	// 2. Upsert show
	showRecord, err := sources.UpsertTVShowRecordTMDB(showID)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error upserting tv show: "+c.Param("id")))
		return
	}
	fmt.Println(showRecord, username)
	helpers.SuccessResponse(c, gin.H{"status": "success"}, 200)
	// 3. get or create latest rewatch

	// 4. get episode record ID

}
