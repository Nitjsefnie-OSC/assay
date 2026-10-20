---
brief: assay:assay:derived-board:10
title: Complete candidate coverage for partial-delivery reconciliation
why: >-
  Excluding recent partial deliveries can expose an older preparatory PR as a false
  completion witness. A complete, independently gathered candidate cohort prevents
  reviewers and downstream workers from trusting a board that overstates delivery.
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
  - https://github.com/medici-finance/assay/pull/735
  - https://github.com/medici-finance/assay/pull/3769
  - docs/contracts.md#semantic-owners
  - freshness-checked 2026-10-20 @ 048ebfb8d
regression-of: medici-finance/assay#3769
exec-tier: strong
exec-tier-why: Cross-artifact candidate completeness and an error-class design trigger require strong execution (questions b and d).
domain: complicated
outcome: none
consumers:
  - "docs/streams/derived-board/brief-07-rollout-backfill.md: follow-up derived-board/10 (source delivery declaration)"
  - "statusgen/witnessadmission_test.go: follow-up derived-board/10 (independently sourced coverage replaces the four-number proof)"
  - "statusgen/briefdelivery.go: out-of-scope (reuse the working declaration reader without semantic changes; exercise it in this brief's flow tests)"
  - "statusgen/briefauthoring.go: out-of-scope (reuse shared trailer admission and Evidence exclusion without a new classifier)"
  - "statusgen/lifecycle.go: out-of-scope (reuse the existing partial fold and retain full witnesses)"
  - "statusgen/reconcilebackfill.go: out-of-scope (existing historical matching is exercised unchanged)"
  - "statusgen/reconcileapply.go: out-of-scope (preserve write refusal and protected cells unchanged)"
  - "statusgen/regen.go: out-of-scope (existing drift consumer is exercised unchanged)"
  - "tools/desk/cmd/fanoutloop/represented.go: out-of-scope (separate S-delivery classifier; no claim this bounded source repair fixes it)"
  - "tools/desk/cmd/deskdispatch/phantom.go: out-of-scope (separate S-delivery classifier; retain ownership gap)"
  - "tools/desk/cmd/desksupervise/reconcile.go: out-of-scope (separate S-delivery classifier; retain ownership gap)"
  - "docs/streams/derived-board/README.md: fixed-here (authoring navigation only; existing generated lifecycle cells unchanged; later lifecycle regeneration remains producer-owned)"
---

# Brief 10 — Complete candidate coverage for partial-delivery reconciliation

## Context
files:
- `docs/streams/derived-board/brief-07-rollout-backfill.md`
- `statusgen/witnessadmission_test.go`
- `statusgen/candidatecoverage_test.go` (planned)
- `statusgen/testdata/witness-admission/candidate-cohort.json` (planned)
- `statusgen/testdata/witness-admission/candidate-pages.json` (planned)
- `statusgen/testdata/witness-admission/candidate-provenance.md` (planned)
- `changelog/partial-delivery-candidate-coverage.md` (planned)

facts:
- snapshot: At 048ebfb8d on 2026-10-20, brief 07 declares #736, #945, #1010 and #1194 partial; TestRolloutSourceAdmission repeats exactly that cohort. Neither includes #735.
- reported-negative: PR #735 merged as d004a75; its body calls it a flag-day prerequisite and excludes the migrated tree. Its five paths change migration, table handling, tests and a changelog, not the whole rollout required by brief 07 Tasks 2–4.
- class: A completion fold accepts an older unsuitable candidate after newer exclusions when source declarations and their regression share the same incomplete candidate list.
- existing-owner: S-delivery in docs/contracts.md declares no canonical semantic owner. The bounded source home is brief 07 delivery metadata; statusgen/briefdelivery.go already resolves declared partials and preserves in-progress without inventing a witness.
- downstream: Issue #2703 records prior merged repairs #2758 and #3769, the strike-two design note, and broader audit bounds. This brief is the design-owed next artifact; authoring it is not an implementation retry or an exception to that stop.

layering: Retain the flat statusgen test seam and existing production admission/fold. Recorded forge responses supply candidates; a separate evidence disposition supplies expected scope. Tasks 1–3 and Verify 1–6 check their join without another runtime classifier or entrypoint.

design-fit:
  owner: unknown (S-delivery); bounded implementation remains the existing brief declarations and statusgen admission tests
  contract: S-delivery
  retires: [the four-number-only cohort in TestRolloutSourceAdmission]
  weight: verbs +0, flags +0, runtime refusals +0, rule-text lines +0; test-only candidate coverage and evidence fixtures replace the incomplete cohort proof
  why-add: n/a; no production mechanism is added, and removing the functioning partial reader would restore false completion

single-point-of-failure: A declaration list cannot prove its own completeness. Independently captured forge pagination and per-candidate scope evidence check omissions; the existing admission reader enforces declared partials. These fail for different reasons, but neither establishes universal semantic completeness outside the recorded cohort. Independent review checks evidence adequacy.

## Read first

- [Design note on issue #2703](https://github.com/medici-finance/assay/issues/2703#issuecomment-6302222772).
- [Brief schema](../../../spec/brief-v1.md), [semantic owners](../../contracts.md#semantic-owners), and [Go engineering guidance](../../go-engineering.md).
- `statusgen/briefdelivery.go`, `briefauthoring.go`, `lifecycle.go`, `reconcilebackfill.go`, `reconcileapply.go` and `regen.go`: existing admission, fold and consumers.
- `statusgen/reconcilepartial_test.go`, `witnessadmission_test.go`, `ghfetch_nocap_test.go` and `reconcileapply_test.go`: existing hermetic seams and protections.

## Ground rules

- No raw push, workflow trigger or infrastructure contact. Publish only through the authorized feature-branch draft-PR path; human merge remains separate.
- Stop at implemented; a nonimplementer owns verification. Do not edit generated board cells, verification stamps or completed Evidence.
- Implement only after independent design review and human merge of the authoring PR. This brief supplies the required design; it grants no bleed exception, runtime qualification or deployment authority.
- Report NEEDS_CONTEXT for contradictory source facts. Missing, capped or unread evidence is could-not-check, never an empty or complete cohort.

## Task

1. **Capture the complete bounded cohort independently.** At one dated public-forge snapshot, use the existing full collection path to inventory every merged trailer candidate and historical branch/body backfill candidate for derived-board/07, not only the winning PR and not numbers read from delivery declarations. Record request parameters, page order, terminal pagination evidence, candidate counts, source URLs, merge SHAs, bodies and required file/patch evidence in the three named fixtures. Preserve every field used by the existing matchers across the scanned pages, including nonmatches, so cohort selection can be replayed rather than asserted. Retain the raw membership inputs separately from the derived disposition list, with hashes and collection limits; redact unrelated content. Each member must have an evidence-backed full/partial/authoring/Evidence-only disposition and a reason referring to actual brief obligations. An independently evidenced full delivery for a different brief is the positive control; use desk-tools/27 #1196 only after checking its scope against #1062 and the real brief, not by copying a test's deliveryFiles stub. Audit the collection's completeness separately from semantic judgments. Capped pages, missing file lists/patches, inconsistent snapshot boundaries or unresolved candidates prevent a clean coverage claim. Refresh before implementation if the snapshot changed; preserve both revisions and the reason.
2. **Repair the input, preserve the reader.** Append the supported partial delivery declaration for assay#735 to brief 07 with its bounded preparatory scope; retain all four existing entries and history. Read every other cohort member before deciding whether another truthful declaration is needed. Do not classify from arbitrary PR prose keywords, special-case numbers in production, edit historical trailers, or add a second reader. If an unexpected candidate would require broader semantics or production changes, hold that scope on #2703 for design amendment. Replace the four-number-only regression with a check deriving membership from the recorded pages through existing collection/matching seams, joined to separately evidenced dispositions and current source declarations. Missing, duplicate, unexplained or contradictory membership must name the exact candidate and fail coverage. A complete cohort here means complete for this brief, snapshot and both matching paths, not all repository history forever.
3. **Make the proof discriminate.** Add the short named tests in Verify, tagged regression #2703. Run the source-backed #735 negative red before its declaration, then green afterward. Across trailer and historical backfill paths, drive the same cohort through admission, lifecycle, drift reporting and isolated apply; require in-progress with no implementation witness or forbidden status write. The positive separately evidenced full delivery survives a later partial and VERIFY FAIL. Plant a second, older preparatory candidate in recorded collection input, absent from source declarations and the expected disposition list: coverage must fail naming it even if newer candidates are excluded. Commit this as a re-runnable positive control, not a production exception. Also test corrupt/missing/capped snapshot input and malformed/unresolved delivery metadata; exercise the underlying reader with coverage bypassed to prove it still excludes a declared partial and refuses unread metadata independently. Preserve authoring-only/Evidence-only exclusions, cached reads, full-over-partial semantics, and existing protected-cell behavior. Record actual red/green commands, revisions and exact failures; never claim a planned test ran.
4. **Bound the conclusion and hand off.** Record each adjacent audit class as covered here, deferred on #2703, or excluded with a reason in candidate-provenance.md. This cohort covers derived-board/07 only. The audit's 34 corrected-witness candidates, 87 machine-classified cases, suspected cases and untraced pre-trailer done rows remain unproven individually; #2655 and other briefs are separate follow-up, not silently cleared. Keep #2703 open. After independently reviewed source merge, the legitimate reconciliation producer regenerates the board; inspect its actual revision and candidate result. Fresh board content requires current human trust admission, independent reviews and human merge before independent graph-execution/28 Verify, then graph-execution/29. That downstream chain is separate from this brief's offline acceptance. Linux cellctl and per-desk containers remain the user outcome; no Windows, runtime Kubernetes qualification or live deployment is delivered here.

## Definition of Done

All cohort members and completeness limits are reviewable independently of the exclusion list; the real #735 negative, full positive, planted second instance, unknown cases and cross-consumer flow discriminate. Production changes are limited to truthful brief delivery inputs. The source change has its own draft PR, review and human merge. No generated content is committed by the implementer. The existing scheduled board producer owns regeneration of stream README lifecycle cells and STATUS.md after that merge; record the actual later producer result separately, without triggering it or promising its outcome.

## Verify

Run from the repository root. Each test row saves output and asserts a real PASS line so a renamed or missing test cannot pass vacuously. Tests are future deliverables, not claims about the authoring PR. Collection is read-only forge access; execution tests replay recorded inputs offline. Row 10 dereferences the real reported merge in local Git history; an unavailable object is could-not-check, requiring ordinary source-history acquisition before rerun. Semantic adequacy of per-candidate evidence remains independent review.

| # | Command | Expect | Class |
|---|---------|--------|-------|
| 1 | `(cd statusgen && o=$(mktemp); go test . -run '^TestCandidateSnapshot$' -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- '--- PASS: TestCandidateSnapshot' "$o")` | exit 0; independently recorded pages replay to the exact candidate cohort through trailer and backfill matching, including #735; hashes, terminal pagination and per-member evidence bindings checked; missing/capped/corrupt inputs cannot certify completeness | check:ci +dereference |
| 2 | `(cd statusgen && o=$(mktemp); go test . -run '^TestCandidateCoverage$' -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- '--- PASS: TestCandidateCoverage' "$o")` | exit 0; every recorded candidate has one supported disposition and required declaration; red before #735 input repair names #735; coverage does not derive membership from declarations | check:ci +mutation |
| 3 | `(cd statusgen && o=$(mktemp); go test . -run '^TestCandidatePlant$' -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- '--- PASS: TestCandidatePlant' "$o")` | exit 0; committed second-older-candidate mutation is rejected with that exact identity, independently of the original five candidates; disabling the coverage join makes this positive control fail | check:ci +mutation |
| 4 | `(cd statusgen && o=$(mktemp); go test . -run '^TestCandidateFlow$' -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- '--- PASS: TestCandidateFlow' "$o")` | exit 0; source declarations → both admission paths → lifecycle → drift → isolated apply retain in-progress and no witness for partial-only cohort; before #735 repair the test fails on false implemented; apply leaves protected cells byte-identical | check:ci +flow |
| 5 | `(cd statusgen && o=$(mktemp); go test . -run '^TestCandidateFull$' -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- '--- PASS: TestCandidateFull' "$o")` | exit 0; separately evidenced full delivery remains implemented and the witness under later partial and VERIFY FAIL; full/partial history and existing read cache remain valid | check:ci +neighbour |
| 6 | `(cd statusgen && o=$(mktemp); go test . -run '^TestCandidateUnknown$' -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- '--- PASS: TestCandidateUnknown' "$o")` | exit 0; cap and missing evidence produce could-not-check coverage; malformed/unresolved declaration produces unknown derivation and apply refusal with no writes; direct reader still protects declared partials with coverage bypassed | check:ci +mutation |
| 7 | `(cd statusgen && for t in TestEvidenceAdmission TestAdmissionCacheAndDelivery TestPartialBackfillSkips TestPartialRegenDriftAgrees TestApplyReconcileWrites TestFetchAllPullsNoSilentCap; do o=$(mktemp); go test . -run "^$t" -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- "--- PASS: $t" "$o"; r=$?; if test "$r" -ne 0; then exit "$r"; fi; done)` | exit 0; existing Evidence, cache, partial, cap and write-boundary controls still execute and pass | check:ci +neighbour |
| 8 | `statusgen --root . --consumers --brief derived-board/10 --base 048ebfb8dd7d9333e37810072ab6b3beba29f56d` | pre-merge authoring check only; exit 0, exactly 3 corroborated, 0 disproved and 9 unchecked. At implementation pickup amend this row, with a version increment, to immutable delivery and first-parent base pins and actual expected routing counts before independent execution verification; unchecked judgments require review | check |
| 9 | `statusgen --root . --lint` | exit 0; no new PROBLEM; generated surfaces remain producer-owned | check |
| 10 | `o=$(mktemp); git diff --name-only d004a756bbe8f0e7920376d96c61d8d50a991914^1 d004a756bbe8f0e7920376d96c61d8d50a991914 > "$o" && python3 -c 'import pathlib,sys; actual=set(pathlib.Path(sys.argv[1]).read_text().splitlines()); expected={"changelog/statusgen-migrate-registers-and-table-cells.md","statusgen/migrate.go","statusgen/migrate_test.go","statusgen/readmetable.go","statusgen/readmetable_test.go"}; assert actual==expected, actual; print("PR 735: exact five preparatory paths confirmed")' "$o"` | exit 0; exact confirmation line; independent local Git object confirms the reported preparatory path set, not merely a fixture repeating it; delivery adequacy remains reviewed against brief 07 scope | check +dereference |

## Pre-mortem

| Failure mode | Caught by |
|---|---|
| Test copies the same incomplete exclusion list and misses an older candidate | 1–3; independent reviewer dereferences raw membership and scope evidence |
| Candidate capture stops at a cap or loses a page | 1 and 6; absent terminal evidence is could-not-check |
| #735 is fixed only in a synthetic fixture | 2 and 4 read current brief declarations |
| A real full delivery is erased or VERIFY FAIL is misread as no delivery | 5 and 7 |
| Backfill, drift or apply bypasses admission | 4 and 7 |
| Coverage failure is hidden by a no-op test selector | PASS assertions in every test row |
| Fixture evidence is plausible but semantically wrong, or broader audit is falsely cleared | no automated semantic verdict; independent review of candidate-provenance.md against primary sources and the stated bounds |
| Source merge is mistaken for board approval or runtime readiness | review-only: inspect the separate producer, trust, review, human-merge and independent-verifier receipts |

## Evidence

Not executed at authoring. Implementer records each red/green run and its bound source; a nonimplementer reruns these rows on merged main before verification.

## Review

Gate: model; all four risks are no because the deliverable is reversible repository input and hermetic tests, with no customer operation, regulated action, irreversible effect or sensitive data. Human merge still applies. Strong execution is required for the recurring class and cross-artifact completeness. Review must independently inspect cohort provenance and scope judgments; a green fixture test alone cannot certify those judgments. No whole-issue closure or runtime qualification follows from this bounded proof.
