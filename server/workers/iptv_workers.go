package workers

import (
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/model"
	"github.com/mcay23/hound/services"
	"github.com/sherif-fanous/xmltv"
)

/*
	The IPTV worker:
	1. Updates epg data for xtream providers if stale
	2. Updates channel playlist for m3u8 providers if stale
	Only 1 worker runs at a time
*/

const (
	workerPollInterval = 1 * time.Minute
	iptvStaleTime      = 12 * time.Hour
)

func InitializeIPTVWorkers() {
	go epgWorker()
}

func epgWorker() {
	slog.Debug("Epg worker started")
	for {
		providers, err := database.GetIPTVProviders()
		if err != nil {
			slog.Error("EPG Worker failed to get IPTV providers", "error", err)
			time.Sleep(1 * time.Minute)
			continue
		}
		for _, provider := range providers {
			// skip if not xtream
			if provider.LastRefresh == nil || provider.LastRefresh.Add(iptvStaleTime).Before(time.Now()) {
				if provider.IPTVProviderType == database.IPTVProviderTypeXtream {
					handleXtream(provider)
					continue
				}
				if provider.IPTVProviderType == database.IPTVProviderTypeM3U8 {
					handleM3U8(provider)
					continue
				}
			}
		}
		time.Sleep(workerPollInterval)
	}
}

func failIPTVTask(task *database.IPTVProvider, err error) {
	slog.Error("Worker failed to refresh iptv provider", "iptvProviderID", task.IPTVProviderID, "iptvProviderType", task.IPTVProviderType, "error", err)
	errMsg := fmt.Sprintf("%v", err)
	task.LastRefreshError = &errMsg
	if err := database.UpdateIPTVProvider(task); err != nil {
		slog.Error("(IPTV Refresh Failed) Worker failed to update IPTV provider", "iptvProviderID", task.IPTVProviderID, "error", err)
	}
}

func handleXtream(provider *database.IPTVProvider) {
	slog.Info("EPG Worker: Downloading EPG...", "iptvProviderID", provider.IPTVProviderID)
	newEPG, err := downloadEPGXtream(provider.IPTVProviderID)
	if err != nil {
		failIPTVTask(provider, err)
		return
	}
	if newEPG == nil {
		slog.Error("EPG Worker: EPG data is empty for provider", "iptvProviderID", provider.IPTVProviderID)
		failIPTVTask(provider, internal.InternalServerError)
		return
	}
	_, err = database.SetCache(fmt.Sprintf(model.ProviderEPGCacheKey, provider.IPTVProviderID), newEPG, -1)
	if err != nil {
		slog.Error("EPG Worker: Failed to set cache", "iptvProviderID", provider.IPTVProviderID, "error", err)
		failIPTVTask(provider, err)
		return
	}
	slog.Info("EPG Worker: Sccessfully downloaded EPG", "iptvProviderID", provider.IPTVProviderID)
	now := time.Now()
	provider.LastRefresh = &now
	provider.LastRefreshError = nil
	if err := database.UpdateIPTVProvider(provider); err != nil {
		slog.Error("(EPG Download Success) But EPG Worker failed to update IPTV provider",
			"iptvProviderID", provider.IPTVProviderID, "error", err)
	}
}

func handleM3U8(provider *database.IPTVProvider) {
	slog.Info("M3U8 Worker: downloading m3u8 playlist...", "iptvProviderID", provider.IPTVProviderID)
	channels, err := services.FetchM3U8Channels(provider.Host)
	if err != nil {
		failIPTVTask(provider, err)
		return
	}
	total, success, resp := model.M3U8ToStandard(provider, channels)
	if total != success && total > 0 {
		slog.Error("M3U8 Worker: Not all channels were succesfully parsed, continuing anyway..", "iptvProviderID",
			provider.IPTVProviderID, "total", total, "success", success)
	}
	err = database.SetM3U8Channels(provider.IPTVProviderID, resp)
	if err != nil {
		slog.Error("M3U8 Worker: Failed to set cache", "iptvProviderID", provider.IPTVProviderID, "error", err)
		failIPTVTask(provider, err)
		return
	}
	slog.Info("M3U8 Worker: Successfully refreshed Playlist", "iptvProviderID", provider.IPTVProviderID)
	now := time.Now()
	provider.LastRefresh = &now
	provider.LastRefreshError = nil
	err = database.UpdateIPTVProvider(provider)
	if err != nil {
		slog.Error("EPG Download Success (But M3U8 Worker: Failed to update IPTV provider",
			"iptvProviderID", provider.IPTVProviderID, "error", err)
	}
}

func downloadEPGXtream(iptvProviderID int64) (*xmltv.EPG, error) {
	provider, err := database.GetIPTVProvider(iptvProviderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get iptv provider: %w", err)
	}
	slog.Info("Downloading new EPG file")
	decryptedPassword, err := internal.DecryptGCM(provider.EncryptedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password: %w", err)
	}
	url := fmt.Sprintf("%s/xmltv.php?username=%s&password=%s", provider.Host, provider.Username, decryptedPassword)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP GET request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad server status: %s: %w", resp.Status, internal.GatewayTimeoutError)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var epg xmltv.EPG
	if err := xml.Unmarshal(body, &epg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XMLTV: %w", err)
	}
	return &epg, nil
}
