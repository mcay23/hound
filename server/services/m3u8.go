package services

import (
	"bufio"
	"net/http"
	"strings"
)

type M3U8Channel struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Group   string `json:"group"`
	TVGID   string `json:"tvg_id"`
	TVGName string `json:"tvg_name"`
	Logo    string `json:"logo"`
}

// this parses an m3u8 playlist that contains multiple channels
func ParseM3U8Channels(playlistURL string) ([]M3U8Channel, error) {
	req, err := http.NewRequest("GET", playlistURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Connection", "keep-alive")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)

	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var channels []M3U8Channel
	var current *M3U8Channel
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// if channel found, parse and set as current
		if strings.HasPrefix(line, "#EXTINF:-") {
			channel := parseEXTINF(line)
			current = &channel
			continue
		} else if strings.HasPrefix(line, "#") {
			continue
		}
		// skip directives that aren't the url
		if strings.HasPrefix(line, "#") {
			continue
		}
		// the next line after current channel is the url
		if current != nil {
			current.URL = line
			channels = append(channels, *current)
			current = nil
		}
	}
	return channels, scanner.Err()
}

// this parses a channel line eg.: #EXTINF:-1 tvg-id="AutoCars.us@SD" tvg-logo="https://abc.com" group-title="Auto",The Auto Channel (720p)
func parseEXTINF(line string) M3U8Channel {
	var ch M3U8Channel
	line = strings.TrimPrefix(line, "#EXTINF:")
	if comma := strings.IndexByte(line, ','); comma >= 0 {
		ch.Name = strings.TrimSpace(line[comma+1:])
		line = line[:comma]
	}
	i := 0
	n := len(line)
	for i < n {
		// skip whitespace
		for i < n && line[i] == ' ' {
			i++
		}
		// skip duration
		if i < n && (line[i] == '-' || (line[i] >= '0' && line[i] <= '9')) {
			for i < n && line[i] != ' ' {
				i++
			}
			continue
		}
		// read key
		keyStart := i
		for i < n && line[i] != '=' && line[i] != ' ' {
			i++
		}
		if i >= n || line[i] != '=' {
			for i < n && line[i] != ' ' {
				i++
			}
			continue
		}
		key := line[keyStart:i]
		i++
		var value string
		// quoted
		if i < n && line[i] == '"' {
			i++
			start := i

			for i < n && line[i] != '"' {
				i++
			}
			value = line[start:i]
			if i < n {
				i++
			}
		} else {
			start := i
			for i < n && line[i] != ' ' {
				i++
			}
			value = line[start:i]
		}
		switch key {
		case "group-title":
			ch.Group = value
		case "tvg-id":
			ch.TVGID = value
		case "tvg-name":
			ch.TVGName = value
		case "tvg-logo":
			ch.Logo = value
		}
	}
	return ch
}
