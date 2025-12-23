package services

import (
	"context"
	"encoding/json"
	"os/exec"
	"time"
)

type FfprobeOutput struct {
	Format struct {
		Filename           string `json:"filename"`
		FileFormat         string `json:"format_name"`
		FileFormatLongName string `json:"format_long_name"`
		Size               string `json:"size"`
		Duration           string `json:"duration"`
		Bitrate            string `json:"bit_rate"`
	} `json:"format"`
	Streams []struct {
		CodecType      string            `json:"codec_type"`
		CodecName      string            `json:"codec_name"`
		CodecLongName  string            `json:"codec_long_name"`
		Channels       int               `json:"channels,omitempty"`
		ChannelLayout  string            `json:"channel_layout,omitempty"`
		PixelFormat    string            `json:"pix_fmt,omitempty"`
		ColorPrimaries string            `json:"color_primaries,omitempty"`
		ColorTransfer  string            `json:"color_transfer,omitempty"`
		ColorSpace     string            `json:"color_space,omitempty"`
		ColorRange     string            `json:"color_range,omitempty"`
		Profile        string            `json:"profile,omitempty"`
		CodecTag       string            `json:"codec_tag_string"`
		Width          int               `json:"width,omitempty"`
		Height         int               `json:"height,omitempty"`
		AvgFrameRate   string            `json:"avg_frame_rate,omitempty"`
		Tags           map[string]string `json:"tags,omitempty"`
	} `json:"streams"`
}

func FFProbe(uri string) (*FfprobeOutput, error) {
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		uri,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, ffmpegEngine.ffprobePath, args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var output FfprobeOutput
	if err := json.Unmarshal(out, &output); err != nil {
		return nil, err
	}
	return &output, nil
}
