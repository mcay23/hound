package v1

import (
	"errors"
	"hound/helpers"
	"hound/model"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func RegistrationHandler(c *gin.Context) {
	if !viper.GetBool("auth.allow-registration") {
		err := errors.New(helpers.BadRequest)
		_ = helpers.LogErrorWithMessage(err, "Registration is currently disabled. Please contact your system admin.")
		helpers.ErrorResponse(c, err)
		return
	}
	userPayload := model.RegistrationUser{}
	if err := c.ShouldBindJSON(&userPayload); err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to bind registration body")
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
	helpers.SuccessResponse(c, gin.H{"username": userPayload.Username}, 200)
}

func LoginHandler(c *gin.Context) {
	userPayload := model.LoginUser{}
	if err := c.ShouldBindJSON(&userPayload); err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to bind registration body")
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
	helpers.SuccessResponse(c, gin.H{"username": userPayload.Username, "token": token}, 200)
}

func validateClientHeaders(c *gin.Context) (string, string, error) {
	clientID := strings.ToLower(c.GetHeader("X-Client-Id"))
	if !slices.Contains(model.SupportedClientIDs, clientID) {
		err := errors.New(helpers.BadRequest)
		_ = helpers.LogErrorWithMessage(err, "Invalid or missing X-Client-Id header")
		return "", "", err
	}
	clientPlatform := strings.ToLower(c.GetHeader("X-Client-Platform"))
	if !slices.Contains(model.SupportedClientPlatforms, clientPlatform) {
		err := errors.New(helpers.BadRequest)
		_ = helpers.LogErrorWithMessage(err, "Invalid or missing X-Client-Platform header")
		return "", "", err
	}
	return clientID, clientPlatform, nil
}
