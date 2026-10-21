package deskkit

import (
	"fmt"
	"io"
	"runtime/debug"

	"github.com/medici-finance/assay/perfrec"
)

// Version stamping. SourceSHA and BuiltAt are injected at build time
// by the Makefile's desk-install / desk-build targets via:
//
//	-ldflags "-X github.com/medici-finance/assay/tools/desk/internal/deskkit.SourceSHA=<short-sha> \
//	          -X github.com/medici-finance/assay/tools/desk/internal/deskkit.BuiltAt=<rfc3339>"
//
// A `go run` / unstamped build leaves them empty; Version() then reports the
// toolchain's vcs.revision / vcs.time when the binary carries them, else "unpinned",
// and WarnIfUnpinned announces the missing stamp on stderr either way — so running
// from source (drift risk) is visible in every transcript.
var (
	SourceSHA = ""
	BuiltAt   = ""
	// ReleaseTag is the artifact-namespaced release tag (`desk-tools/vX.Y.Z`)
	// stamped by .github/workflows/release.yml (the desk-tools build step) via
	//   -X …/internal/deskkit.ReleaseTag=$RELEASE_TAG
	// mirroring release-statusgen.yml's `-X main.statusgenVersion`. It is what
	// maps a RUNNING binary back to the release it was cut from: SourceSHA alone
	// is a commit, not a version, so before this stamp nothing could say which
	// `desk-tools/vX.Y.Z` a binary is. A `go run` / unstamped build leaves it
	// empty, reported as "dev" (ReleaseTagOrDev) — never a fabricated release.
	ReleaseTag = ""
)

const unpinned = "unpinned"

// readBuildInfo is the build-info reader Version falls back to; a var so a test can present
// a binary with or without VCS information.
var readBuildInfo perfrec.BuildInfoFunc = debug.ReadBuildInfo

// devRelease is what an unstamped build answers for its release tag. It mirrors
// statusgen's `dev` default: a binary that was not cut from a release must say so
// rather than claim a version, so a source build can never be mistaken for a pin.
const devRelease = "dev"

// Version returns (sourceSHA, builtAt). A value not stamped in at link time falls back to
// the toolchain-embedded vcs.revision (short) / vcs.time — so a build path that omits the
// ldflags stamp still names its commit — and to "unpinned" when the binary carries no VCS
// information either (`go run`, -buildvcs=false). Both are echoed in every audit record and
// in each tool's --version output. IsPinned is NOT affected: it keys on the link-time stamp
// alone, so an unstamped build still warns.
func Version() (sourceSHA, builtAt string) {
	s := perfrec.RevisionFrom(SourceSHA, readBuildInfo)
	b := perfrec.BuildTimeFrom(BuiltAt, readBuildInfo)
	if s == "" {
		s = unpinned
	}
	if b == "" {
		b = unpinned
	}
	return s, b
}

// ReleaseTagOrDev returns the stamped release tag, or "dev" for an unstamped
// build. Callers print this on `--version` alongside SourceSHA/BuiltAt so a
// running desk-tools binary can be mapped back to its `desk-tools/vX.Y.Z`.
func ReleaseTagOrDev() string {
	if ReleaseTag == "" {
		return devRelease
	}
	return ReleaseTag
}

// IsPinned reports whether the binary was stamped (installed via `sudo make
// desk-install`) rather than run from source. It keys on SourceSHA/BuiltAt
// only — the ReleaseTag stamp is additive and does NOT change which builds are
// pinned, so every build that reported IsPinned() before this stamp still does.
func IsPinned() bool { return SourceSHA != "" && BuiltAt != "" }

// WarnIfUnpinned writes a one-line WARNING to w when the binary is unpinned, so an
// operator sees drift risk in the transcript. It is a no-op for a stamped binary.
func WarnIfUnpinned(w io.Writer) {
	if IsPinned() {
		return
	}
	fmt.Fprintln(w, "desk-tools WARNING: running UNPINNED (go run / unstamped build) — "+
		"sourceSHA/builtAt not embedded, drift is invisible. Install via `sudo make desk-install`.")
}
