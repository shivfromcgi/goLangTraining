package version

import (
	"os"
)

// Version information variables that will be set at build time via -ldflags
// or pulled from environment variables at runtime
var (
	// Version can be set during build time using -ldflags "-X cgi.com/goLangTraining/src/pkg/version.Version=x.y.z"
	Version = "dev"

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

// GetVersion returns the version, preferring environment variable over build-time value
func GetVersion() string {
	if envVersion := os.Getenv("APP_VERSION"); envVersion != "" {
		return envVersion
	}
	return Version
}

// GetBuildDate returns the build date, preferring environment variable over build-time value
func GetBuildDate() string {
	if envBuildDate := os.Getenv("BUILD_DATE"); envBuildDate != "" {
		return envBuildDate
	}
	return BuildDate
}

// GetGitCommit returns the git commit, preferring environment variable over build-time value
func GetGitCommit() string {
	if envGitCommit := os.Getenv("GIT_COMMIT"); envGitCommit != "" {
		return envGitCommit
	}
	return GitCommit
}

// GetVersionInfo returns version information from environment or build-time values
func GetVersionInfo() Info {
	return Info{
		Version:   GetVersion(),
		BuildDate: GetBuildDate(),
		GitCommit: GetGitCommit(),
	}
}
