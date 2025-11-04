package version

// Version information for all services
const (
	// Version is the current version of all services
	Version = "1.0.0"

	// BuildDate can be set during build time using -ldflags
	BuildDate = "unknown"

	// GitCommit can be set during build time using -ldflags
	GitCommit = "unknown"
)

// Info contains version information
type Info struct {
	Version   string `json:"version"`
	BuildDate string `json:"build_date"`
	GitCommit string `json:"git_commit"`
}

// GetVersionInfo returns version information
func GetVersionInfo() Info {
	return Info{
		Version:   Version,
		BuildDate: BuildDate,
		GitCommit: GitCommit,
	}
}
