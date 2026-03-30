package v1

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/model"
)

type CreateAPIKeyRequest struct {
	Name      string `json:"name"`
	ExpiresAt string `json:"expires_at"`
}

// @Router /api/v1/api_keys [get]
// @Summary Get User API Keys
// @Tags API Keys
// @Accept json
// @Produce json
// @Success 200 {object} V1SuccessResponse{data=[]database.APIKey}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetUserAPIKeysHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	apiKeys, err := database.GetUserAPIKeys(userID)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, apiKeys, 200)
}

// @Router /api/v1/api_keys [post]
// @Summary Create API Key
// @Tags API Keys
// @Accept json
// @Produce json
// @Param name query string true "API Key Name"
// @Success 200 {object} V1SuccessResponse{data=database.APIKey}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func CreateAPIKeyHandler(c *gin.Context) {
	var request CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	userID, err := getUserIDFromContext(c)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	var expiresAt *time.Time
	if request.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, request.ExpiresAt)
		if err != nil {
			internal.ErrorResponse(c, err)
			return
		}
		expiresAt = &t
	}
	if request.Name == "" {
		request.Name = "API Key"
	}
	record, err := model.CreateAPIKey(userID, request.Name, expiresAt)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, record, 200)
}

// @Router /api/v1/api_keys/{id} [delete]
// @Summary Revoke API Key
// @Tags API Keys
// @Accept json
// @Produce json
// @Param id path string true "API Key ID"
// @Success 200 {object} V1SuccessResponse
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func RevokeAPIKeyHandler(c *gin.Context) {
	keyIDStr := c.Param("id")
	if keyIDStr == "" {
		internal.ErrorResponse(c, fmt.Errorf("key id is required: %w", internal.UnauthorizedError))
		return
	}
	keyID, err := strconv.ParseInt(keyIDStr, 10, 64)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("invalid key id: %w", internal.UnauthorizedError))
		return
	}
	err = database.RevokeAPIKey(keyID)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, nil, 200)
}
