package helpers

import (
	"github.com/gin-gonic/gin"
)

func SuccessResponse(c *gin.Context, payload interface{}, statusCode int) {
	c.JSON(statusCode, payload)
}

func ErrorResponse(c *gin.Context, err error) {
	c.AbortWithStatusJSON(GetErrorStatusCode(err), gin.H{"error": err.Error()})
}
