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

Historical record — retained as quotation, without current certification.

> > Historical record (superseded for closure while provenance is reviewed):
> > 
> > Historical host run — superseded by the admitted Linux run below; retained as quotation.
> > 
> > > <!-- Independent verifier appends one row per Verify command after implementation. -->
> > > Independent verifier: assay-verifier-app[bot], 2026-10-18, gpt-6.1-sol / strong; this agent did not implement the delivery. Source: 8f1206fe5eb4b4ab19f5f2f7266c515aa71f8eae. Pre-work and final deskdispatch admission both passed with the same run binding; only brief Evidence changed.
> > >
> > > **VERIFY: FAIL** — blocked/could-not-check under the known host hold #1800. Mandatory statusgen verifyrun exited 2: every check:ci row was explicitly unrun because this macOS host lacks Linux unshare network-off execution. No implementation failure was observed and no verified/done transition is justified. Source remains at its admitted detached commit.
> > >
> > > The wrote-to-the-test delivery is #3331, merge e28201023c74b5f407ae98f9833ee4f79999f1db, head 2f4aa073176da03d448accedd7cc40479f39d7ef. Forge/source history establish the genuine pre-delivery merge-base 352d5f26cf7e8d6b43fdab75934d17cacf5ba57d. deskpathguard rederive with that explicit --main exited 0: original rows 1–3 govern; row 4 is author-added and receives no acceptance pass. The delivering first parent f92d2e9f24251bb3ac57c6447f229916d7748dfc contains the same brief; default merged HEAD/origin-main is not a pre-delivery comparison.
> > >
> > > <!-- Independent verifier appends one row per Verify command after implementation. -->
> > >
> > > | # | Command | Result | Output | Date | Runner |
> > > |---|---------|--------|--------|------|--------|
> > > | 1 | `(cd tools/desk && o=$(mktemp) && trap 'rm -f "$o"' 0 && GOWORK=off go test -count=1 -v -run '^TestWorkflowOracle$' ./internal/reviewtest/... > "$o" && grep -F -- '--- PASS: TestWorkflowOracle ' "$o")` | could-not-run exit=- sandbox=unshare — check:ci hermetic execution requires a network-off sandbox, unavailable on this host: the network sandbox uses `unshare --net`, a Linux facility, and this host is darwin. check:ci rows are re-executed network-off by design (verdict-lane/02, R-6 c.6) — run on a Linux runner that provides `unshare --net` | sha256:e3b0c44298fc | 2026-10-18 | assay-verifier-app[bot] @ 8f1206fe5eb4 (on-behalf-of human:ian) (forge-identity) |
> > > | 2 | `(cd tools/desk && o=$(mktemp) && trap 'rm -f "$o"' 0 && GOWORK=off go test -count=1 -v -run '^TestCostCohort$' ./internal/reviewtest/... > "$o" && grep -F -- '--- PASS: TestCostCohort ' "$o")` | could-not-run exit=- sandbox=unshare — check:ci hermetic execution requires a network-off sandbox, unavailable on this host: the network sandbox uses `unshare --net`, a Linux facility, and this host is darwin. check:ci rows are re-executed network-off by design (verdict-lane/02, R-6 c.6) — run on a Linux runner that provides `unshare --net` | sha256:e3b0c44298fc | 2026-10-18 | assay-verifier-app[bot] @ 8f1206fe5eb4 (on-behalf-of human:ian) (forge-identity) |
> > > | 3 | `statusgen --root . --consumers --brief desk-tools/37` | could-not-run exit=- sandbox=unshare — check:ci hermetic execution requires a network-off sandbox, unavailable on this host: the network sandbox uses `unshare --net`, a Linux facility, and this host is darwin. check:ci rows are re-executed network-off by design (verdict-lane/02, R-6 c.6) — run on a Linux runner that provides `unshare --net` | sha256:e3b0c44298fc | 2026-10-18 | assay-verifier-app[bot] @ 8f1206fe5eb4 (on-behalf-of human:ian) (forge-identity) |
> > > | 4 | `(cd tools/desk && GOWORK=off go test -count=1 -v -run '^TestGitPackagesInstallQuiet$' ./internal/gitquiet)` | could-not-run exit=- sandbox=unshare — check:ci hermetic execution requires a network-off sandbox, unavailable on this host: the network sandbox uses `unshare --net`, a Linux facility, and this host is darwin. check:ci rows are re-executed network-off by design (verdict-lane/02, R-6 c.6) — run on a Linux runner that provides `unshare --net` | sha256:e3b0c44298fc | 2026-10-18 | assay-verifier-app[bot] @ 8f1206fe5eb4 (on-behalf-of human:ian) (forge-identity) |
> > > Verification-Attestation: medici-finance/assay#3335 run=79d4705247e9d206708c0ddcc28a00e1cfaf0e27f1caa4521ce582963e9f47fd source=8f1206fe5eb4b4ab19f5f2f7266c515aa71f8eae brief=docs/streams/desk-tools/brief-37-review-workflow-semantic-oracle-and-matched-cost-baseline.md model=gpt-6.1-sol tier=strong
> > >
> > > Witness output hash: sha256:2c2116c68c10985ce3bed85447b28b419d1071a580f346c5059915668d4a206e. Every mandatory row has empty captured output, full sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855. Exact refusal: “check:ci hermetic execution requires a network-off sandbox, unavailable on this host: the network sandbox uses `unshare --net`, a Linux facility, and this host is darwin. check:ci rows are re-executed network-off by design (verdict-lane/02, R-6 c.6) — run on a Linux runner that provides `unshare --net`”.
> > >
> > > Supporting diagnostics ran the original commands against the admitted source after offline-safety inspection, with permission to use local compiler caches. They do not replace mandatory hermetic witnesses:
> > >
> > > | Verify row | Supplemental observation | Output SHA-256 | Date and runner |
> > > |---|---|---|---|
> > > | 1 | exit 0; --- PASS: TestWorkflowOracle (7.93s). Mandatory witness remains unrun. | d3cc4e1fbc2ab8443fb954d4ab31c13a4a1df6ea2150ca12dbf52175342de942 | 2026-10-18 assay-verifier-app[bot] @ 8f1206fe5eb4 |
> > > | 2 | exit 0; --- PASS: TestCostCohort (5.10s). Mandatory witness remains unrun. | 1d05b6784edbffe13cab6aeeb9ce67667da1323a41933cca39cb998e2b36fabc | 2026-10-18 assay-verifier-app[bot] @ 8f1206fe5eb4 |
> > > | 3 | exit 0; 0 corroborated, 0 disproved, 4 unchecked at default merged base; not a pass. Mandatory witness remains unrun. | 2b01a4a505285fb7ea9972d4bebbc6092553aaa56bebf2980b99fd5d53d388f2 | 2026-10-18 assay-verifier-app[bot] @ 8f1206fe5eb4 |
> > > | 4 | author-added; supporting execution exit 0 with named TestGitPackagesInstallQuiet line (0.41s); no acceptance credit. Mandatory witness remains unrun. | 052142cf9b2bf8e8fcf6c4e5124c109936ed8197b34c9466c60c629c47b6deef | 2026-10-18 assay-verifier-app[bot] @ 8f1206fe5eb4 |
> > >
> > > A separate supplemental statusgen --root . --consumers --brief desk-tools/37 --base 352d5f26cf7e8d6b43fdab75934d17cacf5ba57d exited 0 with 4 corroborated, 0 disproved, 0 unchecked (sha256:a69f2dcd575ce456328ed6d5d94dd014ca19c4566c5d40f604c5f336829b8a9e). This is broad path-touch corroboration including subsequent work. Independent inspection of #3331's exact delivering diff confirms each claimed command path gained its test-only production bridge; no absent/uninspected consumer is declared clear.
> > >
> > > Independent Task/DoD judgment: source traces the oracle to production board classification/head reads, planner/output/idle, round preparation/rendering/holds, and issue command/HTTP enumeration. Literal heads/lanes/dispositions, output order/counts, negative controls and named applicability regressions are present. The six cohorts preserve failed/partial outcomes and explicitly unknown CPU/provider components. The dispatcher bridge stops before claims/launcher and the agent is local Git diff, so full dispatcher-plus-model-agent cost and real cache/token/audit effects remain unqualified; the original Task 2–3 complete-workflow obligation is an acceptance limitation, not an efficiency result. No production activation or policy deletion was observed. The stream README still says todo despite merged #3331; recorded only, never edited.
> > >
> > > Literal enumeration covered all 13 delivery-diff files: 1,640 conservative Go/JSON/documentation literal entries, each with source line, retained in the independent report (inventory sha256:2072b2986172552dbcf6b791c6b62dfbdec95196413a9e4fac114ba75217f848). No new irreversible production value or authority was found. Fixture isolation/consent and exact-head values rank first; completeness literals follow; reversible test limits, seeds, concurrency, timeouts and formatting rank last.
> > >
> > > RISK-VALUE: DERIVED — fixtureTelemetry = "ASSAY_TELEMETRY=0" @ tools/desk/internal/reviewtest/qualification_test.go:89 — the telemetry contract activates only for exactly 1; the explicit 0 and constructed environment implement the brief's consent-off fake/local acceptance boundary.
> > >
> > > RISK-VALUE: DERIVED — fixtureToken = "local-fixture" @ tools/desk/cmd/issueboard/qualification_test.go:54 — the credential-free fixture uses a synthetic token and in-memory transport; no provider authority is granted and inherited credentials are omitted by the constructed environment.
> > >
> > > RISK-VALUE: DERIVED — endpointSHACharacters = 40 @ tools/desk/cmd/deskdispatch/qualification_test.go:17 — this repository reports sha1 object format; 160-bit SHA-1 rendered as hexadecimal is 40 characters. Both exact endpoints come from the disposable Git fixture and preserve the brief's exact-head contract.
> > >
> > > RISK-VALUE: DERIVED — fixturePRPopulationCap = 100 @ tools/desk/cmd/deskboard/qualification_test.go:75 — the fixture carries the existing production prListLimit of 100 (board.go:80) and rejects idle on a cap-sized possibly truncated population; it selects no new production cap.
> > >
> > > The container launcher currently mounts only the home at /work and does not establish inner desk admission/common-Git metadata closure for this linked worktree. No unadmitted container/copy was used. Recovery requires an admission-preserving Linux witness runner; #1800 remains the host hold. Parent desk owns landing, lifecycle and issue routing.
> > >
> > > Tool provenance clarification: mandatory verifyrun and both supplemental consumers invocations used the installed statusgen resolved from the explicitly prefixed tool PATH; it reports v1.0.48, SHA-256 17a018969f41ad77da901a9b602a8370c8b3297255c93334ca72dbd7711f8922. Embedded build: go1.27.2, darwin/arm64, vcs.revision ec8d77e56627a7c2269f3d143ef6a31f3e41bd02, vcs.modified=true, version ldflag v1.0.48. Hash/version were measured at handoff, not before each earlier invocation. The installed tool's build source is distinct from the admitted deliverable source. Row3 ran after verifyrun's nine-line Evidence append: the brief therefore appears in the working-tree diff, avoiding the clean-tree exit2 absent-brief refusal, while its inherited unchanged consumers remain four UNCHECKED entries. This explains the observed exit0 without treating it as corroboration. No test rerun was performed for this clarification.
> > 
> > Independent Linux witness follow-up — 2026-10-18, assay-verifier-app[bot], gpt-6.1-sol / strong.
> > 
> > 
> > This follow-up supersedes the earlier macOS could-not-run execution results for the four rows actually executed here. It does not erase those historical observations or close the general host issue #1800. The approved temporary container seccomp exception enabled unshare; each row retained its required network-off namespace. The ephemeral container exited and the exception was not persisted. No additional capabilities, network weakening, image pull, copied credentials, alternative source home or self-stamp was used.
> > 
> > The mandatory statusgen verifyrun completed exit 0 in the immutable cached image sha256:91886f2fd3fbc5acc840d84d5b44a53107fefa6f12bb2a6c618e0eb1d76c23d2, against original admitted source 8f1206fe5eb4b4ab19f5f2f7266c515aa71f8eae. It performed admission before rows and returned the original #3335 run binding. A separate post-run admission check exited 0 with the same binding. The witness's +dirty attribution denotes admitted Evidence-only changes; only the brief is modified and no stream index/lifecycle cell changed.
> > 
> > | # | Command | Result | Output | Date | Runner |
> > |---|---------|--------|--------|------|--------|
> > | 1 | `(cd tools/desk && o=$(mktemp) && trap 'rm -f "$o"' 0 && GOWORK=off go test -count=1 -v -run '^TestWorkflowOracle$' ./internal/reviewtest/... > "$o" && grep -F -- '--- PASS: TestWorkflowOracle ' "$o")` | pass exit=0 sandbox=unshare | sha256:85a301582162 | 2026-10-18 | assay-verifier-app[bot] @ 8f1206fe5eb4+dirty (on-behalf-of human:ian) (forge-identity) |
> > | 2 | `(cd tools/desk && o=$(mktemp) && trap 'rm -f "$o"' 0 && GOWORK=off go test -count=1 -v -run '^TestCostCohort$' ./internal/reviewtest/... > "$o" && grep -F -- '--- PASS: TestCostCohort ' "$o")` | pass exit=0 sandbox=unshare | sha256:112d9c839a00 | 2026-10-18 | assay-verifier-app[bot] @ 8f1206fe5eb4+dirty (on-behalf-of human:ian) (forge-identity) |
> > | 3 | `statusgen --root . --consumers --brief desk-tools/37` | pass exit=0 sandbox=unshare | sha256:2b01a4a50528 | 2026-10-18 | assay-verifier-app[bot] @ 8f1206fe5eb4+dirty (on-behalf-of human:ian) (forge-identity) |
> > | 4 | `(cd tools/desk && GOWORK=off go test -count=1 -v -run '^TestGitPackagesInstallQuiet$' ./internal/gitquiet)` | author-added; execution exit=0 sandbox=unshare; no acceptance credit | sha256:b4e1d8c9725c | 2026-10-18 | assay-verifier-app[bot] @ 8f1206fe5eb4+dirty (on-behalf-of human:ian) (forge-identity) |
> > Verification-Attestation: medici-finance/assay#3335 run=79d4705247e9d206708c0ddcc28a00e1cfaf0e27f1caa4521ce582963e9f47fd source=8f1206fe5eb4b4ab19f5f2f7266c515aa71f8eae brief=docs/streams/desk-tools/brief-37-review-workflow-semantic-oracle-and-matched-cost-baseline.md model=gpt-6.1-sol tier=strong
> > 
> > The native witness publishes SHA-256 prefixes, as shown above. Full runner-output SHA-256: 0f9e436fa63b54734514e07bb216b67d3fd0dcfd5f17a55a01600e2d8a0ca04e. Rows 1 and 2 use the original pre-delivery commands and their successful exits include the required named-PASS grep. The original row 1 expectation, including intended actions, lane independence, holds, ordering and uncapped depths, remains the acceptance source; the delivery author's edited Expect text was not substituted.
> > 
> > Original row 3 is discharged by the genuine Linux witness plus independent inspection of the exact #3331 delivery diff. All four declared consumers gained their stated test-only bridge: deskboard calls production sweepActionsRepo and fetchOpenPRs; reviewloop calls cmdPlan and projects ReadBoard/Coalesce/Idle; deskdispatch calls the existing round preparation and rendering path before claims; issueboard calls run with the production GitHub issue-enumeration adapter. The tool's output remains four UNCHECKED entries because the claims are inherited at merged main; its witness output prefix 2b01a4a50528 matches the earlier recorded diagnostic. That is a limitation of its comparison, not a new numerical acceptance condition or a failure of the independently corroborated routing. The supplemental explicit pre-delivery-base comparison and exact delivering-diff inspection remain recorded in the earlier report. No absent or uninspected consumer is assumed clear.
> > 
> > Pre-delivery provenance is unchanged: #3331 merged as e28201023c74b5f407ae98f9833ee4f79999f1db, PR head 2f4aa073176da03d448accedd7cc40479f39d7ef; genuine merge-base 352d5f26cf7e8d6b43fdab75934d17cacf5ba57d. Explicit deskpathguard rederive recovered original rows 1–3 and marked row 4 author-added. Row 4's genuine Linux runtime exit 0 is recorded as execution only; it cannot authenticate its newly authored acceptance assertion. The exact protected-table rule (plugins/assay/skills/verify-desk/SKILL.md:430–433) says a row present only in the post-change table is `author-added`, reported that way in Evidence — **never** `pass`, whatever it asserts; every other row uses the merge-base command/expect. It does not impose a whole-brief block for the added row. Therefore row 4 is excluded from acceptance credit rather than treated as a completion blocker. The earlier follow-up interpretation that it automatically blocked completion was an overread and is superseded here.
> > 
> > Independent scope judgment: the oracle and six cohorts exercise the stated bounded fake-provider/disposable-Git foundation and preserve declared unknown/partial cost coverage. No real dispatcher-claim-launcher-model-agent cost or complete production cache/token/audit effect was established. This limits future performance claims; this report asserts neither an efficiency saving nor complete production CPU/provider coverage. It does not manufacture an additional numerical success threshold. The original-row checks and independent consumer judgments are checked-clean within that declared fixture scope. No substantive unmet Task/DoD obligation was identified for that bounded fake-provider/disposable-Git foundation: it records every invocation in its declared controlled cycle and explicitly labels unobserved production/descendant/provider work as incomplete or unknown. The unchanged-production-control and unknown-cost requirements are preserved; a later performance comparison must qualify the relevant actual components before claiming a saving.
> > 
> > Container tool provenance, directly read in a mount-free --network none invocation: /usr/local/bin/statusgen reports v1.0.46, SHA-256 8d5c77d5d34662c7de4160fdc98f848aca554f608134c0ce38291c922259b102. /usr/local/bin/deskdispatch reports sourceSHA=25bb50ed, builtAt=2026-10-16T18:19:48Z, releaseTag=v1.0.46, SHA-256 cc6e15cc353391466945ff46adaccc9fcc435689df8a2afb3bb4a7de95c7ea0a. Both embed vcs.revision 25bb50edf6bbde9a8771cc8235891d46061ebcaa with vcs.modified=true; these stamps are not a claim of pristine released-source builds. Runtime is Go 1.27.2 linux/amd64 and Git 2.56.0. Fixture binaries are compiled from the admitted source; the installed witness/consumer tools are distinguished from that source. Earlier host diagnostics used the separately recorded installed v1.0.48 binary.
> > 
> > Risk enumeration and derivations remain unchanged because the implementation source and pre-delivery diff are unchanged. The complete 1,640-entry literal inventory remains in the independent report. No new irreversible production value or authority was introduced by running the witness.
> > 
> > RISK-VALUE: DERIVED — fixtureTelemetry = "ASSAY_TELEMETRY=0" @ tools/desk/internal/reviewtest/qualification_test.go:89 — the telemetry contract activates only for exactly 1; the explicit 0 and constructed environment implement the brief's consent-off fake/local acceptance boundary.
> > 
> > RISK-VALUE: DERIVED — fixtureToken = "local-fixture" @ tools/desk/cmd/issueboard/qualification_test.go:54 — the credential-free fixture uses a synthetic token and in-memory transport; no provider authority is granted and inherited credentials are omitted by the constructed environment.
> > 
> > RISK-VALUE: DERIVED — endpointSHACharacters = 40 @ tools/desk/cmd/deskdispatch/qualification_test.go:17 — this repository reports sha1 object format; 160-bit SHA-1 rendered as hexadecimal is 40 characters. Both exact endpoints come from the disposable Git fixture and preserve the brief's exact-head contract.
> > 
> > RISK-VALUE: DERIVED — fixturePRPopulationCap = 100 @ tools/desk/cmd/deskboard/qualification_test.go:75 — the fixture carries the existing production prListLimit of 100 (board.go:80) and rejects idle on a cap-sized possibly truncated population; it selects no new production cap.
> > 
> > The README discrepancy remains recorded: row 37 still says todo despite the merged delivery. The verifier did not edit it, publish, push, file issues or acknowledge shared scratch.
> > 
> > **VERIFY: PASS** — original pre-delivery rows 1–3 and the bounded fixture Task/DoD are checked-clean. Current row 4 is author-added and receives no acceptance credit; its runtime execution is supplemental only. No production efficiency saving or complete real dispatcher/model-agent cost is certified. The parent desk owns landing and lifecycle.
> > 
> > 
> > Lifecycle closure remains implemented with no Verified stamp: the closure checker cannot represent the protected-table author-added supplemental row. Reproduction: statusgen brief --check-verified desk-tools/37 exits 1 naming row 4 as could-not-run despite its recorded execution and exclusion from acceptance. Tracked by [assay#3342](https://github.com/medici-finance/assay/issues/3342); no result was relabelled to bypass that check.
> > 
> > Review follow-up — 2026-10-18, assay-verifier-app[bot].
> > 
> > The earlier record is preserved above as historical quotation. Its tool-emitted fingerprints do not make the original underlying row stdout/stderr available: the first Docker attempt did not retain those bytes. A new execution cannot authenticate an earlier fingerprint. The retained original reports, literal inventory, rederivation, supplemental outputs and admission logs have been preserved separately for inspection.
> > 
> > An independent verifier has completed a separately identified new attempt with captured row output, exact fingerprint comparison and genuine admission before execution and after normal credential recovery. This observed runtime result is not yet an accepted verification: access to primary artifacts and the authorization record remains under review. The earlier final admission failure from an expired cached credential is retained beside the later successful recovery check; no source or sandbox rule was changed to pass it.
> > 
> > The brief remains implemented, without a Verified stamp, until the provenance findings are independently cleared. No prior or new execution summary substitutes for inspectable primary proof. Publication of the prepared restricted artifact index awaits the driver's authorization; no restricted payload was posted after the publication refusal.
> > 
> > The original three acceptance rows and bounded fixture scope are distinct from supplemental author-added row 4, which receives no acceptance credit. The separate closure-reader defect remains tracked in #3342.
> > 
> > Admission binding (pre-work dispatch provenance only; not an accepted verification result):
> > 
> > Verification-Attestation: medici-finance/assay#3335 run=79d4705247e9d206708c0ddcc28a00e1cfaf0e27f1caa4521ce582963e9f47fd source=8f1206fe5eb4b4ab19f5f2f7266c515aa71f8eae brief=docs/streams/desk-tools/brief-37-review-workflow-semantic-oracle-and-matched-cost-baseline.md model=gpt-6.1-sol tier=strong
> > 
> > **VERIFY: BLOCKED** — provenance review and artifact-access approval pending.
>
> Primary-artifact audit follow-up — 2026-10-18, independent non-implementer inspection.
>
> Independent audit of the newly captured 2026-10-18 attempt inspected the retained primary bytes and original temporary-container approval. All 68 fresh manifest entries, all 12 namespace capture pairs and the 22 original-index entries match full SHA-256/size; direct/recorded qualification confirms normal-completion output/status preservation and real namespace isolation. This audit is separate from execution and does not reconstruct the earlier discarded Linux output.
>
> The actual run observed source `8f1206fe5eb4b4ab19f5f2f7266c515aa71f8eae`. Pre-admission succeeded; the exact witness commands exited 0. The initial final admission exited 5 on credential expiry/masked mint refusal and remains recorded. A separate genuine recovery admission, following normal same-role refresh, succeeded with the identical original run/source binding. No row was repeated for that recovery.
>
> Original protected acceptance rows 1–3 pass their bounded expectations. Row 1 retains the named workflow-oracle PASS, row 2 the named cost-cohort PASS. Row 3's command emits four UNCHECKED entries; independent inspection of delivery #3331's exact diff confirms all four declared production-adapter routes. Exit 0 alone is not consumer corroboration. Supplemental row 4 remains **author-added**, with no acceptance credit.
>
> Source/tool identities and build-modified flags are retained explicitly. These observations qualify the local fixture and preserve unknown/partial process, provider and real-agent cost; they establish no production savings. A read-only comparison with main `06587d38fc8195bde31098616f7143f287a06baa` found the declared delivery inputs unchanged, but the recorded execution remains source 8f.
>
> The driver authorized the restricted evidence index and it was posted for the authorized reviewer audience. A separate non-implementer has now inspected the primary execution bytes, manifests, recorder qualification and original approval excerpt. Artifact access and provenance are established for this fresh attempt; the normal PR review remains independent. No additional Verify command was run for this audit.
>
> Captured native execution witnesses for the separate fresh attempt:
>
> | # | Command | Result | Output | Date | Runner |
> |---|---------|--------|--------|------|--------|
> | 1 | `(cd tools/desk && o=$(mktemp) && trap 'rm -f "$o"' 0 && GOWORK=off go test -count=1 -v -run '^TestWorkflowOracle$' ./internal/reviewtest/... > "$o" && grep -F -- '--- PASS: TestWorkflowOracle ' "$o")` | pass exit=0 sandbox=unshare | sha256:6466c2ce1409 | 2026-10-18 | assay-verifier-app[bot] @ 8f1206fe5eb4+dirty (on-behalf-of human:ian) (forge-identity) |
> | 2 | `(cd tools/desk && o=$(mktemp) && trap 'rm -f "$o"' 0 && GOWORK=off go test -count=1 -v -run '^TestCostCohort$' ./internal/reviewtest/... > "$o" && grep -F -- '--- PASS: TestCostCohort ' "$o")` | pass exit=0 sandbox=unshare | sha256:4256ba3472fe | 2026-10-18 | assay-verifier-app[bot] @ 8f1206fe5eb4+dirty (on-behalf-of human:ian) (forge-identity) |
> | 3 | `statusgen --root . --consumers --brief desk-tools/37` | pass exit=0 sandbox=unshare | sha256:2b01a4a50528 | 2026-10-18 | assay-verifier-app[bot] @ 8f1206fe5eb4+dirty (on-behalf-of human:ian) (forge-identity) |
> | 4 | `(cd tools/desk && GOWORK=off go test -count=1 -v -run '^TestGitPackagesInstallQuiet$' ./internal/gitquiet)` | author-added; execution exit=0 sandbox=unshare; no acceptance credit | sha256:afec4dc4a43c | 2026-10-18 | assay-verifier-app[bot] @ 8f1206fe5eb4+dirty (on-behalf-of human:ian) (forge-identity) |
> Verification-Attestation: medici-finance/assay#3335 run=79d4705247e9d206708c0ddcc28a00e1cfaf0e27f1caa4521ce582963e9f47fd source=8f1206fe5eb4b4ab19f5f2f7266c515aa71f8eae brief=docs/streams/desk-tools/brief-37-review-workflow-semantic-oracle-and-matched-cost-baseline.md model=gpt-6.1-sol tier=strong
>
> Exact command-output excerpts for the named-PASS assertions (complete row streams, not retained inner verbose test logs):
>
> Row 1: 37 bytes, SHA-256 `6466c2ce1409b07dcd7c98d5dc66e6e7d0e06cf7b510932c50424f27084c88c9`.
>
> ```text
> --- PASS: TestWorkflowOracle (9.07s)
> ```
>
> Row 2: 34 bytes, SHA-256 `4256ba3472fe926516fa5ae2e70f3c8cde28ef762b91137f5b50d3a51bc76ae6`.
>
> ```text
> --- PASS: TestCostCohort (11.11s)
> ```
>
> The consumer row contains a configuration-bearing prefix in its raw stream. Its complete raw bytes remain restricted and were independently hashed; public excerpts are separately redacted and do not claim the raw digest. The raw consumer output SHA-256 is `2b01a4a505285fb7ea9972d4bebbc6092553aaa56bebf2980b99fd5d53d388f2`; the actual summary is 0 corroborated, 0 disproved, 4 unchecked. The exact delivery inspection establishes the four declared routes.
>
> Protected-table provenance remains delivery #3331 at `2f4aa073176da03d448accedd7cc40479f39d7ef`, merge `e28201023c74b5f407ae98f9833ee4f79999f1db`, genuine pre-delivery merge-base `352d5f26cf7e8d6b43fdab75934d17cacf5ba57d`. Original rows 1–3 govern acceptance. Supplemental row 4 executed successfully but remains author-added.
>
> **VERIFY: BLOCKED** — the original three rows and bounded fixture expectations have observed passing results with inspected primary proof. Formal closure remains blocked by the author-added-row reader defect [#3342](https://github.com/medici-finance/assay/issues/3342), followed by the separate outcome-history writer defect [#3343](https://github.com/medici-finance/assay/issues/3343), and normal independent PR disposition. Lifecycle remains implemented with no Verified stamp. No acceptance result was relabelled and no guard was bypassed.

Fresh independent verification — 2026-10-18.

The earlier primary-artifact certification claims are withdrawn: the temporary raw execution artifacts referenced above are no longer available for review. Their excerpts and hashes do not establish inspectable execution provenance. This new run does not authenticate or reconstruct those discarded bytes. Historical outcome records remain unchanged.

### Independent re-verification — 2026-10-18

This fresh run observes merged source a944c61762c50835f4fdd40ccc7785a51317f354. Earlier execution claims are not authenticated by this run. Expected behavior was recorded from the brief and pre-delivery production semantics before reading the delivery diff/tests. The delivering change is merged PR #3331, e28201023c74b5f407ae98f9833ee4f79999f1db; its sole parent is f92d2e9f24251bb3ac57c6447f229916d7748dfc. Protected-table re-derivation uses original rows 1–3 and their original Command/Expect/class. The current commands for those rows equal the original commands. Row 4 is author-added, and receives no acceptance credit.

Verification-Table: delivery=e28201023c74b5f407ae98f9833ee4f79999f1db base=f92d2e9f24251bb3ac57c6447f229916d7748dfc original=f92d2e9f24251bb3ac57c6447f229916d7748dfc

Verification-Attestation: medici-finance/assay#3422 run=6cad522fc49066c6d6960dff4fe8f7e8fd4f7cc4e2341c03b6df7f9b6034d0c2 source=a944c61762c50835f4fdd40ccc7785a51317f354 brief=docs/streams/desk-tools/brief-37-review-workflow-semantic-oracle-and-matched-cost-baseline.md model=gpt-6.1-sol tier=strong

| # | Command | Result | Output | Date | Runner |
|---|---------|--------|--------|------|--------|
| 1 | `(cd tools/desk && o=$(mktemp) && trap 'rm -f "$o"' 0 && GOWORK=off go test -count=1 -v -run '^TestWorkflowOracle$' ./internal/reviewtest/... > "$o" && grep -F -- '--- PASS: TestWorkflowOracle ' "$o")` | pass exit=0 sandbox=unshare | sha256:51ebd0282cc9 | 2026-10-18 | assay-verifier-app[bot] @ a944c61762c5 (on-behalf-of human:ian) (git-config) |
| 2 | `(cd tools/desk && o=$(mktemp) && trap 'rm -f "$o"' 0 && GOWORK=off go test -count=1 -v -run '^TestCostCohort$' ./internal/reviewtest/... > "$o" && grep -F -- '--- PASS: TestCostCohort ' "$o")` | pass exit=0 sandbox=unshare | sha256:b692f69c9631 | 2026-10-18 | assay-verifier-app[bot] @ a944c61762c5 (on-behalf-of human:ian) (git-config) |
| 3 | `statusgen --root . --consumers --brief desk-tools/37` | fail exit=2 sandbox=unshare | sha256:600dc7e4d35e | 2026-10-18 | assay-verifier-app[bot] @ a944c61762c5 (on-behalf-of human:ian) (git-config) |
| 4 | `(cd tools/desk && GOWORK=off go test -count=1 -v -run '^TestGitPackagesInstallQuiet$' ./internal/gitquiet)` | author-added; execution exit=0 sandbox=unshare; no acceptance credit | sha256:e45dee7646d7 | 2026-10-18 | assay-verifier-app[bot] @ a944c61762c5 (on-behalf-of human:ian) (git-config) |

Original acceptance observations:

| # | Original Expect | Observed, 2026-10-18 assay-verifier-app[bot] |
|---|---|---|
| 1 | exit 0; named PASS. Literal intended actions, exact heads, lane independence, holds, order and uncapped depths agree with existing production entrypoints; a planted wrong head, missing WAIT row and changed ordering each fail. | exit 0; `--- PASS: TestWorkflowOracle (24.02s)`. Named controls include wrong-head, missing-WAIT, changed-order, wrong-depth, printed-output mutations, production-reviewloop and production-deskdispatch. Independent baseline derivation gives mixed dispositions 2 DISPATCH, 1 WAIT, 1 SURFACE, 1 NO-OP, descending score order and two exact-head proposals; unresolved-head proposals remain suppressed. |
| 2 | exit 0; named PASS. Fixture manifest reproduces cold/steady, skew/growth, concurrency and failed/partial runs; complete parent/child CPU, wall and request coverage is explicit. Unknown costs never become zero. | exit 0; `--- PASS: TestCostCohort (11.35s)`. Six manifests retain invocation counts 9/18/3/38/9/3 and actual fixture issue-HTTP attempts 14/28/15/28/14/1; cap-sized growth, partial and failed cohorts retain failed-or-partial. Parent/direct-process own CPU and overlapping wall intervals are represented; descendant/inclusive CPU, retry lineage, provider points and model/launcher cost remain explicitly unknown. These observations qualify the declared local fixture, with incomplete production workflow coverage; they establish no production saving. |
| 3 | exit 0; implementation consumer routing is corroborated against the actual diff. Out-of-scope judgments remain review obligations. | exit 2; the exact command executed but could not check its consumer claims: the brief is not in the diff against source HEAD, so no entry was corroborated or disproved. Its raw output also reports absent runtime roster. This row does not satisfy expected exit 0; no consumer acceptance is claimed. |

All four witness output fingerprints and process exits were checked against retained exact combined row bytes. Full outer Go test stdout/stderr, the cost manifests, real live admission responses, protected-table derivation, source manifests, namespace evidence, recorder source/qualification and build identities are retained with this run. The fixture's inner command-test binary hashes are in the cost manifests; binary bytes are excluded from the text evidence archive. Go's test launcher can resolve its GOROOT tool directly, so capture is claimed for the actual acceptance Go processes and row boundary, not every nested compiler subprocess.

Execution used a read-only source bind at the admitted absolute home, unchanged index/HEAD and a byte-identical manifest of 6,144 tracked files. Container Git uses a minimal read-only metadata view with the exact index/HEAD and original object store; operator config, hooks and credentials are excluded. Genuine host dispatcher admission is relayed per unique request through a client-writable request directory and host-written, container-read-only responses, with a host-only raw audit. This is a qualified execution mapping, not a claim that the container has the operator's native config. Network isolation used Docker network-none plus the unchanged unshare helper; separate namespace, loopback-only interfaces and empty routes were observed. The recorder preserves actual merged row bytes, NUL-containing streams, nonzero exits and signal metadata; storage failure refuses before child execution. Native row 3's clean-diff/runtime limitation remains visible.

RISK-VALUE: DERIVED — fixture population cap = 100 @ tools/desk/cmd/deskboard/qualification_test.go:75 — the fixture must reproduce the existing production prListLimit = 100 at tools/desk/cmd/deskboard/board.go:80; treating a cap-sized population as complete would permit false idle. No production limit or authority changes in this delivery; other fixture bounds, timeouts and identities are reversible local test parameters.

**VERIFY: FAIL** — original acceptance row 3 exited 2 and could not corroborate the implementation consumer claims. Original rows 1–2 pass their observed fixture checks; supplemental row 4 exited 0 without acceptance credit. The brief must not advance on this run. The post-merge consumer-row class is tracked in [issue #1915](https://github.com/medici-finance/assay/issues/1915); no manufactured Evidence diff or source change was used to select the brief.

### Independent re-verification — 2026-10-19

This new run observes merged source 54c3342f89eddf7d815341c929afb690bb4f60a5 (origin/main at dispatch). It re-runs every row as written in the brief at that source. Earlier blocks above are kept unchanged as the historical record, including the withdrawal; this run neither authenticates nor relies on them. The verifier did not implement the delivery. The delivery is merged PR #3331, merge e28201023c74b5f407ae98f9833ee4f79999f1db, whose sole parent is f92d2e9f24251bb3ac57c6447f229916d7748dfc.

Verify-table provenance: at this source, row 3 is the delivery-pinned form landed by merged PR #3428 (the post-merge consumer-check fix for the issue #1915 class). The pin adds `--delivery`/`--base` and an exact-summary assertion. `deskpathguard rederive --main 352d5f26cf7e8d6b43fdab75934d17cacf5ba57d` exited 0 and reported 4 rows, 1 author-added (row 4). It gives rows 1–2 the same command as at this source. For row 3 it gives the base command `statusgen --root . --consumers --brief desk-tools/37`. Row 4 is author-added and earns no acceptance credit.

Verification-Attestation: medici-finance/assay#3479 run=0140bb5b7bae8fd9ba60e6e63988e0c4755b4dab702206dd95cf3ce0a5b98319 source=54c3342f89eddf7d815341c929afb690bb4f60a5 brief=docs/streams/desk-tools/brief-37-review-workflow-semantic-oracle-and-matched-cost-baseline.md model=claude-opus-5-5 tier=strong

Execution witness (statusgen verifyrun, non-dry, exit 2): this darwin host has no network-off `unshare --net`, so the witness could not run any row. That is the known host limit, issue #1800, not an observed failure.

| # | Command | Result | Output | Date | Runner |
|---|---------|--------|--------|------|--------|
| 1 | `(cd tools/desk && o=$(mktemp) && trap 'rm -f "$o"' 0 && GOWORK=off go test -count=1 -v -run '^TestWorkflowOracle$' ./internal/reviewtest/... > "$o" && grep -F -- '--- PASS: TestWorkflowOracle ' "$o")` | could-not-run exit=- sandbox=unshare — check:ci hermetic execution requires a network-off sandbox, unavailable on this host: the network sandbox uses `unshare --net`, a Linux facility, and this host is darwin. check:ci rows are re-executed network-off by design (verdict-lane/02, R-6 c.6) — run on a Linux runner that provides `unshare --net` | sha256:e3b0c44298fc | 2026-10-19 | assay-verifier-app[bot] @ 54c3342f89ed (on-behalf-of human:ian) (forge-identity) |
| 2 | `(cd tools/desk && o=$(mktemp) && trap 'rm -f "$o"' 0 && GOWORK=off go test -count=1 -v -run '^TestCostCohort$' ./internal/reviewtest/... > "$o" && grep -F -- '--- PASS: TestCostCohort ' "$o")` | could-not-run exit=- sandbox=unshare — check:ci hermetic execution requires a network-off sandbox, unavailable on this host: the network sandbox uses `unshare --net`, a Linux facility, and this host is darwin. check:ci rows are re-executed network-off by design (verdict-lane/02, R-6 c.6) — run on a Linux runner that provides `unshare --net` | sha256:e3b0c44298fc | 2026-10-19 | assay-verifier-app[bot] @ 54c3342f89ed (on-behalf-of human:ian) (forge-identity) |
| 3 | `o=$(statusgen --root . --consumers --brief desk-tools/37 --delivery e28201023c74b5f407ae98f9833ee4f79999f1db --base f92d2e9f24251bb3ac57c6447f229916d7748dfc); s=$?; printf '%s\n' "$o"; test "$s" -eq 0 && case "$o" in *'summary: 4 corroborated, 0 disproved, 0 unchecked, 0 brief(s) claiming nothing'*) ;; *) exit 1;; esac` | could-not-run exit=- sandbox=unshare — check:ci hermetic execution requires a network-off sandbox, unavailable on this host: the network sandbox uses `unshare --net`, a Linux facility, and this host is darwin. check:ci rows are re-executed network-off by design (verdict-lane/02, R-6 c.6) — run on a Linux runner that provides `unshare --net` | sha256:e3b0c44298fc | 2026-10-19 | assay-verifier-app[bot] @ 54c3342f89ed (on-behalf-of human:ian) (forge-identity) |
| 4 | `(cd tools/desk && GOWORK=off go test -count=1 -v -run '^TestGitPackagesInstallQuiet$' ./internal/gitquiet)` | could-not-run exit=- sandbox=unshare — check:ci hermetic execution requires a network-off sandbox, unavailable on this host: the network sandbox uses `unshare --net`, a Linux facility, and this host is darwin. check:ci rows are re-executed network-off by design (verdict-lane/02, R-6 c.6) — run on a Linux runner that provides `unshare --net` | sha256:e3b0c44298fc | 2026-10-19 | assay-verifier-app[bot] @ 54c3342f89ed (on-behalf-of human:ian) (forge-identity) |

By-hand execution (separate from the witness). Each row ran twice at the source: once on the host as written, and once under a darwin sandbox-exec profile that denies outbound network. The Go rows also had GOPROXY=off in the sandboxed run. The host runs added only a `-timeout 500s` to each go test. go1.27.2 darwin/arm64, git 2.56.0, statusgen v1.0.50.

| # | Command | Expected | Observed | Date / Runner |
|---|---------|----------|----------|---------------|
| 1 | row 1 as written (TestWorkflowOracle) | exit 0; named PASS; wrong-head/WAIT/order/depth controls and production overlays fail the oracle | exit 0 on host and under network-deny; `--- PASS: TestWorkflowOracle (7.49s)` (sandboxed 9.16s); 29 subtests PASS and 0 FAIL. They cover 18 oracle cases over board, planner, dispatch and issues, plus wrong-head, missing-WAIT, changed-order, wrong-depth, the printed-output mutations, production-reviewloop and production-deskdispatch | 2026-10-19 assay-verifier-app[bot] (claude-opus-5-5) (on-behalf-of human:ian) |
| 2 | row 2 as written (TestCostCohort) | exit 0; named PASS; cold/steady, skew/growth, concurrency, failed/partial; unknown costs never zero | exit 0 on host and under network-deny; `--- PASS: TestCostCohort (6.27s)` (sandboxed 5.79s); 7 subtests PASS and 0 FAIL: cold-small, steady-small, skew-growth, concurrent, partial, failed, attempts-304-and-unknown | 2026-10-19 assay-verifier-app[bot] (claude-opus-5-5) (on-behalf-of human:ian) |
| 3 | row 3 as written at source (delivery-pinned consumers corroboration) | exit 0; exact 4-corroborated summary from the immutable delivered snapshot | exit 0 on host and under network-deny; output names the range base f92d2e9f..e2820102 (immutable delivered snapshot), shows all 4 consumer entries CORROBORATED, and prints `summary: 4 corroborated, 0 disproved, 0 unchecked, 0 brief(s) claiming nothing` | 2026-10-19 assay-verifier-app[bot] (claude-opus-5-5) (on-behalf-of human:ian) |
| 3b | supplemental: the rederived BASE row-3 command, without --delivery/--base | exit 0; consumer routing corroborated against the actual diff | exit 0, but `summary: 0 corroborated, 0 disproved, 4 unchecked, 0 brief(s) claiming nothing`. The base command still cannot corroborate after merge (issue #1915 class). Recorded for provenance only; it is not this run's row 3 | 2026-10-19 assay-verifier-app[bot] (claude-opus-5-5) (on-behalf-of human:ian) |
| 4 | row 4 as written (TestGitPackagesInstallQuiet), author-added, supplemental only | exit 0; named PASS | exit 0 on host and under network-deny; `--- PASS: TestGitPackagesInstallQuiet (0.42s)`. Earns no acceptance credit | 2026-10-19 assay-verifier-app[bot] (claude-opus-5-5) (on-behalf-of human:ian) |

Risk-bearing value. The trigger fired because the delivery touches the risk-classed path prefix tools/desk/internal/deskkit/ (a lint NOTICE, risk-files-crossread). Enumeration covered the full first-parent diff f92d2e9f..e2820102: 13 files, all test, testdata, docs or changelog. The literals found were:

- The cap-sized population threshold Size >= 100, at reviewtest qualification_test.go:538 and :554, and the fixture cap 100 at deskboard qualification_test.go:75.
- The cohort sizes 3/3/100/8/3/3, concurrency 1/1/1/2/1/1 and cycles 1/2/1/2/1/1 in testdata cohorts.json:2-7.
- The planner unverifiable exit 6 at qualification_test.go:554.
- The oracle population floor of 15 cases at qualification_test.go:231 (18 are present).
- The fixture size bound 1..200 at probe.go:64.
- Timeouts 3m at qualification_test.go:56 and :364, and 45s at :108.
- Seed 37 at qualification_test.go:501, and the HTTP statuses 200/304/503 at :662.
- One CI registry opt-out entry at deskkit citrigger_test.go:2163.

No production constant, limit or authority binding changes. Every entry is a reversible local test parameter. The top-ranked entry is the cap threshold, because a wrong value could let a capped population read as complete, which is a false idle.

RISK-VALUE: DERIVED — fixture truncation threshold Size >= 100 @ tools/desk/internal/reviewtest/qualification_test.go:538 — it matches production prListLimit = 100 @ tools/desk/cmd/deskboard/board.go:80 and the production truncation rule len(prs) >= prListLimit @ tools/desk/cmd/deskboard/stalled.go:557. A read that returns exactly the page ceiling cannot prove completeness, so >= (not >) is the fail-closed boundary. The paired planner exit 6 at line 554 is reviewloop's documented "6 unverifiable" exit code (cmd/reviewloop/doc.go:64).

Landing note: this block lands on the Evidence branch after a two-parent merge of main `2bc6c4229fb90952d63382597d39fdb2ab13311a` into it. The Verify section on the branch is byte-identical to the one at the observed source.

**VERIFY: BLOCKED**: no acceptance obligation failed. Rows 1–3, as written at source 54c3342f, pass by hand both on the host and under the network-deny sandbox, and the pinned row 3 corroborates all 4 consumers. Supplemental row 4 passes and earns no acceptance credit. Closure is blocked because the mandatory verifyrun witness could not run any row on this darwin host (environment, medici-finance/assay#1800). A verified closure needs a witness for every row, so this run advances no lifecycle cell, README row or stamp. The rederived base row-3 command still reports 4 unchecked (the medici-finance/assay#1915 class); it is recorded for provenance only.

Outcome record: `blocked`, environment, medici-finance/assay#1800. It supersedes the newest earlier receipt for this brief, `37-20261018T033804Z-c015447e9a86`. The withdrawal above retracts that receipt's row-pass claims, its primary-byte audit and its medici-finance/assay#3342 blocker. All earlier receipts stay unchanged as history.
### Fresh independent verification — 2026-10-20

Verification-Attestation: medici-finance/assay#3718 run=e6049fd8292576ffefde9f4ffcfcc2f95d176cc0946f742eb27cfbe93fa69a5c source=b267d1120f28c02292a6970e3512a4b1d9c98de4 brief=docs/streams/desk-tools/brief-37-review-workflow-semantic-oracle-and-matched-cost-baseline.md model=gpt-6-astra tier=strong

Verification-Table: delivery=e28201023c74b5f407ae98f9833ee4f79999f1db base=f92d2e9f24251bb3ac57c6447f229916d7748dfc original=f92d2e9f24251bb3ac57c6447f229916d7748dfc

This run is blocked at the admitted executor preflight. Native dispatcher admission exited 0 for the source and model above. A supplemental parser-only preflight against the unchanged brief and pinned adapter exited 1 with `admitted.rows refused: unsupported Verify prose or multiline row`. No acceptance command or container executed. The adapter supports neither this Verify section's introductory prose nor the original `check:ci +flow +mutation` class spelling. Its receipt allowlist also excludes the native admitted `autocrlf=input eol=native` checkout. The latter two mismatches are source/receipt inspection, not additional execution results.

The current original commands are preserved, including restored row 3. Original row 1 retains its pre-delivery Expect below; the later expansion does not replace that obligation. Row 4 is author-added supplemental work, excluded from original acceptance credit. Existing Evidence, refusals and receipts remain historical and are not withdrawn by this blocked run.

| # | Command | Expected original obligation | Observed | Runner |
|---|---------|------------------------------|----------|--------|
| 1 | `(cd tools/desk && o=$(mktemp) && trap 'rm -f "$o"' 0 && GOWORK=off go test -count=1 -v -run '^TestWorkflowOracle$' ./internal/reviewtest/... > "$o" && grep -F -- '--- PASS: TestWorkflowOracle ' "$o")` | exit 0; named PASS. Literal intended actions, exact heads, lane independence, holds, order and uncapped depths agree with existing production entrypoints; a planted wrong head, missing WAIT row and changed ordering each fail. | could-not-run: executor preflight rejected the unchanged Verify section before row execution; no exit code or output exists for this row | 2026-10-20 gpt-6-astra |
| 2 | `(cd tools/desk && o=$(mktemp) && trap 'rm -f "$o"' 0 && GOWORK=off go test -count=1 -v -run '^TestCostCohort$' ./internal/reviewtest/... > "$o" && grep -F -- '--- PASS: TestCostCohort ' "$o")` | exit 0; named PASS. Fixture manifest reproduces cold/steady, skew/growth, concurrency and failed/partial runs; complete parent/child CPU, wall and request coverage is explicit. Unknown costs never become zero. | could-not-run: executor preflight rejected the unchanged Verify section before row execution; no exit code or output exists for this row | 2026-10-20 gpt-6-astra |
| 3 | `statusgen --root . --consumers --brief desk-tools/37` | exit 0; implementation consumer routing is corroborated against the actual diff. Out-of-scope judgments remain review obligations. | could-not-run: executor preflight rejected the unchanged Verify section before row execution; no exit code or output exists for this row | 2026-10-20 gpt-6-astra |
| 4 | `(cd tools/desk && GOWORK=off go test -count=1 -v -run '^TestGitPackagesInstallQuiet$' ./internal/gitquiet)` | exit 0; named PASS; the new Git-fixture package uses the shared quiet runner without weakening fixture controls or cleanup. | author-added; could-not-run: same executor preflight; supplemental only, no acceptance credit | 2026-10-20 gpt-6-astra |

Accounting: original acceptance 0 passed, 0 failed, 3 could-not-run / 3 total. Supplemental row 4: 0 executed / 1 total. No bound execution witness, binding file or fail-first evidence was produced. The parser diagnostic is not an execution witness. Substance and risk-value derivation remain incomplete; this run supplies no qualification or closure decision.

Needed remedy: reviewed support for the unchanged table and native checkout in the admission adapter, regression coverage, required release/pin/installation audit, then fresh independent dispatch/admission and the exact sealed execution bracket. Preserve the original acceptance table, historical findings and all existing human/closure gates. No lifecycle cell or stamp advances.

**VERIFY: BLOCKED** — admitted executor cannot represent this unchanged brief; every original acceptance row remains unrun.

## Review

Retain the derived frontmatter gate. Approval of the source spec is recorded; this brief
adds implementation acceptance rather than reopening that decision. Existing adopter
writer/custody/equivalence and collection consent gates remain separate.
