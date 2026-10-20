### Changed
- The unreleased per-call attribution flag `--perf-desk-role` is now `--perf-loop-role` (desk tools, statusgen, skills and docs), because the old literal packed next to other string data in the desk-base binaries and failed the release image secret scan; a new regression test builds the shipped binary set and applies that scanner's own patterns on every pull request.
