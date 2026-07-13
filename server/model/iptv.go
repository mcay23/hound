package model

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/sources"
	"github.com/sherif-fanous/xmltv"
)

var EPGFilepath = filepath.Join(internal.HoundIPTVDownloadsPath, "epg.xml")
var epg xmltv.EPG

type LiveIPTVChannel struct {
	IPTVProfileID    int64     `json:"iptv_profile_id"`
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

func InitializeIPTV() {
	err := os.MkdirAll(internal.HoundIPTVDownloadsPath, 0755)
	if err != nil {
		_ = internal.LogErrorWithMessage(err, "Failed to create IPTV downloads directory")
		panic(fmt.Errorf("fatal error creating IPTV downloads directory %w", err))
	}
	err = sources.DownloadEPGXtream(EPGFilepath)
	if err != nil {
		_ = internal.LogErrorWithMessage(err, "Failed to download EPG file")
		panic(fmt.Errorf("fatal error downloading EPG file %w", err))
	}
	data, err := os.ReadFile(EPGFilepath)
	if err != nil {
		panic(fmt.Errorf("fatal error reading EPG file %w", err))
	}
	if err := xml.Unmarshal(data, &epg); err != nil {
		panic(fmt.Errorf("fatal error parsing XMLTV %w", err))
	}
	fmt.Printf("Found %d channels and %d programmes\n", len(epg.Channels), len(epg.Programmes))
	for i, channel := range epg.Channels {
		fmt.Printf("Channel id: %s\n", channel.ID)
		for _, displayName := range channel.DisplayNames {
			fmt.Printf("Channel name: %s\n", displayName.Text)
		}
		if i > 10 {
			break
		}
	}
	fmt.Println("--------------cats---------------")
	cats, err := sources.GetLiveChannelsIPTV("2352")
	if err != nil {
		_ = internal.LogErrorWithMessage(err, "Failed to get live categories")
		return
	}
	for i, cat := range cats {
		fmt.Printf("Channel id: %d, name: %s\n", cat.StreamID, cat.Name)
		if i == 10 {
			break
		}
	}
}

// TODO encrypt password
func AddIPTVProfileXtream(name string, host string, username string, password string) error {
	if name == "" || host == "" || username == "" || password == "" {
		return fmt.Errorf("iptv profile name, host, username and password must not be empty: %w", internal.BadRequestError)
	}
	profile := &database.IPTVProfile{
		Name:              name,
		Host:              host,
		Username:          username,
		EncryptedPassword: password,
		ProxyStream:       false,
		IPTVStreamType:    database.IPTVStreamTypeXTREAM,
	}
	_, err := database.AddIPTVProfile(profile)
	if err != nil {
		return err
	}
	return nil
}

func GetLiveChannelsIPTV(iptvProfileID int64, categoryID string) ([]LiveIPTVChannel, error) {
	if categoryID == "" {
		return nil, fmt.Errorf("category id must not be empty: %w", internal.BadRequestError)
	}
	profile, err := database.GetIPTVProfile(iptvProfileID)
	if err != nil {
		return nil, err
	}
	channels, err := sources.GetLiveChannelsIPTV(categoryID)
	if err != nil {
		return nil, err
	}
	var resp []LiveIPTVChannel
	for _, channel := range channels {
		temp := LiveIPTVChannel{
			IPTVProfileID:    iptvProfileID,
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
		if !profile.ProxyStream {
			streamURL, err := sources.GetXtreamStreamLink(channel.StreamID)
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

func GetChannelEPGs(iptvProfileID int64, EPGChannelIDs []string) ([]EPGProgramme, error) {
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
