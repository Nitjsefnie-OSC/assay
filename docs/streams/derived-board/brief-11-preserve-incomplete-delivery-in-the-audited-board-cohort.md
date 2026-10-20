---
brief: assay:assay:derived-board:11
title: Preserve incomplete delivery in the audited board cohort
why: >-
  The board promotes an unfinished performance brief because its merged slice lacks
  a partial-delivery declaration. Binding that source repair to the independently
  reviewed board cohort prevents another plausible completion claim from hiding owed work.
wave: 0
depends: []
unblocks: []
effort: M
gate: model
risk: {regulatory: no, customer: no, irreversible: no, sensitive-data: no}
issues: []
schema: brief-v2
version: 1
authored: 2026-10-20 by assay-worker-app
sources:
  - https://github.com/medici-finance/assay/issues/2703#issuecomment-6302222772
  - https://github.com/medici-finance/assay/pull/3200#pullrequestreview-5592762972
  - https://github.com/medici-finance/assay/pull/3200#pullrequestreview-5592834194
  - https://github.com/medici-finance/assay/pull/3759
  - https://github.com/medici-finance/assay/pull/3745
  - https://github.com/medici-finance/assay/commit/44829891e8f74b1520b64f4180910d96e17be96a
  - docs/contracts.md#semantic-owners
  - freshness-checked 2026-10-20 @ c8c4020d1
regression-of: medici-finance/assay#3787
exec-tier: strong
exec-tier-why: Cross-consumer delivery meaning and the recurring error-class design require strong execution (questions b and d).
domain: complicated
outcome: none
consumers:
  - "docs/streams/statusgen/brief-26-extend-statusgen-phase-measurements-through-the-shared-performance-recorder.md: follow-up derived-board/11 (add the supported partial declaration only; preserve owed tasks and verification)"
  - "statusgen/witnessadmission_test.go: follow-up derived-board/11 (production-owner regression against independent cohort evidence)"
  - "statusgen/boardcohort_test.go: follow-up derived-board/11 (bounded test-only membership/disposition join)"
  - "statusgen/testdata/witness-admission/: follow-up derived-board/11 (public primary evidence and recorded matcher fields)"
  - "docs/streams/derived-board/README.md: fixed-here (authoring navigation and native registration only)"
  - "statusgen/briefdelivery.go: out-of-scope (reuse existing reader and malformed/unresolved metadata floors unchanged)"
  - "statusgen/briefauthoring.go: out-of-scope (reuse shared admission and Evidence classification unchanged)"
  - "statusgen/lifecycle.go: out-of-scope (existing partial fold and legitimate full-witness meaning preserved)"
  - "statusgen/reconcile.go: out-of-scope (existing shared admission caller exercised unchanged)"
  - "statusgen/reconcilebackfill.go: out-of-scope (existing historical matcher exercised unchanged)"
  - "statusgen/reconcileapply.go: out-of-scope (existing refusal and protected-cell boundary preserved)"
  - "statusgen/regen.go: out-of-scope (existing drift consumer exercised unchanged)"
  - "docs/streams/statusgen/README.md: out-of-scope (legitimate board producer alone regenerates lifecycle cells after source merge)"
  - "tools/desk/cmd/fanoutloop/represented.go: out-of-scope (separate delivery classifier; canonical S-delivery owner remains unknown)"
  - "tools/desk/cmd/deskdispatch/phantom.go: out-of-scope (separate delivery classifier remains outside this bounded repair)"
  - "tools/desk/cmd/desksupervise/reconcile.go: out-of-scope (separate delivery classifier remains outside this bounded repair)"
---

# Brief 11 — Preserve incomplete delivery in the audited board cohort

## Context

files:
- `docs/streams/statusgen/brief-26-extend-statusgen-phase-measurements-through-the-shared-performance-recorder.md`
- `statusgen/witnessadmission_test.go`
- `statusgen/boardcohort_test.go` (planned)
- `statusgen/testdata/witness-admission/board-cohort.json` (planned)
- `statusgen/testdata/witness-admission/board-cohort-provenance.md` (planned)
- `changelog/audited-board-partial-delivery.md` (planned)

facts:
- source: At c8c4020d1, statusgen/26 has no delivery declaration for #3759. Its Tasks 3/4 and Verify 2 remain owed: snapshot reuse inventory/parity and synthetic workloads. The workload script/directory and the named parity test are absent. PR #3759 merged as 44829891e8f74b1520b64f4180910d96e17be96a on 2026-10-20T20:13:29Z; its actual body explicitly delivers sub-split (a), retains in-progress, and disclaims the remaining work. The forge and local merge object agree on its eleven changed files.
- bounded-membership: Independent correctness review 5592762972 and security review 5592834194 each inspected all ten changed rows at board head a1d5037539688ef46cf2dbc8be380a40ecfa1d35, against main c8c4020d1 and merge base 0d6a1bb5bc835c2fe80d0b23da53faba23964abf. The actual board diff/body plus both reviews fix membership independently of declarations and regression expectations. This is not every candidate in repository history. Correctness F-impl-witness-2 retains its implemented-without-delivery class and round history; security S-impl-witness-1 is the same occurrence, not a reset or separate class.
- class: A merged slice is treated as full delivery despite owed acceptance tasks; a regression built only from known declarations cannot detect a missing declaration at another brief.
- owner: S-delivery has no declared canonical semantic owner. Existing statusgen/briefdelivery.go reads supported declarations, and shared admission/fold already enforce partial meaning. No reader or classifier change is needed for this source repair.
- reproduction: Correctness independently attempted read-only reconciliation but timed out reading the pull-list body: lookedAt=false, wouldRefuse present, could-not-check. Security separately built from main c8c4020d1 and reproduced all ten actual pairs: lookedAt=true, no unread input or wouldRefuse, within 74 wouldApply / 57 held / 525 cells / ten unknown cells. That successful scoped run does not rewrite the correctness timeout or certify the other potential promotions.
- neighbor: PR #3745 (merge 4cd89433272a47d0ba99f9f01f4c559885871fb8) describes an interim perf-v2 step, carries Issue #3109 rather than a Brief trailer, and leaves statusgen/26 owed. The body matches the existing backfillCandidates substring rule, but prRecordsFromPulls excludes its issue-only trailer. backfillBase normally does not revisit a partial-trailer in-progress cell; preserve and exercise this caller distinction. Refresh actual membership and eligibility at pickup, rather than assuming body citations are delivery.
- limits: The 34 corrected-witness candidates, 87 machine-classified cases, suspected cases and untraced history in issue #2703 remain individually unproven by this ten-row audit. Existing derived-board/07 and desk-tools/34 partial repairs and their metadata/guard floors remain intact.
- scheduling: At c8c4020d1 the source reader, admission and reconcile seams below are landed; no dependency on the unmerged candidate-coverage helper planned by Authors derived-board/10 in PR #3794. Brief 11 is independently Wave 0, one bounded M/strong source-and-regression unit.

The fixed membership is:

| Brief | Reviewed witness | Review disposition |
|---|---|---|
| apps-installer/10 | #3171 | full delivery examined |
| brief-quality/01 | #1719 | full design delivery examined; no implementation of its follow-ons claimed |
| contributor-trust/09 | #980 | full delivery examined |
| desk-containers/06 | #3729 | source and offline checker delivered; no live deployment claimed |
| desk-tools/27 | #1196 | full delivery examined; VERIFY FAIL does not erase implementation |
| desk-tools/37 | #3331 | full bounded test/oracle delivery examined |
| graph-execution/28 | #3728 | full implementation examined; independent Verify still owed |
| graph-execution/32 | #3307 | bounded corpus/report delivered; no production mechanism claimed |
| measured-status/01 | #1726 | full delivery examined |
| statusgen/26 | #3759 | partial; Tasks 3/4 and Verify 2 still owed |

layering: Keep the existing source declaration and shared admission/fold. A bounded test-only membership/disposition join, adapting existing source/admission/reconcile test seams, checks coverage; production admission enforces the declaration. Tests call those actual owners with hermetic forge seams (Tasks 1–3, Verify 1–5), not a new runtime parser or prose classifier.

design-fit:
  owner: unknown (S-delivery); bounded source owner is statusgen/26 delivery metadata
  contract: S-delivery
  retires: []
  weight: verbs +0, flags +0, runtime refusals +0, rule-text lines +0; one source declaration and bounded tests/fixtures
  why-add: n/a; the existing reader already expresses the repair without a production mechanism

single-point-of-failure: Metadata alone cannot establish completeness. The independent fixed-cohort review supplies membership and evidence; the existing reader supplies enforcement. Missing source membership and malformed metadata fail differently. Semantic adequacy remains independent review; neither layer clears candidates outside this cohort.

## Read first

- The source issue design note and both full independent board reviews linked in sources; this is design-first follow-up to the class stop, not a waiver.
- PR #3759's full body, merged files and statusgen/26 Task/DoD/Verify; compare the exact source obligations with actual deliverables.
- `statusgen/briefdelivery.go`, `briefauthoring.go`, `lifecycle.go`, `reconcile.go`, `reconcilebackfill.go`, `reconcileapply.go`, `regen.go`.
- `statusgen/witnessadmission_test.go`, `reconcilepartial_test.go`, `reconcileapply_test.go`; `docs/contracts.md`, `spec/brief-v1.md`, `docs/go-engineering.md`.

## Ground rules

- Implement only after independent review and human merge of this authoring PR. No production change is delivered by the authoring PR.
- Keep all work in Assay. No infrastructure contact, workflow trigger, raw push, ready flip, generated-cell edit or verification stamp. Publish source through the normal draft PR path; human merge remains separate.
- Preserve statusgen/26 Tasks 3/4 and Verify 2 verbatim; no moving, waiving or inventing child delivery. This brief does not implement them.
- Missing/contradictory provenance or new scope is NEEDS_CONTEXT/could-not-check; do not manufacture a full/partial decision to make the board pass.

## Interface and consumer flow

| Consumer | Existing seam and future proof |
|---|---|
| Brief declaration and alias registry | readDeclaredPartials/resolveDeliveryRef read the real statusgen/26 source; no alternate parser (rows 1, 4) |
| Trailer and historical candidates | prRecordsFromPulls/backfillCandidates feed newAuthoringClassifier, nonDelivery and markDeclaredPartials under backfillBase eligibility; actual in-progress #3759 case does not enter fallback (rows 1–3, 6) |
| Lifecycle and verification | deriveOne/classifyPRs retain in-progress for partial-only and a genuine full winner under later partial plus VERIFY FAIL (rows 2, 5) |
| Drift and writer | regenDriftNotices and runReconcileCaptured exercise the real flow and existing apply/no-write boundary in isolated fixture roots (rows 2, 4, 6) |
| Generated board | Normal producer after source merge; fresh trust/reviews/human merge remain separate receipts, never an agent-triggered test |

## Task

1. **Bind the bounded cohort to primary evidence.** Record the ten pairs above from the actual pinned board diff/body and both independent full reviews, separately from expected dispositions and declarations. Store only the bounded public primary excerpts and matcher fields needed, with source URLs, revisions and hashes. Compare the independently extracted identities before accepting any disposition. Re-read each cited delivery's actual body, merge/file evidence and brief obligations at pickup; keep exact matching fields and only the public scope evidence needed, with hashes/URLs and reviewed revisions. For #3759 independently confirm the real eleven-file merge set and the still-missing named workload/parity deliverables. Source evidence for #1196 must establish the real full delivery against the scoped Task/DoD, not copy the generic deliveryFiles test stub; #2295 remains the Evidence-only counterexample. Scrub fixtures before public commit; unpublishable required evidence holds the proof. Derive the exact ten-row membership from the pinned board tree diff and compare with the independent review manifest; missing, extra, duplicated or mismatched pairs fail by identity. If the board changes, preserve this historical bound and report a separate amendment; do not silently expand it or infer every other candidate is full. Separately enumerate actual merged trailer and historical backfill candidates for statusgen/26 at pickup through existing prRecordsFromPulls and backfillCandidates, retaining collection completeness/unread information and the cutoff. Distinguish raw potential body matches from candidates eligible under the real current cell/base state. Route #3745 as a potential body match but not a current witness after the #3759 partial trailer establishes in-progress; its actual issue-only trailer and interim scope are independently recorded above. A newly discovered actually eligible candidate without supported scope is a named unresolved candidate and could-not-check/hold on #2703. A hypothetical different base state authorizes neither another declaration nor a semantic change.
2. **Declare the partial truth and exercise the real owner.** Add the supported delivery entry `{in: "assay#3759", covers: partial, note: "sub-split (a); snapshot reuse, workloads and Verify 2 remain owed"}` to statusgen/26, using its existing registry identity. Preserve all Task/Verify/Evidence content and existing declarations; add only this delivery entry and increment the brief version if required by its metadata contract. Do not weaken any other acceptance or provenance field. Production Go code stays byte-identical. Add a regression that reads current source declarations rather than building them in the test: before the entry, #3759 can become the false implemented witness; after it, statusgen/26 remains in-progress with no implementation witness. Use actual recorded matcher fields/merge identity through trailer admission and the existing backfill eligibility boundary: the real #3759 partial trailer yields in-progress and does not enter fallback. Exercise existing legitimate historical backfill controls separately, preserving their distinct fixture/base conditions; never strip the real trailer to pretend that historical mutation is the current production case. Record pre-fix red and restored green; isolated omission mutations remain re-runnable after the source fix. No PR-number special case, arbitrary prose heuristic, duplicate classifier, new runtime verb or changed meaning of full delivery.
3. **Guard the bounded class across consumers.** Check every member of the independently fixed cohort against its evidence disposition and source through shared admission, lifecycle, drift and isolated apply, retaining each caller's actual eligibility conditions. Require #3759 excluded as a completion witness without erasing any reviewed genuine delivery. Plant a second, older partial candidate in an isolated unaffected brief's recorded matcher input, absent from BOTH the expected disposition record and source declarations; the coverage join must fail naming the exact unresolved brief/PR, even when newer known partials are excluded. Do not pre-label the plant in the expected list, derive membership from declarations, or assume an unlisted candidate is full. Retain that mutation as a positive control; disabling the coverage join must make its test fail. It is explicitly synthetic, never evidence that a real full delivery is partial. Also bypass the coverage guard and show that the existing reader alone excludes a declared partial; malformed/unresolved metadata or missing forge/file evidence must retain unknown/refusal and zero writes. Preserve Evidence-only and authoring-only exclusions, full-after-partial behavior, a genuine full delivery despite later partial and VERIFY FAIL, cached reads, provenance identity, and all protected lifecycle cells. Partial alongside an earlier genuine full delivery must not erase that full witness.
4. **Hand off source, not a board stamp.** Record actual commands, revisions, red/green and limitations in the implementation PR. Stop at implemented after source review/human merge. Only the legitimate producer regenerates the board; this brief neither triggers it nor commits generated cells. Fresh regenerated board content requires current human trust admission, independent full reviews and human merge before independent graph-execution/28 Verify and then graph-execution/29. Retain both existing independent reconciliation outcomes: correctness timed out; security reproduced the ten pairs at the recorded head. Neither an offline fixture nor that old successful snapshot certifies the later repaired board; obtain fresh receipts there. Keep #2703 open. This ten-row acceptance neither depends on nor certifies derived-board/10's independent derived-board/07 historical-cohort proof. It does not clear the 34/87 broader audit or qualify Linux containers/Kubernetes runtime deployment.

If proof needs a production semantic change, new classifier, owner, flag, rule, refusal, runtime gate or broad new collection/parser mechanism, stop with NEEDS_CONTEXT. Test-only evidence joins may compare identities and reviewed dispositions; they must not interpret prose as a semantic classifier. Reuse landed seams first; no dependence on an unmerged sibling.

## Definition of Done

The supported #3759 declaration and production-owner regression stop the false promotion without weakening legitimate full witnesses. Every fixed-cohort member has primary evidence and an explicit tested disposition, including a discriminating second-site plant and independent lower-reader failure tests. Production Go, other declarations, and all unfinished statusgen/26 obligations remain unchanged. Source implementation receives its own review and human merge; generated board and runtime acceptance remain separately owned.

## Verify

Run from the repository root. Named tests are future deliverables and use hermetic fixtures; no planned tests are claimed executed here. Each command asserts a real PASS. The existing exact-class runner skips +qualified rows: a nonimplementer must explicitly execute every applicable command and record its output; parser enumeration is not execution. Unavailable Git objects/evidence are could-not-check. Semantic evidence adequacy requires independent review.

| # | Command | Expect | Class |
|---|---|---|---|
| 1 | `(cd statusgen && o=$(mktemp); go test . -run '^TestBoardCohortSource$' -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- '--- PASS: TestBoardCohortSource' "$o")` | exit 0; exactly ten pairs from the independently pinned board diff and review manifest, not declaration enumeration; source-bound #3759 test is red before metadata and green after; missing/extra pair fails with identity | check:ci +dereference +mutation |
| 2 | `(cd statusgen && o=$(mktemp); go test . -run '^TestBoardCohortFlow$' -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- '--- PASS: TestBoardCohortFlow' "$o")` | exit 0; actual trailer → admission → lifecycle → drift → isolated apply keeps statusgen/26 in-progress without completion witness or fallback; separate legitimate historical controls exercise backfill; omission mutation catches false implemented; protected cells unchanged | check:ci +flow |
| 3 | `(cd statusgen && o=$(mktemp); go test . -run '^TestBoardCohortPlant$' -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- '--- PASS: TestBoardCohortPlant' "$o")` | exit 0; committed older second-site candidate omitted from both expected/source records fails coverage naming its brief/PR; disabling the join makes the control fail; bypassed guard still leaves the real lower reader enforcing a declared partial | check:ci +mutation |
| 4 | `(cd statusgen && o=$(mktemp); go test . -run '^TestBoardCohortFloors$' -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- '--- PASS: TestBoardCohortFloors' "$o")` | exit 0; malformed/unresolved declaration, missing file evidence or mismatched provenance cannot produce a clean witness or apply; unknown/refusal and zero writes; source Task/Verify/Evidence bytes and existing delivery declarations unchanged | check:ci +mutation |
| 5 | `(cd statusgen && o=$(mktemp); go test . -run '^TestBoardCohortFull$' -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- '--- PASS: TestBoardCohortFull' "$o")` | exit 0; actual evidenced #1196 full remains implemented despite later partial and VERIFY FAIL; Evidence-only #2295 cannot replace it; all nine bounded full dispositions retain their reviewed meaning and cache behavior | check:ci +neighbour |
| 6 | `(cd statusgen && for t in TestRolloutSourceAdmission TestCISliceSourceAdmission TestOwedTaskSourceAdmission TestEvidenceAdmission TestAdmissionCacheAndDelivery TestPartialBackfillSkips TestPartialRegenDriftAgrees TestApplyReconcileWrites; do o=$(mktemp); go test . -run "^$t" -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- "--- PASS: $t" "$o"; r=$?; if test "$r" -ne 0; then exit "$r"; fi; done)` | exit 0; existing neighbors pass, including all five apply-prefix tests; no existing partial declaration or protected-cell floor weakened | check:ci +neighbour |
| 7 | `o=$(mktemp); git diff --name-only 44829891e8f74b1520b64f4180910d96e17be96a^1 44829891e8f74b1520b64f4180910d96e17be96a > "$o" && python3 -c 'import pathlib,sys; actual=set(pathlib.Path(sys.argv[1]).read_text().splitlines()); expected={"changelog/statusgen-26-v3.md","docs/streams/statusgen/README.md","docs/telemetry.md","statusgen/brieffile.go","statusgen/dispatchqueue.go","statusgen/forgeread.go","statusgen/load.go","statusgen/main.go","statusgen/parse.go","statusgen/perfrecord.go","statusgen/perfrecord_test.go"}; assert actual==expected, actual; print("PR 3759: exact eleven delivery paths confirmed")' "$o"` | exit 0; exact confirmation; real Git merge independently corroborates the fixture's delivery paths; semantic completeness still reviewed against the actual brief | check +dereference |
| 8 | `statusgen --root . --consumers --brief derived-board/11 --base c8c4020d1ffaeee8610b13e12071c2d8c66b9365` | authoring check only; exit 0, 5 corroborated, 0 disproved, 11 unchecked; at implementation pickup increment version and pin actual delivery/base and measured routing counts; unchecked scope remains review | check |
| 9 | `statusgen --root . --lint` | exit 0; LINT PASS; no new PROBLEM or generated-cell write | check |

## Pre-mortem

| Failure mode | Caught by |
|---|---|
| Regression copies declarations and omits another reviewed candidate | 1 derives the fixed cohort from independent board diff/review; 3 plants a second site |
| Only trailer admission is fixed while backfill/drift/apply promotes the slice | 2 and 6 |
| A full witness or verification failure is confused with incomplete delivery | 5 and 6 |
| Missing data or malformed metadata silently defaults to full | 4; lower reader tested with coverage bypassed in 3 |
| Metadata repair rewrites owed tasks or other protections | 4 byte checks and 6 neighbor controls |
| A fixture's semantic disposition is wrong despite valid formatting | review-only: dereference primary scope, obligations and actual files; row 7 independently binds #3759 paths |
| Ten reviewed rows are mistaken for all history, or source merge for approved board/runtime | review-only: enforce fixed membership and separate producer/trust/review/merge/verifier receipts |
| Empty selector or skipped qualified row is presented as execution | PASS assertions and explicit nonimplementer command receipts |

## Evidence

Not executed at authoring. Implementation records source-bound red/green; a nonimplementer reruns every applicable row on merged main before verification.

## Review

Gate model: all four risks are no for a reversible public source declaration and hermetic controls; no live operation, sensitive data, regulatory or irreversible change is authorized. Human merge remains required. Strong execution follows the cross-artifact flow and recurring class. The fixed cohort and per-member semantic dispositions require full independent review. This is the required bounded design amendment, not an implementation retry or a class-stop waiver.
