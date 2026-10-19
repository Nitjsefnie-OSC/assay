---
brief: assay:assay:desktools-v2:65
title: "Measure own forge request timing and GraphQL cost"
why: "HTTP counts alone cannot explain forge latency or quota consumption. Existing transport and provider response seams must expose actual attempts, durations and attributable GraphQL cost without issuing monitoring requests or mistaking shared quota deltas for per-request charges."
wave: 2
depends: ["desktools-v2/63"]
unblocks: ["desktools-v2/62", "performance-collection/02"]
effort: M
gate: model
risk: {regulatory: no, customer: no, irreversible: no, sensitive-data: no}
issues: []
schema: brief-v3
version: 1
id: 72fd07fd-5323-4631-8807-fc07c277ca3c
authored: 2026-10-19 by performance attribution authoring
sources:
  - docs/performance-attribution-spec.md
  - docs/streams/desktools-v2/spec.md
  - docs/telemetry.md
  - docs/library-first.md
  - "https://docs.github.com/en/graphql/overview/rate-limits-and-query-limits-for-the-graphql-api#returning-the-point-value-of-a-query (checked 2026-10-19)"
  - "https://docs.gitlab.com/api/graphql/getting_started/ (checked 2026-10-19; complexity is not GitHub quota points)"
  - "freshness-checked 2026-10-19 @ 2e02111ed; perf-v2 implementation 2013b8dc6, no perf-v3 or perfdata module"
exec-tier: strong
exec-tier-why: "Cross-process identity, compatibility and negative-path coverage require design judgment across shared consumers."
domain: complicated
outcome: none
consumers:
  - "tools/desk/internal/deskkit/forgetransport.go and forge backends: fixed-here"
  - "tools/desk/cmd/deskboard/readcost.go: fixed-here"
  - "tools/desk/internal/credential/: fixed-here (observe existing authentication HTTP separately)"
  - "perfrec/: follow-up desktools-v2/63"
  - "tools/desk/cmd/deskboard and deskdispatch: follow-up desktools-v2/62"
  - "statusgen/: out-of-scope (child owns provider calls; no authenticated backend import)"
  - "perfdata/adapters/: follow-up performance-collection/02"
  - "tools/desk/internal/forgeservice/: follow-up forge-neutral/45"
---

# Brief 65 — Measure own forge request timing and GraphQL cost

## Context

files:
- `tools/desk/internal/deskkit/forgetransport.go`
- `tools/desk/internal/deskkit/forge_github.go`
- `tools/desk/internal/deskkit/forge_gitlab.go`
- `tools/desk/internal/deskkit/perf.go`
- `tools/desk/internal/deskkit/perfcoverage_test.go`
- `tools/desk/internal/deskkit/perfforge_test.go (planned)`
- `tools/desk/cmd/deskboard/readcost.go`
- `tools/desk/internal/credential/`
- `docs/telemetry.md`
- `tools/desk/README.md`
- `changelog/perf-v3-forge-cost.md (planned)`

facts (source and date in frontmatter):
- The counting transport at 2e02111ed counts attempts but does not time or inspect bodies. deskkit PerfRun observes it as a lower-bound http counter.
- Existing backend/board seams already classify actual reads, 304s and local quota-floor refusals. Reuse them; do not create a second incompatible request counter.
- Provider responses may omit attributable cost. Remaining/used rate gauges are shared and cannot prove one request's point consumption.

layering: perfrec owns typed optional observations; existing CLI, launcher and provider adapters retain effects and authority. No new service or credential boundary.

single-point-of-failure: a producer can mislabel optional context or coverage; independent writer/profile validation and production-flow negative controls expose this without granting operational authority.

design-fit:
  owner: perfrec and existing producer adapters
  contract: docs/performance-attribution-spec.md
  retires: [unattributed new performance samples and implicit unknown-as-zero accounting]
  weight: "verbs +0; profiles +2; shared flags +8 across the track, owned by /64; no new sender, service or credential holder"
  why-add: "Reuse the local recorder and effect seams; three bounded M briefs separate schema, propagation and provider observation."

## Read first

Read docs/performance-attribution-spec.md, the Context sources and the implemented /61.
Refresh current source and active work before pickup; all new paths and named tests below are
deliverables, not existing implementation or acceptance evidence.

## Ground rules

Feature branch and draft PR only; no merge, release, workflow dispatch or live infrastructure.
Synthetic providers, isolated homes and disposable local processes only. Preserve custody,
claims, budgets, audit, help/version and operation results. Optional collection cannot become
required accounting. Stop at implemented; independent verification owns Evidence.

## Interface contract

Implement docs/performance-attribution-spec.md's own-process request contract. Both forges may share totals with bounded protocol/resource/unit classes. Count retries/304s/errors only at the actual transport leaf; measure dispatch through body EOF/close or terminal error with explicit incomplete coverage. Provider adapters supply bounded actual cost when already observed; estimates require method/version, and unknown/unsupported/failed is never zero. No separate monitoring request, expanded business selection or new credentials. Qualified bounded cost metadata may join the same existing operation; disclose its overhead. Parent processes do not scrape or charge their children's calls.

## Task

1. Reuse the transport leaf and request-meter owner to observe count/outcome and monotonic lifetime through response-body EOF/close/error. Preserve bytes, read/close/cancellation/error behavior without eager read/buffer. Wrapper composition and retry paths count once per actual attempt; cached/local refusals cost zero. Bound samples or use documented mergeable histograms with explicit coverage.
2. At existing GitHub/GitLab response decoding seams, extract attributable GraphQL cost from that operation's existing response path; when supported, add bounded cost metadata to the same operation (GitHub rateLimit.cost, GitLab query complexity in its own unit), never an additional request. Record numeric zero only when measured, count missing observations, and separate protocol/resource/unit and estimated method/version. Shared quota gauges carry window/reset provenance and are never differenced or summed as actual cost. Do not add separate cost/dry-run requests, expand business selections, re-read bodies or acquire new credentials. Qualify supported backend metadata/error semantics without failing the operation for missing telemetry.
3. Feed typed observations to /63's recorder and existing deskperf reader. Keep auth/service HTTP distinct from forge-provider operations, preserve retry/304 distinctions and own-process accounting. Unsupported custom clients and killed/unfinished requests remain coverage gaps. Concurrent request time is work, not additive elapsed time.
4. Qualify through deterministic synthetic transports/provider fixtures, compiled command records and the real report reader. Compare instrumentation-on/off calls, response bytes, exits and cancellation; exercise body errors/early close, retries, zero/missing cost and concurrent unrelated quota usage. Publish actual coverage and on/off overhead; never claim complete fleet coverage from one transport.

## Verify

Named tests below are planned. Each command requires its exact PASS line; a missing test
cannot pass. Plant the stated defect through the production boundary, not a duplicate checker.

| # | Class | Command | Expect |
|---|---|---|---|
| 1 | check:ci +flow +mutation | `(cd tools/desk && out=$(mktemp) && trap 'rm -f "$out"' 0 && GOWORK=off go test -count=1 -v -run '^TestPerfV3ForgeAttemptTimingFlow$' ./internal/deskkit/ > "$out" && grep -F -- '--- PASS: TestPerfV3ForgeAttemptTimingFlow ' "$out")` | exit 0; named PASS. Production backend/transport/record/report counts exact retry,304,error and canceled attempts with body-lifetime timing; local cache/refusal and auth HTTP are distinct; planted double-count/timer-at-headers fails. |
| 2 | check:ci +flow +mutation | `(cd tools/desk && out=$(mktemp) && trap 'rm -f "$out"' 0 && GOWORK=off go test -count=1 -v -run '^TestPerfV3GraphQLCostCoverage$' ./internal/deskkit/ > "$out" && grep -F -- '--- PASS: TestPerfV3GraphQLCostCoverage ' "$out")` | exit 0; named PASS. Existing response cost yields measured zero/nonzero while missing/unsupported/failed/estimated remain distinct; unrelated concurrent quota changes cannot become actual cost; raw queries/URLs/bodies never enter records. |
| 3 | check:ci +flow +mutation | `(cd tools/desk && out=$(mktemp) && trap 'rm -f "$out"' 0 && GOWORK=off go test -count=1 -v -run '^TestPerfV3ForgeObservationParity$' ./internal/deskkit/ > "$out" && grep -F -- '--- PASS: TestPerfV3ForgeObservationParity ' "$out")` | exit 0; named PASS. On/off has identical provider calls and business selections/results, preserves provider body bytes through observation and the original errors/cancellation/result, with bounded metadata differences documented; parent/child fixture counts each request only in its emitting process and reports uninstrumented child coverage. |
| 4 | check +dereference | `statusgen --root . --consumers --brief desktools-v2/65` | exit 0; changed consumer routing corroborates the actual diff; unavailable is not pass. |

## Pre-mortem and detection

Rows 1–3 plant the counterexamples named in Expect: identity/consent or request-count bugs,
false-zero/compatibility bugs, and changed operational behavior. Row 4 checks shared consumer
routing. Sufficiency of real-world coverage and local disclosure remains reviewer judgment;
fixture success never asserts live adoption or permission.

## Evidence

<!-- Independent verifier records subject revision, environment, command, exit and named result. -->

## Review

Gate: model. Reject changed legacy meanings, guessed identity/cost, lost coverage or optional measurement that changes an operation.
