package sources

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/mcay23/hound/internal"
	"github.com/sherif-fanous/xtreamcodes"
)

var iptvHost = ""
var iptvUsername = ""
var iptvPassword = ""

type IPTVLiveCategory struct {
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	ParentID     string `json:"parent_id"`
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
	client := xtreamcodes.NewClient(iptvHost, iptvUsername, iptvPassword)
	cats, err := client.ListLiveCategories(context.Background())
	if err != nil {
		return nil, sanitizeHTTPError(err)
	}
	return cats, nil
}

func GetLiveChannelsIPTV(categoryID string) ([]xtreamcodes.LiveStream, error) {
	client := xtreamcodes.NewClient(iptvHost, iptvUsername, iptvPassword)
	streams, err := client.ListLiveStreamsInCategory(context.Background(), categoryID)
	if err != nil {
		return nil, sanitizeHTTPError(err)
	}
	return streams, nil
}

// prevent full iptv urls being exposed, since they contain username/password
func sanitizeHTTPError(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return fmt.Errorf("%s failed: %w", urlErr.Op, urlErr.Err)
	}
	return err
}

func GetXtreamStreamLink(streamID int) (string, error) {
	url := fmt.Sprintf("%s/live/%s/%s/%d.m3u8", iptvHost, iptvUsername, iptvPassword, streamID)
	return url, nil
}
