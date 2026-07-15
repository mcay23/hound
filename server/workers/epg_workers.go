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
	"github.com/sherif-fanous/xmltv"
)

/*
	The EPG worker searches xtream providers, and updates if the last downloaded EPG
	is past the stale time.
	Only 1 epg worker runs at a time
*/

const (
	WorkerPollInterval = 1 * time.Minute
	EPGStaleTime       = 12 * time.Hour
)

func InitializeEPGWorkers() {
	go epgWorker()
}

func epgWorker() {
	slog.Debug("Epg worker started")
	for {
		tasks, err := database.GetIPTVProviders()
		if err != nil {
			slog.Error("EPG Worker failed to get IPTV providers", "error", err)
			time.Sleep(1 * time.Minute)
			continue
		}
		for _, task := range tasks {
			// skip if not xtream
			if task.IPTVStreamType != database.IPTVStreamTypeXTREAM {
				continue
			}
			if task.LastEPGDownload == nil || task.LastEPGDownload.Add(EPGStaleTime).Before(time.Now()) {
				slog.Info("EPG Worker downloading EPG...", "iptvProviderID", task.IPTVProviderID)
				if newEPG, err := downloadEPGXtream(task.IPTVProviderID); err != nil {
					failEPGTask(task, err)
				} else if newEPG == nil {
					slog.Error("EPG data is empty for provider provider", "iptvProviderID", task.IPTVProviderID)
				} else {
					_, err := database.SetCache(fmt.Sprintf(model.ProviderEPGCacheKey, task.IPTVProviderID), newEPG, -1)
					if err != nil {
						slog.Error("(EPG Download Failed) EPG Worker failed to set cache", "iptvProviderID", task.IPTVProviderID, "error", err)
						failEPGTask(task, err)
					} else {
						slog.Info("EPG Worker successfully downloaded EPG", "iptvProviderID", task.IPTVProviderID)
						now := time.Now()
						task.LastEPGDownload = &now
						task.LastEPGDownloadError = nil
						if err := database.UpdateIPTVProvider(task); err != nil {
							slog.Error("(EPG Download Success) EPG Worker failed to update IPTV provider", "iptvProviderID", task.IPTVProviderID, "error", err)
						}
					}
				}
			}
		}
		time.Sleep(WorkerPollInterval)
	}
}

func failEPGTask(task *database.IPTVProvider, err error) {
	slog.Error("EPG Worker failed to download EPG", "iptvProviderID", task.IPTVProviderID, "error", err)
	errMsg := fmt.Sprintf("%v", err)
	task.LastEPGDownloadError = &errMsg
	if err := database.UpdateIPTVProvider(task); err != nil {
		slog.Error("(EPG Download Failed) EPG Worker failed to update IPTV provider", "iptvProviderID", task.IPTVProviderID, "error", err)
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
