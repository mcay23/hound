package database

import "time"

const (
	mediaFilesTable = "media_files"
)

type MediaFileMetadata struct {
	FileID           int64  `xorm:"pk autoincr 'file_id'" json:"file_id"`
	Filepath         string `xorm:"not null unique 'file_path'" json:"file_path"`
	OriginalFilename string `xorm:"'original_file_name'" json:"original_file_name"`
	RecordID         int64  `xorm:"'record_id'" json:"record_id"`
	VideoMetadata    `xorm:"extends"`
	CreatedAt        time.Time `xorm:"timestampz created" json:"created_at"`
	UpdatedAt        time.Time `xorm:"timestampz updated" json:"updated_at"`
}

type VideoMetadata struct {
	Filename           string        `xorm:"'file_name'" json:"file_name"`
	Filesize           int64         `xorm:"'file_size'" json:"file_size"`
	FileFormat         string        `xorm:"'format_name'" json:"format_name"`
	FileFormatLongName string        `xorm:"'format_long_name'" json:"format_long_name"`
	Duration           time.Duration `xorm:"'duration'" json:"duration"`
	Bitrate            string        `xorm:"'bit_rate'" json:"bit_rate"`
	Width              int           `xorm:"'width'" json:"width,omitempty"`
	Height             int           `xorm:"'height'" json:"height,omitempty"`
	Framerate          string        `xorm:"'framerate'" json:"framerate,omitempty"`
	Streams            []Stream      `xorm:"json 'streams'" json:"streams"`
}

type Stream struct {
	Title          string `json:"title,omitempty"`
	CodecType      string `json:"codec_type"` // video, audio, subtitle
	CodecName      string `json:"codec_name"`
	CodecLongName  string `json:"codec_long_name"`
	Profile        string `json:"profile,omitempty"`
	Channels       int    `json:"channels,omitempty"`
	ChannelLayout  string `json:"channel_layout,omitempty"`
	PixelFormat    string `json:"pix_fmt,omitempty"`
	ColorPrimaries string `json:"color_primaries,omitempty"`
	ColorTransfer  string `json:"color_transfer,omitempty"`
	ColorSpace     string `json:"color_space,omitempty"`
	ColorRange     string `json:"color_range,omitempty"`
	Language       string `json:"language,omitempty"`
}

func instantiateMediaFilesTable() error {
	return databaseEngine.Table(mediaFilesTable).Sync2(new(MediaFileMetadata))
}

func InsertMediaFile(recordID int64, filepath string, originalFilename string, videoMetadata *VideoMetadata) (*MediaFileMetadata, error) {
	mediaFileMetadata := MediaFileMetadata{
		Filepath:         filepath,
		OriginalFilename: originalFilename,
		RecordID:         recordID,
		VideoMetadata:    *videoMetadata,
	}
	_, err := databaseEngine.Table(mediaFilesTable).Insert(&mediaFileMetadata)
	return &mediaFileMetadata, err
}

func GetMediaFile(fileID int64) (*MediaFileMetadata, error) {
	var metadata MediaFileMetadata
	_, err := databaseEngine.Table(mediaFilesTable).Where("file_id = ?", fileID).Get(&metadata)
	return &metadata, err
}

func GetMediaFileByRecordID(recordID int64) ([]*MediaFileMetadata, error) {
	var metadata []*MediaFileMetadata
	err := databaseEngine.Table(mediaFilesTable).Where("record_id = ?", recordID).Find(&metadata)
	return metadata, err
}
