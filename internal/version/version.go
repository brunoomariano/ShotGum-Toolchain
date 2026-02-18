// Package version exposes the build version, overridable at link time.
//
//	go build -ldflags="-X 'github.com/shotgum/stg/internal/version.Version=v1.2.3'"
package version

// Version is set by the build system via ldflags. Falls back to "dev" for
// local builds where no tag is injected.
var Version = "dev"
