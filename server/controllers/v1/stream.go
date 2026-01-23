package v1

import (
	"errors"
	"hound/database"
	"hound/helpers"
	"hound/model"
	"hound/model/providers"
	"hound/model/sources"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

/*
Proxies links through the server
*/
func StreamHandler(c *gin.Context) {
	streamDetails, err := providers.DecodeJsonStreamAES(c.Param("encodedString"))
	if err != nil || streamDetails == nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError),
			"Failed to parse encoded string:"+c.Param("encodedString")))
		return
	}
	slog.Info("Initializing Stream ", "infohash", streamDetails.InfoHash,
		"filename", streamDetails.Filename)

	if streamDetails.StreamProtocol == database.ProtocolP2P {
		handleP2PStream(c, streamDetails)
		return
	}
	if streamDetails.StreamProtocol == database.ProtocolFileHTTP {
		handleFileStream(c, streamDetails)
		return
	}
	// Direct stream case, just proxy url
	handleProxyStream(c, streamDetails)
}

func handleFileStream(c *gin.Context, streamDetails *providers.StreamObjectFull) {
	filePath := streamDetails.URI
	if filePath == "" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"File path not provided"))
		return
	}
	// Verify file exists
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"File not found: "+filePath))
		} else {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Error accessing file"))
		}
		return
	}
	c.Writer.Header().Set("Content-Type", model.GetMimeType(filePath))
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Accept-Ranges", "bytes")
	http.ServeFile(c.Writer, c.Request, filePath)
}

func handleP2PStream(c *gin.Context, streamDetails *providers.StreamObjectFull) {
	if streamDetails.InfoHash == "" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Infohash not provided"))
		return
	}
	// fileIdx can sometimes be null, gettorrentfile will automatically grab
	// largest video file in that case
	file, fileIdx, _, err := model.GetTorrentFile(streamDetails.InfoHash,
		streamDetails.FileIdx, streamDetails.Sources)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	// GetTorrentFile could return nil
	if file == nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Could not find file in torrent"+streamDetails.InfoHash))
		return
	}
	c.Writer.Header().Set("Content-Type", model.GetMimeType(file.DisplayPath()))
	// if file already exists, serve that instead
	// this is an edge case, completed files
	// aren't served properly by the reader if the torrent session is restarted
	// and files are still in the download path
	// ideally, dropped torrents should delete its download folder immediately/
	// but on restarts, this would be an issue since we want to resume downloads
	stat, err := os.Stat(filepath.Join(model.HoundP2PDownloadsPath, streamDetails.InfoHash, file.Path()))
	if err == nil {
		f, err := os.Open(filepath.Join(model.HoundP2PDownloadsPath, streamDetails.InfoHash, file.Path()))
		if err != nil {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err,
				"Failed to open file"))
			return
		}
		if file.Length() != stat.Size() {
			helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err,
				"File exists but size mismatch"))
			return
		}
		_ = model.AddActiveTorrentStream(streamDetails.InfoHash, fileIdx)
		defer model.RemoveActiveTorrentStream(streamDetails.InfoHash, fileIdx)
		defer f.Close()
		http.ServeContent(
			c.Writer,
			c.Request,
			stat.Name(),
			stat.ModTime(),
			f,
		)
		return
	}
	// if file doesn't exist, serve it from torrent
	reader := file.NewReader()
	defer func() {
		if closer, ok := reader.(io.Closer); ok {
			closer.Close()
		}
	}()
	// add/remove active streams for this index for cleanup tracking
	// remove active torrent streams extends session lifetime by a few minutes for cleanup grace
	_ = model.AddActiveTorrentStream(streamDetails.InfoHash, fileIdx)
	defer model.RemoveActiveTorrentStream(streamDetails.InfoHash, fileIdx)
	slog.Info("Streaming file", "file", file.DisplayPath())
	http.ServeContent(c.Writer, c.Request, file.DisplayPath(), time.Time{}, reader)
}

func handleProxyStream(c *gin.Context, streamDetails *providers.StreamObjectFull) {
	videoURL := streamDetails.URI
	if videoURL == "" {
		c.String(http.StatusBadRequest, "Video URL not provided")
		return
	}
	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", videoURL, nil)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Error creating URL: "+err.Error()))
		return
	}
	if rangeHeader := c.GetHeader("Range"); rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}
	// mock browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
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
	streamDetails, err := providers.DecodeJsonStreamAES(c.Param("encodedString"))
	if err != nil || streamDetails == nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError),
			"Failed to parse encoded string:"+c.Param("encodedString")))
		return
	}
	if streamDetails.StreamProtocol != database.ProtocolP2P {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Invalid stream protocol, has to be p2p: "+streamDetails.StreamProtocol))
		return
	}
	// may want to be more lax in the future
	if streamDetails.FileIdx == nil || streamDetails.InfoHash == "" {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Torrent hash, File Index and/or File name not provided"))
		return
	}
	err = model.AddTorrent(streamDetails.InfoHash, streamDetails.Sources)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}
	helpers.SuccessResponse(c, nil, 200)
}

// This downloads the media file to the server, not the client
func DownloadHandler(c *gin.Context) {
	streamDetails, err := providers.DecodeJsonStreamAES(c.Param("encodedString"))
	if err != nil || streamDetails == nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError),
			"Failed to parse encoded string:"+c.Param("encodedString")))
		return
	}
	if streamDetails.StreamProtocol == database.ProtocolFileHTTP {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"This file should already be downloaded"))
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

// Cancel downloads
func CancelIngestTaskHandler(c *gin.Context) {
	taskIDStr := c.Param("taskID")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Invalid task ID Param"))
		return
	}
	task, err := database.GetIngestTask(database.IngestTask{IngestTaskID: int64(taskID)})
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to get task"))
		return
	}
	if task.Status != database.IngestStatusDownloading &&
		task.Status != database.IngestStatusPendingDownload {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Only tasks that are downloading or pending_download can be canceled"))
		return
	}
	updatedTask := database.IngestTask{
		IngestTaskID: int64(taskID),
		Status:       database.IngestStatusCanceled,
	}
	_, err = database.UpdateIngestTask(&updatedTask)
	if err != nil {
		helpers.ErrorResponse(c, helpers.LogErrorWithMessage(err, "Failed to cancel task"))
		return
	}
	helpers.SuccessResponse(c, gin.H{"ingest_task_id": taskID, "status": "pending_cancel"}, 200)
}
