package middlewares

import (
	"fmt"
	"strings"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/model"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(c *gin.Context) {
	apiKey := c.GetHeader("X-Api-Key")
	// user call is cached, speed is vital
	var user *database.User
	// API Key case
	if apiKey != "" {
		key, err := model.ValidateAPIKey(apiKey)
		if err != nil {
			internal.ErrorResponse(c, fmt.Errorf("failed to validate API Key: %w", internal.UnauthorizedError))
			return
		}
		user, err = database.GetUser(key.UserID)
		if err != nil {
			internal.ErrorResponse(c, fmt.Errorf("failed to get user: %w", internal.UnauthorizedError))
			return
		}
		c.Set("userID", key.UserID)
		c.Set("clientID", "api")
		c.Set("clientPlatform", "api")
		c.Set("deviceID", "")
	} else {
		// Access Token case
		sessionID, err := c.Cookie("token")
		if err != nil {
			sessionID, err = ExtractBearerToken(c.GetHeader("Authorization"))
			if err != nil {
				// no auth provided
				internal.ErrorResponse(c, fmt.Errorf("failed to get auth token: %w", internal.UnauthorizedError))
				return
			}
		}
		sess, err := model.ParseAuthSession(sessionID)
		if err != nil {
			internal.ErrorResponse(c, fmt.Errorf("failed to parse auth session: %w", internal.UnauthorizedError))
			return
		}
		user, err = database.GetUser(sess.UserID)
		if err != nil {
			internal.ErrorResponse(c, fmt.Errorf("failed to get user: %w", internal.UnauthorizedError))
			return
		}
		c.Set("userID", sess.UserID)
		c.Set("clientID", sess.ClientID)
		c.Set("clientPlatform", sess.ClientPlatform)
		c.Set("deviceID", sess.DeviceID)
	}
	if user == nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get user: %w", internal.UnauthorizedError))
		return
	}
	if user.IsAdmin {
		c.Set("role", "admin")
	} else {
		c.Set("role", "user")
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
	c.Writer.Header().Set("Access-Control-Allow-Headers", "User-Agent, Content-Type, Content-Length, Accept-Ranges, Content-Range, Range, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Client-Id, X-Client-Platform, X-Device-Id")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, HEAD")
	c.Writer.Header().Set("Accept-Ranges", "bytes")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}
	c.Next()
}

func ExtractBearerToken(header string) (string, error) {
	if header == "" {
		return "", fmt.Errorf("no auth token in header: %w", internal.UnauthorizedError)
	}
	token := strings.Split(header, " ")
	if len(token) != 2 {
		return "", fmt.Errorf("invalid header token: %w", internal.UnauthorizedError)
	}
	return token[1], nil
}
