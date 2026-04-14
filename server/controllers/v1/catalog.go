package v1

import (
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/model"

	"github.com/gin-gonic/gin"
)

// @Router /api/v1/catalog/{id} [get]
// @Summary Get Catalog
// @ID get-catalog
// @Tags Catalog
// @Accept json
// @Produce json
// @Param id path string true "Catalog ID"
// @Success 200 {object} V1SuccessResponse{data=[]view.MediaRecordCatalog}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetCatalogHandler(c *gin.Context) {
	idParam := c.Param("id")
	catalogID := idParam
	// lock to page 1 for now
	page := 1
	viewArray, err := model.GetInternalCatalog(catalogID, &page)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, viewArray, 200)
}
