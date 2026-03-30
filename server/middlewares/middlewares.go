package middlewares

import (
	"fmt"
	"strings"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/model"

	"github.com/gin-gonic/gin"
)

func extractBearerToken(header string) (string, error) {
	if header == "" {
		return "", fmt.Errorf("no auth token in header: %w", internal.UnauthorizedError)
	}
	jwtToken := strings.Split(header, " ")
	if len(jwtToken) != 2 {
		return "", fmt.Errorf("invalid header token: %w", internal.UnauthorizedError)
	}
	return jwtToken[1], nil
}

func AuthMiddleware(c *gin.Context) {
	apiKey := c.GetHeader("X-Api-Key")
	// API Key case
	if apiKey != "" {
		key, err := model.ValidateAPIKey(apiKey)
		if err != nil {
			internal.ErrorResponse(c, err)
			return
		}
		user, err := database.GetUser(key.UserID)
		if err != nil {
			internal.ErrorResponse(c, fmt.Errorf("failed to get user: %w", err))
			return
		}
		role := "user"
		if user.IsAdmin {
			role = "admin"
		}
		c.Set("userID", user.UserID)
		c.Set("clientID", "api")
		c.Set("clientPlatform", "api")
		c.Set("role", role)
	} else {
		// Access Token case
		jwtToken, err := c.Cookie("token")
		if err != nil {
			jwtToken, err = extractBearerToken(c.GetHeader("Authorization"))
			if err != nil {
				// no auth provided
				internal.ErrorResponse(c, err)
				return
			}
		}
		claims, err := model.ParseAccessToken(jwtToken)
		if err != nil {
			internal.ErrorResponse(c, err)
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("clientID", claims.ClientID)
		c.Set("clientPlatform", claims.ClientPlatform)
		c.Set("role", claims.Role)
	}
	c.Next()
}

func AdminMiddleware(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
		internal.ErrorResponse(c, fmt.Errorf("admin role required: %w", internal.UnauthorizedError))
		return
	}
	c.Next()
}

func CORSMiddleware(c *gin.Context) {
	// origin := c.Request.Header.Get("Origin")
	// if origin != "" {
	// 	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	// 	c.Writer.Header().Set("Vary", "Origin")
	// }
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "User-Agent, Content-Type, Content-Length, Accept-Ranges, Content-Range, Range, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Client-Id, X-Client-Platform")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, HEAD")
	c.Writer.Header().Set("Accept-Ranges", "bytes")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}
	c.Next()
}
