package v1

import (
	"hound/helpers"
	"hound/model"

	"github.com/gin-gonic/gin"
)

func GetCatalogHandler(c *gin.Context) {
	idParam := c.Param("id")
	catalogID := idParam
	// lock to page 1 for now
	page := 1
	viewArray, err := model.GetInternalCatalog(catalogID, &page)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	helpers.SuccessResponse(c, viewArray, 200)
}
