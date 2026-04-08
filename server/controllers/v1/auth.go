package v1

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/model"

	"github.com/gin-gonic/gin"
	"github.com/mcay23/hound/middlewares"
)

// @Router /api/v1/users [post]
// @Summary Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.RegistrationUser true "Registration Details"
// @Success 200 {object} V1SuccessResponse{data=database.User}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func RegistrationHandler(c *gin.Context) {
	userPayload := model.RegistrationUser{}
	if err := c.ShouldBindJSON(&userPayload); err != nil {
		err := fmt.Errorf("%w: Failed to bind registration body", internal.BadRequestError)
		internal.ErrorResponse(c, err)
		return
	}
	newUser, err := model.RegisterNewUser(&userPayload, false)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, newUser, 200)
}

type LoginResponse struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Token       string `json:"token"`
	Role        string `json:"role"`
}

// @Router /api/v1/auth/login [post]
// @Summary User login
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.LoginUser true "Login Details"
// @Success 200 {object} V1SuccessResponse{data=LoginResponse}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func LoginHandler(c *gin.Context) {
	userPayload := model.LoginUser{}
	if err := c.ShouldBindJSON(&userPayload); err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	clientID, clientPlatform, deviceID, err := validateClientHeaders(c)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	userID, err := database.GetUserIDFromUsername(userPayload.Username)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	token, displayName, role, err := model.AuthenticateUser(userID, userPayload.Password, clientID, clientPlatform, deviceID)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("Failed to generate access token: %w", err))
		return
	}
	cookie := &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		MaxAge:   int(database.AuthSessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   c.Request.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(c.Writer, cookie)
	internal.SuccessResponse(c,
		LoginResponse{
			Username:    userPayload.Username,
			DisplayName: displayName,
			Token:       token,
			Role:        role,
		}, 200)
}

// @Router /api/v1/auth/logout [post]
// @Summary User logout
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} V1SuccessResponse
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func LogoutHandler(c *gin.Context) {
	sessionID, err := c.Cookie("token")
	if err != nil {
		sessionID, err = middlewares.ExtractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			internal.ErrorResponse(c, err)
			return
		}
	}
	if sessionID == "" {
		internal.ErrorResponse(c, fmt.Errorf("no auth token: %w", internal.UnauthorizedError))
		return
	}
	userID, err := getUserIDFromContext(c)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get user ID: %w", internal.UnauthorizedError))
		return
	}
	err = database.DeleteAuthSession(userID, sessionID)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to delete auth session: %w", err))
		return
	}
	c.SetCookie("token", "", -1, "/", "", true, true)
	internal.SuccessResponse(c, nil, 200)
}

func validateClientHeaders(c *gin.Context) (string, string, string, error) {
	clientID := strings.ToLower(c.GetHeader("X-Client-Id"))
	if !slices.Contains(model.SupportedClientIDs, clientID) {
		return "", "", "", fmt.Errorf("%w: Invalid or missing X-Client-Id header", internal.BadRequestError)
	}
	clientPlatform := strings.ToLower(c.GetHeader("X-Client-Platform"))
	if !slices.Contains(model.SupportedClientPlatforms, clientPlatform) {
		return "", "", "", fmt.Errorf("%w: Invalid or missing X-Client-Platform header", internal.BadRequestError)
	}
	deviceID := c.GetHeader("X-Device-Id")
	if deviceID == "" {
		return "", "", "", fmt.Errorf("%w: Invalid or missing X-Device-Id header", internal.BadRequestError)
	}
	return clientID, clientPlatform, deviceID, nil
}
