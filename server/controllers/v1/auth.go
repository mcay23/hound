package v1

import (
	"fmt"
	"hound/helpers"
	"hound/model"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type RegistrationResponse struct {
	Username string `json:"username"`
}

// @Router /api/v1/auth/register [post]
// @Summary Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.RegistrationUser true "Registration Details"
// @Success 200 {object} V1SuccessResponse{data=RegistrationResponse}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func RegistrationHandler(c *gin.Context) {
	if !viper.GetBool("auth.allow-registration") {
		err := fmt.Errorf("%w: Registration is currently disabled. Please contact your system admin.", helpers.BadRequestError)
		helpers.ErrorResponse(c, err)
		return
	}
	userPayload := model.RegistrationUser{}
	if err := c.ShouldBindJSON(&userPayload); err != nil {
		err := fmt.Errorf("%w: Failed to bind registration body", helpers.BadRequestError)
		helpers.ErrorResponse(c, err)
		return
	}
	err := model.RegisterNewUser(&userPayload, false)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	clientID, clientPlatform, err := validateClientHeaders(c)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	tokenPayload := model.LoginUser{
		Username: userPayload.Username,
		Password: userPayload.Password,
	}
	token, err := model.GenerateAccessToken(tokenPayload, clientID, clientPlatform)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	c.SetCookie("token", token, viper.GetInt("auth.jwt-access-token-expiration"), "/", "", true, true)
	helpers.SuccessResponse(c, RegistrationResponse{Username: userPayload.Username}, 200)
}

type LoginResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
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
		helpers.ErrorResponse(c, err)
		return
	}
	clientID, clientPlatform, err := validateClientHeaders(c)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	token, err := model.GenerateAccessToken(userPayload, clientID, clientPlatform)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	cookie := &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		MaxAge:   viper.GetInt("auth.jwt-access-token-expiration"),
		HttpOnly: true,
		Secure:   c.Request.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(c.Writer, cookie)
	helpers.SuccessResponse(c, LoginResponse{Username: userPayload.Username, Token: token}, 200)
}

func validateClientHeaders(c *gin.Context) (string, string, error) {
	clientID := strings.ToLower(c.GetHeader("X-Client-Id"))
	if !slices.Contains(model.SupportedClientIDs, clientID) {
		return "", "", fmt.Errorf("%w: Invalid or missing X-Client-Id header", helpers.BadRequestError)
	}
	clientPlatform := strings.ToLower(c.GetHeader("X-Client-Platform"))
	if !slices.Contains(model.SupportedClientPlatforms, clientPlatform) {
		return "", "", fmt.Errorf("%w: Invalid or missing X-Client-Platform header", helpers.BadRequestError)
	}
	return clientID, clientPlatform, nil
}
