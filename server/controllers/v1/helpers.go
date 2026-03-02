package v1

import (
	"errors"
	"hound/database"
	"hound/helpers"
	"hound/sources"
	"strconv"
	"strings"
)

func validateMediaParams(mediaType string, mediaSource string) error {
	validType := mediaType == database.MediaTypeTVShow || mediaType == database.MediaTypeMovie || mediaType == database.MediaTypeGame
	if !validType {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid media type")
	}
	validSource := mediaSource == sources.MediaSourceTMDB || mediaSource == sources.SourceIGDB
	if !validSource {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid media source")
	}
	return nil
}

func getSourceIDFromParams(tmdbParam string) (string, int, error) {
	split := strings.Split(tmdbParam, "-")
	if len(split) != 2 {
		return "", -1, errors.New(helpers.BadRequest + "Invalid source id parameters")
	}
	id, err := strconv.ParseInt(split[1], 10, 64)
	// only accept tmdb ids for now
	if err != nil || split[0] != sources.MediaSourceTMDB && split[0] != sources.SourceIGDB {
		return "", -1, errors.New(helpers.BadRequest + "Invalid source id parameters")
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
			_ = helpers.LogErrorWithMessage(err, "Invalid limit query param")
			return -1, -1, errors.New(helpers.BadRequest)
		}
	}
	if offsetQuery != "" {
		var err error
		offset, err = strconv.Atoi(offsetQuery)
		if err != nil {
			_ = helpers.LogErrorWithMessage(err, "Invalid offset query param")
			return -1, -1, errors.New(helpers.BadRequest)
		}
	}
	return limit, offset, nil
}
