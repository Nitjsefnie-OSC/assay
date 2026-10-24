---
brief: assay:assay:desk-tools:55
title: Adopt parity-checked prune mutation selection
why: "Coverage selection and isolated workers already exist, but the prune lane runs every selected test serially for every mutation. A parity-gated selected lane can remove that repeated work while the classic CI lane independently preserves broad mutation coverage."
wave: 1
depends: []
unblocks: []
effort: M
gate: model
risk: {regulatory: no, customer: no, irreversible: no, sensitive-data: no}
issues: []
schema: brief-v2
version: 1
authored: 2026-10-24 by maintainer-directed mutation review
sources:
  - medici-finance/assay#4359
  - tools/desk/cmd/muhar/selection.go
  - freshness-checked 2026-10-24 @ 61eb62a4a (origin/main)
exec-tier: strong
exec-tier-why: "Selection changes test execution semantics and requires host-portability and per-mutant parity reasoning."
domain: complex
consumers:
  - "tools/desk/cmd/deskwt/prune-mutations.json: follow-up desk-tools/55 (this brief; flips to fixed-here when selection configuration lands)"
  - "docs/streams/desk-tools/brief-26-deskwt-prune-one-walk-and-a-lock.md: follow-up desk-tools/55 (this brief; flips to fixed-here when row 11 adopts the selected lane)"
---

# Brief 55 — Adopt parity-checked prune mutation selection

## Context
files: tools/desk/cmd/muhar/selection.go, tools/desk/cmd/muhar/selection_test.go, tools/desk/cmd/deskwt/prune-mutations.json, docs/streams/desk-tools/brief-26-deskwt-prune-one-walk-and-a-lock.md, tools/desk/cmd/muhar/README.md, changelog/muhar-prune-selection.md
facts:
- The existing `-select` implementation narrows to tests that execute the mutated function and proves each distinct selected set green on pristine source.
- Selection documents a residual false-CAUGHT risk from shared test state and changed ordering; adoption therefore requires per-mutant parity, not merely a green selected run.
- On 2026-10-24 the untouched base's selection tests failed on macOS with zero mutated files instrumented; do not enable selection until those tests pass on the host lane.
- Brief 54 may add a `-skip` flag to this manifest. Selection must preserve that flag if present; pickup reconciles whether that change has landed rather than duplicating it.
- `.github/workflows/truth-suite.yml` remains a full classic mutation lane. This brief changes the local desk-tools/26 row, not the independent broad CI control.

design-fit:
  owner: muhar coverage selection and the prune verification command
  contract: none — this extends the existing selector; no new semantic owner
  retires: [unconditional full-suite execution in the local prune lane]
  weight: "No new CLI verb or flag; carry the manifest's existing skip semantics into selection."
  why-add: "The existing selector is the right owner; a second selection mechanism was rejected."

## Ground rules
- NEVER git push / trigger workflows / run mutating infra commands. Commit only per the task instructions.
- Stop at `implemented` — you do not set verified/done (a different, non-implementing identity does).
- If anything is unclear or contradicts repo state: report NEEDS_CONTEXT, don't guess.
- Never treat a selected CAUGHT as sufficient without the classic parity required below. Do not remove the classic CI lane or shorten timeout policy to obtain parity.

## Task
1. Diagnose and fix the current host-side selection failures, including temporary-module path handling and coverage attribution. Add regressions that fail on the untouched base and pass after the fix; do not silence the tests or accept zero instrumented files.
2. Teach selection to carry `go test -skip` with the same subset-preserving semantics as `-run`. Add `TestSelectCarriesSkip` proving skipped performance tests stay excluded and unknown flags remain loud refusals.
3. Configure the prune manifest's `select.always` list only for tests that can catch a mutation without executing the mutated function. Leave it empty when no such test exists; do not use it to re-add the broad suite.
4. Change desk-tools/26 row 11 to run `muhar -j 2 -select` behind the exact 55-caught totals gate. Keep `.github/workflows/truth-suite.yml` classic and unchanged.
5. On one immutable revision, run classic and selected sweeps and diff every per-mutant verdict. Require identical verdicts and exactly 55 caught in both. If parity fails, do not adopt the selected row; land only the portability/test fix and report the mismatch.
6. Document the parity and re-parity triggers in `tools/desk/cmd/muhar/README.md`. Change both consumer entries to `fixed-here` only if the selected row is adopted; otherwise amend the brief scope and routing rather than claiming delivery.

## Verify (executable — no prose-only DoD items)
| # | Command | Expect | Class |
|---|---------|--------|-------|
| 1 | `cd tools/desk && go test ./cmd/muhar -run '^TestCoverSelector' -count=1 -v > "${TMPDIR:-/tmp}/dt55-cover.log" 2>&1 && grep -F -- '--- PASS: TestCoverSelectorEndToEnd ' "${TMPDIR:-/tmp}/dt55-cover.log"` | exit 0; named PASS and mutated files are instrumented on the host | check:ci +mutation |
| 2 | `cd tools/desk && go test ./cmd/muhar -run '^TestSelect' -count=1 -v > "${TMPDIR:-/tmp}/dt55-select.log" 2>&1 && grep -F -- '--- PASS: TestSelectCarriesSkip ' "${TMPDIR:-/tmp}/dt55-select.log" && grep -F -- '--- PASS: TestSelectSubsetRedOnPristine ' "${TMPDIR:-/tmp}/dt55-select.log"` | exit 0; skip preservation and pristine-set refusal both pass | check:ci +mutation |
| 3 | `cd tools/desk && D=$(mktemp -d "${TMPDIR:-/tmp}/dt55.XXXXXX") && go run ./cmd/muhar -spec cmd/deskwt/prune-mutations.json > "$D/classic.out" && go run ./cmd/muhar -j 2 -select -spec cmd/deskwt/prune-mutations.json > "$D/selected.out" && grep '^  [A-Z]' "$D/classic.out" > "$D/classic.verdicts" && grep '^  [A-Z]' "$D/selected.out" > "$D/selected.verdicts" && diff -u "$D/classic.verdicts" "$D/selected.verdicts" && grep -F 'Totals: 55 caught, 0 NOT CAUGHT, 0 could-not-mutate.' "$D/classic.out" && grep -F 'Totals: 55 caught, 0 NOT CAUGHT, 0 could-not-mutate.' "$D/selected.out"` | exit 0; byte-identical per-mutant verdict sequences and exact totals in both lanes | check +mutation +flow +dereference |
| 4 | `test -n "${DELIVERY_HEAD:-}" && test -n "${DELIVERY_BASE:-}" && test "$(git rev-parse HEAD)" = "$DELIVERY_HEAD" && git merge-base --is-ancestor "$DELIVERY_BASE" "$DELIVERY_HEAD" && ! grep -q -e '[:] follow-up ' docs/streams/desk-tools/brief-55-adopt-parity-checked-prune-mutation-selection.md && o=$(mktemp) && trap 'rm -f "$o"' 0 && statusgen --root . --consumers --brief desk-tools/55 --base "$DELIVERY_BASE" > "$o" 2>&1 && grep -F -e 'summary: 2 corroborated, 0 disproved, 0 unchecked' "$o"` | exit 0; manifest and row 11 are the adopted delivery diff and both routes corroborate | check +flow +dereference |

## Evidence
<!-- Independent verifier appends exact commands, outputs, source/run identity, date and runner. This authoring PR claims no future implementation test executed. -->

## Failure map
- Selection silently instruments nothing: row 1 fails rather than falling back to a deceptively green full run.
- The skip flag disappears during narrowing: row 2 fails.
- Shared state or ordering changes a verdict: row 3's per-mutant diff fails; totals alone are insufficient.
- The selected row is claimed without adoption: row 4 fails.
- A speed improvement removes the independent broad control: review rejects any classic CI removal.

## Review
Gate: model. Review the portability fix and every parity result; a faster selected run is not evidence by itself.
