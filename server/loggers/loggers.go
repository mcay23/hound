package loggers

import "github.com/mcay23/hound/helpers"

func InitializeLoggers() {
	err := initializeIngestLogger()
	if err != nil {
		helpers.LogErrorWithMessage(err, "Failed to initialize external library logger")
		panic(err)
	}
}
