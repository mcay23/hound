package middlewares

import (
	"errors"
	"hound/helpers"
	"hound/model"
	"strings"

	"github.com/gin-gonic/gin"
)

func extractBearerToken(header string) (string, error) {
	if header == "" {
		return "", helpers.LogErrorWithMessage(errors.New(helpers.Unauthorized), "No auth token in header")
	}
	jwtToken := strings.Split(header, " ")
	if len(jwtToken) != 2 {
		return "", helpers.LogErrorWithMessage(errors.New(helpers.Unauthorized), "Invalid header token")
	}
	return jwtToken[1], nil
}

func JWTMiddleware(c *gin.Context) {
	jwtToken, err := c.Cookie("token")
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Cookie not found, checking auth header")
		jwtToken, err = extractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			helpers.ErrorResponse(c, err)
			return
		}
	}
	claims, err := model.ParseAccessToken(jwtToken)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	// set headers from auth token, overwrite current headers
	//c.Request.Header.Del("X-Username")
	//c.Request.Header.Del("X-Client")
	c.Request.Header.Add("X-Username", claims.Username)
	c.Request.Header.Add("X-Client", claims.Client)
	c.Next()
}

func CORSMiddleware(c *gin.Context) {
	origin := c.Request.Header.Get("Origin")
	if origin != "" {
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Vary", "Origin")
	}
	// c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "User-Agent, Content-Type, Content-Length, Accept-Ranges, Content-Range, Range, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Client")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, HEAD")
	c.Writer.Header().Set("Accept-Ranges", "bytes")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}
	c.Next()
}
