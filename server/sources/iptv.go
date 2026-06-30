package sources

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/mcay23/hound/internal"
	"github.com/sherif-fanous/xtreamcodes"
)

var iptvHost = "http://cf.cdn-90.me"
var iptvUsername = "112ee65d0d"
var iptvPassword = "e1af615116"

type IPTVLiveCategory struct {
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	ParentID     string `json:"parent_id"`
}

var xtreamClient *xtreamcodes.Client

func InitializeXtream() {
	xtreamClient = xtreamcodes.NewClient(iptvHost, iptvUsername, iptvPassword)
}

func DownloadEPGXtream(epgFilePath string) error {
	_, err := os.Stat(epgFilePath)
	if err == nil {
		return nil
	}
	slog.Info("Downloading new EPG file")
	url := fmt.Sprintf("%s/xmltv.php?username=%s&password=%s", iptvHost, iptvUsername, iptvPassword)
	out, err := os.Create(epgFilePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to send HTTP GET request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad server status: %s: %w", resp.Status, internal.GatewayTimeoutError)
	}
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file content: %w", err)
	}
	slog.Info("EPG download successful", "filepath", epgFilePath)
	return nil
}

func GetLiveCategoriesIPTV() ([]xtreamcodes.LiveCategory, error) {
	return xtreamClient.ListLiveCategories(context.Background())
}

func GetLiveChannelsIPTV(categoryID string) ([]xtreamcodes.LiveStream, error) {
	return xtreamClient.ListLiveStreamsInCategory(context.Background(), categoryID)
}
