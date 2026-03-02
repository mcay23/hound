package v1

import (
	"hound/database"
	"hound/helpers"
	"hound/view"

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
	collectionView, err := getHoundDownloadedRecords(limit, offset)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to get hound downloaded records"))
		return
	}
	helpers.SuccessResponse(c, collectionView, 200)
}

func getHoundDownloadedRecords(limit int, offset int) (view.CollectionView, error) {
	records, total_records, err := database.GetDownloadedParentRecords(limit, offset)
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
