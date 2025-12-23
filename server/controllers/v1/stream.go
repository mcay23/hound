package v1

import (
	"errors"
	"hound/helpers"
	"hound/model"
	"hound/model/sources"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

/*
Proxies links through the server
*/
func StreamHandler(c *gin.Context) {
	streamDetails, err := model.DecodeJsonStreamJWT(c.Param("encodedString"))
	if err != nil || streamDetails == nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError),
			"Failed to parse encoded string:"+c.Param("encodedString")))
		return
	}
	slog.Info("Initializing Stream ", "infohash", streamDetails.InfoHash,
		"filename", streamDetails.Filename)
	// Torrent/P2P Streaming Case
	if streamDetails.Cached == "false" && streamDetails.P2P == "p2p" {
		file, _, err := model.GetTorrentFile(streamDetails.InfoHash,
			streamDetails.FileIndex, streamDetails.Filename, streamDetails.Sources)
		if err != nil {
			helpers.ErrorResponse(c, err)
			return
		}
		// GetTorrentFile could return nil
		if file == nil {
			helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"Could not find file in torrent"+streamDetails.InfoHash)
			return
		}
		c.Writer.Header().Set("Content-Type", model.GetMimeType(file.DisplayPath()))
		reader := file.NewReader()
		defer func() {
			if closer, ok := reader.(io.Closer); ok {
				closer.Close()
			}
		}()
		http.ServeContent(c.Writer, c.Request, file.DisplayPath(), time.Time{}, reader)
		return
	}
	// Direct stream case, just proxy url
	videoURL := streamDetails.URL
	if videoURL == "" {
		c.String(http.StatusBadRequest, "Video URL not provided")
		return
	}
	req, err := http.NewRequest("GET", videoURL, nil)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Error creating URL: "+err.Error()))
		return
	}
	if rangeHeader := c.GetHeader("Range"); rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}
	if userAgent := c.GetHeader("User-Agent"); userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"HTTP error fetching URL: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	// Copy all headers from the remote response
	for name, values := range resp.Header {
		for _, value := range values {
			c.Header(name, value)
		}
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Accept-Ranges", "bytes")
	//c.Writer.Header().Set("Cache-Control", "no-store")
	c.Status(resp.StatusCode)

	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"IO copy error: "+err.Error()))
		return
	}
}

func AddTorrentHandler(c *gin.Context) {
	streamDetails, err := model.DecodeJsonStreamJWT(c.Param("encodedString"))
	if err != nil || streamDetails == nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError),
			"Failed to parse encoded string:"+c.Param("encodedString")))
		return
	}
	// may want to be more lax in the future
	if streamDetails.FileIndex == -1 || streamDetails.Filename == "" || streamDetails.InfoHash == "" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Torrent hash, File Index and/or File name not provided"))
		return
	}
	err = model.AddTorrent(streamDetails.InfoHash, streamDetails.Sources)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	helpers.SuccessResponse(c, gin.H{"status": "success"}, 200)
}

// This downloads the torrent to the server, not the client
func DownloadTorrentHandler(c *gin.Context) {
	streamDetails, err := model.DecodeJsonStreamJWT(c.Param("encodedString"))
	if err != nil || streamDetails == nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError),
			"Failed to parse encoded string:"+c.Param("encodedString")))
		return
	}
	if streamDetails.MediaSource != sources.MediaSourceTMDB {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Invalid media source: "+streamDetails.MediaSource))
		return
	}
	err = model.CreateIngestTaskDownload(streamDetails)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to download torrent"))
		return
	}
	helpers.SuccessResponse(c, gin.H{"status": "started"}, 200)
}
