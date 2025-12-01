package constants

// AppVersion is the current version of ShipShipShip
// This should match the version in admin/package.json
const AppVersion = "1.2.4"

// VersionInfo contains additional version metadata
type VersionInfo struct {
	Version string
	Build   string
	Commit  string
}

// GetVersionInfo returns the current version information
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version: AppVersion,
		Build:   "production",
		Commit:  "", // Can be set during build with -ldflags
	}
}
