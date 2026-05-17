package v1

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/model"
)

// @Router /api/v1/home [get]
// @Summary Get User Home Rows
// @ID get-user-home-rows
// @Tags Home Rows
// @Accept json
// @Produce json
// @Success 200 {object} V1SuccessResponse{data=database.UserHomeRows}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetUserHomeRowsHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	homeRows, err := database.GetUserHomeRows(userID)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, homeRows, 200)
}

// @Router /api/v1/home/default [get]
// @Summary Get Default Home Rows
// @ID get-default-home-rows
// @Tags Home Rows
// @Accept json
// @Produce json
// @Success 200 {object} V1SuccessResponse{data=database.UserHomeRows}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetDefaultHomeRowsHandler(c *gin.Context) {
	homeRows, err := database.GetDefaultHomeRows()
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, homeRows, 200)
}

// @Router /api/v1/home/{rowIndex} [get]
// @Summary Get Home Row by Index
// @ID get-home-row-index
// @Tags Home Rows
// @Accept json
// @Produce json
// @Param rowIndex path int true "Row Index"
// @Success 200 {object} V1SuccessResponse{data=[]view.MediaRecordCatalog}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetHomeRowIndexHandler(c *gin.Context) {
	rowIndexStr := c.Param("rowIndex")
	if rowIndexStr == "" {
		internal.ErrorResponse(c, fmt.Errorf("rowIndex is required: %w", internal.BadRequestError))
		return
	}
	rowIndex, err := strconv.Atoi(rowIndexStr)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("invalid rowIndex: %s: %w", rowIndexStr, internal.BadRequestError))
		return
	}
	userID, err := getUserIDFromContext(c)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	viewArray, err := model.GetHomeRow(userID, rowIndex)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, viewArray, 200)
}

// @Router /api/v1/home [put]
// @Summary Update User Home Rows
// @ID update-user-home-rows
// @Tags Home Rows
// @Accept json
// @Produce json
// @Param homeRows body database.UserHomeRows true "Home Rows"
// @Success 200 {object} V1SuccessResponse{}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func UpdateUserHomeRowsHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	var homeRows database.UserHomeRows
	if err := c.ShouldBindJSON(&homeRows); err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	homeRows.UserID = userID
	if err := database.UpdateUserHomeRows(homeRows); err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, nil, 200)
}

// @Router /api/v1/home [delete]
// @Summary Reset User Home Rows to Server Defaults
// @ID reset-user-home-rows
// @Tags Home Rows
// @Accept json
// @Produce json
// @Success 200 {object} V1SuccessResponse{}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func ResetUserHomeRowsHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	if err := database.ResetUserHomeRows(userID); err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, nil, 200)
}

// @Router /api/v1/home/default [put]
// @Summary Update Default Home Rows
// @ID update-default-home-rows
// @Tags Home Rows
// @Accept json
// @Produce json
// @Param homeRows body database.UserHomeRows true "Home Rows"
// @Success 200 {object} V1SuccessResponse{}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func UpdateDefaultHomeRowsHandler(c *gin.Context) {
	var homeRows database.UserHomeRows
	if err := c.ShouldBindJSON(&homeRows); err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	if err := database.UpdateDefaultHomeRows(homeRows); err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, nil, 200)
}
