package model

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/providers"
)

var (
	srtDetectionRegex = regexp.MustCompile(`(?m)^\d+\s*\n\s*\d{1,2}:\d{2}:\d{2}([,.]\d{1,3})?\s*-->`)
	srtToVttRegex     = regexp.MustCompile(`(\d{1,2}:\d{2}:\d{2}),(\d{3})`)
)

const (
	SubtitleTypeUnknown = "text/plain"
	SubtitleTypeASS     = "text/x-ass"
	SubtitleTypeSRT     = "application/x-subrip"
	SubtitleTypeVTT     = "text/vtt"
	SubtitleCacheKey    = "providers|subtitles|uri:%s"
)

func GetSubtitle(uri string, convert string) (string, string) {
	valid := internal.IsValidURL(uri)
	if !valid {
		return getFallbackSubtitle(uri, "invalid", false)
	}
	var content string
	exists, err := database.GetCache(fmt.Sprintf(SubtitleCacheKey, uri), &content)
	// If not in cache, fetch from remote
	if err != nil || !exists {
		// check if current service is blocked for this
		// soft lock, if there's an error grabbing we don't want to block forever
		blocked, err := providers.IsServiceBlocked(uri)
		if err == nil && blocked {
			return getFallbackSubtitle(uri, "remote", false)
		}
		// this is considered a failure state, since AIOStreams sometimes returns
		// this url when there is an issue in the upstream provider
		if strings.Contains(uri, "https://github.com/Viren070/AIOStreams") {
			return getFallbackSubtitle(uri, "remote", false)
		}
		client := &http.Client{
			Timeout: 5 * time.Second,
		}
		req, err := http.NewRequest("GET", uri, nil)
		if err != nil {
			slog.Error("error creating request", "uri", uri, "error", err)
			return getFallbackSubtitle(uri, "remote", true)
		}
		setMockBrowserHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			slog.Error("http error fetching subtitle", "uri", uri, "error", err)
			return getFallbackSubtitle(uri, "remote", true)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			slog.Error("error fetching subtitle, status not OK", "uri", uri, "status", resp.StatusCode)
			return getFallbackSubtitle(uri, "remote", true)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("io read error reading subtitle", "uri", uri, "error", err)
			return getFallbackSubtitle(uri, "remote", true)
		}
		content = string(body)
		// at this point, call was successful, clear failure states
		providers.ClearServiceFailures(uri)
		// cache raw subtitle for 24 hours
		_, err = database.SetCache(fmt.Sprintf(SubtitleCacheKey, uri), content, 24*time.Hour)
		if err != nil {
			slog.Error("error caching subtitle", "uri", uri, "error", err)
		}
	}
	subtitleType := getSubtitleType(content)
	// SRT to VTT conversion if requested, ignored for other formats
	if convert == SubtitleTypeVTT && subtitleType == SubtitleTypeSRT {
		content = "WEBVTT\n\n" + srtToVttRegex.ReplaceAllString(content, "$1.$2")
		subtitleType = SubtitleTypeVTT
	}
	return content, subtitleType
}

/*
Handles remote fetch failures and invalid requests with fallback subtitles.
Currently, some players don't handle not being able to grab subtitles well,
and will block playback until all subtitles have been fetched or timeout.
If there are 5+ subtitles, this can block playback for a while.
Ideally, clients would handle subtitle fetching failures better, but this
should be fine for now.
*/
func getFallbackSubtitle(url string, fallbackType string, incrementFailure bool) (string, string) {
	slog.Error("Failed to fetch subtitle", "url", url, "type", fallbackType, "incrementFailure", incrementFailure)
	var fallbackSubtitle string
	switch fallbackType {
	case "remote":
		fallbackSubtitle = "WEBVTT\n\n00:00:00.000 --> 03:00:00.000\nHound: Failed to fetch remote subtitles, you may be rate-limited or the service is down"
	case "invalid":
		fallbackSubtitle = "WEBVTT\n\n00:00:00.000 --> 03:00:00.000\nHound: Invalid subtitle requested"
	default:
		fallbackSubtitle = "WEBVTT\n\n00:00:00.000 --> 03:00:00.000\nHound: Failed to fetch remote subtitles, you may be rate-limited or the service is down"
	}
	if incrementFailure {
		providers.IncrementServiceFailure(url)
	}
	return fallbackSubtitle, SubtitleTypeVTT
}

func getSubtitleType(sub string) string {
	if len(sub) > 4096 {
		sub = sub[:4096]
	}
	// Handle UTF-8 BOM if present
	sub = strings.TrimPrefix(sub, "\uFEFF")
	sub = strings.TrimSpace(strings.ToLower(sub))

	if strings.HasPrefix(sub, "webvtt") {
		return SubtitleTypeVTT
	}
	if strings.Contains(sub, "[script info]") || strings.Contains(sub, "dialogue:") {
		return SubtitleTypeASS
	}
	if srtDetectionRegex.MatchString(sub) {
		return SubtitleTypeSRT
	}
	return SubtitleTypeUnknown
}

func setMockBrowserHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
}
