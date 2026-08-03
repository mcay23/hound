package database

import (
	"time"
)

type IPTVCollection struct {
	OwnerUserID    string    `xorm:"owner_user_id", json:"owner_user_id"`
	CollectionName string    `xorm:"channel_name", json:"channel_name"`
	URL            string    `xorm:"url", json:"url"`
	IconURL        string    `xorm:"icon_url", json:"icon_url"`
	CountryCode    string    `xorm:"country_code", json:"country_code"`
	CountryName    string    `xorm:"country_name", json:"country_name"`
	CategoryName   string    `xorm:"category_name", json:"category_name"`
	AddedAt        time.Time `xorm:"added_at", json:"added_at"`
	UpdatedAt      time.Time `xorm:"updated_at", json:"updated_at"`
}
