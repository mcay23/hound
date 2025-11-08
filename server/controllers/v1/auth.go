package v1

import (
	"errors"
	"hound/helpers"
	"hound/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func RegistrationHandler(c *gin.Context) {
	if !viper.GetBool("auth.allow-registration") {
		err := errors.New(helpers.BadRequest)
		_ = helpers.LogErrorWithMessage(err, "Registration is currently disabled. Please contact your system admin.")
		helpers.ErrorResponse(c, err)
	}
	userPayload := model.RegistrationUser{}
	if err := c.ShouldBindJSON(&userPayload); err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to bind registration body")
		helpers.ErrorResponse(c, err)
		return
	}
	err := model.RegisterNewUser(&userPayload)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	client := c.GetHeader("X-Client")
	tokenPayload := model.LoginUser{
		Username: userPayload.Username,
		Password: userPayload.Password,
	}
	token, err := model.GenerateAccessToken(tokenPayload, client)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	c.SetCookie("token", token, viper.GetInt("auth.jwt-access-token-expiration"), "/", "", true, true)
	helpers.SuccessResponse(c, gin.H{"status": "success", "username": userPayload.Username}, 200)
}

func LoginHandler(c *gin.Context) {
	userPayload := model.LoginUser{}
	if err := c.ShouldBindJSON(&userPayload); err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to bind registration body")
		helpers.ErrorResponse(c, err)
		return
	}
	client := c.GetHeader("X-Client")
	if client == "" {
		err := errors.New(helpers.BadRequest)
		_ = helpers.LogErrorWithMessage(err, "Failed to get client from header")
		helpers.ErrorResponse(c, err)
		return
	}
	token, err := model.GenerateAccessToken(userPayload, client)
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
	helpers.SuccessResponse(c, gin.H{"status": "success", "username": userPayload.Username}, 200)
}
