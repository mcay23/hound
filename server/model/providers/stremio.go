package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"hound/database"
	"hound/helpers"
	"net/http"
	"time"
)

// const BASE_URL = "http://localhost:3500/stremio/6bc22eb8-2cf3-4c1c-b264-3dd47b5972b0/eyJpdiI6InU5NEpFOFJGaVUrMEI0U216VStzaFE9PSIsImVuY3J5cHRlZCI6IldzUlYxbWR4dUdlb1B2RGxES2htTnc9PSIsInR5cGUiOiJhaW9FbmNyeXB0In0"

const BASE_URL = "https://aiostreamsfortheweebs.midnightignite.me/stremio/be1079b5-bc45-4338-aba5-12797469ae95/eyJpIjoiZzVBWHlRUEUzdVRYa0o1MzBrTzRmQT09IiwiZSI6IkdHN3FWcjRRRVdjVisyU29ITTU4a1E9PSIsInQiOiJhIn0"
const MANIFEST_PATH = "/manifest.json"
const TV_STREAMS_PATH = "/stream/series/%s:%d:%d.json"
const MOVIE_STREAMS_PATH = "/stream/movie/%s.json"

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

func getStremioStreams(request ProvidersQueryRequest, details StreamMediaDetails) (*ProviderObject, error) {
	url := ""
	providerName := "Stremio"
	switch request.MediaType {
	case database.MediaTypeMovie:
		url = BASE_URL + fmt.Sprintf(MOVIE_STREAMS_PATH, request.IMDbID)
	case database.MediaTypeTVShow:
		if request.SeasonNumber == nil || request.EpisodeNumber == nil {
			return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid season/episode number")
		}
		url = BASE_URL + fmt.Sprintf(TV_STREAMS_PATH, request.IMDbID, *request.SeasonNumber, *request.EpisodeNumber)
	default:
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid media type")
	}
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError),
			"Error querying stremio plugin"+resp.Status)
	}
	var stremioResp StremioStreamResponse
	if err := json.NewDecoder(resp.Body).Decode(&stremioResp); err != nil {
		return nil, err
	}
	var streamResponse []*StreamObject
	for _, stream := range stremioResp.Streams {
		obj, err := stream.toStreamObject(details, providerName)
		// if unexpected response in an object, skip instead of blocking
		if err != nil {
			helpers.LogErrorWithMessage(err,
				"Error converting stremio stream to generic stream object")
			continue
		}
		streamResponse = append(streamResponse, obj)
	}
	providerObject := &ProviderObject{
		Provider: providerName,
		Streams:  streamResponse,
	}
	return providerObject, nil
}

// convert stremio results to a generic stream object
func (stremioStream *StremioStreamObject) toStreamObject(details StreamMediaDetails,
	providerName string) (*StreamObject, error) {
	if stremioStream == nil {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Invalid stremio stream")
	}
	uri := ""
	infoHash := ""
	streamProtocol := ""
	// http case
	if stremioStream.URL != nil {
		streamProtocol = database.ProtocolProxyHTTP
		uri = *stremioStream.URL
		tempInfoHash, ok := helpers.ExtractInfoHashFromURL(*stremioStream.URL)
		if ok {
			infoHash = tempInfoHash
		}
	} else {
		// p2p case
		if stremioStream.InfoHash == nil {
			return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
				"Invalid stremio stream, no url/infohash provided")
		}
		streamProtocol = database.ProtocolP2P
		infoHash = *stremioStream.InfoHash
		uri = helpers.GetMagnetURI(infoHash, stremioStream.Sources)
	}
	// last sanity check
	if uri == "" {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.BadRequest),
			"Invalid stremio stream, uri is empty")
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
	streamObject := &StreamObject{
		Provider:       providerName,
		StreamProtocol: streamProtocol,
		URI:            uri,
		Title:          title,
		Description:    description,
		InfoHash:       infoHash,
		Filename:       stremioStream.BehaviorHints.Filename,
		FileIdx:        stremioStream.FileIdx,
		FileSize:       stremioStream.BehaviorHints.VideoSize,
		Sources:        stremioStream.Sources,
		VideoMetadata:  nil,
	}
	// create encoding from full object
	streamObjectFull := StreamObjectFull{
		StreamObject:       *streamObject,
		StreamMediaDetails: details,
	}
	encodedData, err := EncodeJsonStreamAES(streamObjectFull)
	if err != nil {
		return nil, err
	}
	streamObject.EncodedData = encodedData
	return streamObject, nil
}
