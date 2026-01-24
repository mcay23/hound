package database

import (
	"errors"
	"fmt"
	"hound/helpers"
	"time"
)

const (
	mediaFilesTable = "media_files"
)

type MediaFile struct {
	FileID           int64   `xorm:"pk autoincr 'file_id'" json:"file_id"`
	Filepath         string  `xorm:"text not null unique 'file_path'" json:"file_path"`
	OriginalFilename string  `xorm:"text 'original_file_name'" json:"original_file_name"`
	RecordID         int64   `xorm:"index 'record_id'" json:"record_id"`
	SourceURI        *string `xorm:"text 'source_uri'" json:"source_uri"`
	FileIdx          *int    `xorm:"'file_idx'" json:"file_idx"`
	VideoMetadata    `xorm:"extends"`
	CreatedAt        time.Time `xorm:"timestampz created" json:"created_at"`
	UpdatedAt        time.Time `xorm:"timestampz updated" json:"updated_at"`
}

type VideoMetadata struct {
	Filename           string        `xorm:"text 'file_name'" json:"file_name"`
	Filesize           int64         `xorm:"'file_size'" json:"file_size"`
	FileFormat         string        `xorm:"text 'format_name'" json:"format_name"`
	FileFormatLongName string        `xorm:"text 'format_long_name'" json:"format_long_name"`
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
	return databaseEngine.Table(mediaFilesTable).Sync2(new(MediaFile))
}

func InsertMediaFile(mediaFileMetadata *MediaFile) (*MediaFile, error) {
	_, err := databaseEngine.Table(mediaFilesTable).Insert(mediaFileMetadata)
	return mediaFileMetadata, err
}

func GetMediaFile(fileID int) (*MediaFile, error) {
	var metadata MediaFile
	_, err := databaseEngine.Table(mediaFilesTable).Where("file_id = ?", fileID).
		Get(&metadata)
	return &metadata, err
}

func GetMediaFileByRecordID(recordID int) ([]*MediaFile, error) {
	var metadata []*MediaFile
	err := databaseEngine.Table(mediaFilesTable).
		Where("record_id = ?", recordID).Find(&metadata)
	return metadata, err
}

func GetMediaFiles(limit *int, offset *int) (int, []*MediaFile, error) {
	var files []*MediaFile
	sess := databaseEngine.Table(mediaFilesTable)
	// total number of all media files
	count, err := sess.Count()
	if err != nil {
		return 0, nil, err
	}
	// necessary for second run
	sess = sess.Table(mediaFilesTable)
	if limit != nil {
		if offset != nil {
			sess = sess.Limit(*limit, *offset)
		} else {
			sess = sess.Limit(*limit, 0)
		}
	}
	err = sess.Find(&files)
	return int(count), files, err
}

func DeleteMediaFileRecord(fileID int) error {
	affected, err := databaseEngine.Table(mediaFilesTable).ID(fileID).Delete(&MediaFile{})
	if affected == 0 {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			fmt.Sprintf("Media file with id %d not found", fileID))
	}
	return err
}
