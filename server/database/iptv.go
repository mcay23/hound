package database

import (
	"fmt"
	"time"
)

const IPTVConfigTypeXTREAM = "xtream"
const IPTVConfigTypeM3U = "m3u"

type IPTVConfig struct {
	IPTVConfigID      int64      `xorm:"pk autoincr 'iptv_config_id'" json:"iptv_config_id"`
	IPTVConfigType    string     `xorm:"'iptv_config_type'" json:"iptv_config_type"`
	Name              string     `xorm:"'title'" json:"title"`
	Host              string     `xorm:"text" json:"host"`
	Username          string     `xorm:"'username'" json:"username"`
	EncryptedPassword string     `xorm:"text 'encrypted_password'" json:"encrypted_password,omitempty"`
	LastDownload      *time.Time `xorm:"timestampz" json:"last_download,omitempty"`
	LastDownloadError *string    `xorm:"text" json:"last_download_error,omitempty"`
}

const iptvConfigTable = "iptv_configs"

func InstantiateIPTVConfigsTable() error {
	err := databaseEngine.Table(iptvConfigTable).Sync2(new(IPTVConfig))
	if err != nil {
		return fmt.Errorf("sync iptv configs table: %w", err)
	}
	return nil
}

func GetIPTVConfigs() ([]*IPTVConfig, error) {
	var configs []*IPTVConfig
	err := databaseEngine.Table(iptvConfigTable).Find(&configs)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPTV configs: %w", err)
	}
	return configs, nil
}

func GetIPTVConfig(iptvConfigID int64) (*IPTVConfig, error) {
	var config IPTVConfig
	has, err := databaseEngine.Table(iptvConfigTable).ID(iptvConfigID).Get(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPTV config: %w", err)
	}
	if !has {
		return nil, fmt.Errorf("IPTV config not found")
	}
	return &config, nil
}

func AddIPTVConfig(config *IPTVConfig) (int64, error) {
	_, err := databaseEngine.Table(iptvConfigTable).Insert(config)
	if err != nil {
		return 0, fmt.Errorf("failed to add IPTV config: %w", err)
	}
	return config.IPTVConfigID, nil
}

func UpdateIPTVConfig(config *IPTVConfig) error {
	_, err := databaseEngine.Table(iptvConfigTable).ID(config.IPTVConfigID).Update(config)
	if err != nil {
		return fmt.Errorf("failed to update IPTV config: %w", err)
	}
	return nil
}

func DeleteIPTVConfig(iptvConfigID int64) error {
	_, err := databaseEngine.Table(iptvConfigTable).ID(iptvConfigID).Delete()
	if err != nil {
		return fmt.Errorf("failed to delete IPTV config: %w", err)
	}
	return nil
}
