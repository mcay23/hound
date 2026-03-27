package v1

import (
	"fmt"
	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/helpers"
	"github.com/mcay23/hound/view"
	"strconv"

	"github.com/gin-gonic/gin"
)

// @Router /api/v1/collection/hound-library [get]
// @Summary Get Hound Library
// @Description Get content downloaded to Hound
// @Tags Collections
// @Accept json
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param media_type query string false "Media Type eg. tvshow or movie"
// @Param genre_id query []int false "Genre IDs"
// @Success 200 {object} V1SuccessResponse{data=view.CollectionView}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
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
			helpers.ErrorResponse(c, fmt.Errorf("invalid media type %s: %w", mediaTypeQuery, helpers.BadRequestError))
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
		helpers.ErrorResponse(c, fmt.Errorf("failed to get hound downloaded records: %w", err))
		return
	}
	helpers.SuccessResponse(c, collectionView, 200)
}

func getHoundDownloadedRecords(limit int, offset int, mediaType string, genreIDs []int64) (view.CollectionView, error) {
	records, total_records, err := database.GetDownloadedParentRecords(limit, offset, mediaType, genreIDs)
	if err != nil {
		return view.CollectionView{}, fmt.Errorf("failed to get downloaded records: %w", err)
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
