package v1

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/model"
	"github.com/mcay23/hound/sources"
)

func GetLiveCategoriesHandler(c *gin.Context) {
	cats, err := sources.GetLiveCategoriesIPTV()
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, cats, 200)
}

type AddIPTVProfileRequest struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func CreateIPTVProfileHandler(c *gin.Context) {
	var req AddIPTVProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	err := model.AddIPTVProfileXtream(req.Name, req.Host, req.Username, req.Password)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, nil, 200)
}

func GetIPTVProfilesHandler(c *gin.Context) {
	profiles, err := database.GetIPTVProfiles()
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	if len(profiles) == 0 {
		internal.SuccessResponse(c, []database.IPTVProfile{}, 200)
		return
	}
	internal.SuccessResponse(c, profiles, 200)
}

func DeleteIPTVProfileHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("id must be int64: %w: %w", err, internal.BadRequestError))
		return
	}
	if err := database.DeleteIPTVProfile(id); err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, nil, 200)
}

func GetLiveChannelsHandler(c *gin.Context) {
	ipTVProfileID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("iptv profile id must be int64: %w: %w", err, internal.BadRequestError))
		return
	}
	categoryID := c.Query("category_id")
	if categoryID == "" {
		internal.ErrorResponse(c, fmt.Errorf("category id must be defined: %w", internal.BadRequestError))
		return
	}
	channels, err := model.GetLiveChannelsIPTV(ipTVProfileID, categoryID)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, channels, 200)
}

// for proxied iptv streams in the future ?
// func StreamLiveTVHandler(c *gin.Context) {
// 	encodedData := c.Param("encodedString")
// 	if encodedData == "" {
// 		internal.ErrorResponse(c, fmt.Errorf("encoded data must be defined: %w", internal.BadRequestError))
// 		return
// 	}
// 	channel, err := model.DecodeXtreamChannelData(encodedData)
// 	if err != nil {
// 		internal.ErrorResponse(c, err)
// 		return
// 	}
// 	url, err := sources.GetXtreamStreamLink(channel.StreamID)
// 	if err != nil {
// 		internal.ErrorResponse(c, fmt.Errorf("failed to get xtream stream link: %w: %w", err, internal.BadRequestError))
// 		return
// 	}
// 	handleProxyStream(c, url)
// }
