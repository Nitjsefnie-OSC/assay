# Verifyloop association scan qualification

This follow-on implements the association-scan part of
[the verification performance specification](verifyloop-performance-spec.md).
It reuses gitcore and the perfrec perf-v3 recorder introduced for verifyloop in
[#3738](https://github.com/medici-finance/assay/pull/3738). Other provider, startup,
publication and future service work retains its existing owners.

The algorithm previously restarted checkout identity, HEAD and registry discovery for
each declared brief, then repeated tracking identity/HEAD for cross-repository items.
`SelectQueue` now gives the entire scan one pass-scoped reader. Work is proportional to
distinct roots, root/commit registries and root/commit/path declarations. Handles,
observed failures and parsed declarations are discarded at the end of the scan.
No permanent cache, new service or alternative declaration parser is introduced.

## Reproducible neutral benchmark

`BenchmarkAssociationQueue` constructs a complete offline fixture: one tracking checkout,
one source checkout, both at fetched main. One third of briefs are legacy, one third declare
the tracking repository and one third declare the source repository. All rows are
implemented, have an empty Evidence section and resolve successfully. Fixture creation
and Git commits are outside timing; the measured operation is production `SelectQueue`,
without a forge source. This exercises planning, not preflight or execution admission.

Baseline: `c7afc51121d636c0a6200f454dc45287ad35e32f`. Use the benchmark and its fixture
helpers from this change with unchanged baseline production files, then with this change's
production files. The baseline was built using a Go overlay restoring those production
files from the named commit; test inputs and fixture helpers were identical.

From `tools/desk`, run each binary with the equivalent of:

```sh
go test ./cmd/verifyloop -run '^$' -bench '^BenchmarkAssociationQueue$' -benchtime=1x -count=5
```

Five alternating before/after samples per size, Apple M4 Max, Go 1.27.2, medians:

| Briefs | Baseline scan | Pass-scoped gitcore scan | Reduction |
|---|---:|---:|---:|
| 12 | 356.7 ms | 8.5 ms | 97.6% |
| 96 | 2,905.3 ms | 69.9 ms | 97.6% |

A separate perf-v3 observation of the 96-row fixture measured 448 local Git subprocesses
before and zero after, with zero HTTP operations on both sides. It recorded two root
opens, one registry validation and 64 declaration resolutions after the change.

These are synthetic component timings, not an end-to-end or fleet SLA. Filesystem cache,
repository object layout, brief mix and machine load affect results; the local benchmark
is not a cold-disk trial. Provider requests and points saved by this change: zero. Normal
preflight, credential work, pending-Evidence reads, row execution and publication remain.

## Correctness and freshness

The regression suite compares production queue items against the original fresh reader
on all 96 fixture rows. The production scan still succeeds with Git absent from PATH and
hostile caller Git variables, proving that its local association path is in-process and
root-qualified. Telemetry records 64 declaration resolutions, two root opens and one
registry read for that fixture, using fixed labels only.

A second selection observes new source commits and recovers after a failed registry is
repaired. Immutable declaration keys separate commits, paths and roots, including linked
worktrees sharing objects. Returned declaration slices cannot mutate memoized results.
Open failures remain stable for one pass and are retried on a new pass.

After planning, move source or tracking HEAD, fetched main, origin, a declaration, the graph
registry or the configured source root. Both native and interim dispatch must refuse before
creating a worktree or invoking a feeder. Those checks still use the existing fresh Git
reader; the pass reader is never stored on a queue item or loop. The independent receipt,
source-closure and landing checks remain in place.

Gitcore reads repository-local remote identity. The configured-URL seam preserves raw
identity rather than applying fetch rewrites or falling back to pushurl. Remote overrides
from other config scopes are not qualified by this planner; dispatch retains its existing
full fresh config semantics. Missing input remains unavailable and invalid declarations
retain their policy refusals. Lower-level I/O diagnostic text may name gitcore rather
than a Git subprocess failure.

## Diagnostics

The existing `ASSAY_TELEMETRY=1 ASSAY_PERF_PROFILE=tool` opt-in records the new fixed phases
`association`, `association-root` and `association-registry` in perf-v3. Their counts show
committed declaration queries, distinct open attempts and registry validation attempts.
Nested timings are inclusive; do not add them to `queue-scan` or process wall time.
No new exporter, spool, metric schema or high-cardinality label is added. Missing provider
coverage remains missing; local reuse does not imply a provider-cache hit.

## Sibling call sites

| Caller | Treatment |
|---|---|
| `SelectQueue`, single-root and multi-root scans | One reader per scan, shared across all roots. |
| `scanAwaiting` used by verdict commands | Fresh pass per scan through the same scan helper. |
| Native/interim dispatch, attested receipt admission, `VerifierDeliverables` admission callers | Existing fresh Git reads and shared declaration policy retained. |
| Pending-Evidence source reads and forge observations | Existing owners/caches retained; other reuse work remains in the parent specification. |

## Validation

The affected `cmd/verifyloop`, `internal/deskkit` and `internal/gitcore` suites passed,
along with documentation lint. Seven compiling overlay mutations were caught by assertion
failures: omit the dispatch recheck, omit the pass's fetched-main gate, collapse declaration
revision keys, reread registries, collapse root keys, accept mutable revision keys, and
allocate a fresh reader per scanned item. The unmodified tests pass. The anchors and test
command are in `tools/desk/cmd/verifyloop/association-pass-mutations.json`.
