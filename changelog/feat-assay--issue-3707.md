### Fixed
- The cell Go cache build tests no longer fail intermittently with "TempDir RemoveAll cleanup: directory not empty": each scratch HOME they hand to `go build` now has Go telemetry turned off first, so no detached telemetry child is left writing into the test's temporary directory while it is removed. A package guard keeps any new test from building a scratch HOME any other way.
