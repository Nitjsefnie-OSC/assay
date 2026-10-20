- Make verifyloop association scanning reuse repository observations across one queue
  pass and immutable graph registries across briefs, using the shared gitcore reader.
  Dispatch and admission retain their fresh Git checks.
- Add fixed perf-v3 association phases, a reproducible offline queue benchmark, and
  parity/freshness regressions. Reuse qualified blocker and offline input observations
  per scan, refresh mutable scheduling observations on every selection, and skip queue
  and provider work when every multi-root checkout is rejected by preflight.
- Preserve effective origin include/worktree config semantics with one identity read per
  opened root; retain fresh dispatch/admission checks. Add the mutation harness control.
