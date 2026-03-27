package database

import (
	"fmt"
	"github.com/mcay23/hound/helpers"
	"time"
)

type ProviderProfile struct {
	ProviderProfileID    int64     `xorm:"pk autoincr 'provider_profile_id'" json:"provider_profile_id"`
	Name                 string    `xorm:"not null 'name'" json:"name"`                             // profile name, eg. Performance, Quality, etc.
	ManifestURL          string    `xorm:"text not null unique 'manifest_url'" json:"manifest_url"` // url to manifest.json for stremio providers
	IsDefaultStreaming   bool      `xorm:"'is_default_streaming'" json:"is_default_streaming"`      // default profile for streaming
	IsDefaultDownloading bool      `xorm:"'is_default_downloading'" json:"is_default_downloading"`  // default profile for downloading
	CreatedAt            time.Time `xorm:"timestampz created" json:"created_at"`
	UpdatedAt            time.Time `xorm:"timestampz updated" json:"updated_at"`
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
	if len(providers) == 0 {
		return []ProviderProfile{}, nil
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
	// if first provider being inserted, set as default for streaming and downloading
	profiles, err := GetProviderProfiles()
	if err != nil {
		return provider, fmt.Errorf("query all providers: %w", err)
	}
	if len(profiles) == 0 {
		provider.IsDefaultStreaming = true
		provider.IsDefaultDownloading = true
	}
	_, err = databaseEngine.Table(providerProfilesTable).Insert(&provider)
	if err != nil {
		return provider, fmt.Errorf("insert provider: %w", err)
	}
	DeleteCache("provider_profiles|all")
	return provider, nil
}

// Note that only true fields are updated
// eg. If isDefaultStreaming is true and isDefaultDownloading is false,
// then only the default streaming profile is updated
// This is meant to set a default, not to unset one
func UpdateDefaultProviderProfile(defaultProviderID int, isDefaultStreaming bool, isDefaultDownloading bool) error {
	sess := databaseEngine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}
	if isDefaultStreaming {
		if _, err := sess.
			Table(providerProfilesTable).
			Cols("is_default_streaming").
			Update(&ProviderProfile{IsDefaultStreaming: false}); err != nil {
			sess.Rollback()
			return fmt.Errorf("unset default streaming profile %d: %w", defaultProviderID, err)
		}
	}
	if isDefaultDownloading {
		if _, err := sess.
			Table(providerProfilesTable).
			Cols("is_default_downloading").
			Update(&ProviderProfile{IsDefaultDownloading: false}); err != nil {
			sess.Rollback()
			return fmt.Errorf("unset default downloading profile %d: %w", defaultProviderID, err)
		}
	}
	update := ProviderProfile{}
	cols := []string{}
	if isDefaultStreaming {
		update.IsDefaultStreaming = isDefaultStreaming
		cols = append(cols, "is_default_streaming")
	}
	if isDefaultDownloading {
		update.IsDefaultDownloading = isDefaultDownloading
		cols = append(cols, "is_default_downloading")
	}
	if len(cols) > 0 {
		if _, err := sess.
			Table(providerProfilesTable).
			Where("provider_profile_id = ?", defaultProviderID).
			Cols(cols...).
			Update(&update); err != nil {
			sess.Rollback()
			return fmt.Errorf("update default provider profile %d: %w", defaultProviderID, err)
		}
	}
	DeleteCache(fmt.Sprintf("provider_profiles|id|%d", defaultProviderID))
	DeleteCache("provider_profiles|all")
	return sess.Commit()
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
