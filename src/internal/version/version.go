// Package version exposes build and repository metadata for ShotGum.
package version

// DefaultVersion is the single source of truth for the project version.
// To bump the project version, change only this constant.
const DefaultVersion = "0.2.0"

// Version is what the app renders and what Cobra reports via --version.
// It intentionally tracks DefaultVersion exactly.
var Version = DefaultVersion

// Repo is the canonical "<owner>/<name>" slug on GitHub.
// Update this single constant when the repository is renamed or transferred.
const Repo = "brunoomariano/ShotGum-Toolchain"

// RepoURL is the full GitHub URL derived from Repo.
const RepoURL = "https://github.com/" + Repo
