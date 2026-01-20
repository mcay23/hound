package providers

import (
	"errors"
	"fmt"
	"hound/helpers"
)

// This maps tmdb source ids to tmdb episode groups that
// match tvdb ordering. Especially useful for anime, this list
// should be updated over time
// If there's a better way to convert tmdb season/episode -> tvdb,
// please let me know

type MapEntry struct {
	EpisodeGroupID string `json:"episode_group_id"`
	Title          string `json:"_title"`
	Comment        string `json:"_comment"`
}

var episodeGroupMapping map[string]MapEntry

func init() {
	episodeGroupMapping = make(map[string]MapEntry)
	episodeGroupMapping["tmdb-209867"] = MapEntry{
		EpisodeGroupID: "679231eba8ce3489ceb57efc",
		Title:          "Frieren: Beyond Journey's End",
		Comment:        "",
	}
	episodeGroupMapping["tmdb-95479"] = MapEntry{
		EpisodeGroupID: "64a3fc4fe9da6900ae2fa807",
		Title:          "JUJUTSU KAISEN",
		Comment:        "",
	}
}

func GetEpisodeGroupMapping(mediaSource string, sourceID string) (string, error) {
	if episodeGroupMapping == nil {
		return "", helpers.LogErrorWithMessage(errors.New(helpers.BadRequest), "episode group mapping not initialized")
	}
	id := fmt.Sprintf("%s-%s", mediaSource, sourceID)
	if entry, ok := episodeGroupMapping[id]; ok {
		return entry.EpisodeGroupID, nil
	}
	// not mapping found, don't return error
	return "", nil
}
