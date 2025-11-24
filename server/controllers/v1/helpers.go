package v1

import (
	"errors"
	"hound/helpers"
	"hound/model/database"
	"hound/model/sources"
	"strconv"
	"strings"
)

func ValidateMediaParams(mediaType string, mediaSource string) error {
	validType := mediaType == database.MediaTypeTVShow || mediaType == database.MediaTypeMovie || mediaType == database.MediaTypeGame
	if !validType {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid media type")
	}
	validSource := mediaSource == sources.SourceTMDB || mediaSource == sources.SourceIGDB
	if !validSource {
		return helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "Invalid media source")
	}
	return nil
}

func GetSourceIDFromParams(tmdbParam string) (string, int, error) {
	split := strings.Split(tmdbParam, "-")
	if len(split) != 2 {
		return "", -1, errors.New(helpers.BadRequest + "Invalid source id parameters")
	}
	id, err := strconv.ParseInt(split[1], 10, 64)
	// only accept tmdb ids for now
	if err != nil || split[0] != sources.SourceTMDB && split[0] != sources.SourceIGDB {
		return "", -1, errors.New(helpers.BadRequest + "Invalid source id parameters")
	}
	return split[0], int(id), nil
}
