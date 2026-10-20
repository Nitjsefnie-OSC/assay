---
brief: assay:assay:statusgen:36
title: "Content-bound witness schema v2: the reader and the one digest owner"
why: "Coverage still binds an execution witness to commit ancestry, so a squash merge orphans a valid witness and a whole-tree scope refuses witnesses whose real inputs never changed. Thirty verified briefs are stuck this way. This brief makes a witness's identity the delivering change request (provenance) plus a digest of the content the row read (staleness), judged in one pure library, with the existing v1 rule kept unchanged and any unknown schema refused."
wave: 1
depends: []
unblocks: ["statusgen/37"]
effort: L
gate: human
risk: {regulatory: no, customer: yes, irreversible: no, sensitive-data: no}
issues: [2026, 3105, 3658]
regression-of: "medici-finance/assay#2304"
schema: brief-v2
outcome: none
version: 1
authored: "2026-10-20 by the-desk (design brief for the witness-staleness class)"
gate-why: "It narrows a verification control. A v2 witness is released on content equality instead of commit ancestry, so a wrong digest scope or a wrong environment rule would release a stale pass. The driver signs off on the release policy in Human decision."
decision-trigger: start
sources:
  - "medici-finance/assay#2026 — witnesses refused as wrong-revision: cause 1 squash-orphaned witness commits, cause 2 the whole-tree scope when no dependency manifest exists"
  - "medici-finance/assay#3105 — the stuck verified-to-done report (twelve briefs at filing)"
  - "medici-finance/assay#3658 — 95 of 379 verify-outcome record shas unreachable from main after squash merges; options content binding, merge commits, classification"
  - "medici-finance/assay#2304 — the earlier fix: the landed-witness search and the Verify-row command parser (coverage_repo.go witnessLanded, coverage_inputs.go)"
  - "medici-finance/assay#3429 — merged 2026-10-19; the shared `files:` grammar in docs/declared-input-scope.md"
  - "docs/contracts.md — S-evidence-coverage (owner: statusgen/evaluation, EvaluateCoverage)"
  - "docs/streams/graph-execution/work-input-amendment.md — WI-2: conservative invalidation first; file non-overlap alone is insufficient"
  - "docs/streams/decisions/DR-desk-tools-45.md — additive, versioned output fields; a missing or unknown version is a gap"
  - "`statusgen --root . --auto-flip-model --dry-run` at 75cc11f14: 30 REFUSED, 0 flipped, 267 wrong-revision claims"
  - "freshness-checked 2026-10-20 @ 75cc11f1423301ea8ba0b111d48c8d5b8a303537 (origin/main)"
exec-tier: strong
exec-tier-why: "It changes what releases a verification gate. Negative controls (a changed byte, a missing digest, an unknown schema, an old reader) must each be shown to refuse at more than one site, and the v1 rule must be proven byte-for-byte unchanged."
domain: complicated
consumers:
  - "statusgen/evaluation/ (coverage_api.go, coverage.go, coverage_inputs.go, declared_api.go, coverage_repo.go; new witness.go): follow-up statusgen/36"
  - "statusgen/coverage.go, statusgen/coverage_repo.go, statusgen/coverage_repomemo.go (adapter: batched per-revision listing, one dependency source for all call sites): follow-up statusgen/36"
  - "statusgen/autoflip.go and statusgen/reconcilefold.go (the two evaluateCoverage call sites that pass no Dependencies today): follow-up statusgen/36"
  - "statusgen/verifyrun.go (reader side only: witnessCells and witnessTreeOf hand the cell to the library; the writer is statusgen/37): follow-up statusgen/36"
  - "statusgen/main.go (the `witnessdigest --json` verb and the versioned witness field on `--coverage --json`): follow-up statusgen/36"
  - "statusgen readers of the Runner cell or witness state — attribution.go, citationcorroborate.go, evidenceactor.go, principal.go, verifyflip.go, verifyissues.go, verifyclosure.go, verifyoutcomes.go, outcomessplit.go, awaiting_bucket.go, evidenceshape.go, protectedclosure.go, failfirst.go, acceptancetransition.go, lifecycle.go, load.go, verifysandbox.go, verifiedrunneragree.go: follow-up statusgen/36 (each gets a disposition in the reader census)"
  - "docs/contracts.md (S-evidence-coverage row and a new R-witness-schema row): follow-up statusgen/36"
  - "spec/brief-v1.md, docs/brief-template.md and docs/declared-input-scope.md (the `inputs:` Context field): follow-up statusgen/36"
  - "tools/desk readers and writers: follow-up statusgen/37"
---

# Brief 36 — Content-bound witness schema v2: the reader and the one digest owner

## Context

files:
- `statusgen/evaluation/coverage_api.go`
- `statusgen/evaluation/coverage.go`
- `statusgen/evaluation/coverage_inputs.go`
- `statusgen/evaluation/coverage_repo.go`
- `statusgen/evaluation/declared_api.go`
- statusgen/evaluation/witness.go (planned)
- statusgen/evaluation/witness_test.go (planned)
- `statusgen/coverage.go`
- `statusgen/coverage_repo.go`
- `statusgen/coverage_repomemo.go`
- `statusgen/autoflip.go`
- `statusgen/reconcilefold.go`
- `statusgen/verifyrun.go`
- `statusgen/main.go`
- statusgen/testdata/witnessv2/ (planned)
- statusgen/witnessv2-mutations.json (planned)
- docs/witness-readers.md (planned)
- `docs/contracts.md`
- `docs/declared-input-scope.md`
- `docs/brief-template.md`
- `spec/brief-v1.md`
- changelog/statusgen-witness-v2-reader.md (planned)

facts (source and date in frontmatter):
- The class. A witness is judged against the item's revision through commit history: the witness commit must be in the object store, share an ancestor with the item, and (since medici-finance/assay#2304) have a landed twin whose tree matches outside `docs/streams/**` and `STATUS.md`. Every instance of the defect is a witness whose inputs are unchanged but whose commit is unreachable (squash merge) or whose scope is wider than its inputs (no manifest, so every path outside bookkeeping counts). #2304 closed instances (a landing search, a command-text parser) and left the class open.
- At 75cc11f14, `statusgen --root . --auto-flip-model --dry-run` refuses 30 verified briefs with 0 flipped, all on coverage, with 267 wrong-revision claims. The first differing paths are bookkeeping-adjacent files no Verify row reads: `.assay-surfaces` (103 claims), `.cursor/assay-managed.json` (78), `.claude-plugin/marketplace.json` (77), `.gitattributes` (7), plus one SKILL.md and `docs/contracts.md`. Four claims, all in build-less-brittle 04, 05, 06 and 09, are real: the Verify row's command changed after the witness.
- The worker design note on #2026 buckets the refused rows at an earlier main as 76 underivable (the command parser cannot read the command, for example a `$` in a `-run` pattern), 48 `+dirty`, 6 `.gitattributes` and 4 command-edited.
- `evaluateCoverage` has three production call sites (`autoflip.go:1077`, `reconcilefold.go:138`, `coverage.go:323`). None passes `Dependencies`, so `DependencyManifest`/`WorkInputDep` (which already carry `Recorded` and `Current` per input) are never populated.
- #3658: 95 of 379 verify-outcome records name a `sha` unreachable from main because of squash merges. The records already carry an `inputs` map (sha256 of the brief file and the tool tag) and a `prs` list.
- The `files:` grammar from #3429 (`docs/declared-input-scope.md`) lives in `tools/desk/internal/deskkit`, an internal package of another Go module, so statusgen cannot import it. It parses the implementation-touch set, not a row's read set, and refuses 302 of 485 briefs. Its grammar is reusable as a specification; its parser is not a source for this brief.
- A Runner cell is `<runner> @ <commit-short>[+dirty|+unknown]`. Today's reader lifts the token after `@` with `@\s*([^\s(]+)` and resolves it as a git object; a token that is not a valid object name cannot release.
- The existing `--coverage --json` output is the cross-module seam tools/desk already reads; DR-desk-tools-45 governs it: fields are additive and versioned, and a missing or unknown version is a gap.

layering: the pure library `statusgen/evaluation` owns the v2 schema, digest computation and the v1/v2 judgment, reading git only through its `Repo` port. The statusgen adapter supplies one batched, memoized listing per revision. Cross-module consumers read the result through `statusgen --coverage --json` or `statusgen witnessdigest --json`, never by importing statusgen or re-parsing cells. No new module, service or daemon.

single-point-of-failure: the brief's `inputs:` declaration decides which content a v2 digest covers; an under-declared input could release a stale pass. Behind it: (1) implicit inputs added mechanically by the library (module and lock files on each input's path, resolved git attributes, the tool pin file), not by the author; (2) a whole-tree digest when `inputs:` is absent or fails to parse; (3) the environment rule, under which a row that can reach the network, forge or a cluster is never released by digest; (4) the human gate on this brief's release policy. An observed-read check (recording what a row actually opened, inside the hermetic runner) is the stronger independent layer and is named as follow-up work, not claimed here.

design-fit:
  owner: statusgen/evaluation (S-evidence-coverage)
  contract: "R-witness-schema, a new contract row owned by S-evidence-coverage: witness schema v2 (change-request id, per-input blob-id map, implicit inputs, tool and runner identity) judged by digest equality; schema v1 judged by the unchanged ancestry rule; any other schema a named coverage gap"
  retires: ["the #2304 command-text parser as a staleness authority for v2 witnesses", "`+dirty` special-casing for v2 witnesses (the digest is of what was read)", "commit-ancestry, object-presence and landed-twin checks for v2 witnesses (the commit becomes informational)"]
  weight: "verbs +1 (`witnessdigest`, JSON only); flags +0; one versioned field on `--coverage --json`; one Context field (`inputs:`); one contract row; no new module, service or daemon"
  why-add: "Ancestry cannot survive squash merges and cannot narrow scope; content equality does both, and the existing DependencyManifest already models recorded-versus-current inputs, so the addition is the schema and the one digest owner, not a second judge."
  reuse-check: "Reuses DependencyManifest/WorkInputDep (graph-execution WI-2 owner; do not fork), the memoRepo pattern in coverage_repomemo.go, the existing `--coverage --json` seam and the `files:` grammar as a written specification. Does not reuse the deskkit parser (wrong module, wrong set)."

## Human decision

A v2 witness records a digest of the files a Verify row read and the change request that delivered it. Coverage releases the row when every recorded digest still equals the content at the item's revision. A digest cannot capture a row that depends on the network, a code-hosting service or a cluster, so those rows are never released by digest and are reported as needing a re-run. The open question is which runs count as environment-free.

Options:
1. Recommended: a run counts as environment-free when it ran in the pinned, network-less container runner, or when it ran on a host (or was back-filled) and its row is neither marked environment-dependent by its author nor names a network tool in a form the library can read. A host row whose command cannot be read is reported as needing a re-run.
2. Strictest: only runs in the pinned, network-less container runner count. Host runs and back-filled records are recorded for audit but always reported as needing a re-run.
3. Widest (not recommended): any run not marked environment-dependent by its author counts, including commands the library cannot read.

Default if no answer: option 2. No wider release is inferred.

## Read first

Read `EvaluateCoverage`, `CoverageInput`, `DependencyManifest` and `WorkInputDep` in `statusgen/evaluation/coverage_api.go`; `classifyRevisionDetail`, `witnessScope`, `newWitnessScope` and `depFingerprintSame` in `statusgen/evaluation/coverage.go`; `witnessLanded` in `statusgen/evaluation/coverage_repo.go`; `forVerifyRow` in `statusgen/evaluation/coverage_inputs.go`; `memoRepo` in `statusgen/coverage_repomemo.go`; `observeBrief` in `statusgen/coverage.go`; `witnessCells`, `witnessTreeOf` and `treeSHA` in `statusgen/verifyrun.go`; and `docs/declared-input-scope.md`. Re-run `statusgen --root . --auto-flip-model --dry-run` on fresh main. If the refused set no longer matches Context, report NEEDS_CONTEXT.

## Ground rules

- Feature branch and draft PR only. Never push to main, merge, release or dispatch a workflow. Merge main in, never rebase.
- Stop at `implemented`. A different, non-implementing identity owns verification and Evidence.
- Fixture repositories only; no network in tests.
- Schema v1 is never loosened. Every existing witness is judged exactly as today. A change that releases any v1 witness that main refuses is a defect.
- Do not edit any Evidence row, verify-outcome record or lifecycle cell. Records are append-only.
- If anything is unclear or contradicts repo state, report NEEDS_CONTEXT. Do not guess.

## Interface contract

**Schema marker.** A v2 Runner cell is `<runner-class> @ [w2] cr:<id> rec:<record-id-16>`. The bracketed marker is not a valid git object name or ref, so a reader that predates v2 lifts `[w2]`, fails to resolve it, and refuses (the version-skew rule). A cell with any other `[w<N>]` marker is an unknown schema. A cell with no bracketed marker is v1.

**Witness record (v2).** One JSON file per run at `docs/streams/witnesses/<stream>/<NN>/<record-id>.json`, keys sorted, written by statusgen/37 and parsed here:

- `schema: 2`, `brief`, `row`, `command_sha256` (SHA-256 of the Verify row's Command cell text),
- `cr` (the forge's change-request id that delivers the witness, or `pending`),
- `runner_class` (for example `host`, `harness-container`, `k8s-pod`), `network` (`none` or `available`),
- `tools` (statusgen and desk-tools release, Go version, harness image digest when containerised),
- `scope` (`declared` or `whole-tree`),
- `inputs` (repo-relative, slash-separated path → git blob id; an absent path recorded as `absent`),
- `implicit` (the same shape, for implicit inputs), `attributes` (path → SHA-256 of the resolved attribute set), and `tree_digest` for whole-tree scope,
- `informational` (`commit`, `dirty`); never read by the judgment.

`record-id` is the SHA-256 of the canonical record with `cr` removed, so binding the change request later does not change it. The reader recomputes it and refuses on mismatch.

**Digest.** A git blob id of the normalised content: for a committed revision, the id from `git ls-tree -r`; for a working tree, `git hash-object --stdin-paths` (which applies the repository's clean filters and end-of-line attributes). CRLF checkouts and Windows path separators therefore produce the same digest. Paths are repo-relative and slash-separated.

**Declared inputs.** A new `## Context` field `inputs:` lists the files and directories the Verify rows read, in the `files:` grammar of `docs/declared-input-scope.md`, reimplemented in `statusgen/evaluation` against the same fixtures. Paths under `docs/streams/**` and `STATUS.md` stay bookkeeping exactly as today and are a lint PROBLEM inside `inputs:`. An absent or unparseable `inputs:` means scope `whole-tree`.

**Implicit inputs**, derived by the library, never declared: for every input path, each `go.mod`, `go.sum`, `go.work`, `go.work.sum`, `package.json`, `package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`, `Cargo.toml`, `Cargo.lock`, `pyproject.toml`, `uv.lock`, `poetry.lock` and `requirements*.txt` on the path from that input to the root; the resolved git attributes of each input (`git check-attr -a --source=<rev>`); and `plugins/assay/paired-versions.yaml` (the tool and harness pins).

**Judgment of a v2 claim at revision R** (all must hold, else the named gap):
1. The record exists, parses, and its id matches the cell (`witness-record-missing`, `witness-record-mismatch`).
2. `cr` is set and not `pending` (`witness-cr-unbound`).
3. `command_sha256` equals the current Command cell (`witness-command-changed`).
4. The environment rule per Human decision (`witness-rerun-owed`); a row whose Class carries `+env` is always `witness-rerun-owed`.
5. The required set at R (declared plus implicit, or whole-tree) is a subset of the recorded keys (`witness-input-undeclared-at-record`).
6. Every required path's blob id at R equals the recorded one, `absent` included; every attribute digest matches; the whole-tree digest matches for whole-tree scope (`witness-input-changed: <first path>`).
7. The recorded tools agree with the pins in `plugins/assay/paired-versions.yaml` at R: the harness image digest for a containerised run; the statusgen and desk-tools release for any run (`witness-tool-changed`). A tool with no pin at R is `witness-rerun-owed`.

The commit, the dirty flag, ancestry and the landed-twin search are not consulted for v2. The v2 judgment populates `DependencyManifest{Complete: true}` with one `WorkInputDep` per input (kind `source`), implicit input (`build`) and tool (`environment`), and `EvaluateCoverage`'s existing `depFingerprintSame` decides. No second judge.

**Seams.** `statusgen --coverage --json` gains, per claim, `witness: {schema: 1|2, judgment, gap, cr}` with a top-level `witness_schema_version: 1`. `statusgen witnessdigest --json --brief <path> [--rev <rev> | --worktree]` prints the canonical v2 record minus run fields, computed by the library. Cross-module consumers use only these.

**Performance budget.** One `git ls-tree -r` and one `git check-attr --stdin` per distinct revision per run, memoized in the existing `memoRepo`; no per-row git process. `--auto-flip-model --dry-run` on main's corpus stays within 1.10× the wall-clock median of current main's binary.

## Task

1. Add `witness.go` to `statusgen/evaluation`: the cell parser (v1, v2, unknown), the record parser and id check, the `inputs:` parser, the implicit-input derivation, the digest computation over the `Repo` port, and the v2 judgment mapped onto `DependencyManifest`. Extend the `Repo` port with batched `ListTree(rev)`, `CheckAttr(rev, paths)` and `HashWorktree(paths)`; implement them in `statusgen/coverage_repo.go` and memoize them in `coverage_repomemo.go`.
2. Route the three `evaluateCoverage` call sites through one function that builds `CoverageInput.Dependencies` from the library, so no call site passes an empty manifest by omission.
3. Move Runner-cell parsing out of `verifyrun.go` into the library; `witnessTreeOf` becomes a thin call. v1 judgment is unchanged, including `witnessLanded` and `forVerifyRow`.
4. Add the `witnessdigest` verb and the versioned `witness` field on `--coverage --json`.
5. Write `docs/witness-readers.md` (planned): every non-test file in statusgen and tools/desk that reads a Runner cell, a witness record or a verify-outcome record, each with a disposition (judges through the library, reads informational fields only, or migrates in statusgen/37). `TestWitnessReaderCensus` (planned) fails when a file matches the census grep but is not listed.
6. Add the `R-witness-schema` contract row under S-evidence-coverage in `docs/contracts.md`, the `inputs:` field to `spec/brief-v1.md` and `docs/brief-template.md`, and a pointer from `docs/declared-input-scope.md`.
7. Build `statusgen/testdata/witnessv2/` fixture repositories with a squash-merged witness, a one-byte change, a missing record, a mismatched record id, a `pending` change request, an unknown `[w3]` schema, a `go.sum` change, a pin change, a CRLF checkout, an attribute change on an input, an `+env` row and a v1 witness that main refuses.
8. Write `statusgen/witnessv2-mutations.json` for `tools/desk/cmd/muhar`: a failing control, and mutants that skip the record-id check, accept `pending`, drop the implicit set, skip the attribute digest, treat an unknown schema as v1, skip the tool pin check, and release an `+env` row. Each must be caught by a named test.
9. Add changelog/statusgen-witness-v2-reader.md. Replace each `follow-up statusgen/36` entry in `consumers:` with `fixed-here` in the same change.

## Verify

Named tests, fixtures, the census and the mutation manifest are planned deliverables. A missing selector is a failure.

| # | Class | Command | Expect |
|---|---|---|---|
| 1 | check:ci +mutation +flow | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r1.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestWitnessV2DigestRelease$' ./evaluation/ > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessV2DigestRelease ' "$out")` | exit 0; named PASS. Negative control for cause 1: a witness whose commit was squash-merged (absent from the object store and not an ancestor of the item) with unchanged inputs releases. A whole-tree-scope witness releases only when the whole-tree digest is unchanged. |
| 2 | check:ci +mutation | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r2.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestWitnessV2ByteChange$' ./evaluation/ > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessV2ByteChange ' "$out")` | exit 0; named PASS. One changed byte in a declared input refuses with `witness-input-changed` naming the path; so does a deleted input and a newly created file at a path recorded `absent`. |
| 3 | check:ci +mutation | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r3.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestWitnessV2MissingDigest$' ./evaluation/ > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessV2MissingDigest ' "$out")` | exit 0; named PASS. A missing record, a record with an input whose digest is absent or malformed, a record whose id does not match the cell, and a `pending` change request each refuse with their named gap. |
| 4 | check:ci +mutation | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r4.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestWitnessSchemaUnknown$' ./evaluation/ > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessSchemaUnknown ' "$out")` | exit 0; named PASS. `[w3]`, `[w]` and `[w2x]` cells are named coverage gaps, never read as v1 or v2. |
| 5 | check +mutation +flow | `d=; trap 'test -n "$d" && git worktree remove --force "$d/old" 2>/dev/null; test -n "$d" && rm -rf "$d"' EXIT; d=$(mktemp -d "${TMPDIR:-/tmp}/sg36-skew.XXXXXX") && git worktree add --detach "$d/old" v1.0.52 && (cd "$d/old/statusgen" && GOWORK=off go build -o "$d/sg-old" .) && (cd statusgen && SG_OLD_BIN="$d/sg-old" GOWORK=off go test -count=1 -v -run '^TestWitnessSkewOldReader$' . > "$d/out" 2>&1) && grep -F -- '--- PASS: TestWitnessSkewOldReader ' "$d/out"` | exit 0; named PASS. A fixture brief that the new reader releases on a v2 witness is REFUSED by the release before this brief (v1.0.52 at authoring; the implementer re-pins to the latest tag that does not contain this brief's merge). The test fails, never skips, when `SG_OLD_BIN` is unset. |
| 6 | check:ci +mutation | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r6.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestWitnessV1Unchanged$' . > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessV1Unchanged ' "$out")` | exit 0; named PASS. Over a frozen copy of main's corpus at the authoring base, every v1 claim's verdict and reason string equal those of the base binary, and the v1 witness that main refuses is still refused. |
| 7 | check:ci +mutation | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r7.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestWitnessImplicitInputs$' ./evaluation/ > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessImplicitInputs ' "$out")` | exit 0; named PASS. With every declared input unchanged, a `go.sum` change on an input's module path refuses, a changed harness digest or statusgen tag in the pin file refuses, and a changed `eol` attribute on an input refuses. A `go.sum` change in an unrelated module does not. |
| 8 | check:ci +mutation | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r8.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestWitnessEnvRerunOwed$' ./evaluation/ > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessEnvRerunOwed ' "$out")` | exit 0; named PASS. An `+env` row, a host row naming a network tool, a host row whose command cannot be read, and a recorded tool with no pin at R each report `witness-rerun-owed` with unchanged inputs, under the policy the Human decision selects. |
| 9 | check:ci | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r9.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestWitnessDigestPortable$' ./evaluation/ > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessDigestPortable ' "$out")` | exit 0; named PASS. The same content checked out with LF and with CRLF (`core.autocrlf=true`, `eol=crlf` attributes) yields the same digest from the committed listing and from the working tree; a path given with a backslash is refused, not normalised silently. |
| 10 | check:ci +mutation | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r10.XXXXXX") && GOWORK=off go test -count=1 -v -run '^(TestWitnessDigestOneOwner\|TestWitnessOwnerFlagsPlant)$' . > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessDigestOneOwner ' "$out" && grep -F -- '--- PASS: TestWitnessOwnerFlagsPlant ' "$out")` | exit 0; both PASS. Class guard (fail-first): no non-test Go file in statusgen or tools/desk outside `statusgen/evaluation` parses a Runner cell's `@` token, calls `ls-tree`, `hash-object` or `check-attr` for witness purposes, or reads a record's `inputs` map, except entries on a committed allow-list with a reason. The plant fixture (a second parser outside the library) is flagged. The Evidence records the pre-fix red run with a planted site. |
| 11 | check:ci +mutation | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r11.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestCoverageGitCallsBounded$' . > "$out" 2>&1 && grep -F -- '--- PASS: TestCoverageGitCallsBounded ' "$out")` | exit 0; named PASS. Over a fixture of 20 briefs × 10 rows at 3 revisions, git invocations through the `Repo` port are at most 2 per distinct revision plus a constant, independent of row count. |
| 12 | check | `d=; trap 'test -n "$d" && git worktree remove --force "$d/base" 2>/dev/null; test -n "$d" && rm -rf "$d"' EXIT; d=$(mktemp -d "${TMPDIR:-/tmp}/sg36-perf.XXXXXX") && git worktree add --detach "$d/base" "$(git merge-base refs/remotes/origin/main HEAD)" && (cd "$d/base/statusgen" && GOWORK=off go build -o "$d/sg-base" .) && (cd statusgen && GOWORK=off go build -o "$d/sg-new" .) && for b in base new; do for i in 1 2 3 4 5; do s=$(perl -MTime::HiRes=time -e 'print time'); "$d/sg-$b" --root . --auto-flip-model --dry-run > /dev/null 2>&1; perl -MTime::HiRes=time -e "print time-$s, qq(\n)" >> "$d/$b"; done; done; mb=$(sort -n "$d/base" \| sed -n 3p); mn=$(sort -n "$d/new" \| sed -n 3p); echo "base=$mb new=$mn"; perl -e "exit(($mn <= 1.10*$mb) ? 0 : 1)"` | exit 0; prints both medians. The new binary's median wall-clock over five dry runs on this tree is at most 1.10× the merge-base binary's. |
| 13 | check:ci +dereference | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r13.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestWitnessReaderCensus$' . > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessReaderCensus ' "$out")` | exit 0; named PASS. Every non-test file in statusgen and tools/desk that matches the census grep (Runner-cell parse, witness record, verify-outcome record fields) is listed in `docs/witness-readers.md` (planned) with a disposition, and every listed file exists. |
| 14 | check +mutation | `d=; trap 'test -n "$d" && rm -rf "$d"' EXIT; d=$(mktemp -d "${TMPDIR:-/tmp}/sg36-muhar.XXXXXX") && (cd tools/desk && GOWORK=off go build -o "$d/muhar" ./cmd/muhar) && (cd statusgen && "$d/muhar" -spec witnessv2-mutations.json -j 4 > "$d/out" 2>&1); rc=$?; tail -1 "$d/out"; test "$rc" -eq 0 && grep -qE '^Totals: [1-9][0-9]* caught, 0 NOT CAUGHT, 0 could-not-mutate' "$d/out"` | exit 0. The failing control is caught and every mutant is caught, with no survivors. |
| 15 | check | `d=; trap 'test -n "$d" && rm -rf "$d"' EXIT; d=$(mktemp -d "${TMPDIR:-/tmp}/sg36-lint.XXXXXX") && (cd statusgen && GOWORK=off go build -o "$d/sg" .) && "$d/sg" --root . --lint > "$d/out" 2>&1; rc=$?; tail -1 "$d/out"; exit $rc` | exit 0; `LINT: PASS`. The `inputs:` field parses in the brief template and in every fixture brief. |
| 16 | check +dereference | `statusgen --root . --consumers --brief statusgen/36 --base "$(git merge-base refs/remotes/origin/main HEAD)"` | pre-merge only; exit 0. The implementation diff corroborates the routing in `consumers:`. At implementation replace this provisional command with the actual delivery and first-parent pins and exact summary counts, per `docs/consumer-delivery-checks.md`. |

## Pre-mortem and detection

- A squash-merged witness still refuses because some ancestry check survived on the v2 path. Caught by row 1.
- A changed input releases (scope too narrow, digest computed at the wrong revision). Caught by row 2 and by the mutants in row 14.
- A missing or tampered record releases, or `pending` is treated as bound. Caught by row 3.
- A future schema is read as v1 or v2. Caught by row 4.
- An older reader releases a v2 witness during a version skew. Caught by row 5, against a real prior release binary.
- The change loosens v1 by accident while sharing code. Caught by row 6's frozen-corpus comparison.
- A dependency or tool bump passes a stale witness. Caught by row 7.
- A network-dependent row is released by digest. Caught by row 8, within the selected policy. An unmarked host row whose network use the library cannot see remains possible under option 1; that is the stated residual risk, and option 2 closes it.
- CRLF or Windows paths change the digest, so witnesses refuse on another platform. Caught by row 9.
- A second parser or digest site appears outside the library. Caught by row 10's allow-list and plant.
- The board slows down with per-row git calls. Caught by rows 11 and 12.
- A reader of cells or records is missed and keeps its own interpretation. Caught by row 13's census.
- An author under-declares `inputs:` and a stale pass is released on a file the row read but did not declare. Not caught by any row here: the implicit set and the whole-tree fallback narrow it, review checks the declaration, and an observed-read layer is the follow-up that would catch it.
- Rows whose subject is `docs/streams/**` content are judged with bookkeeping excluded, as today. Carried-over limit; review only.

## Evidence

<!-- Independent verifier records the subject revision, environment, command, exit, named result and date. This authoring change executes no implementation check. -->

## Review

Gate: human. The reviewer answers, in the verdict, which single control stands between an under-declared input and a stale release, and whether any Verify row proves a lower layer catches a fault with the upper one bypassed (rows 7 and 8 are the candidates). Reject any of these:
- a v1 witness that main refuses being released;
- a v2 release that consults commit ancestry or the dirty flag;
- a missing digest, an unbound change request or an unknown schema that does not refuse;
- a digest computed outside `statusgen/evaluation`;
- an environment-dependent row released by digest.
