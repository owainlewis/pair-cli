// Package version exposes the pair CLI version.
package version

import "runtime/debug"

// Version is the release version, injected at build time via
//
//	-ldflags "-X github.com/owainlewis/pair-cli/internal/version.Version=v1.2.3"
//
// It is empty for local and `go install` builds, where String falls back
// to VCS build info.
var Version = ""

// String returns the resolved version: the ldflags value if set, otherwise
// the module version recorded by the Go toolchain (e.g. for `go install`),
// and finally "dev".
func String() string {
	if Version != "" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return "dev"
}
