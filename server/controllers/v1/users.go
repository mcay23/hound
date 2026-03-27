package v1

import (
	"fmt"
	"hound/database"
	"hound/helpers"
	"strconv"

	"github.com/gin-gonic/gin"
)

// @Router /api/v1/users [get]
// @Summary Get all users
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} V1SuccessResponse{data=[]database.User}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetUsersHandler(c *gin.Context) {
	users, err := database.GetUsers()
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	helpers.SuccessResponse(c, users, 200)
}

// @Router /api/v1/users/{id} [delete]
// @Summary Delete a user
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
		helpers.ErrorResponse(c, fmt.Errorf("invalid user id: %w", helpers.BadRequestError))
		return
	}
	// prevent self-deletion, for now only one admin can exist
	currentUsername := c.GetHeader("X-Username")
	currentUserID, err := database.GetUserIDFromUsername(currentUsername)
	if err == nil && currentUserID == userID {
		helpers.ErrorResponse(c, fmt.Errorf("cannot delete admin user: %w", helpers.BadRequestError))
		return
	}
	err = database.DeleteUser(userID)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	helpers.SuccessResponse(c, "user deleted", 200)
}
