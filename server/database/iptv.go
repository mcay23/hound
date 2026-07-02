package database

import (
	"fmt"
	"time"
)

const IPTVStreamTypeXTREAM = "xtream"
const IPTVStreamTypeM3U = "m3u"

type IPTVProfile struct {
	IPTVProfileID        int64      `xorm:"pk autoincr 'iptv_profile_id'" json:"iptv_profile_id"`
	IPTVStreamType       string     `xorm:"'iptv_stream_type'" json:"iptv_stream_type"`
	ProxyStream          bool       `xorm:"'proxy_stream'" json:"proxy_stream"` // whether or not to proxy the stream, always false for now
	Name                 string     `xorm:"'name'" json:"name"`
	Host                 string     `xorm:"text" json:"host"`
	Username             string     `xorm:"'username'" json:"username"`
	EncryptedPassword    string     `xorm:"text 'encrypted_password'" json:"encrypted_password,omitempty"`
	LastEPGDownload      *time.Time `xorm:"timestampz" json:"last_epg_download,omitempty"`
	LastEPGDownloadError *string    `xorm:"text" json:"last_epg_download_error,omitempty"`
}

const iptvProfilesTable = "iptv_profiles"

func instantiateIPTVProfilesTable() error {
	err := databaseEngine.Table(iptvProfilesTable).Sync2(new(IPTVProfile))
	if err != nil {
		return fmt.Errorf("sync iptv_profiles table: %w", err)
	}
	return nil
}

func GetIPTVProfiles() ([]*IPTVProfile, error) {
	var profiles []*IPTVProfile
	err := databaseEngine.Table(iptvProfilesTable).Find(&profiles)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPTV profiles: %w", err)
	}
	for _, profile := range profiles {
		profile.EncryptedPassword = ""
	}
	return profiles, nil
}

func GetIPTVProfile(iptvProfileID int64) (*IPTVProfile, error) {
	var config IPTVProfile
	has, err := databaseEngine.Table(iptvProfilesTable).ID(iptvProfileID).Get(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPTV profile: %w", err)
	}
	if !has {
		return nil, fmt.Errorf("IPTV profile not found")
	}
	return &config, nil
}

func AddIPTVProfile(profile *IPTVProfile) (int64, error) {
	_, err := databaseEngine.Table(iptvProfilesTable).Insert(profile)
	if err != nil {
		return 0, fmt.Errorf("failed to add IPTV profile: %w", err)
	}
	return profile.IPTVProfileID, nil
}

func UpdateIPTVProfile(profile *IPTVProfile) error {
	_, err := databaseEngine.Table(iptvProfilesTable).ID(profile.IPTVProfileID).Update(profile)
	if err != nil {
		return fmt.Errorf("failed to update IPTV profile: %w", err)
	}
	return nil
}

func DeleteIPTVProfile(iptvProfileID int64) error {
	_, err := databaseEngine.Table(iptvProfilesTable).Where("iptv_profile_id = ?", iptvProfileID).Delete()
	if err != nil {
		return fmt.Errorf("failed to delete IPTV profile: %w", err)
	}
	return nil
}
