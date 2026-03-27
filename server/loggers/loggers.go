package loggers

import "github.com/mcay23/hound/internal"

func InitializeLoggers() {
	err := initializeIngestLogger()
	if err != nil {
		internal.LogErrorWithMessage(err, "Failed to initialize external library logger")
		panic(err)
	}
}
