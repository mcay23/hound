package model

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/sources"
	"github.com/sherif-fanous/xmltv"
)

var EPGFilepath = filepath.Join(internal.HoundIPTVDownloadsPath, "epg.xml")
var epg xmltv.EPG

type LiveIPTVChannel struct {
	IPTVConfigID int       `json:"iptv_config_id"`
	StreamID     int       `json:"stream_id"`
	Name         string    `json:"name"`
	StreamType   string    `json:"stream_type"`
	ThumbnailURL string    `json:"thumbnail_url"`
	EPGChannelID string    `json:"epg_channel_id"`
	CategoryID   string    `json:"category_id"`
	AddedAt      time.Time `json:"added_at"`
	EncodedData  string    `json:"encoded_data"`
}

func InitializeIPTV() {
	sources.InitializeXtream()
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

func GetLiveChannelsIPTV(categoryID string) []LiveIPTVChannel {
	channels, err := sources.GetLiveChannelsIPTV(categoryID)
	if err != nil {
		_ = internal.LogErrorWithMessage(err, "Failed to get live channels")
		return nil
	}
	var resp []LiveIPTVChannel
	for _, channel := range channels {
		temp := LiveIPTVChannel{
			StreamID:     channel.StreamID,
			Name:         channel.Name,
			StreamType:   channel.StreamType,
			ThumbnailURL: channel.StreamIcon,
			CategoryID:   categoryID,
			AddedAt:      channel.AddedOn,
		}
		if channel.EPGChannelID != nil {
			temp.EPGChannelID = *channel.EPGChannelID
		}
		resp = append(resp, temp)
	}
	return resp
}
