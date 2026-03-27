package providers

import (
	"fmt"
	"github.com/mcay23/hound/helpers"
	"net/http"
	"strings"
)

// check if provider is online/valid
func PingProviderManifest(manifestURL string) error {
	if manifestURL == "" {
		return fmt.Errorf("manifest url is empty: %w", helpers.BadRequestError)
	}
	if !strings.Contains(manifestURL, "manifest.json") {
		if !strings.HasSuffix(manifestURL, "/") {
			manifestURL += "/"
		}
		manifestURL += "manifest.json"
	}
	req, err := http.NewRequest("GET", manifestURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("provider manifest not found: %s: %s", manifestURL, resp.Status)
	}
	return nil
}
