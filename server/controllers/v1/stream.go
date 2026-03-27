package v1

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/model"
	"github.com/mcay23/hound/providers"

	"github.com/gin-gonic/gin"
)

/*
Proxies links through the server
*/
// @Router /api/v1/stream/{encodedString} [get]
// @Summary Stream Video
// @Description A streamable link for a video defined by the encodedString
// @Tags Stream
// @Accept json
// @Produce json
// @Param encodedString path string true "Encoded Stream Details"
// @Success 200 {object} V1SuccessResponse{data=object}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func StreamHandler(c *gin.Context) {
	streamDetails, err := providers.DecodeJsonStreamAES(c.Param("encodedString"))
	if err != nil || streamDetails == nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to decode aes stream with encodedString %s: %w", c.Param("encodedString"), err))
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
		internal.ErrorResponse(c, fmt.Errorf("invalid filePath: %w", internal.BadRequestError))
		return
	}
	// Verify file exists
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			internal.ErrorResponse(c, fmt.Errorf("file not found: %s: %w", filePath, err))
		} else {
			internal.ErrorResponse(c, fmt.Errorf("error accessing file: %w", err))
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
		internal.ErrorResponse(c, fmt.Errorf("invalid infohash: %w", internal.BadRequestError))
		return
	}
	// fileIdx can sometimes be null, gettorrentfile will automatically grab
	// largest video file in that case
	file, fileIdx, _, err := model.GetTorrentFile(streamDetails.InfoHash,
		streamDetails.FileIdx, streamDetails.Sources)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	// GetTorrentFile could return nil
	if file == nil {
		internal.ErrorResponse(c, fmt.Errorf("could not find file in torrent %s: %w", streamDetails.InfoHash, internal.BadRequestError))
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
			internal.ErrorResponse(c, fmt.Errorf("failed to open file: %w", err))
			return
		}
		if file.Length() != stat.Size() {
			internal.ErrorResponse(c, fmt.Errorf("file exists but size mismatch: %w", err))
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
		internal.ErrorResponse(c, fmt.Errorf("error creating URL: %w", err))
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
		internal.ErrorResponse(c, fmt.Errorf("http error fetching url: %w", err))
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
		internal.ErrorResponse(c, fmt.Errorf("io copy error: %w", err))
		return
	}
}

// @Router /api/v1/torrent/{encodedString} [post]
// @Summary Add Torrent
// @Description Adds a p2p torrent to the server for streaming/download.
// @Description Not strictly necessary, as calling stream directly invokes this. May be deprecated in the future.
// @Tags Stream
// @Accept json
// @Produce json
// @Param encodedString path string true "Encoded Stream Details"
// @Success 200 {object} V1SuccessResponse{data=object}
// @Failure 400 {object} V1ErrorResponse
// @Failure 500 {object} V1ErrorResponse
func AddTorrentHandler(c *gin.Context) {
	streamDetails, err := providers.DecodeJsonStreamAES(c.Param("encodedString"))
	if err != nil || streamDetails == nil {
		internal.ErrorResponse(c, fmt.Errorf("failed to decode aes stream with encodedString %s: %w", c.Param("encodedString"), err))
		return
	}
	if streamDetails.StreamProtocol != database.ProtocolP2P {
		internal.ErrorResponse(c, fmt.Errorf("invalid stream protocol, has to be p2p: %s: %w", streamDetails.StreamProtocol, internal.BadRequestError))
		return
	}
	// may want to be more lax in the future
	if streamDetails.FileIdx == nil || streamDetails.InfoHash == "" {
		internal.ErrorResponse(c, fmt.Errorf("torrent hash, file index and/or file name not provided: %w", internal.BadRequestError))
		return
	}
	err = model.AddTorrent(streamDetails.InfoHash, streamDetails.Sources)
	if err != nil {
		internal.ErrorResponse(c, err)
		return
	}
	internal.SuccessResponse(c, nil, 200)
}
