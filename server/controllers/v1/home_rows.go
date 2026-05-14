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
	homeRows, err := database.GetDefaultHomeRows(-1)
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
