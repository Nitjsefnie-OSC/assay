# Bounded reviewed-head observations

Deskboard's `RE-REVIEW` fallback remains authoritative whenever the reviewed-head
interval cannot be established completely. The shared GitHub adapter can now read
more than the former single-response commit/file windows without adding a second
review classifier or observation store.

## Placement and bounds

`GitHubForge.CompareRefs` collects up to ten numeric pages of 100 commits. A stable,
nonnegative declared total, unique commit identities, matching status/divergence
metadata and terminal pagination evidence are required for `CommitsComplete`.
Further pages require both requested endpoints to be immutable commit identifiers;
shared callers passing moving branch names retain a conservative incomplete result.
Only the first page's aggregate file window is projected into `RefComparison.Files`:
that existing field is not newly claimed to be complete.

`GitHubForge.GetCommit` collects up to thirty numeric pages of 100 files. A branch
argument is pinned to the first response's immutable SHA before further pages. SHA
and parent ordering must remain stable, and filenames must be present and unique.
Rename sources are preserved. Missing file arrays, repeated data, transport/JSON
errors and continuation without progress never become a complete observation.
An exactly full final file page without a next link requires another bounded read;
GitHub's 3000-file ceiling therefore remains incomplete even at an exact boundary.

Every page uses the existing authenticated client, request observer and deskboard's
conditional transport/quota floor. Link headers signal continuation only: provider
URLs are never followed. No permission, identity, quota-floor, decision, approval,
trust, admission, breaker, audit or telemetry-consent setting changes.

Deskboard still walks only the first-parent chain and checks each commit's files.
Edits later reverted and merge conflict resolutions remain visible; unrelated
second-parent history is not fetched as commit detail. Its existing per-PR fact
store stores only complete intervals and retains its existing invalidation key.
No new disk cache, TTL, service or global mutable observation owner is introduced.
The future shared review view can consume these existing typed completeness values;
collection remains in the credentialed adapter and action derivation remains pure.

The limits are per operation, not a promise of a small whole-sweep request budget:
a long first-parent chain can still need many commit reads, and the existing sweep
quota floor can refuse the board before that work finishes. Such a refusal stays a
refusal. The incomplete-data fallback is not waived to obtain a cheaper result.

## Consumer routing

- Fixed here: shared GitHub `GetCommit` and `CompareRefs` pagination and evidence.
- Preserved callers: deskboard, deskdispatch review preparation, deskevidence,
  deskpost, deskflip and deskautolane use the same Forge methods and ordinary errors.
  Metadata consumers still receive commit metadata when file completeness is false;
  callers that require file/history evidence must honor the completeness fields.
- Unchanged: GitLab adapter, aggregate compare-file semantics, review queue snapshot,
  actor/exact-head policies, local git readers and decision-admission sweep reuse.
- Existing perf-v4 codec/readers and producer migration belong to desktools-v2/67–68;
  this patch adds no recorder, schema, emitter migration or remote collection.

## Qualification and measurement

Native adapter regressions cover complete multi-page reads and exact boundaries;
short/full pages with and without a continuation signal; stable counts and identities;
renames, duplicate/missing entries, changing refs, malformed/cut responses, later-page
errors and bounded exhaustion. Native deskboard regressions exercise first-parent
classification and complete-only interval reuse, including retry after partial data.
The dedicated mutation spec qualifies with a green baseline, caught positive control,
13 caught faults, zero survivors and zero unmutatable faults. All seven shared adapter
and consumer suites pass (deskkit 419.611s, deskboard 23.532s, deskdispatch 78.407s,
deskevidence 85.295s, deskpost 8.281s, deskflip 1.012s, deskautolane 0.512s). Focused race,
vet, source lint, mutation-anchor and committed test-ledger checks pass. No original
test or control is retired. Full suites exercised identical production bytes before
the final measurement-only fixture extension; that extension was exercised in the
paired processes and focused race run.

Matched measurement uses the actual native adapter, first-parent reader and shared
`PerfRunMeasured` writer/reader with only remote HTTP replaced by a fixture. Fixture
requests have an injected 8 ms delay. Each variant is a fresh process; small and large
cohorts are measured separately. A complete read and an incomplete fallback are not
equivalent work, so extra large-cohort requests are reported as a completeness cost,
not an invocation-speed improvement or an estimate of reviewer/model savings.

**perf-v4 qualification pending:** baseline main 917772b3 contains the v4 design and
migration briefs but no v4 codec or producer. Actual records must retain their real
`perf-v3` schema; relabeling those records or creating a second v4 writer would not
satisfy the requested measurement. Re-run this fixture with the canonical v4 producer
once that separately owned migration lands. No v4, production or rollout claim is made.

Matched runs (five fresh processes per variant/cohort, alternating order):

| Interval fixture | HTTP before → after | Median wall ms | Observed nearest-rank p95 wall ms | Median own CPU ms | Evidence |
|---|---:|---:|---:|---:|---|
| Small cold: 1 commit, 2 files | 2 → 2 | 47 → 44 | 396 → 186 | 14 → 15 | complete → complete |
| Large cold: 251 commits, 301 files | 1 → 7 | 66 → 83 | 91 → 85 | 16 → 21 | incomplete → complete |
| Eight unchanged large-interval sweeps | 8 → 7 | 130 → 96 | 536 → 208 | 35 → 26 | incomplete → complete |

The repeat cohort replays one unchanged externally observed PR key through actual
fact-store save/load. All five request-count repetitions agree. It saves one interval
request across eight such sweeps (12.5%) and recovers complete evidence on all eight;
repo snapshots, checks/reviews and consequential rechecks remain outside this interval
fixture. There is no whole-board, reviewer-cost, fleet, rate-limit-point or production
speedup claim. The large cold cohort does additional work, not equivalent cheaper work.
Host scheduling was noisy (small-before wall samples span 34–396 ms);
these five-sample quantiles describe only the observations, not a latency guarantee.
CPU values and unsupported/unknown point coverage are retained without a general savings claim.

Baseline production: `917772b3ee4709bc47e5d0888cb5cc698444b13e`.
After source and identical fixture: `7c0b4a1608c8d99ae15ff947dfdb81fd588e7814`.
Before binary SHA256: `efe0ab2870a7ba33f411a57ff6350c06b69802d6c619a674fdd45cebbb86898f`.
After binary SHA256: `6b1d09b1bd44d24d00575fe8bf9163c465021eeabb53ddf0c56e9314f9b83855`.
Each record's native HTTP attempts matched a separate fixture-server witness. No
historical record was relabeled and no v4 or repeated mutation/performance run is invented.

API contract: [GitHub commit and comparison endpoints](https://docs.github.com/en/rest/commits/commits).
