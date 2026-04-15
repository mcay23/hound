package v1

import (
	"fmt"
	"strconv"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/model"

	"github.com/gin-gonic/gin"
)

// @Router /api/v1/users [get]
// @Summary Get all users
// @ID get-users
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} V1SuccessResponse{data=[]database.User}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetUsersHandler(c *gin.Context) {
	users, err := database.GetUsers()
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, users, 200)
}

// @Router /api/v1/users/{id} [delete]
// @Summary Delete a user
// @ID delete-user
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} V1SuccessResponse{data=string}
// @Failure 400 {object} V1ErrorResponse
// @Failure 401 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func DeleteUserHandler(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("invalid user id: %w", internal.BadRequestError))
		return
	}
	// prevent self-deletion, for now only one admin can exist
	currentUserID, err := getUserIDFromContext(c)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	if currentUserID == userID {
		internal.ErrorResponse(c, fmt.Errorf("cannot delete admin user: %w", internal.BadRequestError))
		return
	}
	err = database.DeleteUser(userID)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, "user deleted", 200)
}

type AdminResetPasswordPayload struct {
	NewPassword string `json:"new_password" binding:"required,gte=8"`
}

// @Router /api/v1/users/{id}/password [post]
// @Summary Admin reset user password
// @ID admin-reset-password
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body AdminResetPasswordPayload true "New Password Details"
// @Success 200 {object} V1SuccessResponse
// @Failure 400 {object} V1ErrorResponse
// @Failure 401 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func AdminResetPasswordHandler(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("invalid user id: %w", internal.BadRequestError))
		return
	}
	var payload AdminResetPasswordPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		internal.ErrorResponse(c, fmt.Errorf("invalid payload: %w", internal.BadRequestError))
		return
	}
	err = model.ResetPassword(userID, payload.NewPassword)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, nil, 200)
}
