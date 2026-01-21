package providers

import (
	"hound/database"
	"hound/model/sources"
	"os"
	"strconv"
)

func GetLocalStreamsForMovie(sourceID int) ([]*StreamObject, error) {
	// set title, but if error should not block response
	// we want to be able to serve files without a connection
	// need to check if latency is high in this case
	title := ""
	movieDetails, err := sources.GetMovieFromIDTMDB(sourceID)
	if err == nil {
		title = movieDetails.Title
		if movieDetails.ReleaseDate != "" && len(movieDetails.ReleaseDate) >= 4 {
			title += " (" + movieDetails.ReleaseDate[0:4] + ")"
		}
	}
	has, record, err := database.GetMediaRecord(database.MediaTypeMovie, "tmdb", strconv.Itoa(sourceID))
	if err != nil {
		return nil, err
	}
	if !has {
		return []*StreamObject{}, nil
	}
	mediaFiles, err := database.GetMediaFileByRecordID(record.RecordID)
	if err != nil {
		return nil, err
	}
	streamObjects := []*StreamObject{}
	for _, file := range mediaFiles {
		if _, err := os.Stat(file.Filepath); os.IsNotExist(err) {
			continue
		}
		streamObj, err := mapMediaFileToStreamObject(strconv.Itoa(sourceID), file, record, title)
		if err != nil {
			continue
		}
		streamObjects = append(streamObjects, streamObj)
	}
	return streamObjects, nil
}

func GetLocalStreamsForTVShow(showID int, seasonNumber int, episodeNumber int) ([]*StreamObject, error) {
	// check note on above
	title := ""
	showDetails, err := sources.GetTVShowFromIDTMDB(showID)
	if err == nil {
		title = showDetails.Name
		if showDetails.FirstAirDate != "" && len(showDetails.FirstAirDate) >= 4 {
			title += " (" + showDetails.FirstAirDate[0:4] + ")"
		}
		title += " - S" + strconv.Itoa(seasonNumber) + "E" + strconv.Itoa(episodeNumber)
	}
	episodeRecord, err := database.GetEpisodeMediaRecord("tmdb", strconv.Itoa(showID),
		&seasonNumber, episodeNumber)
	if err != nil {
		return nil, err
	}
	mediaFiles, err := database.GetMediaFileByRecordID(episodeRecord.RecordID)
	if err != nil {
		return nil, err
	}
	streamObjects := []*StreamObject{}
	for _, file := range mediaFiles {
		if _, err := os.Stat(file.Filepath); os.IsNotExist(err) {
			continue
		}

		streamObj, err := mapMediaFileToStreamObject(strconv.Itoa(showID), file, episodeRecord, title)
		if err != nil {
			continue
		}
		streamObjects = append(streamObjects, streamObj)
	}
	return streamObjects, nil
}

// we don't need to include everything in the encode, since uri is usually enough
// re-add metadata for the json response
func mapMediaFileToStreamObject(sourceID string, file *database.MediaFile, record *database.MediaRecord, title string) (*StreamObject, error) {
	fileSize := int(file.Filesize)
	if title == "" {
		title = file.VideoMetadata.Filename
	}
	streamObj := &StreamObject{
		Provider:       "Hound",
		StreamProtocol: database.ProtocolFileHTTP,
		URI:            file.Filepath,
		Title:          title,
		Description:    "Local file: " + file.Filepath,
		FileSize:       &fileSize,
		VideoMetadata:  nil,
	}
	details := StreamMediaDetails{
		MediaType:   record.RecordType,
		MediaSource: record.MediaSource,
		SourceID:    sourceID,
	}
	if record.RecordType == database.RecordTypeEpisode {
		details.MediaType = database.MediaTypeTVShow
		details.SeasonNumber = record.SeasonNumber
		details.EpisodeNumber = record.EpisodeNumber
		details.EpisodeSourceID = &record.SourceID
		epID := strconv.Itoa(int(record.RecordID))
		details.EpisodeSourceID = &epID
	}
	streamObjectFull := StreamObjectFull{
		StreamObject:       *streamObj,
		StreamMediaDetails: details,
	}
	encodedData, err := EncodeJsonStreamAES(streamObjectFull)
	if err != nil {
		return nil, err
	}
	streamObj.EncodedData = encodedData
	streamObj.VideoMetadata = &file.VideoMetadata
	return streamObj, nil
}
