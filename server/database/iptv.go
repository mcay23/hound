package database

import (
	"fmt"
	"time"
)

const IPTVStreamTypeXTREAM = "xtream"
const IPTVStreamTypeM3U = "m3u"

type IPTVProvider struct {
	IPTVProviderID       int64      `xorm:"pk autoincr 'iptv_provider_id'" json:"iptv_provider_id"`
	IPTVStreamType       string     `xorm:"'iptv_stream_type'" json:"iptv_stream_type"`
	ProxyStream          bool       `xorm:"'proxy_stream'" json:"proxy_stream"` // whether or not to proxy the stream, always false for now
	Name                 string     `xorm:"'name'" json:"name"`
	Host                 string     `xorm:"text 'host'" json:"host"`
	Username             string     `xorm:"'username'" json:"username"`
	IsDefault            bool       `xorm:"'is_default'" json:"is_default"`
	EncryptedPassword    string     `xorm:"text 'encrypted_password'" json:"encrypted_password,omitempty"`
	LastEPGDownload      *time.Time `xorm:"timestampz 'last_epg_download'" json:"last_epg_download,omitempty"`
	LastEPGDownloadError *string    `xorm:"text 'last_epg_download_error'" json:"last_epg_download_error,omitempty"`
}

const iptvProvidersTable = "iptv_providers"

const (
	allProvidersCacheKey = "iptv_providers|all"
	iptvProviderCacheKey = "iptv_providers|%d|provider"
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
