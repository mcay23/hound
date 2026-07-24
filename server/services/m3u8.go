package services

import (
	"bufio"
	"net/http"
	"strings"

	"github.com/mcay23/hound/internal"
)

type M3U8Channel struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Group   string `json:"group"`
	TVGID   string `json:"tvg_id"`
	TVGName string `json:"tvg_name"`
	LogoURL string `json:"logo_url"`
}

// this parses an m3u8 playlist that contains multiple channels
func FetchM3U8Channels(playlistURL string) ([]M3U8Channel, error) {
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

	channels := []M3U8Channel{}
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

// check if all channels have a valid url
func CleanM3U8Channels(channels []M3U8Channel) []M3U8Channel {
	resp := []M3U8Channel{}
	for _, channel := range channels {
		if channel.Name == "" || channel.URL == "" {
			continue
		}
		if !internal.IsValidURL(channel.URL) {
			continue
		}
		resp = append(resp, channel)
	}
	return resp
}

// this parses a channel line eg.: #EXTINF:-1 tvg-id="AutoCars.us@SD" tvg-logo="https://abc.com" group-title="Auto",The Auto Channel (720p)
func parseEXTINF(line string) M3U8Channel {
	var ch M3U8Channel
	line = strings.TrimPrefix(line, "#EXTINF:")
	i := 0
	n := len(line)
	// skip duration
	for i < n && line[i] != ' ' && line[i] != ',' {
		i++
	}
	for i < n {
		// skip whitespace
		for i < n && line[i] == ' ' {
			i++
		}
		// comma means the remainder is the channel name.
		if i < n && line[i] == ',' {
			ch.Name = strings.TrimSpace(line[i+1:])
			break
		}
		// parse key
		keyStart := i
		for i < n && line[i] != '=' && line[i] != ' ' && line[i] != ',' {
			i++
		}
		// No = means we're done with attributes.
		if i >= n || line[i] != '=' {
			break
		}
		key := line[keyStart:i]
		// skip =
		i++
		var value string
		if i < n && line[i] == '"' {
			i++
			start := i
			for i < n {
				if line[i] == '"' {
					break
				}
				i++
			}
			value = line[start:i]
			if i < n {
				i++
			}
		} else {
			start := i

			for i < n && line[i] != ' ' && line[i] != ',' {
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
			ch.LogoURL = value
		}
	}
	return ch
}
