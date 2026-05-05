package providers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
)

const MANIFEST_PATH = "/manifest.json"
const TV_SERIES_PATH = "/series/%s:%d:%d.json"
const MOVIES_PATH = "/movie/%s.json"
const TV_STREAMS_PATH = "/stream" + TV_SERIES_PATH
const MOVIE_STREAMS_PATH = "/stream" + MOVIES_PATH
const TV_SUBTITLES_PATH = "/subtitles" + TV_SERIES_PATH
const MOVIE_SUBTITLES_PATH = "/subtitles" + MOVIES_PATH

type StremioStreamBehaviorHints struct {
	BingeGroup *string `json:"bingeGroup,omitempty"`
	VideoHash  *string `json:"videoHash,omitempty"`
	Filename   *string `json:"filename,omitempty"`
	VideoSize  *int    `json:"videoSize,omitempty"` // size in bytes
}

// Pretty much everything is optional per Stremio docs,
// but url/infohash are required
// only http/p2p streams are supported for now
type StremioStreamObject struct {
	Name          *string                     `json:"name,omitempty"`
	Title         *string                     `json:"title,omitempty"`       // will be deprecated in stremio according to docs
	Description   *string                     `json:"description,omitempty"` // title will be replaced with description
	URL           *string                     `json:"url,omitempty"`
	InfoHash      *string                     `json:"infoHash,omitempty"`
	FileIdx       *int                        `json:"fileIdx,omitempty"`
	Sources       *[]string                   `json:"sources,omitempty"`
	BehaviorHints *StremioStreamBehaviorHints `json:"behaviorHints,omitempty"`
}

type StremioStreamResponse struct {
	Streams []StremioStreamObject `json:"streams,omitempty"`
}

type StremioSubtitleObject struct {
	ID   string `json:"id"`
	Lang string `json:"lang"`
	URL  string `json:"url"`
}

type StremioSubtitlesResponse struct {
	Subtitles []StremioSubtitleObject `json:"subtitles,omitempty"`
}

func getStremioStreams(query ProvidersQueryRequest, details StreamMediaDetails) (*ProviderStreamObject, error) {
	if query.ProviderProfileID == nil {
		return nil, fmt.Errorf("nil provider profile id: %w", internal.BadRequestError)
	}
	url := ""
	provider, err := database.GetProviderProfile(*query.ProviderProfileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider profile: %w", err)
	}
	url += provider.ManifestURL
	switch query.MediaType {
	case database.MediaTypeMovie:
		url += fmt.Sprintf(MOVIE_STREAMS_PATH, query.IMDbID)
	case database.MediaTypeTVShow:
		if query.SeasonNumber == nil || query.EpisodeNumber == nil {
			return nil, fmt.Errorf("query %s invalid season/episode number", query.MediaType)
		}
		url += fmt.Sprintf(TV_STREAMS_PATH, query.IMDbID, *query.SeasonNumber, *query.EpisodeNumber)
	default:
		return nil, fmt.Errorf("query %s invalid media type", query.MediaType)
	}
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query %s-%s non-200 response received from stremio plugin status %s: %w",
			query.MediaSource, query.SourceID, resp.Status, internal.GatewayTimeoutError)
	}
	var stremioResp StremioStreamResponse
	if err := json.NewDecoder(resp.Body).Decode(&stremioResp); err != nil {
		return nil, fmt.Errorf("query %s-%s error decoding stremio plugin response: %w",
			query.MediaSource, query.SourceID, err)
	}
	streamResponse := []*StreamObject{}
	for _, stream := range stremioResp.Streams {
		obj, err := stream.toStreamObject(details, provider.Name, int(provider.ProviderProfileID))
		// if unexpected response in an object, skip instead of blocking
		if err != nil {
			slog.Debug("convert stremio stream to generic stream object",
				"stream", stream, "error", err)
			continue
		}
		streamResponse = append(streamResponse, obj)
	}
	providerObject := &ProviderStreamObject{
		Provider: provider.Name,
		Streams:  streamResponse,
	}
	return providerObject, nil
}

// convert stremio results to a generic stream object
func (stremioStream *StremioStreamObject) toStreamObject(details StreamMediaDetails,
	providerName string, providerID int) (*StreamObject, error) {
	if stremioStream == nil {
		return nil, fmt.Errorf("nil stremio stream: %w", internal.BadRequestError)
	}
	uri := ""
	infoHash := ""
	streamProtocol := ""
	// http case
	if stremioStream.URL != nil {
		streamProtocol = database.ProtocolProxyHTTP
		uri = *stremioStream.URL
		tempInfoHash, ok := internal.ExtractInfoHashFromURL(*stremioStream.URL)
		if ok {
			infoHash = tempInfoHash
		}
	} else {
		// p2p case
		if stremioStream.InfoHash == nil {
			slog.Debug("Bad stream found", "stream", stremioStream)
			return nil, fmt.Errorf("invalid stremio stream, infohash is nil for type p2p: %w", internal.BadRequestError)
		}
		streamProtocol = database.ProtocolP2P
		infoHash = *stremioStream.InfoHash
		uri = internal.GetMagnetURI(infoHash, stremioStream.Sources)
	}
	// last sanity check
	if uri == "" {
		return nil, fmt.Errorf("invalid stremio stream, uri is empty: %w", internal.BadRequestError)
	}
	// stremio description is either the title (deprecated soon) or description
	// for our object, the title is not the stremio 'title' field but the name
	title := ""
	description := ""
	if stremioStream.Name != nil {
		title = *stremioStream.Name
	}
	if stremioStream.Description != nil {
		description = *stremioStream.Description
	} else if stremioStream.Title != nil {
		description = *stremioStream.Title
	}
	// next, we encode the full stream object, including metadata such as movie/show, season, episode,
	// and stream details such as source uri and which provider profile it's from.
	// this encoded data is used in the final stream link
	streamObject := &StreamObject{
		ProviderProfileName: providerName,
		ProviderProfileID:   providerID,
		StreamProtocol:      streamProtocol,
		URI:                 uri,
		Title:               title,
		Description:         description,
		InfoHash:            infoHash,
		Filename:            stremioStream.BehaviorHints.Filename,
		FileIdx:             stremioStream.FileIdx,
		FileSize:            stremioStream.BehaviorHints.VideoSize,
		Sources:             stremioStream.Sources,
		VideoMetadata:       nil,
	}
	// create encoding from full object
	streamObjectFull := StreamObjectFull{
		StreamObject:       *streamObject,
		StreamMediaDetails: details,
	}
	encodedData, err := EncodeJsonStreamAES(streamObjectFull)
	if err != nil {
		return nil, fmt.Errorf("aes encoding: %w", err)
	}
	streamObject.EncodedData = encodedData
	return streamObject, nil
}

func getStremioSubtitles(query ProvidersQueryRequest) (*ProviderSubtitleObject, error) {
	url := ""
	provider, err := database.GetProviderProfile(*query.ProviderProfileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider profile: %w", err)
	}
	url += provider.ManifestURL
	switch query.MediaType {
	case database.MediaTypeMovie:
		url += fmt.Sprintf(MOVIE_SUBTITLES_PATH, query.IMDbID)
	case database.MediaTypeTVShow:
		if query.SeasonNumber == nil || query.EpisodeNumber == nil {
			return nil, fmt.Errorf("query %s invalid season/episode number", query.MediaType)
		}
		url += fmt.Sprintf(TV_SUBTITLES_PATH, query.IMDbID, *query.SeasonNumber, *query.EpisodeNumber)
	default:
		return nil, fmt.Errorf("query %s invalid media type", query.MediaType)
	}
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query %s-%s non-200 response received from stremio plugin status %s: %w",
			query.MediaSource, query.SourceID, resp.Status, internal.GatewayTimeoutError)
	}
	var stremioResp StremioSubtitlesResponse
	if err := json.NewDecoder(resp.Body).Decode(&stremioResp); err != nil {
		return nil, fmt.Errorf("query %s-%s error decoding stremio plugin response: %w",
			query.MediaSource, query.SourceID, err)
	}
	validSubs := []SubtitleObject{}
	for _, sub := range stremioResp.Subtitles {
		if !internal.IsValidURL(sub.URL) {
			slog.Debug("invalid subtitle url, skipping", "subtitle", sub)
			continue
		}
		encodedURI, err := EncodeURIAES(sub.URL)
		if err != nil {
			slog.Debug("error encoding subtitle url, skipping", "subtitle", sub, "error", err)
			continue
		}
		validSubs = append(validSubs, SubtitleObject{
			ProviderProfileID:   int(provider.ProviderProfileID),
			ProviderProfileName: provider.Name,
			URI:                 sub.URL,
			EncodedData:         encodedURI,
			Language:            sub.Lang,
			Title:               sub.Lang + " | " + provider.Name,
		})
	}
	return &ProviderSubtitleObject{
		ProviderProfileID:   int(provider.ProviderProfileID),
		ProviderProfileName: provider.Name,
		Subtitles:           validSubs,
	}, nil
}
