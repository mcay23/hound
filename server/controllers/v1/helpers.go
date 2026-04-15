package v1

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/sources"
)

func validateMediaParams(mediaType string, mediaSource string) error {
	validType := mediaType == database.MediaTypeTVShow || mediaType == database.MediaTypeMovie || mediaType == database.MediaTypeGame
	if !validType {
		return fmt.Errorf("invalid media type: %w", internal.BadRequestError)
	}
	validSource := mediaSource == sources.MediaSourceTMDB
	if !validSource {
		return fmt.Errorf("invalid media source: %w", internal.BadRequestError)
	}
	return nil
}

func getUserIDFromContext(c *gin.Context) (int64, error) {
	temp, exists := c.Get("userID")
	if !exists || temp == nil {
		return -1, fmt.Errorf("invalid user: %w", internal.BadRequestError)
	}
	userID, ok := temp.(int64)
	if !ok {
		return -1, fmt.Errorf("invalid user: %w", internal.BadRequestError)
	}
	// since id is autoincremented from 1 in postgres, treat 0 as invalid
	if userID == 0 {
		return -1, fmt.Errorf("invalid user id 0 parsed: %w", internal.BadRequestError)
	}
	return userID, nil
}

func getSourceIDFromParams(tmdbParam string) (string, int, error) {
	split := strings.Split(tmdbParam, "-")
	if len(split) != 2 {
		return "", -1, fmt.Errorf("invalid source id parameters: %w", internal.BadRequestError)
	}
	id, err := strconv.ParseInt(split[1], 10, 64)
	// only accept tmdb ids for now
	if err != nil || split[0] != sources.MediaSourceTMDB {
		return "", -1, fmt.Errorf("invalid source id parameters: %w", internal.BadRequestError)
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
			return -1, -1, fmt.Errorf("invalid limit query param: %w", internal.BadRequestError)
		}
	}
	if offsetQuery != "" {
		var err error
		offset, err = strconv.Atoi(offsetQuery)
		if err != nil {
			return -1, -1, fmt.Errorf("invalid offset query param: %w", internal.BadRequestError)
		}
	}
	return limit, offset, nil
}

func getSeasonEpisode(seasonQuery, episodeQuery string) (int, int, error) {
	if seasonQuery == "" || episodeQuery == "" {
		return -1, -1, fmt.Errorf("invalid season query param: %w", internal.BadRequestError)
	}
	seasonNumber, err := strconv.Atoi(seasonQuery)
	if err != nil {
		return -1, -1, fmt.Errorf("invalid season query param: %w", internal.BadRequestError)
	}
	episodeNumber, err := strconv.Atoi(episodeQuery)
	if err != nil {
		return -1, -1, fmt.Errorf("invalid episode query param: %w", internal.BadRequestError)
	}
	return seasonNumber, episodeNumber, nil
}
