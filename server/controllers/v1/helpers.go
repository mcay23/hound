package v1

import (
	"errors"
	"hound/helpers"
	"hound/model/database"
	"hound/model/sources"
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
