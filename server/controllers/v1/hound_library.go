package v1

import (
	"errors"
	"hound/database"
	"hound/helpers"
	"hound/view"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetHoundLibraryHandler(c *gin.Context) {
	limitQuery := c.Query("limit")
	offsetQuery := c.Query("offset")
	limit, offset, err := getLimitOffset(limitQuery, offsetQuery)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	mediaType := ""
	mediaTypeQuery := c.Query("media_type")
	if mediaTypeQuery != "" {
		switch mediaTypeQuery {
		case database.MediaTypeMovie:
			mediaType = database.MediaTypeMovie
		case database.MediaTypeTVShow:
			mediaType = database.MediaTypeTVShow
		default:
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"Invalid media type, needs to be 'movie' or 'tvshow'"))
			return
		}
	}
	// multiple genre_id params, get records with at least of the genres
	var genreIDs []int64
	genreIDQueries := c.QueryArray("genre_id")
	for _, idStr := range genreIDQueries {
		if idStr != "" {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				genreIDs = append(genreIDs, id)
			}
		}
	}
	collectionView, err := getHoundDownloadedRecords(limit, offset, mediaType, genreIDs)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to get hound downloaded records"))
		return
	}
	helpers.SuccessResponse(c, collectionView, 200)
}

func getHoundDownloadedRecords(limit int, offset int, mediaType string, genreIDs []int64) (view.CollectionView, error) {
	records, total_records, err := database.GetDownloadedParentRecords(limit, offset, mediaType, genreIDs)
	if err != nil {
		return view.CollectionView{}, helpers.LogErrorWithMessage(err, "Failed to get downloaded records")
	}
	var viewArray []view.MediaRecordCatalog
	for _, item := range records {
		viewObject := createMediaRecordCatalogObject(item)
		viewArray = append(viewArray, viewObject)
	}
	collectionView := view.CollectionView{
		Records: viewArray,
		Collection: &view.CollectionObject{
			CollectionID:    -1,
			CollectionTitle: "Hound Library",
			Description:     "Downloaded Content in Hound",
			OwnerUsername:   "Hound",
			IsPublic:        true,
			ThumbnailURI:    "",
		},
		TotalRecords: total_records,
		Limit:        limit,
		Offset:       offset,
	}
	return collectionView, nil
}
