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
)

type LiveIPTVChannel struct {
	IPTVProviderID   int64     `json:"iptv_provider_id"`
	Order            int       `json:"order"`
	StreamID         int       `json:"stream_id"`
	Name             string    `json:"name"`
	XtreamStreamType string    `json:"xtream_stream_type"`
	ThumbnailURL     string    `json:"thumbnail_url"`
	EPGChannelID     string    `json:"epg_channel_id"`
	CategoryID       string    `json:"category_id"`
	AddedAt          time.Time `json:"added_at"`
	StreamURL        string    `json:"stream_url"`
}

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
	channels, err := services.ParseM3U8Channels("https://iptv-org.github.io/iptv/categories/auto.m3u")
	if err != nil {
		slog.Error("Failed to parse M3U8 channels", "error", err.Error())
		return
	}
	for _, ch := range channels {
		fmt.Println(ch.Name)
		fmt.Println(ch.URL)
		fmt.Println("-----------------")
	}
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

// TODO encrypt password
func AddIPTVProviderXtream(name string, host string, username string, password string) error {
	if name == "" || host == "" || username == "" || password == "" {
		return fmt.Errorf("iptv provider name, host, username and password must not be empty: %w", internal.BadRequestError)
	}
	encryptedPassword, err := internal.EncryptGCM([]byte(password))
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}
	provider := &database.IPTVProvider{
		Name:              name,
		Host:              host,
		Username:          username,
		EncryptedPassword: encryptedPassword,
		ProxyStream:       false,
		IPTVStreamType:    database.IPTVStreamTypeXTREAM,
	}
	_, err = database.AddIPTVProvider(provider)
	if err != nil {
		return err
	}
	return nil
}

func GetLiveChannelsIPTV(iptvProviderID int64, categoryID string) ([]LiveIPTVChannel, error) {
	if categoryID == "" {
		return nil, fmt.Errorf("category id must not be empty: %w", internal.BadRequestError)
	}
	provider, err := database.GetIPTVProvider(iptvProviderID)
	if err != nil {
		return nil, err
	}
	channels, err := sources.GetLiveChannelsIPTV(iptvProviderID, categoryID)
	if err != nil {
		return nil, err
	}
	var resp []LiveIPTVChannel
	for _, channel := range channels {
		temp := LiveIPTVChannel{
			IPTVProviderID:   iptvProviderID,
			Order:            channel.Number,
			StreamID:         channel.StreamID,
			Name:             channel.Name,
			XtreamStreamType: channel.StreamType,
			ThumbnailURL:     channel.StreamIcon,
			CategoryID:       categoryID,
			AddedAt:          channel.AddedOn,
		}
		if channel.EPGChannelID != nil {
			temp.EPGChannelID = *channel.EPGChannelID
		}
		// only no-proxy for now
		if !provider.ProxyStream {
			streamURL, err := sources.GetXtreamStreamLink(iptvProviderID, channel.StreamID)
			if err != nil {
				slog.Error("Failed to get stream link", "channel", channel, "error", err.Error())
				continue
			}
			temp.StreamURL = streamURL
		}
		resp = append(resp, temp)
	}
	return resp, nil
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

func encodeData(channel LiveIPTVChannel) (string, error) {
	data, err := json.Marshal(channel)
	if err != nil {
		return "", err
	}
	return internal.EncryptGCM(data)
}

func DecodeXtreamChannelData(encodedData string) (LiveIPTVChannel, error) {
	var channel LiveIPTVChannel
	raw, err := internal.DecryptGCM(encodedData)
	if err != nil {
		return channel, err
	}
	if err := json.Unmarshal(raw, &channel); err != nil {
		return channel, err
	}
	return channel, nil
}
