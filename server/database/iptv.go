package database

import (
	"fmt"
	"time"
)

const IPTVProviderTypeXtream = "xtream"
const IPTVProviderTypeM3U8 = "m3u8"

// LastRefresh: for xtream, when epg was updated, for m3u8, when channels were updated
type IPTVProvider struct {
	IPTVProviderID    int64      `xorm:"pk autoincr 'iptv_provider_id'" json:"iptv_provider_id"`
	IPTVProviderType  string     `xorm:"'iptv_provider_type'" json:"iptv_provider_type"`
	ProxyStream       bool       `xorm:"'proxy_stream'" json:"proxy_stream"` // whether or not to proxy the stream, always false for now
	Name              string     `xorm:"'name'" json:"name"`
	Host              string     `xorm:"text 'host'" json:"host"`
	Username          string     `xorm:"'username'" json:"username"`
	EncryptedPassword string     `xorm:"text 'encrypted_password'" json:"encrypted_password,omitempty"`
	IsDefault         bool       `xorm:"'is_default'" json:"is_default"`
	LastRefresh       *time.Time `xorm:"timestampz 'last_refresh'" json:"last_refresh,omitempty"`
	LastRefreshError  *string    `xorm:"text 'last_refresh_error'" json:"lash_refresh_error,omitempty"`
}

type LiveIPTVChannel struct {
	IPTVProviderID   int64     `json:"iptv_provider_id"`
	IPTVProviderType string    `json:"iptv_provider_type"`
	Order            int       `json:"order"`
	StreamID         int       `json:"stream_id,omitempty"`
	Group            string    `json:"group,omitempty"` // mostly for m3u8
	Name             string    `json:"name"`
	StreamType       string    `json:"stream_type"`
	ThumbnailURL     string    `json:"thumbnail_url"`
	EPGChannelID     string    `json:"epg_channel_id"`
	CategoryID       string    `json:"category_id,omitempty"`
	AddedAt          time.Time `json:"added_at"`
	StreamURL        string    `json:"stream_url"`
}

const iptvProvidersTable = "iptv_providers"

const (
	allProvidersCacheKey         = "iptv_providers|all"
	iptvProviderCacheKey         = "iptv_providers|%d|provider"
	M3u8ProviderChannelsCacheKey = "iptv_providers|%d|m3u8_channels"
)

func instantiateIPTVProvidersTable() error {
	err := databaseEngine.Table(iptvProvidersTable).Sync2(new(IPTVProvider))
	if err != nil {
		return fmt.Errorf("sync iptv_providers table: %w", err)
	}
	return nil
}

func GetIPTVProviders() ([]*IPTVProvider, error) {
	var providers []*IPTVProvider
	if found, _ := GetCache(allProvidersCacheKey, &providers); found {
		return providers, nil
	}
	err := databaseEngine.Table(iptvProvidersTable).Find(&providers)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPTV providers: %w", err)
	}
	if len(providers) > 0 {
		SetCache(allProvidersCacheKey, providers, time.Hour)
	}
	return providers, nil
}

func GetIPTVProvider(iptvProviderID int64) (*IPTVProvider, error) {
	var config IPTVProvider
	cacheKey := fmt.Sprintf(iptvProviderCacheKey, iptvProviderID)
	if found, _ := GetCache(cacheKey, &config); found {
		return &config, nil
	}
	has, err := databaseEngine.Table(iptvProvidersTable).ID(iptvProviderID).Get(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPTV provider: %w", err)
	}
	if !has {
		return nil, fmt.Errorf("IPTV provider not found")
	}
	SetCache(cacheKey, config, time.Hour)
	return &config, nil
}

func AddIPTVProvider(provider *IPTVProvider) (int64, error) {
	_, err := databaseEngine.Table(iptvProvidersTable).Insert(provider)
	if err != nil {
		return 0, fmt.Errorf("failed to add IPTV provider: %w", err)
	}
	DeleteCache(allProvidersCacheKey)
	return provider.IPTVProviderID, nil
}

func UpdateIPTVProvider(provider *IPTVProvider) error {
	cacheKey := fmt.Sprintf(iptvProviderCacheKey, provider.IPTVProviderID)
	DeleteCache(cacheKey)
	DeleteCache(allProvidersCacheKey)
	_, err := databaseEngine.Table(iptvProvidersTable).ID(provider.IPTVProviderID).Update(provider)
	if err != nil {
		return fmt.Errorf("failed to update IPTV provider: %w", err)
	}
	return nil
}

func DeleteIPTVProvider(iptvProviderID int64) error {
	cacheKey := fmt.Sprintf(iptvProviderCacheKey, iptvProviderID)
	DeleteCache(cacheKey)
	DeleteCache(allProvidersCacheKey)
	_, err := databaseEngine.Table(iptvProvidersTable).Where("iptv_provider_id = ?", iptvProviderID).Delete()
	if err != nil {
		return fmt.Errorf("failed to delete IPTV provider: %w", err)
	}
	return nil
}

/*
Store/Get M3U8 playlist channels
*/
func GetM3U8Channels(iptvProviderID int64) ([]LiveIPTVChannel, error) {
	var channels []LiveIPTVChannel
	cacheKey := fmt.Sprintf(M3u8ProviderChannelsCacheKey, iptvProviderID)
	if found, _ := GetCache(cacheKey, &channels); found {
		return channels, nil
	}
	return channels, nil
}

func SetM3U8Channels(iptvProviderID int64, channels []LiveIPTVChannel) error {
	cacheKey := fmt.Sprintf(M3u8ProviderChannelsCacheKey, iptvProviderID)
	_, err := SetCache(cacheKey, channels, -1)
	if err != nil {
		return fmt.Errorf("failed to set M3U8 channels: %w", err)
	}
	return nil
}
