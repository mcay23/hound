package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/sources"
)

func GetLiveCategoriesHandler(c *gin.Context) {
	cats, err := sources.GetLiveCategoriesIPTV()
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, cats, 200)
}

func GetLiveChannelsHandler(c *gin.Context) {
	categoryID := c.Query("category_id")
	channels, err := sources.GetLiveChannelsIPTV(categoryID)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, channels, 200)
}
