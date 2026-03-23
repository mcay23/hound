package database

import (
	"fmt"
	"hound/helpers"
	"time"
)

type ProviderProfile struct {
	ProviderID  int64     `json:"provider_profile_id" xorm:"pk autoincr 'provider_profile_id'"`
	Name        string    `json:"name" xorm:"not null 'name'"`                             // profile name, eg. Performance, Quality, etc.
	ManifestURL string    `json:"manifest_url" xorm:"text not null unique 'manifest_url'"` // url to manifest.json for stremio providers
	CreatedAt   time.Time `json:"created_at" xorm:"timestampz created"`
	UpdatedAt   time.Time `json:"updated_at" xorm:"timestampz updated"`
}

const providerProfilesTable = "provider_profiles"

func instantiateProviderProfilesTable() error {
	return databaseEngine.Table(providerProfilesTable).Sync2(new(ProviderProfile))
}

func GetProviderProfiles() ([]ProviderProfile, error) {
	var providers []ProviderProfile
	cacheKey := "provider_profiles|all"
	cacheExists, _ := GetCache(cacheKey, &providers)
	if cacheExists {
		return providers, nil
	}
	err := databaseEngine.Table(providerProfilesTable).Find(&providers)
	if err != nil {
		return nil, fmt.Errorf("query all providers: %w", err)
	}
	if len(providers) > 0 {
		SetCache(cacheKey, providers, 12*time.Hour)
	}
	return providers, nil
}

func GetProviderProfile(providerID int) (ProviderProfile, error) {
	var provider ProviderProfile
	cacheKey := fmt.Sprintf("provider_profiles|id|%d", providerID)
	cacheExists, _ := GetCache(cacheKey, &provider)
	if cacheExists {
		return provider, nil
	}
	has, err := databaseEngine.Table(providerProfilesTable).ID(providerID).Get(&provider)
	if err != nil {
		return ProviderProfile{}, fmt.Errorf("query provider %d: %w", providerID, err)
	}
	if !has {
		return ProviderProfile{}, fmt.Errorf("query provider for provider_id %d not found: %w", providerID, helpers.NotFoundError)
	}
	SetCache(cacheKey, provider, 12*time.Hour)
	return provider, nil
}

func InsertProviderProfile(name string, manifestURL string) (ProviderProfile, error) {
	provider := ProviderProfile{
		Name:        name,
		ManifestURL: manifestURL,
	}
	_, err := databaseEngine.Table(providerProfilesTable).Insert(&provider)
	if err != nil {
		return provider, fmt.Errorf("insert provider: %w", err)
	}
	DeleteCache("provider_profiles|all")
	return provider, nil
}

func DeleteProviderProfile(providerID int) error {
	_, err := databaseEngine.Table(providerProfilesTable).Where("provider_profile_id = ?", providerID).Delete()
	if err != nil {
		return fmt.Errorf("delete provider %d: %w", providerID, err)
	}
	DeleteCache(fmt.Sprintf("provider_profiles|id|%d", providerID))
	DeleteCache("provider_profiles|all")
	return nil
}
