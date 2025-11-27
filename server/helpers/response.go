package helpers

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func SuccessResponse(c *gin.Context, payload interface{}, statusCode int) {
	c.JSON(statusCode, payload)
}

func ErrorResponse(c *gin.Context, err error) {
	c.AbortWithStatusJSON(GetErrorStatusCode(err), gin.H{"error": err.Error()})
}

// data can be strings, or others
func ErrorResponseWithMessage(c *gin.Context, err error, data interface{}) {
	LogErrorWithMessage(err, fmt.Sprint(data))
	c.AbortWithStatusJSON(GetErrorStatusCode(err), gin.H{"error": data})
}
