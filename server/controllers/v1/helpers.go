package v1

import (
	"fmt"
	"hound/database"
	"hound/helpers"
	"hound/sources"
	"strconv"
	"strings"
)

func validateMediaParams(mediaType string, mediaSource string) error {
	validType := mediaType == database.MediaTypeTVShow || mediaType == database.MediaTypeMovie || mediaType == database.MediaTypeGame
	if !validType {
		return fmt.Errorf("invalid media type: %w", helpers.BadRequestError)
	}
	validSource := mediaSource == sources.MediaSourceTMDB || mediaSource == sources.SourceIGDB
	if !validSource {
		return fmt.Errorf("invalid media source: %w", helpers.BadRequestError)
	}
	return nil
}

func getSourceIDFromParams(tmdbParam string) (string, int, error) {
	split := strings.Split(tmdbParam, "-")
	if len(split) != 2 {
		return "", -1, fmt.Errorf("invalid source id parameters: %w", helpers.BadRequestError)
	}
	id, err := strconv.ParseInt(split[1], 10, 64)
	// only accept tmdb ids for now
	if err != nil || split[0] != sources.MediaSourceTMDB && split[0] != sources.SourceIGDB {
		return "", -1, fmt.Errorf("invalid source id parameters: %w", helpers.BadRequestError)
	}
	return split[0], int(id), nil
}

// returns -1 if empty params which can be taken as no limit/offset
func getLimitOffset(limitQuery, offsetQuery string) (int, int, error) {
	limit := -1
	offset := -1
	if limitQuery != "" {
		var err error
		limit, err = strconv.Atoi(limitQuery)
		if err != nil {
			return -1, -1, fmt.Errorf("invalid limit query param: %w", helpers.BadRequestError)
		}
	}
	if offsetQuery != "" {
		var err error
		offset, err = strconv.Atoi(offsetQuery)
		if err != nil {
			return -1, -1, fmt.Errorf("invalid offset query param: %w", helpers.BadRequestError)
		}
	}
	return limit, offset, nil
}

func getSeasonEpisode(seasonQuery, episodeQuery string) (int, int, error) {
	var seasonNumber int
	if seasonQuery != "" {
		s, err := strconv.Atoi(seasonQuery)
		if err != nil {
			return -1, -1, fmt.Errorf("invalid season query param: %w", helpers.BadRequestError)
		}
		seasonNumber = s
	}
	var episodeNumber int
	if episodeQuery != "" {
		e, err := strconv.Atoi(episodeQuery)
		if err != nil {
			return -1, -1, fmt.Errorf("invalid episode query param: %w", helpers.BadRequestError)
		}
		episodeNumber = e
	}
	return seasonNumber, episodeNumber, nil
}
