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
version: 3
authored: 2026-10-20 by assay-worker-app
sources:
  - https://github.com/medici-finance/assay/issues/2703#issuecomment-6302222772
  - https://github.com/medici-finance/assay/pull/735
  - https://github.com/medici-finance/assay/pull/3769
  - docs/contracts.md#semantic-owners
  - https://github.com/medici-finance/assay/commit/c45ef4ec6d3c3f1220cd72d0932ae10ffd77bda7
  - https://github.com/medici-finance/assay/pull/3787
  - freshness-checked 2026-10-20 @ 0d6a1bb5b
regression-of: medici-finance/assay#3787
exec-tier: strong
exec-tier-why: Cross-artifact candidate completeness and an error-class design trigger require strong execution (questions b and d).
domain: complicated
outcome: none
consumers:
  - "docs/streams/derived-board/brief-07-rollout-backfill.md: out-of-scope (c45ef4ec6 already added the missing partial declaration; read unchanged, mutate only an isolated test copy)"
  - "statusgen/witnessadmission_test.go: follow-up derived-board/10 (independently sourced coverage replaces the five-number-only proof)"
  - "statusgen/ghfetch.go: out-of-scope (reuse collector and injected httpDoer unchanged; enforce opt-in read budgets in test-only code)"
  - "statusgen/briefdelivery.go: out-of-scope (reuse the working declaration reader without semantic changes; exercise it in this brief's flow tests)"
  - "statusgen/briefauthoring.go: out-of-scope (reuse shared trailer admission and Evidence exclusion without a new classifier)"
  - "statusgen/lifecycle.go: out-of-scope (reuse the existing partial fold and retain full witnesses)"
  - "statusgen/reconcilebackfill.go: out-of-scope (existing historical matching is exercised unchanged)"
  - "statusgen/reconcileapply.go: out-of-scope (preserve write refusal and protected cells unchanged)"
  - "statusgen/reconcile.go: out-of-scope (direct caller of shared admission remains unchanged; exercised by the source-to-apply flow)"
  - "statusgen/regen.go: out-of-scope (existing drift consumer is exercised unchanged)"
  - "tools/desk/cmd/fanoutloop/represented.go: out-of-scope (separate S-delivery classifier; no claim this bounded proof fixes it)"
  - "tools/desk/cmd/deskdispatch/phantom.go: out-of-scope (separate S-delivery classifier; retain ownership gap)"
  - "tools/desk/cmd/desksupervise/reconcile.go: out-of-scope (separate S-delivery classifier; retain ownership gap)"
  - "docs/streams/derived-board/README.md: fixed-here (authoring navigation only; existing generated lifecycle cells unchanged; later lifecycle regeneration remains producer-owned)"
---

# Brief 10 — Complete candidate coverage for partial-delivery reconciliation

## Context
files:
- `statusgen/witnessadmission_test.go`
- `statusgen/candidatecoverage_test.go` (planned)
- `statusgen/testdata/witness-admission/candidate-cohort.json` (planned)
- `statusgen/testdata/witness-admission/candidate-pages.json` (planned)
- `statusgen/testdata/witness-admission/candidate-provenance.md` (planned)
- `changelog/partial-delivery-candidate-coverage.md` (planned)

facts:
- snapshot: At 0d6a1bb5b on 2026-10-20, merged c45ef4ec6 declares #735 partial and expands TestRolloutSourceAdmission to five hardcoded numbers (735, 736, 945, 1010, 1194). The source omission is repaired; independently gathered membership is still absent. At historical 048ebfb8d both lists omitted #735; that is the origin example, not current work owed.
- reported-negative: PR #735 merged as d004a75; its body calls it a flag-day prerequisite and excludes the migrated tree. Its five paths change migration, table handling, tests and a changelog, not the whole rollout required by brief 07 Tasks 2–4.
- class: A completion fold accepts an older unsuitable candidate after newer exclusions when source declarations and their regression share the same incomplete candidate list.
- existing-owner: S-delivery in docs/contracts.md declares no canonical semantic owner. The bounded source home is brief 07 delivery metadata; statusgen/briefdelivery.go already resolves declared partials and preserves in-progress without inventing a witness.
- downstream: Issue #2703 records prior repairs #2758, #3769 and now #3787, the strike-two design note, and broader audit bounds. Commit 552bc5aaa separately adds owed-task partial metadata for harness-portability/13 and desk-supervision/08; this brief preserves it. The remaining design concerns independent cohort proof, not reapplying source repairs or waiving the stop.

layering: Retain the flat statusgen test seam and existing production admission/fold. Recorded forge responses supply candidates; a separate evidence disposition supplies expected scope. Tasks 1–4 and Verify 1–6, 11–12 check their join without another runtime classifier or entrypoint.

design-fit:
  owner: unknown (S-delivery); bounded implementation is statusgen admission tests reading the existing brief declarations
  contract: S-delivery
  retires: [the five-number-only cohort in TestRolloutSourceAdmission]
  weight: verbs +0, flags +0, runtime refusals +0, rule-text lines +0; test-only candidate coverage and evidence fixtures replace the incomplete cohort proof
  why-add: n/a; no production mechanism is added; the functioning partial reader and all five repaired declarations remain intact

single-point-of-failure: A declaration list cannot prove its own completeness. A verifier opts into a separately collected forge view to check recorded membership; hermetic controls prove it detects an omitted page even with a consistent forged terminal receipt. Per-candidate scope evidence checks semantics; the existing admission reader enforces declared partials. These fail for different reasons, but neither establishes universal semantic completeness outside the recorded cohort. Independent review checks semantic evidence adequacy. Two equal bounded forge walks support current membership for pre-cutoff merges; they do not prove an atomic historical snapshot or recover changed historical matcher inputs.

## Read first

- [Design note on issue #2703](https://github.com/medici-finance/assay/issues/2703#issuecomment-6302222772).
- [Brief schema](../../../spec/brief-v1.md), [semantic owners](../../contracts.md#semantic-owners), and [Go engineering guidance](../../go-engineering.md).
- `statusgen/ghfetch.go`, `briefdelivery.go`, `briefauthoring.go`, `lifecycle.go`, `reconcilebackfill.go`, `reconcileapply.go` and `regen.go`: existing admission, fold and consumers.
- `statusgen/reconcilepartial_test.go`, `witnessadmission_test.go`, `ghfetch_nocap_test.go` and `reconcileapply_test.go`: existing hermetic seams and protections.

## Ground rules

- No raw push, workflow trigger or infrastructure contact. Publish only through the authorized feature-branch draft-PR path; human merge remains separate.
- Stop at implemented; a nonimplementer owns verification. Do not edit generated board cells, verification stamps or completed Evidence.
- Implement only after independent design review and human merge of the authoring PR. This brief supplies the required design; it grants no bleed exception, runtime qualification or deployment authority.
- Report NEEDS_CONTEXT for contradictory source facts. Missing, capped or unread evidence is could-not-check, never an empty or complete cohort.

## Task

1. **Capture the complete bounded cohort independently.** At one dated public-forge snapshot, use the existing full collection path to inventory every merged trailer candidate and historical branch/body backfill candidate for derived-board/07, not only the winning PR and not numbers read from delivery declarations. Record request parameters, page order, terminal pagination evidence, candidate counts, source URLs, merge SHAs, bodies and required file/patch evidence in the three named fixtures. Preserve every field used by the existing matchers across the scanned pages, including nonmatches, so cohort selection can be replayed rather than asserted. Retain the raw membership inputs separately from the derived disposition list, with hashes, a UTC merged-before cutoff and collection limits. Keep only fields consumed by existing matchers and identity checks: number, state, draft, body, merged_at, merge_commit_sha and head.sha/ref. Preserve exact matcher bytes, including nonmatches; omit unrelated response fields. Manually scrub before public commitment; if required matcher bytes cannot safely be published, hold this proof could-not-check for a reviewed evidence amendment rather than silently redacting them or claiming exact replay. Each member must have an evidence-backed full/partial/authoring/Evidence-only disposition and a reason referring to actual brief obligations. An independently evidenced full delivery for a different brief is the positive control; use desk-tools/27 #1196 only after checking its scope against #1062 and the real brief, not by copying a test's deliveryFiles stub. Audit the collection's completeness separately from semantic judgments. Capped pages, missing file lists/patches, inconsistent snapshot boundaries or unresolved candidates prevent a clean coverage claim. Refresh before implementation if the snapshot changed; preserve both revisions and the reason.
2. **Prove coverage of the repaired input, preserve the reader.** Keep the #735 declaration delivered by c45ef4ec6 and the other four partial entries byte-identical in the production brief. Replace the five-number-only regression with a check deriving membership from the recorded pages through existing collection/matching seams, joined to separately evidenced dispositions and current source declarations. Missing, duplicate, unexplained or contradictory membership must name the exact candidate and fail coverage. Do not classify from arbitrary PR prose keywords, special-case numbers in production, edit historical trailers, or add a second reader. If a newly discovered candidate is not truthfully covered, report it on #2703 and hold that affected proof for a reviewed design amendment; this brief does not authorize new production declarations or semantics. A complete cohort here means complete for this brief, snapshot and both matching paths, not all repository history forever. Preserve the separate owed-task controls added by 552bc5aaa.
3. **Independently recheck membership with explicit opt-in.** Implement TestCandidateRecollect in the same test file, using the existing ghClient.fetchAllPulls/httpDoer injection seam, prRecordsFromPulls and backfillCandidates; do not copy their parsers, predicates or runtime classifier. Default execution skips this network-only test before reading credentials or constructing a network transport. Only the one-command ASSAY_COHORT_RECHECK=1 opt-in in Verify 11 enables it, with an explicitly supplied GITHUB_TOKEN from the existing authorized role; never discover ambient accounts or acquire wider scopes. Print the fixed target https://api.github.com/repos/medici-finance/assay/pulls and read bounds before first contact. Allow only GET collection requests to that origin/path, reject redirects, and reuse the existing collector with a test-only bounded httpDoer. Make two fresh complete walks, state=all/per_page=100, recording each requested page, response hash/count, terminal short page, elapsed time and actual request/byte totals in a verifier-owned temporary receipt bound to source SHA and fixture hash. Hard ceilings across both walks: 200 requests, 100 pages per walk, 2 MiB per response, 64 MiB total, and 180 seconds wall time; no retries beyond these budgets. Reaching any bound, HTTP/auth/rate-limit error, malformed or duplicate record, inconsistent terminal, or stderr yields could-not-check and no clean marker. Filter each walk to merged_at at or before the recorded cutoff; later merges cannot enter the cohort. Require both eligible record sets and matcher-input bytes to agree, then independently recompute both matching paths and compare against the fixture's complete eligible record inventory and cohort. An independently returned eligible record missing from the capture (including a whole omitted page), missing candidate or contradictory cohort is a named checked-failed coverage result, even if the fixture's own pagination appears valid. A captured historical record now absent, or changed historical matcher bytes, is could-not-check: the API cannot establish what a mutable body/ref contained at capture time. Distinguish this current-equivalence failure from a claim about unknowable historical semantics. New merges may move page boundaries; disagreement between walks is could-not-check, never permission to splice a plausible snapshot. This optional network path is test-only, not part of ordinary CI, runtime reconciliation or downstream deployment gates. A nonimplementer must perform it for this brief's verification; unavailable access remains could-not-check, not a fixture substitute.
4. **Make the proof discriminate.** Add the short named tests in Verify, tagged regression #2703. Run the source-backed #735 negative against the recorded pre-fix source or an isolated copy of current declarations with only #735 removed: observe the false implemented witness and named coverage failure, restore the declaration, and observe green. Never remove it from the production brief; prove the test copy was restored and production inputs stayed byte-identical. Across trailer and historical backfill paths, drive the same cohort through admission, lifecycle, drift reporting and isolated apply; require in-progress with no implementation witness or forbidden status write. The positive separately evidenced full delivery survives a later partial and VERIFY FAIL. Plant a second, older preparatory candidate in recorded collection input, absent from source declarations and the expected disposition list: coverage must fail naming it even if newer candidates are excluded. Commit this as a re-runnable positive control, not a production exception. Also test corrupt/missing/capped snapshot input and malformed/unresolved delivery metadata; exercise the underlying reader with coverage bypassed to prove it still excludes a declared partial and refuses unread metadata independently. Preserve authoring-only/Evidence-only exclusions, cached reads, full-over-partial semantics, and existing protected-cell behavior. Record actual red/green commands, revisions and exact failures; never claim a planned test ran.
5. **Bound the conclusion and hand off.** Record each adjacent audit class as covered here, deferred on #2703, or excluded with a reason in candidate-provenance.md. This cohort covers derived-board/07 only. The audit's 34 corrected-witness candidates, 87 machine-classified cases, suspected cases and untraced pre-trailer done rows remain unproven individually; #2655 and other briefs are separate follow-up, not silently cleared. Keep #2703 open. Source repair #3787 is already merged. Its downstream chain still requires the legitimate reconciliation producer to regenerate the board; inspect its actual revision and candidate result. This proof brief neither triggers that producer nor substitutes for any downstream receipt. Fresh board content requires current human trust admission, independent reviews and human merge before independent graph-execution/28 Verify, then graph-execution/29. That downstream chain is separate from this brief's offline acceptance. Linux cellctl and per-desk containers remain the user outcome; no Windows, runtime Kubernetes qualification or live deployment is delivered here.

## Definition of Done

All cohort members and completeness limits are reviewable independently of the exclusion list; the real #735 negative, full positive, planted second instance, unknown cases and cross-consumer flow discriminate. A nonimplementer independently rechecks bounded forge membership; unavailable or changed historical inputs prevent a clean result. Production declarations and the working reader remain unchanged. The cohort-proof implementation has its own draft PR, review and human merge. No generated content is committed by the implementer. The existing scheduled board producer owns regeneration of stream README lifecycle cells and STATUS.md from repaired main; record the actual producer result separately, without triggering it, promising its outcome or adding this proof as a new deployment gate.

## Verify

Run from the repository root. Each test row saves output and asserts a real PASS line so a renamed or missing test cannot pass vacuously. Tests are future deliverables, not claims about the authoring PR. Collection is read-only forge access. Rows 1–7 and 12 replay recorded inputs offline; row 11 alone explicitly opts into independent forge reads and is excluded from ordinary CI. A skip is never success for row 11. Row 10 dereferences the real reported merge in local Git history; an unavailable object is could-not-check, requiring ordinary source-history acquisition before rerun. Semantic adequacy of per-candidate evidence remains independent review; the current API cannot reconstruct historical mutable bodies. Row 11 proves bounded current equivalence for pre-cutoff merges, not universal or atomic historical completeness.

| # | Command | Expect | Class |
|---|---------|--------|-------|
| 1 | `(cd statusgen && o=$(mktemp); go test . -run '^TestCandidateSnapshot$' -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- '--- PASS: TestCandidateSnapshot' "$o")` | exit 0; independently recorded pages replay to the exact candidate cohort through trailer and backfill matching, including #735; hashes, terminal pagination and per-member evidence bindings checked; missing/capped/corrupt inputs cannot certify completeness | check:ci +dereference |
| 2 | `(cd statusgen && o=$(mktemp); go test . -run '^TestCandidateCoverage$' -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- '--- PASS: TestCandidateCoverage' "$o")` | exit 0; every recorded candidate has one supported disposition and required declaration; isolated declaration-omission mutation or recorded pre-fix source fails naming #735; restoring the already-merged declaration passes; coverage does not derive membership from declarations | check:ci +mutation |
| 3 | `(cd statusgen && o=$(mktemp); go test . -run '^TestCandidatePlant$' -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- '--- PASS: TestCandidatePlant' "$o")` | exit 0; committed second-older-candidate mutation is rejected with that exact identity, independently of the original five candidates; disabling the coverage join makes this positive control fail | check:ci +mutation |
| 4 | `(cd statusgen && o=$(mktemp); go test . -run '^TestCandidateFlow$' -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- '--- PASS: TestCandidateFlow' "$o")` | exit 0; source declarations → both admission paths → lifecycle → drift → isolated apply retain in-progress and no witness for partial-only cohort; with #735 removed only from the isolated source copy, the test fails on false implemented; restored current source passes; apply leaves protected cells byte-identical | check:ci +flow |
| 5 | `(cd statusgen && o=$(mktemp); go test . -run '^TestCandidateFull$' -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- '--- PASS: TestCandidateFull' "$o")` | exit 0; separately evidenced full delivery remains implemented and the witness under later partial and VERIFY FAIL; full/partial history and existing read cache remain valid | check:ci +neighbour |
| 6 | `(cd statusgen && o=$(mktemp); go test . -run '^TestCandidateUnknown$' -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- '--- PASS: TestCandidateUnknown' "$o")` | exit 0; cap and missing evidence produce could-not-check coverage; malformed/unresolved declaration produces unknown derivation and apply refusal with no writes; direct reader still protects declared partials with coverage bypassed | check:ci +mutation |
| 7 | `(cd statusgen && for t in TestEvidenceAdmission TestAdmissionCacheAndDelivery TestPartialBackfillSkips TestPartialRegenDriftAgrees TestOwedTaskSourceAdmission TestApplyReconcileWrites TestFetchAllPullsNoSilentCap; do o=$(mktemp); go test . -run "^$t" -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- "--- PASS: $t" "$o"; r=$?; if test "$r" -ne 0; then exit "$r"; fi; done)` | exit 0; existing Evidence, cache, partial, cap and write-boundary controls still execute and pass; TestApplyReconcileWrites intentionally selects all five prefixed tests, and any failure fails go test | check:ci +neighbour |
| 8 | `statusgen --root . --consumers --brief derived-board/10 --base 0d6a1bb5bc835c2fe80d0b23da53faba23964abf` | pre-merge authoring check only; exit 0, exactly 2 corroborated, 0 disproved and 12 unchecked. At implementation pickup amend this row, with a version increment, to immutable delivery and first-parent base pins and actual expected routing counts before independent execution verification; unchecked judgments require review | check |
| 9 | `statusgen --root . --lint` | exit 0; no new PROBLEM; generated surfaces remain producer-owned | check |
| 10 | `o=$(mktemp); git diff --name-only d004a756bbe8f0e7920376d96c61d8d50a991914^1 d004a756bbe8f0e7920376d96c61d8d50a991914 > "$o" && python3 -c 'import pathlib,sys; actual=set(pathlib.Path(sys.argv[1]).read_text().splitlines()); expected={"changelog/statusgen-migrate-registers-and-table-cells.md","statusgen/migrate.go","statusgen/migrate_test.go","statusgen/readmetable.go","statusgen/readmetable_test.go"}; assert actual==expected, actual; print("PR 735: exact five preparatory paths confirmed")' "$o"` | exit 0; exact confirmation line; independent local Git object confirms the reported preparatory path set, not merely a fixture repeating it; delivery adequacy remains reviewed against brief 07 scope | check +dereference |

| 11 | `(cd statusgen && test -n "$GITHUB_TOKEN" && o=$(mktemp) && ASSAY_COHORT_RECHECK=1 go test . -run '^TestCandidateRecollect$' -count=1 -timeout 210s -v > "$o" 2>&1 && grep -Fx -- 'RECOLLECT checked-clean' "$o" && grep -F -- '--- PASS: TestCandidateRecollect' "$o")` | exit 0 only after two bounded independent forge walks agree with the recorded pre-cutoff eligible inventory and both candidate paths; receipt binds revision, fixture hash, page/response hashes, terminal evidence and actual read cost; omitted-page membership is named checked-failed; changed historical inputs, caps or unavailable access are could-not-check with no clean marker; credentials never printed | check +dereference |
| 12 | `(cd statusgen && o=$(mktemp); go test . -run '^TestRecollectControls$' -count=1 -timeout 60s -v > "$o" 2>&1 && grep -F -- '--- PASS: TestRecollectControls' "$o")` | exit 0; independent httptest responses expose a fixture with an omitted whole page and recomputed self-consistent terminal receipt/hashes, naming the missing member; disabling independent comparison makes the control fail; changed old body/ref, disagreement between walks, every read bound, error and missing token yield could-not-check; post-cutoff merges are excluded; without opt-in a transport that fails on any call proves zero network/credential reads | check:ci +mutation |

## Pre-mortem

| Failure mode | Caught by |
|---|---|
| Test copies the same incomplete exclusion list and misses an older candidate | 1–3; independent reviewer dereferences raw membership and scope evidence |
| Candidate capture stops at a cap or loses a page, with a plausible terminal receipt | 1, 6, 11 and 12; fresh independent membership detects a self-consistent omission; bounds or missing evidence remain could-not-check |
| Historical matcher inputs change, pagination shifts, or recheck silently contacts the network | 11 and 12; double-walk agreement, explicit opt-in and enforced read budgets; historical reconstruction remains unproven |
| Tests assume #735 is still missing or accidentally revert the merged fix | 2 and 4 read current declarations and confine the negative mutation to an isolated copy; production inputs must stay byte-identical |
| A real full delivery is erased or VERIFY FAIL is misread as no delivery | 5 and 7 |
| Backfill, drift or apply bypasses admission | 4 and 7 |
| Coverage failure is hidden by a no-op test selector | PASS assertions in every test row |
| Fixture evidence is plausible but semantically wrong, or broader audit is falsely cleared | no automated semantic verdict; independent review of candidate-provenance.md against primary sources and the stated bounds |
| Source merge is mistaken for board approval or runtime readiness | review-only: inspect the separate producer, trust, review, human-merge and independent-verifier receipts |

## Evidence

Not executed at authoring. Implementer records each red/green run and its bound source; a nonimplementer reruns these rows on merged main before verification.

## Review

Gate: model; all four risks are no because the deliverable is reversible recorded evidence, hermetic tests and an explicitly requested bounded public-forge read, with no customer operation, regulated action, irreversible effect or sensitive data. Human merge still applies. Strong execution is required for the recurring class and cross-artifact completeness. Review must independently inspect cohort provenance and scope judgments; a green fixture test alone cannot certify those judgments. No whole-issue closure or runtime qualification follows from this bounded proof.
