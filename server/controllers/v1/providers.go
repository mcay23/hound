package v1

import (
	"github.com/gin-gonic/gin"
	"hound/helpers"
	"hound/providers"
)

func SearchProvidersHandler(c *gin.Context) {
	res, err := providers.SearchProviders(nil)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to search providers")
	}
	helpers.SuccessResponse(c, gin.H{"status": "success", "data": res}, 200)
}
