package model

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/services"
	"github.com/mcay23/hound/sources"
	"github.com/sherif-fanous/xmltv"
	"github.com/sherif-fanous/xtreamcodes"
)

type EPGProgramme struct {
	EPGChannelID string                 `json:"epg_channel_id"`
	StartTime    time.Time              `json:"start_time"`
	StopTime     time.Time              `json:"stop_time"`
	Titles       []EPGProgrammeLanguage `json:"titles"`
	Descriptions []EPGProgrammeLanguage `json:"descriptions"`
}

type EPGProgrammeLanguage struct {
	Text     string  `json:"text"`
	Language *string `json:"lang,omitempty"`
}

const (
	ProviderEPGCacheKey = "iptv_providers|%d|epg"
)

func InitializeIPTV() {

}

func GetXtreamProviderEPG(iptvProviderID int64) (xmltv.EPG, error) {
	cacheKey := fmt.Sprintf(ProviderEPGCacheKey, iptvProviderID)
	var epg xmltv.EPG
	found, err := database.GetCache(cacheKey, &epg)
	if err != nil {
		return xmltv.EPG{}, fmt.Errorf("failed to get epg from cache for provider %d: %w", iptvProviderID, err)
	}
	if !found {
		return xmltv.EPG{}, fmt.Errorf("epg not found for provider %d (still downloading?): %w", iptvProviderID, internal.NotFoundError)
	}
	return epg, nil
}

func AddIPTVProviderXtream(name string, host string, username string, password string) (*database.IPTVProvider, error) {
	if name == "" || host == "" || username == "" || password == "" {
		return nil, fmt.Errorf("iptv provider name, host, username and password must not be empty: %w", internal.BadRequestError)
	}
	valid := internal.IsValidURL(host)
	if !valid {
		return nil, fmt.Errorf("invalid url: %s: %w", host, internal.BadRequestError)
	}
	encryptedPassword, err := internal.EncryptGCM([]byte(password))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}
	provider := &database.IPTVProvider{
		Name:              name,
		Host:              host,
		Username:          username,
		EncryptedPassword: encryptedPassword,
		ProxyStream:       false,
		IPTVProviderType:  database.IPTVProviderTypeXtream,
	}
	_, err = database.AddIPTVProvider(provider)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

func AddIPTVProviderM3U8(name string, host string) (*database.IPTVProvider, int, int, error) {
	if name == "" || host == "" {
		return nil, -1, -1, fmt.Errorf("iptv provider name, host, must not be empty: %w", internal.BadRequestError)
	}
	if !internal.IsValidURL(host) {
		return nil, -1, -1, fmt.Errorf("invalid url: %s: %w", host, internal.BadRequestError)
	}
	tempChannels, err := services.FetchM3U8Channels(host)
	if err != nil {
		return nil, -1, -1, err
	}
	cleanedChannels := services.CleanM3U8Channels(tempChannels)
	now := time.Now()
	provider := &database.IPTVProvider{
		Name:             name,
		Host:             host,
		ProxyStream:      false,
		IPTVProviderType: database.IPTVProviderTypeM3U8,
		LastRefresh:      &now,
	}
	_, err = database.AddIPTVProvider(provider)
	if err != nil {
		return nil, -1, -1, err
	}
	return provider, len(tempChannels), len(cleanedChannels), nil
}

func GetLiveChannelsIPTV(iptvProviderID int64, categoryID string) (int, int, []database.LiveIPTVChannel, error) {
	provider, err := database.GetIPTVProvider(iptvProviderID)
	if err != nil {
		return -1, -1, nil, err
	}
	switch provider.IPTVProviderType {
	case database.IPTVProviderTypeXtream:
		if categoryID == "" {
			return -1, -1, nil, fmt.Errorf("category id must not be empty: %w", internal.BadRequestError)
		}
		channels, err := sources.GetXtreamChannelsIPTV(iptvProviderID, categoryID)
		if err != nil {
			return -1, -1, nil, err
		}
		total, success, resp := xtreamToStandard(provider, channels, categoryID)
		return total, success, resp, nil
	case database.IPTVProviderTypeM3U8:
		channels, err := services.FetchM3U8Channels(provider.Host)
		if err != nil {
			return -1, -1, nil, err
		}
		total, success, resp := M3U8ToStandard(provider, channels)
		return total, success, resp, nil
	}
	return -1, -1, nil, fmt.Errorf("Invalid provider type in db, please report this issue to github: %w", internal.InternalServerError)
}

func GetChannelEPGs(iptvProviderID int64, EPGChannelIDs []string) ([]EPGProgramme, error) {
	// May return not found error if epg is not downloaded yet
	epg, err := GetXtreamProviderEPG(iptvProviderID)
	if err != nil {
		return nil, err
	}
	// normalize
	for i := range EPGChannelIDs {
		EPGChannelIDs[i] = strings.ToLower(EPGChannelIDs[i])
	}
	var resp []EPGProgramme
	for _, programme := range epg.Programmes {
		titles := []EPGProgrammeLanguage{}
		for _, title := range programme.Titles {
			titles = append(titles, EPGProgrammeLanguage{
				Text:     title.Text,
				Language: title.Lang,
			})
		}
		descriptions := []EPGProgrammeLanguage{}
		for _, description := range programme.Descriptions {
			descriptions = append(descriptions, EPGProgrammeLanguage{
				Text:     description.Text,
				Language: description.Lang,
			})
		}
		if slices.Contains(EPGChannelIDs, strings.ToLower(programme.Channel)) {
			resp = append(resp, EPGProgramme{
				EPGChannelID: programme.Channel,
				StartTime:    programme.Start.Time,
				StopTime:     programme.Stop.Time,
				Titles:       titles,
				Descriptions: descriptions,
			})
		}
	}
	return resp, nil
}

func encodeData(channel database.LiveIPTVChannel) (string, error) {
	data, err := json.Marshal(channel)
	if err != nil {
		return "", err
	}
	return internal.EncryptGCM(data)
}

func DecodeXtreamChannelData(encodedData string) (database.LiveIPTVChannel, error) {
	var channel database.LiveIPTVChannel
	raw, err := internal.DecryptGCM(encodedData)
	if err != nil {
		return channel, err
	}
	if err := json.Unmarshal(raw, &channel); err != nil {
		return channel, err
	}
	return channel, nil
}

// Returns total channels, succesful channels, and standardized channel list
func xtreamToStandard(provider *database.IPTVProvider, channels []xtreamcodes.LiveStream, categoryID string) (int, int, []database.LiveIPTVChannel) {
	resp := []database.LiveIPTVChannel{}
	succesfulChannels := 0
	for _, channel := range channels {
		temp := database.LiveIPTVChannel{
			IPTVProviderID:   provider.IPTVProviderID,
			IPTVProviderType: provider.IPTVProviderType,
			Order:            channel.Number,
			StreamID:         channel.StreamID,
			Name:             channel.Name,
			StreamType:       channel.StreamType,
			ThumbnailURL:     channel.StreamIcon,
			CategoryID:       categoryID,
			AddedAt:          channel.AddedOn,
		}
		if channel.EPGChannelID != nil {
			temp.EPGChannelID = *channel.EPGChannelID
		}
		// only no-proxy for now
		if !provider.ProxyStream {
			streamURL, err := getXtreamStreamLink(provider.IPTVProviderID, channel.StreamID)
			if err != nil {
				slog.Error("Failed to get stream link", "channel", channel, "error", err.Error())
				continue
			}
			temp.StreamURL = streamURL
		}
		if channel.Name == "" {
			slog.Error("No channel name provided", "channel", channel)
			continue
		}
		resp = append(resp, temp)
		succesfulChannels++
	}
	return len(channels), succesfulChannels, resp
}

func M3U8ToStandard(provider *database.IPTVProvider, channels []services.M3U8Channel) (int, int, []database.LiveIPTVChannel) {
	cleanedChannels := services.CleanM3U8Channels(channels)
	resp := []database.LiveIPTVChannel{}
	for idx, channel := range cleanedChannels {
		temp := database.LiveIPTVChannel{
			IPTVProviderID:   provider.IPTVProviderID,
			IPTVProviderType: provider.IPTVProviderType,
			Order:            idx,
			StreamID:         idx,
			Name:             channel.Name,
			Group:            channel.Group,
			StreamType:       "live",
			StreamURL:        channel.URL,
			ThumbnailURL:     channel.LogoURL,
			AddedAt:          time.Now(),
		}
		if channel.LogoURL != "" {
			temp.ThumbnailURL = channel.LogoURL
		}
		resp = append(resp, temp)
	}
	return len(channels), len(cleanedChannels), resp
}

func getXtreamStreamLink(iptvProviderID int64, streamID int) (string, error) {
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
