package v1

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/providers"

	"github.com/gin-gonic/gin"
)

type CreateProviderRequest struct {
	Name        string `json:"name"`
	ManifestURL string `json:"manifest_url"`
}

// @Router /api/v1/provider_profiles [get]
// @Summary Get all provider profiles
// @ID get-provider-profiles
// @Tags Provider Profiles
// @Accept json
// @Produce json
// @Success 200 {object} V1SuccessResponse{data=[]database.ProviderProfile}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func GetProviderProfilesHandler(c *gin.Context) {
	providerProfiles, err := database.GetProviderProfiles()
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get provider profiles: %w", err))
		return
	}
	internal.SuccessResponse(c, providerProfiles, 200)
}

// @Router /api/v1/provider_profiles [post]
// @Summary Create a provider profile
// @ID create-provider-profile
// @Tags Provider Profiles
// @Accept json
// @Produce json
// @Param provider_profile body CreateProviderRequest true "Provider Profile"
// @Success 200 {object} V1SuccessResponse{data=database.ProviderProfile}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func CreateProviderProfileHandler(c *gin.Context) {
	var body CreateProviderRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to bind provider profile: %w", err))
		return
	}
	if !strings.Contains(body.ManifestURL, "http://") && !strings.Contains(body.ManifestURL, "https://") {
		internal.ErrorResponse(c, fmt.Errorf("invalid provider manifest url, prepend http:// or https:// : %w", internal.BadRequestError))
		return
	}
	err := providers.PingProviderManifest(body.ManifestURL)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to ping provider manifest: %w", err))
		return
	}
	// store url without manifest.json
	provider, err := database.InsertProviderProfile(body.Name,
		strings.TrimSuffix(body.ManifestURL, "/manifest.json"))
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to create provider profile: %w", err))
		return
	}
	internal.SuccessResponse(c, provider, 200)
}

type UpdateProviderProfileRequest struct {
	IsDefaultStreaming   bool `json:"is_default_streaming"`
	IsDefaultDownloading bool `json:"is_default_downloading"`
}

// @Router /api/v1/provider_profiles/{id} [put]
// @Summary Update a provider profile
// @ID update-provider-profile
// @Description Set the default provider profiles for streaming, downloading. Note that clients may choose to override their own defaults.
// @Tags Provider Profiles
// @Accept json
// @Produce json
// @Param id path int true "Provider ID"
// @Param provider_profile body UpdateProviderProfileRequest true "Provider Profile Body"
// @Success 200 {object} V1SuccessResponse{data=database.ProviderProfile}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func UpdateProviderProfileHandler(c *gin.Context) {
	providerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("invalid provider id: %w: %w", internal.BadRequestError, err))
		return
	}
	var body UpdateProviderProfileRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to bind provider profile: %w", err))
		return
	}
	if err := database.UpdateDefaultProviderProfile(providerID, body.IsDefaultStreaming, body.IsDefaultDownloading); err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to update provider profile: %w", err))
		return
	}
	provider, err := database.GetProviderProfile(providerID)
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to get provider profile: %w", err))
		return
	}
	internal.SuccessResponse(c, provider, 200)
}

// @Router /api/v1/provider_profiles/{id} [delete]
// @Summary Delete a provider profile
// @ID delete-provider-profile
// @Tags Provider Profiles
// @Accept json
// @Produce json
// @Param id path int true "Provider ID"
// @Success 200 {object} V1SuccessResponse{data=object}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func DeleteProviderProfileHandler(c *gin.Context) {
	providerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		internal.ErrorResponse(c, fmt.Errorf("invalid provider id: %w: %w", internal.BadRequestError, err))
		return
	}
	if err := database.DeleteProviderProfile(providerID); err != nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to delete provider profile: %w", err))
		return
	}
	internal.SuccessResponse(c, nil, 200)
}
