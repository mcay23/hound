package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	tmdb "github.com/cyruzin/golang-tmdb"
)

var invalidFilenameChars = regexp.MustCompile(`[<>:"/\\|?*]`)

func SanitizeFilename(filename string) string {
	return invalidFilenameChars.ReplaceAllString(filename, "")
}

func PrettyPrint(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "  ")
	fmt.Println(InfoMsg(string(s)))
}

// GetMagnetURI returns magnet: uri from hash and trackers
func GetMagnetURI(infoHash string, trackers *[]string) string {
	if infoHash == "" {
		return ""
	}
	magnetURI := fmt.Sprintf("magnet:?xt=urn:btih:%s", strings.ToLower(infoHash))
	if trackers == nil {
		return magnetURI
	}
	uniqueTrackers := make(map[string]struct{})
	for _, tracker := range *trackers {
		parts := strings.SplitN(tracker, ":", 2)
		if len(parts) != 2 {
			continue
		}
		sourceType := parts[0]
		value := parts[1]
		if sourceType == "tracker" {
			if _, exists := uniqueTrackers[value]; !exists {
				uniqueTrackers[value] = struct{}{}
			}
		} else {
			slog.Debug("Skipping tracker: " + sourceType)
		}
	}
	// append trackers
	var trackerParts []string
	for tracker := range uniqueTrackers {
		escapedTracker := url.QueryEscape(tracker)
		trackerParts = append(trackerParts, fmt.Sprintf("tr=%s", escapedTracker))
	}
	if len(trackerParts) > 0 {
		magnetURI += "&" + strings.Join(trackerParts, "&")
	}
	return magnetURI
}

// given a http url, extract infohash from it if it's in the url
func ExtractInfoHashFromURL(url string) (string, bool) {
	re := regexp.MustCompile(
		`(?i)[-/\[\(;:&]([a-f0-9]{40})[-\]\)/:;&]`,
	)
	m := re.FindStringSubmatch(url)
	if len(m) < 2 {
		return "", false
	}
	return strings.ToLower(m[1]), true
}

func GetTMDBImageURL(path string, size string) string {
	if path == "" {
		return ""
	}
	return tmdb.GetImageURL(path, size)
}

func IsValidURL(str string) bool {
	u, err := url.ParseRequestURI(str)
	if err != nil {
		return false
	}
	if u.Scheme == "" || u.Host == "" {
		return false
	}
	return true
}

func SetMockBrowserHeaders(req *http.Request) {
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
