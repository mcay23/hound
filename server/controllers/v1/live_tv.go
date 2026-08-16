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

// @Router /api/v1/{id}/categories [get]
// @Summary Get Live TV Categories for a Provider
// @ID get-live-tv-categories
// @Description Get live TV categories from the specified IPTV provider
// @Tags Live TV
// @Accept json
// @Produce json
// @Param id path int true "IPTV Provider ID"
// @Success 200 {object} V1SuccessResponse{data=[]xtreamcodes.LiveCategory}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
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

type AddIPTVProviderResponse struct {
	Provider      database.IPTVProvider `json:"provider"`
	TotalChannels int                   `json:"total_channels,omitempty"`
	AddedChannels int                   `json:"added_channels,omitempty"`
}

type GetLiveChannelsResponse struct {
	Total    int                        `json:"total"`
	Added    bool                       `json:"added"`
	Channels []database.LiveIPTVChannel `json:"channels"`
}

type UpdateIPTVProviderRequest struct {
	Name              *string `json:"name"`
	Host              *string `json:"host"`
	Username          *string `json:"username"`
	Password          *string `json:"password"`
	IsDefaultProvider *bool   `json:"is_default_provider"`
}

// @Router /api/v1/iptv_providers [post]
// @Summary Add IPTV Provider
// @ID add-iptv-provider
// @Description Add a new IPTV provider (Xtream or M3U8)
// @Tags Live TV
// @Accept json
// @Produce json
// @Param request body AddIPTVProviderRequest true "IPTV Provider creation details"
// @Success 200 {object} V1SuccessResponse{data=AddIPTVProviderResponse}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func AddIPTVProviderHandler(c *gin.Context) {
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

// @Router /api/v1/iptv_providers [get]
// @Summary Get IPTV Providers
// @ID get-iptv-providers
// @Description Get a list of all configured IPTV providers
// @Tags Live TV
// @Accept json
// @Produce json
// @Success 200 {object} V1SuccessResponse{data=[]database.IPTVProvider}
// @Failure 500 {object} V1ErrorResponse
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

// @Router /api/v1/iptv_providers/{id} [put]
// @Summary Update IPTV Provider
// @ID update-iptv-provider
// @Description Update an existing IPTV provider's details or default status
// @Tags Live TV
// @Accept json
// @Produce json
// @Param id path int true "IPTV Provider ID"
// @Param request body UpdateIPTVProviderRequest true "IPTV Provider update details"
// @Success 200 {object} V1SuccessResponse
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func UpdateIPTVProviderHandler(c *gin.Context) {
	iptvProviderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("iptv provider id must be int64: %w: %w", err, internal.BadRequestError))
		return
	}
	var req UpdateIPTVProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	update := database.IPTVProvider{
		IPTVProviderID: iptvProviderID,
	}
	if req.Name != nil {
		update.Name = *req.Name
	}
	if req.Host != nil {
		update.Host = *req.Host
	}
	if req.Username != nil {
		update.Username = *req.Username
	}
	if req.Password != nil {
		encryptedPassword, err := internal.EncryptGCM([]byte(*req.Password))
		if err != nil {
			internal.ErrorResponse(c, fmt.Errorf("failed to encrypt password: %w",
				internal.InternalServerError))
		}
		update.EncryptedPassword = encryptedPassword
	}
	// will set current default to false before updating
	if req.IsDefaultProvider != nil && *req.IsDefaultProvider {
		if err := database.UpdateDefaultIPTVProvider(int(iptvProviderID)); err != nil {
			internal.ErrorResponse(c, err)
			return
		}
	}
	if err := database.UpdateIPTVProvider(&update); err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, nil, 200)
}

// @Router /api/v1/iptv_providers/{id} [delete]
// @Summary Delete IPTV Provider
// @ID delete-iptv-provider
// @Description Delete an IPTV provider by ID
// @Tags Live TV
// @Accept json
// @Produce json
// @Param id path int true "IPTV Provider ID"
// @Success 200 {object} V1SuccessResponse
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
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

// @Router /api/v1/live/{id}/channels [get]
// @Summary Get Live TV Channels
// @ID get-live-tv-channels
// @Description Get live TV channels for a specified IPTV provider, filtered by category (xtream)
// @Tags Live TV
// @Accept json
// @Produce json
// @Param id path int true "IPTV Provider ID"
// @Param category_id query string false "Category ID filter"
// @Success 200 {object} V1SuccessResponse{data=GetLiveChannelsResponse}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
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

// @Router /api/v1/live/{id}/epg [post]
// @Summary Get Channel EPGs
// @ID get-channel-epgs
// @Description Get Electronic Program Guide (EPG) data for specified channels under an IPTV provider
// @Tags Live TV
// @Accept json
// @Produce json
// @Param id path int true "IPTV Provider ID"
// @Param request body GetChannelEPGsRequest true "Channel EPG request details"
// @Success 200 {object} V1SuccessResponse{data=[]model.EPGProgramme}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
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
