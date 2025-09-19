package v1

import (
	"github.com/gin-gonic/gin"
	"hound/helpers"
	"io"
	"log"
	"net/http"
)

func StreamHandler(c *gin.Context) {
	//videoURL := "https://repo.jellyfin.org/archive/jellyfish/media/jellyfish-15-mbps-hd-h264.mkv"
	videoURL := "https://sgp1-4.download.real-debrid.com/d/UL2WN7D3LQ4BK40/Tastefully%20Yours%20S01E08%201080p%20NF%20WEB-DL%20AAC2%200%20H%20264-Kitsune.mkv"
	if videoURL == "" {
		c.String(http.StatusBadRequest, "Video URL not provided")
		return
	}
	req, err := http.NewRequest("GET", videoURL, nil)
	if err != nil {
		helpers.ErrorResponse(c, err)
		return
	}

	// Forward the client's Range header to the remote request.
	if rangeHeader := c.GetHeader("Range"); rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	if userAgent := c.GetHeader("User-Agent"); userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		helpers.ErrorResponse(c, err)
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
	c.Writer.Header().Set("Cache-Control", "no-store")
	c.Status(resp.StatusCode)
	// Stream the body of the remote response directly to the client.
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		log.Printf("io.Copy error: %v", err)
		return
	}
}