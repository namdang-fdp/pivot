// Package buildinfo exposes immutable metadata about the Pivot binary.
package buildinfo

// These values are replaced at release time with linker flags.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// Info is the stable structured representation of build metadata.
type Info struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

// Current returns metadata for the running Pivot binary.
func Current() Info {
	return Info{
		Name:    "pivot",
		Version: Version,
		Commit:  Commit,
		Date:    Date,
	}
}
