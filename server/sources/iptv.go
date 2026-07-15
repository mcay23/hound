package sources

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/sherif-fanous/xtreamcodes"
)

const (
	liveCategoriesCacheKey = "iptv_providers|%d|categories"
	channelsCacheKey       = "iptv_providers|%d|category|%s"
	categoryCacheTTL       = time.Hour
	channelsCacheTTL       = 30 * time.Minute
)

type IPTVLiveCategory struct {
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	ParentID     string `json:"parent_id"`
}

func GetLiveCategoriesIPTV(iptvProviderID int64) ([]xtreamcodes.LiveCategory, error) {
	cacheKey := fmt.Sprintf(liveCategoriesCacheKey, iptvProviderID)
	var cats []xtreamcodes.LiveCategory
	if found, _ := database.GetCache(cacheKey, &cats); found {
		return cats, nil
	}
	provider, err := database.GetIPTVProvider(int64(iptvProviderID))
	if err != nil {
		return nil, err
	}
	password, err := internal.DecryptGCM(provider.EncryptedPassword)
	if err != nil {
		return nil, err
	}
	client := xtreamcodes.NewClient(provider.Host, provider.Username, string(password))
	cats, err = client.ListLiveCategories(context.Background())
	if err != nil {
		return nil, sanitizeHTTPError(err)
	}
	database.SetCache(cacheKey, cats, categoryCacheTTL)
	return cats, nil
}

func GetLiveChannelsIPTV(iptvProviderID int64, categoryID string) ([]xtreamcodes.LiveStream, error) {
	cacheKey := fmt.Sprintf(channelsCacheKey, iptvProviderID, categoryID)
	var chans []xtreamcodes.LiveStream
	if found, _ := database.GetCache(cacheKey, &chans); found {
		return chans, nil
	}
	provider, err := database.GetIPTVProvider(iptvProviderID)
	if err != nil {
		return nil, err
	}
	password, err := internal.DecryptGCM(provider.EncryptedPassword)
	if err != nil {
		return nil, err
	}
	client := xtreamcodes.NewClient(provider.Host, provider.Username, string(password))
	chans, err = client.ListLiveStreamsInCategory(context.Background(), categoryID)
	if err != nil {
		return nil, sanitizeHTTPError(err)
	}
	database.SetCache(cacheKey, chans, channelsCacheTTL)
	return chans, nil
}

// prevent full iptv urls being exposed, since they contain username/password
func sanitizeHTTPError(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return fmt.Errorf("%s failed: %w", urlErr.Op, urlErr.Err)
	}
	return err
}

func GetXtreamStreamLink(iptvProviderID int64, streamID int) (string, error) {
	provider, err := database.GetIPTVProvider(iptvProviderID)
	if err != nil {
		return "", err
	}
	password, err := internal.DecryptGCM(provider.EncryptedPassword)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/live/%s/%s/%d.m3u8", provider.Host, provider.Username, string(password), streamID)
	return url, nil
}
