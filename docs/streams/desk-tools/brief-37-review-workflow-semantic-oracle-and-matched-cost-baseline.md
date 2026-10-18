---
brief: assay:assay:desk-tools:37
title: Review workflow semantic oracle and matched cost baseline
why: "Uncontrolled historical samples and incomplete process counters cannot prove a workflow improvement. Freeze intended behavior and reproducible complete-cycle workloads before changing collection."
wave: 1
depends: []
unblocks: ["desk-tools/38", "desk-tools/39"]
effort: M
gate: model
risk: {regulatory: no, customer: no, irreversible: no, sensitive-data: no}
issues: []
schema: brief-v2
version: 3
authored: 2026-10-17 by statusgen newbrief
sources:
  - docs/desk-workflow-efficiency-spec.md
  - docs/desk-workflow-efficiency-change-map.md
  - freshness-checked 2026-10-17 @ fb85924c039c27e7d2aa80e055bfce118e46125d (App-authenticated main fetch)
  - 'desk-tools: Review workflow semantic oracle and matched cost baseline'
id: 202e9121-a4f6-49e5-90cb-dd5c8acaed1a
exec-tier: strong
exec-tier-why: Semantic extraction, cost coverage and cross-caller failure behavior require independent contract judgment.
domain: complicated
consumers:
  - "tools/desk/cmd/deskboard/: fixed-here (test-only production sweep adapter)"
  - "tools/desk/cmd/reviewloop/: fixed-here (test-only production plan adapter)"
  - "tools/desk/cmd/deskdispatch/: fixed-here (test-only pre-claim round and hold adapter)"
  - "tools/desk/cmd/issueboard/: fixed-here (test-only command and HTTP enumeration adapter)"
---

# Brief 37 — Review workflow semantic oracle and matched cost baseline

## Context

files: `tools/desk/internal/reviewtest/`, `tools/desk/internal/reviewtest/testdata/`, `tools/desk/cmd/{deskboard,reviewloop,deskdispatch,issueboard}/qualification_test.go`, `tools/desk/internal/deskkit/citrigger_test.go`, `docs/review-efficiency-qualification.md`, `changelog/desk-tools-37.md`

Implementation scope clarification (v2): the commands are separate `package main`
packages, so the fixture reaches their production handlers through test-binary adapters.
The four `_test.go` bridges add no shipped flags or alternative policy implementation.
The v3 observation boundary includes captured planner stdout and dispatch round text;
literal output assertions and production-rendering overlays supplement internal values.
The CI reader classification records that compiling those adapters stays within the desk
module. Cost coverage is explicit: local issue enumeration reaches the production HTTP
adapter; other typed fakes and the fixed Git reviewer do not establish full provider,
model, launcher or descendant CPU cost. Filling those counters remains desk-tools/38.

facts: 17 October 2026 at the source pin above. Existing implementation is described in
approved spec §2. All named tests and paths marked planned below are deliverables, not
passing claims. desk-tools/23 remains implemented, /24–25 done, /27 retains its own
verification debt; desktools-v2/09 is verified and /45 implemented. Refresh actual source,
claims and open PRs at pickup. #3248 and #3302 are active CLI ports; #3325 is proposed
supervision guidance and #3305 proposed generic improvement planning. Neither draft is
assumed shipped. The initial blocker is the missing complete observation and attribution,
not the full future runtime.

layering: existing desk read/planning semantics and performance-collection meanings. Contracts and pure computation remain separated from
credentialed collection/effects; adapters consume values through legal module APIs.
No new shipped service, runtime store or credential owner is authorized. The flow and boundary
mutation cases below enforce this placement.

design-fit:
  owner: existing desk read/planning semantics and performance-collection meanings
  contract: docs/desk-workflow-efficiency-spec.md §§2, 5–8; existing action table is the oracle
  retires: duplicate work only after the applicable caller and equivalence acceptance
  weight: verbs +0, flags +0, refusals +0, rule-text +0; local fixture only
  why-add: reuse existing contracts/runner/typed reads; a complete observation is required before removing the second read, and a new cache would preserve that duplication

## Read first

- [Approved source](../../desk-workflow-efficiency-spec.md) and [consolidation](../../desk-workflow-efficiency-change-map.md).
- [Delivery waves](README.md#review-efficiency-follow-ons) and existing source files declared above.
- `docs/library-first.md`, `docs/brief-rules.md` and `docs/telemetry.md` for the applicable boundary/consent rules.

## Ground rules

- Preserve exact-head independent verdicts, claims, trust, budgets, audit and consequential rechecks.
- Use fake providers and disposable local Git for acceptance; live activation and collection are separate gates.
- Stop at implemented; independent verification supplies Evidence. Do not rewrite completed brief Evidence.

## Task

1. Capture independent expected observations/actions from existing deskboard classification and
   reviewloop planning, including body-only edits, check/review applicability, overflow, access
   failure, superseded heads, duplicate lanes, WAIT/SURFACE/NO-OP, freshness and final idle gates.
   Describe which historical defects are corrected intent, rather than freezing them as truth.
   Exercise production adapters; a self-comparison of two copies is insufficient.
2. Build a local fake-provider/disposable-Git workload runner over existing public command
   entrypoints. Record revision/tool/plugin versions, fixture hash, host/method, seeds, cache state,
   scope/size, concurrency and result. Keep current token memoization, audit tail, pools and
   conditional reads in both arms. No real credentials or live provider data are required.
3. Declare the cost vector and permitted tradeoffs before comparisons: own vs descendant CPU,
   inclusive coverage, overlapping wall intervals, actual endpoint attempts/retries/304s, and actual
   vs estimated/unknown provider points. Record tool invocations and complete dispatcher-plus-agent
   work, never sum parent summaries and child events twice. No money conversion or invented
   numerical SLO.
4. Freeze a tool-only arm with supervision held constant. Add separate caller-only and combined arms
   when #3325 is qualified; plugin/launcher adoption and dispatch-context policy are explicit
   factors. Do not claim reduced context tokens equal CPU/provider savings. Keep this a bounded
   technical fixture, not the reusable scorecard/experiment system proposed in #3305.
5. Update affected command/operator contracts and the declared changelog; regenerate applicable
   documentation with the existing statusgen README generator. No generated global board is branch-
   committed. All planned tests must be wired to actual production paths, not a parallel test
   implementation.

## Verify

Run from the public repository root; each named test is a planned deliverable and its PASS
line is mandatory (zero selected tests is a failure). No authoring-time execution Evidence
is asserted. Each test covers production flow plus the stated wrong-but-plausible mutation.

| # | Class | Command | Expect |
|---|-------|---------|--------|
| 1 | check:ci +flow +mutation | `(cd tools/desk && o=$(mktemp) && trap 'rm -f "$o"' 0 && GOWORK=off go test -count=1 -v -run '^TestWorkflowOracle$' ./internal/reviewtest/... > "$o" && grep -F -- '--- PASS: TestWorkflowOracle ' "$o")` | exit 0; named PASS. Literal actions, exact heads, lanes and holds agree with production paths; captured command output preserves visible row order, counts, proposals, idle and round text. Wrong-head/WAIT/order/depth controls and production planner-row/dispatch-scope overlays each fail the oracle. |
| 2 | check:ci +flow +mutation | `(cd tools/desk && o=$(mktemp) && trap 'rm -f "$o"' 0 && GOWORK=off go test -count=1 -v -run '^TestCostCohort$' ./internal/reviewtest/... > "$o" && grep -F -- '--- PASS: TestCostCohort ' "$o")` | exit 0; named PASS. Fixture manifest reproduces cold/steady, skew/growth, concurrency and failed/partial runs; complete parent/child CPU, wall and request coverage is explicit. Unknown costs never become zero. |
| 3 | check:ci +dereference | `statusgen --root . --consumers --brief desk-tools/37` | exit 0; implementation consumer routing is corroborated against the actual diff. Out-of-scope judgments remain review obligations. |
| 4 | check:ci | `(cd tools/desk && GOWORK=off go test -count=1 -v -run '^TestGitPackagesInstallQuiet$' ./internal/gitquiet)` | exit 0; named PASS; the new Git-fixture package uses the shared quiet runner without weakening fixture controls or cleanup. |

## Pre-mortem and detection

| Failure mode | Caught by |
|---|---|
| Adapter produces plausible output while duplicating or omitting work | row 1, independent oracle and actual entrypoint/count instrumentation |
| A partial/unsupported observation or hidden consumer creates a false success | row 2, planted invalid input and explicit refusal/coverage |
| Operational activation or control deletion inferred from source qualification | review-only: review scope and separate existing adoption gates; no live effects in the test harness |

## Evidence

<!-- Independent verifier appends one row per Verify command after implementation. -->
Independent verifier: assay-verifier-app[bot], 2026-10-18, gpt-6.1-sol / strong; this agent did not implement the delivery. Source: 8f1206fe5eb4b4ab19f5f2f7266c515aa71f8eae. Pre-work and final deskdispatch admission both passed with the same run binding; only brief Evidence changed.

**VERIFY: FAIL** — blocked/could-not-check under the known host hold #1800. Mandatory statusgen verifyrun exited 2: every check:ci row was explicitly unrun because this macOS host lacks Linux unshare network-off execution. No implementation failure was observed and no verified/done transition is justified. Source remains at its admitted detached commit.

The wrote-to-the-test delivery is #3331, merge e28201023c74b5f407ae98f9833ee4f79999f1db, head 2f4aa073176da03d448accedd7cc40479f39d7ef. Forge/source history establish the genuine pre-delivery merge-base 352d5f26cf7e8d6b43fdab75934d17cacf5ba57d. deskpathguard rederive with that explicit --main exited 0: original rows 1–3 govern; row 4 is author-added and receives no acceptance pass. The delivering first parent f92d2e9f24251bb3ac57c6447f229916d7748dfc contains the same brief; default merged HEAD/origin-main is not a pre-delivery comparison.

<!-- Independent verifier appends one row per Verify command after implementation. -->

| # | Command | Result | Output | Date | Runner |
|---|---------|--------|--------|------|--------|
| 1 | `(cd tools/desk && o=$(mktemp) && trap 'rm -f "$o"' 0 && GOWORK=off go test -count=1 -v -run '^TestWorkflowOracle$' ./internal/reviewtest/... > "$o" && grep -F -- '--- PASS: TestWorkflowOracle ' "$o")` | could-not-run exit=- sandbox=unshare — check:ci hermetic execution requires a network-off sandbox, unavailable on this host: the network sandbox uses `unshare --net`, a Linux facility, and this host is darwin. check:ci rows are re-executed network-off by design (verdict-lane/02, R-6 c.6) — run on a Linux runner that provides `unshare --net` | sha256:e3b0c44298fc | 2026-10-18 | assay-verifier-app[bot] @ 8f1206fe5eb4 (on-behalf-of human:ian) (forge-identity) |
| 2 | `(cd tools/desk && o=$(mktemp) && trap 'rm -f "$o"' 0 && GOWORK=off go test -count=1 -v -run '^TestCostCohort$' ./internal/reviewtest/... > "$o" && grep -F -- '--- PASS: TestCostCohort ' "$o")` | could-not-run exit=- sandbox=unshare — check:ci hermetic execution requires a network-off sandbox, unavailable on this host: the network sandbox uses `unshare --net`, a Linux facility, and this host is darwin. check:ci rows are re-executed network-off by design (verdict-lane/02, R-6 c.6) — run on a Linux runner that provides `unshare --net` | sha256:e3b0c44298fc | 2026-10-18 | assay-verifier-app[bot] @ 8f1206fe5eb4 (on-behalf-of human:ian) (forge-identity) |
| 3 | `statusgen --root . --consumers --brief desk-tools/37` | could-not-run exit=- sandbox=unshare — check:ci hermetic execution requires a network-off sandbox, unavailable on this host: the network sandbox uses `unshare --net`, a Linux facility, and this host is darwin. check:ci rows are re-executed network-off by design (verdict-lane/02, R-6 c.6) — run on a Linux runner that provides `unshare --net` | sha256:e3b0c44298fc | 2026-10-18 | assay-verifier-app[bot] @ 8f1206fe5eb4 (on-behalf-of human:ian) (forge-identity) |
| 4 | `(cd tools/desk && GOWORK=off go test -count=1 -v -run '^TestGitPackagesInstallQuiet$' ./internal/gitquiet)` | could-not-run exit=- sandbox=unshare — check:ci hermetic execution requires a network-off sandbox, unavailable on this host: the network sandbox uses `unshare --net`, a Linux facility, and this host is darwin. check:ci rows are re-executed network-off by design (verdict-lane/02, R-6 c.6) — run on a Linux runner that provides `unshare --net` | sha256:e3b0c44298fc | 2026-10-18 | assay-verifier-app[bot] @ 8f1206fe5eb4 (on-behalf-of human:ian) (forge-identity) |
Verification-Attestation: medici-finance/assay#3335 run=79d4705247e9d206708c0ddcc28a00e1cfaf0e27f1caa4521ce582963e9f47fd source=8f1206fe5eb4b4ab19f5f2f7266c515aa71f8eae brief=docs/streams/desk-tools/brief-37-review-workflow-semantic-oracle-and-matched-cost-baseline.md model=gpt-6.1-sol tier=strong

Witness output hash: sha256:2c2116c68c10985ce3bed85447b28b419d1071a580f346c5059915668d4a206e. Every mandatory row has empty captured output, full sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855. Exact refusal: “check:ci hermetic execution requires a network-off sandbox, unavailable on this host: the network sandbox uses `unshare --net`, a Linux facility, and this host is darwin. check:ci rows are re-executed network-off by design (verdict-lane/02, R-6 c.6) — run on a Linux runner that provides `unshare --net`”.

Supporting diagnostics ran the original commands against the admitted source after offline-safety inspection, with permission to use local compiler caches. They do not replace mandatory hermetic witnesses:

| Verify row | Supplemental observation | Output SHA-256 | Date and runner |
|---|---|---|---|
| 1 | exit 0; --- PASS: TestWorkflowOracle (7.93s). Mandatory witness remains unrun. | d3cc4e1fbc2ab8443fb954d4ab31c13a4a1df6ea2150ca12dbf52175342de942 | 2026-10-18 assay-verifier-app[bot] @ 8f1206fe5eb4 |
| 2 | exit 0; --- PASS: TestCostCohort (5.10s). Mandatory witness remains unrun. | 1d05b6784edbffe13cab6aeeb9ce67667da1323a41933cca39cb998e2b36fabc | 2026-10-18 assay-verifier-app[bot] @ 8f1206fe5eb4 |
| 3 | exit 0; 0 corroborated, 0 disproved, 4 unchecked at default merged base; not a pass. Mandatory witness remains unrun. | 2b01a4a505285fb7ea9972d4bebbc6092553aaa56bebf2980b99fd5d53d388f2 | 2026-10-18 assay-verifier-app[bot] @ 8f1206fe5eb4 |
| 4 | author-added; supporting execution exit 0 with named TestGitPackagesInstallQuiet line (0.41s); no acceptance credit. Mandatory witness remains unrun. | 052142cf9b2bf8e8fcf6c4e5124c109936ed8197b34c9466c60c629c47b6deef | 2026-10-18 assay-verifier-app[bot] @ 8f1206fe5eb4 |

A separate supplemental statusgen --root . --consumers --brief desk-tools/37 --base 352d5f26cf7e8d6b43fdab75934d17cacf5ba57d exited 0 with 4 corroborated, 0 disproved, 0 unchecked (sha256:a69f2dcd575ce456328ed6d5d94dd014ca19c4566c5d40f604c5f336829b8a9e). This is broad path-touch corroboration including subsequent work. Independent inspection of #3331's exact delivering diff confirms each claimed command path gained its test-only production bridge; no absent/uninspected consumer is declared clear.

Independent Task/DoD judgment: source traces the oracle to production board classification/head reads, planner/output/idle, round preparation/rendering/holds, and issue command/HTTP enumeration. Literal heads/lanes/dispositions, output order/counts, negative controls and named applicability regressions are present. The six cohorts preserve failed/partial outcomes and explicitly unknown CPU/provider components. The dispatcher bridge stops before claims/launcher and the agent is local Git diff, so full dispatcher-plus-model-agent cost and real cache/token/audit effects remain unqualified; the original Task 2–3 complete-workflow obligation is an acceptance limitation, not an efficiency result. No production activation or policy deletion was observed. The stream README still says todo despite merged #3331; recorded only, never edited.

Literal enumeration covered all 13 delivery-diff files: 1,640 conservative Go/JSON/documentation literal entries, each with source line, retained in the independent report (inventory sha256:2072b2986172552dbcf6b791c6b62dfbdec95196413a9e4fac114ba75217f848). No new irreversible production value or authority was found. Fixture isolation/consent and exact-head values rank first; completeness literals follow; reversible test limits, seeds, concurrency, timeouts and formatting rank last.

RISK-VALUE: DERIVED — fixtureTelemetry = "ASSAY_TELEMETRY=0" @ tools/desk/internal/reviewtest/qualification_test.go:89 — the telemetry contract activates only for exactly 1; the explicit 0 and constructed environment implement the brief's consent-off fake/local acceptance boundary.

RISK-VALUE: DERIVED — fixtureToken = "local-fixture" @ tools/desk/cmd/issueboard/qualification_test.go:54 — the credential-free fixture uses a synthetic token and in-memory transport; no provider authority is granted and inherited credentials are omitted by the constructed environment.

RISK-VALUE: DERIVED — endpointSHACharacters = 40 @ tools/desk/cmd/deskdispatch/qualification_test.go:17 — this repository reports sha1 object format; 160-bit SHA-1 rendered as hexadecimal is 40 characters. Both exact endpoints come from the disposable Git fixture and preserve the brief's exact-head contract.

RISK-VALUE: DERIVED — fixturePRPopulationCap = 100 @ tools/desk/cmd/deskboard/qualification_test.go:75 — the fixture carries the existing production prListLimit of 100 (board.go:80) and rejects idle on a cap-sized possibly truncated population; it selects no new production cap.

The container launcher currently mounts only the home at /work and does not establish inner desk admission/common-Git metadata closure for this linked worktree. No unadmitted container/copy was used. Recovery requires an admission-preserving Linux witness runner; #1800 remains the host hold. Parent desk owns landing, lifecycle and issue routing.

Tool provenance clarification: mandatory verifyrun and both supplemental consumers invocations used the installed statusgen resolved from the explicitly prefixed tool PATH; it reports v1.0.48, SHA-256 17a018969f41ad77da901a9b602a8370c8b3297255c93334ca72dbd7711f8922. Embedded build: go1.27.2, darwin/arm64, vcs.revision ec8d77e56627a7c2269f3d143ef6a31f3e41bd02, vcs.modified=true, version ldflag v1.0.48. Hash/version were measured at handoff, not before each earlier invocation. The installed tool's build source is distinct from the admitted deliverable source. Row3 ran after verifyrun's nine-line Evidence append: the brief therefore appears in the working-tree diff, avoiding the clean-tree exit2 absent-brief refusal, while its inherited unchanged consumers remain four UNCHECKED entries. This explains the observed exit0 without treating it as corroboration. No test rerun was performed for this clarification.

## Review

Retain the derived frontmatter gate. Approval of the source spec is recorded; this brief
adds implementation acceptance rather than reopening that decision. Existing adopter
writer/custody/equivalence and collection consent gates remain separate.
