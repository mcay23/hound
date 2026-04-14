package internal

var (
	Version   = "development"
	BuildTime = "unknown"
	Commit    = "unknown"
)

type BuildInfo struct {
	Version   string `json:"version"`
	BuildTime string `json:"buildTime"`
	Commit    string `json:"commit"`
}

func GetBuildInfo() BuildInfo {
	return BuildInfo{
		Version:   Version,
		BuildTime: BuildTime,
		Commit:    Commit,
	}
}
