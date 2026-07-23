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
	iptvProviderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("iptv provider id must be int64: %w: %w", err, internal.BadRequestError))
		return
	}
	cats, err := sources.GetLiveCategoriesXtream(iptvProviderID)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, cats, 200)
}

type AddIPTVProviderRequest struct {
	IPTVProviderType string `json:"iptv_provider_type" binding:"required"`
	Name             string `json:"name" binding:"required"`
	Host             string `json:"host" binding:"required"`
	Username         string `json:"username"`
	Password         string `json:"password"`
}

type GetChannelEPGsRequest struct {
	EPGChannelIDs []string `json:"epg_channel_ids" binding:"required"`
}

func CreateIPTVProviderHandler(c *gin.Context) {
	var req AddIPTVProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	switch req.IPTVProviderType {
	case database.IPTVProviderTypeXtream:
		provider, err := model.AddIPTVProviderXtream(req.Name, req.Host, req.Username, req.Password)
		if err != nil {
			internal.ErrorResponse(c, err)
			return
		}
		internal.SuccessResponse(c, gin.H{
			"provider": provider,
		}, 200)
		return
	case database.IPTVProviderTypeM3U8:
		provider, totalChannels, cleanedChannels, err := model.AddIPTVProviderM3U8(req.Name, req.Host)
		if err != nil {
			internal.ErrorResponse(c, err)
			return
		}
		internal.SuccessResponse(c, gin.H{
			"provider":       provider,
			"total_channels": totalChannels,
			"added_channels": cleanedChannels,
		}, 200)
		return
	default:
		internal.ErrorResponse(c, fmt.Errorf("invalid iptv stream type: %s: %w", req.IPTVProviderType, internal.BadRequestError))
		return
	}
}

func GetIPTVProvidersHandler(c *gin.Context) {
	providers, err := database.GetIPTVProviders()
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	for _, p := range providers {
		p.EncryptedPassword = ""
	}
	if len(providers) == 0 {
		internal.SuccessResponse(c, []database.IPTVProvider{}, 200)
		return
	}
	internal.SuccessResponse(c, providers, 200)
}

func DeleteIPTVProviderHandler(c *gin.Context) {
	iptvProviderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("iptv provider id must be int64: %w: %w", err, internal.BadRequestError))
		return
	}
	if err := database.DeleteIPTVProvider(iptvProviderID); err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, nil, 200)
}

func GetLiveChannelsHandler(c *gin.Context) {
	iptvProviderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("iptv provider id must be int64: %w: %w", err, internal.BadRequestError))
		return
	}
	categoryID := c.Query("category_id")
	total, added, channels, err := model.GetLiveChannelsIPTV(iptvProviderID, categoryID)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, gin.H{
		"total":    total,
		"added":    added,
		"channels": channels,
	}, 200)
}

func GetChannelEPGsHandler(c *gin.Context) {
	iptvProviderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("iptv provider id must be int64: %w: %w", err, internal.BadRequestError))
		return
	}
	var req GetChannelEPGsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		internal.ErrorResponse(c, fmt.Errorf("request body must be defined: %w", internal.BadRequestError))
		return
	}
	epg, err := model.GetChannelEPGs(iptvProviderID, req.EPGChannelIDs)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, epg, 200)
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
