package helpers

import (
	"encoding/json"
	"fmt"
	"log/slog"
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
			slog.Info("Skipping tracker: " + sourceType)
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
