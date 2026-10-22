### Added
- `deskfile new --for-pr <owner>/<repo>#<N> [--brief <id>]` files a workflow-bound issue (one an
  open PR's review or verification requires) that the per-session filing budget does not count.
  The PR must be open and the brief must be the PR's own brief. Bound filings audit as
  `new-bound`, are capped per role and per PR by the new `issue.file.bound_per_pr` knob (default 0,
  off), still charge the per-PR and repo-wide outward-write tiers, and back-link the new issue on
  the PR. `--for-pr` does not bind the filer to the PR; the counting is the bound.
