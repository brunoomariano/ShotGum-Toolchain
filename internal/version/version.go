// Package version exposes the build version, overridable at link time.
//
//	go build -ldflags="-X 'github.com/brunoomariano/ShotGum-Toolchain/internal/version.Version=v0.2.0'"
package version

// DefaultVersion is the baseline version for local and dev builds.
// Releases override this at link time via ldflags (see Makefile LDFLAGS).
// To bump the project version, change only this constant.
const DefaultVersion = "0.2.0"

// Version is injected by the build system via ldflags; falls back to
// DefaultVersion so local builds always carry a meaningful version string.
var Version = DefaultVersion

// Repo is the canonical "<owner>/<name>" slug on GitHub.
// Update this single constant when the repository is renamed or transferred.
const Repo = "brunoomariano/ShotGum-Toolchain"

// RepoURL is the full GitHub URL derived from Repo.
const RepoURL = "https://github.com/" + Repo
