//nolint:gochecknoglobals
package version

var (
	// BuildTime is a time label of the moment when the binary was built. Sets at compile-time.
	BuildTime = "unset"
	// Commit is a last commit hash at the moment when the binary was built. Sets at compile-time.
	Commit = "unset"
)

// GetBuildInfo returns build time and commit hash.
func GetBuildInfo() (string, string) {
	return BuildTime, Commit
}
