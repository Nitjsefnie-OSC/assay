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
  - "statusgen/autoflip.go and statusgen/reconcilefold.go (two of the three evaluateCoverage call sites, none of which passes Dependencies today; the third, coverage.go, is routed on the line above): follow-up statusgen/36"
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
- statusgen/witnessv2_test.go (planned)
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
- A Runner cell is `<runner> @ <commit-short>[+dirty|+unknown]`, optionally followed by an on-behalf-of and a runner-source parenthetical (`statusgen/verifyrun.go`, the cell writer). Today's reader lifts the token after `@` with `@\s*([^\s(]+)` and resolves it as a git object; a token that is not a valid object name cannot release.
- v1 guards that v2 must keep. Coverage refuses a witness when the Verify row's Expect text at the witness commit differs from the current Expect (`statusgen/evaluation/coverage.go`, TestCoverageExpectChangedSinceWitnessRan). The identity reads take the token before `@`: `implementerAttributed` and `evidenceHasIndependentRow` (`statusgen/attribution.go`), the verifier floor (`belowFloorRunner`, `humanRunnerName`, same file) and `runnerKey` (`statusgen/verifiedrunneragree.go`, which drops everything from the first parenthesis). The `+dirty` tolerance is documented in `coverage.go` as a witness-trust gap that only a witness-format change can close.
- Evidence authorship is checked by `statusgen/evidenceactor.go`, which describes itself as tamper-evident, not tamper-proof.
- The existing `--coverage --json` output is the cross-module seam tools/desk already reads; DR-desk-tools-45 governs it: fields are additive and versioned, and a missing or unknown version is a gap.

layering: the pure library `statusgen/evaluation` owns the v2 schema, digest computation and the v1/v2 judgment, reading git only through its `Repo` port. The statusgen adapter supplies one batched, memoized listing per revision. Cross-module consumers read the result through `statusgen --coverage --json` or `statusgen witnessdigest --json`, never by importing statusgen or re-parsing cells. No new module, service or daemon.

trust-assumption: v2 detects staleness, not forgery. The writer computes every digest and self-reports `runner_class`, `network`, `tools` and the outcome, so anyone who can write a record the reader parses can write one it accepts. A record's authenticity rests on the same controls as a v1 cell today: who authored the Evidence change (`evidenceactor.go`, tamper-evident, not tamper-proof), review of the Evidence change request, and the change request's verified landing (judgment step 2). v2 therefore must not weaken any v1 identity, acceptance-text or dirty-tree check, and binds the executing identity, the acceptance text and the outcome inside the record id so that a record cannot be replayed onto another row or outcome.

single-point-of-failure: the author's `inputs:` declaration, checked only by human review of the brief, is the one control against an under-declared source file: a file the row reads but the declaration omits is not hashed, so a change to it cannot refuse the witness. Behind it for that fault: NONE in this split. The other layers cover different faults and do not catch this one: implicit inputs cover module and lock files, attributes and tool pins; the whole-tree digest applies only when `inputs:` is absent or unparseable; the environment rule covers network, forge and cluster rows. The mechanical lower bound below (every path the Command cell names literally must be declared) catches only the omission of a path the command itself names, not of a file it reads transitively. A second layer for transitive reads is not feasible in this brief: a shell command's read set cannot be derived statically (the #2304 command parser could not read 76 rows), and the alternative, recording the files a row actually opened, needs the network-less container runner that statusgen/37 builds plus a file-access tracer on every supported host. That observed-read check is the named follow-up. Until it exists, Human decision option 2 keeps host and back-filled runs from releasing, which narrows who can produce a releasable record but does not detect under-declaration.

design-fit:
  owner: statusgen/evaluation (S-evidence-coverage)
  contract: "R-witness-schema, a new contract row owned by S-evidence-coverage: witness schema v2 (verified change-request id, executing identity, acceptance-text digests, outcome, dirty paths, per-input blob-id map, implicit inputs, tool and runner class) judged by digest equality plus the v1 identity, acceptance-text and dirty-tree guards; schema v1 judged by the unchanged ancestry rule; any other schema a named coverage gap"
  retires: ["the #2304 command-text parser as a staleness authority for v2 witnesses", "the `+dirty` base-commit tolerance for v2 witnesses (the digest is of what was read, and a dirty path outside the required set refuses)", "commit-ancestry, object-presence and landed-twin checks for v2 witnesses (the commit becomes informational)"]
  weight: "verbs +1 (`witnessdigest`, JSON only); flags +0; one versioned field on `--coverage --json`; one Context field (`inputs:`); one lint check (literal command paths inside `inputs:`); one contract row; no new module, service or daemon"
  why-add: "Ancestry cannot survive squash merges and cannot narrow scope; content equality does both, and the existing DependencyManifest already models recorded-versus-current inputs, so the addition is the schema and the one digest owner, not a second judge."
  reuse-check: "Reuses DependencyManifest/WorkInputDep (graph-execution WI-2 owner; do not fork), the memoRepo pattern in coverage_repomemo.go, the existing `--coverage --json` seam and the `files:` grammar as a written specification. Does not reuse the deskkit parser (wrong module, wrong set)."

## Human decision

A v2 witness records a digest of the files a Verify row read and the change request that delivered it. Coverage releases the row when every recorded digest still equals the content at the item's revision. A digest cannot capture a row that depends on the network, a code-hosting service or a cluster, so those rows are never released by digest and are reported as needing a re-run. The open question is which runs count as environment-free.

Options:
1. Recommended: a run counts as environment-free when it ran in the pinned, network-less container runner, or when it ran on a host (or was back-filled) and its row is neither marked environment-dependent by its author nor names a network tool in a form the library can read. A host row whose command cannot be read is reported as needing a re-run.
2. Strictest: only runs in the pinned, network-less container runner count. Host runs and back-filled records are recorded for audit but always reported as needing a re-run.
3. Widest (not recommended): any run not marked environment-dependent by its author counts, including commands the library cannot read.

Default if no answer: option 2 after 2026-11-03.

Option 2 is the strictest choice, so no wider release is inferred while the question is open.

## Read first

Read `EvaluateCoverage`, `CoverageInput`, `DependencyManifest` and `WorkInputDep` in `statusgen/evaluation/coverage_api.go`; `classifyRevisionDetail`, `witnessScope`, `newWitnessScope` and `depFingerprintSame` in `statusgen/evaluation/coverage.go`; `witnessLanded` in `statusgen/evaluation/coverage_repo.go`; `forVerifyRow` in `statusgen/evaluation/coverage_inputs.go`; `memoRepo` in `statusgen/coverage_repomemo.go`; `observeBrief` in `statusgen/coverage.go`; `witnessCells`, `witnessTreeOf` and `treeSHA` in `statusgen/verifyrun.go`; and `docs/declared-input-scope.md`. Re-run `statusgen --root . --auto-flip-model --dry-run` on fresh main and record the refused set in the PR. The set is expected to drift as Evidence lands (some of the 30 briefs have received Evidence since 75cc11f14), so a smaller or different set is not a stop condition. Report NEEDS_CONTEXT only if no brief is refused on coverage any more, or if a refusal reason appears that is outside the causes in Context (squash-orphaned commit, whole-tree scope, `+dirty`, attributes, command changed).

## Ground rules

- Feature branch and draft PR only. Never push to main, merge, release or dispatch a workflow. Merge main in, never rebase.
- Stop at `implemented`. A different, non-implementing identity owns verification and Evidence.
- Fixture repositories only; no network in tests.
- Schema v1 is never loosened. Every existing witness is judged exactly as today. A change that releases any v1 witness that main refuses is a defect.
- Do not edit any Evidence row, verify-outcome record or lifecycle cell. Records are append-only.
- If anything is unclear or contradicts repo state, report NEEDS_CONTEXT. Do not guess.

## Interface contract

**Schema marker.** A v2 Runner cell is `<runner> @ [w2] cr:<id> rec:<record-id>`, followed by the same on-behalf-of and runner-source parentheticals a v1 cell carries. `<runner>` is the executing identity exactly as verifyrun writes it in a v1 cell today, so every identity read (`implementerAttributed`, `evidenceHasIndependentRow`, the verifier floor, `runnerKey`) keeps working unchanged. The runner class is recorded in the record, never in place of the identity. `<record-id>` is the full 64-hex id; there is no prefix resolution. The bracketed marker is not a valid git object name or ref, so a reader that predates v2 lifts `[w2]`, fails to resolve it, and refuses (the version-skew rule). A cell with any other `[w<N>]` marker is an unknown schema. A cell with no bracketed marker is v1.

**Witness record (v2).** One JSON file per run at `docs/streams/witnesses/<stream>/<NN>/<record-id>.json`, keys sorted, written by statusgen/37 and parsed here. Every field below is inside the record id:

- `schema: 2`, `brief` (the brief id), `row` (the Verify row id),
- `command_sha256`, `expect_sha256`, `class_sha256` (SHA-256 of the Verify row's Command, Expect and Class cell text at the revision the run read),
- `runner` and `runner_source` (the executing identity and its source, as written in the cell),
- `runner_class` (for example `host`, `harness-container`, `k8s-pod`), `network` (`none` or `available`), `tools` (statusgen and desk-tools release, Go version, harness image digest when containerised) and `tools_source` (`observed`, `inferred-from-pins`, or `none` when the tools are unknown),
- `origin` (`run`, or `backfill` for a record derived by statusgen/38); a back-filled record also carries `source_cr` (the merged change request whose commit list held the witnessed commit) and `source_cell` (the v1 cell text),
- `result` (the Result cell, which carries the exit status), `exit`, `output_sha256` (the output digest cell) and `date`,
- `cr` (the forge's change-request id that delivers this record, or `pending`),
- `scope` (`declared` or `whole-tree`),
- `inputs` (repo-relative, slash-separated path → git blob id; an absent path recorded as `absent`),
- `implicit` (the same shape, for implicit inputs), `attributes` (path → SHA-256 of the resolved attribute set), and `tree_digest` for whole-tree scope,
- `dirty_paths` (every non-bookkeeping path that was modified, deleted or untracked in the working tree when the row ran; empty for a clean tree),
- `informational` (`commit`, `dirty`); never read by the judgment, though inside the id.

`record-id` is the SHA-256 of the canonical record. Because `cr` is inside it, binding a `pending` record to its change request writes a new record and a new id (statusgen/37's `--bind-cr`, allowed only before the record reaches the default branch). The reader recomputes the id and refuses on mismatch.

**Digest.** A git blob id of the normalised content: for a committed revision, the id from `git ls-tree -r`; for a working tree, `git hash-object --stdin-paths` (which applies the repository's clean filters and end-of-line attributes). CRLF checkouts and Windows path separators therefore produce the same digest. Paths are repo-relative and slash-separated.

**Declared inputs.** A new `## Context` field `inputs:` lists the files and directories the Verify rows read, in the `files:` grammar of `docs/declared-input-scope.md`, reimplemented in `statusgen/evaluation` against the same fixtures. Paths under `docs/streams/**` and `STATUS.md` stay bookkeeping exactly as today and are a lint PROBLEM inside `inputs:`. An absent or unparseable `inputs:` means scope `whole-tree`.

**Literal-path lower bound.** Every repo-relative path that a Verify row's Command cell names literally and that the library can read (a `cd` target, a package path such as `./evaluation/`, a file argument) must lie inside the declared set. `--lint` reports a violation as a PROBLEM, and the reader refuses with `witness-input-undeclared`. This is a partial check: it catches an omitted path the command names, not a file the command reads transitively (see single-point-of-failure).

**Implicit inputs**, derived by the library, never declared: for every input path, each `go.mod`, `go.sum`, `go.work`, `go.work.sum`, `package.json`, `package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`, `Cargo.toml`, `Cargo.lock`, `pyproject.toml`, `uv.lock`, `poetry.lock` and `requirements*.txt` on the path from that input to the root; the resolved git attributes of each input (`git check-attr -a --source=<rev>`); and `plugins/assay/paired-versions.yaml` (the tool and harness pins).

**Judgment of a v2 claim at revision R** (all must hold, else the named gap):
1. The record exists, parses, and its id matches the cell (`witness-record-missing`, `witness-record-mismatch`).
2. `cr` is set, not `pending`, equal in the cell and the record, and verified: the first-parent commit on R's history that added the record names that change request in a recognised landing form, and its diff also changes the citing brief (`witness-cr-unbound`, `witness-cr-mismatch`, `witness-cr-unverified`). The recognised landing forms are one table in the library: the GitHub merge-commit subject, the GitHub squash subject suffix, and the GitLab merge-commit trailer. A landing in no recognised form (a fast-forward or rebase merge, for example) is `witness-cr-unverified` offline; with `--forge`, the forge seam may confirm that the change request is merged and that its landed commit is the adding commit. A commit on R's first-parent history is the offline proof that the change request merged.
3. The record's `brief` and `row` equal the citing brief and Verify row, and the citing witness row's Result, output-digest and Date cells equal the record's `result`, `output_sha256` and `date` (`witness-record-replayed`).
4. The record's `runner` equals the cell's identity token (`witness-runner-mismatch`).
5. `command_sha256`, `expect_sha256` and `class_sha256` equal the current Command, Expect and Class cells (`witness-command-changed`, `witness-expect-changed`, `witness-class-changed`). This keeps v1's Expect-changed refusal: a loosened Expect never releases an old pass.
6. The environment rule per Human decision (`witness-rerun-owed`); a row whose Class carries `+env` is always `witness-rerun-owed`.
7. The required set at R (declared plus implicit, or whole-tree) is a subset of the recorded keys (`witness-input-undeclared-at-record`), and the literal-path lower bound holds (`witness-input-undeclared`).
8. Every entry of `dirty_paths` lies inside the required set; otherwise `witness-dirty-outside-inputs`. A dirty edit to a recorded input is covered by its digest; a dirty edit anywhere else refuses.
9. Every required path's blob id at R equals the recorded one, `absent` included; every attribute digest matches; the whole-tree digest matches for whole-tree scope (`witness-input-changed: <first path>`).
10. The recorded tools agree with the pins in `plugins/assay/paired-versions.yaml` at R: the harness image digest for a containerised run; the statusgen and desk-tools release for any run (`witness-tool-changed`). A tool with no pin at R, or a record with `tools_source: none`, is `witness-rerun-owed`.

The commit, ancestry and the landed-twin search are not consulted for v2. The v2 judgment populates `DependencyManifest{Complete: true}` with one `WorkInputDep` per input (kind `source`), implicit input (`build`) and tool (`environment`), and `EvaluateCoverage`'s existing `depFingerprintSame` decides the digest comparison. Steps 1–8 and 10 are gates in front of it, not a second judge.

**Seams.** `statusgen --coverage --json` gains, per claim, `witness: {schema: 1|2, judgment, gap, cr, runner, independent, evidence_actor}` with a top-level `witness_schema_version: 1`. `runner` is the cell's identity token; `independent` is the verdict of the existing offline identity reads for that row (the independent-row read and the verifier floor); `evidence_actor` is the existing Evidence-actor classification of the commit that authored the citing Evidence line (`verifier`, `other` or `unknown`). These are reported from the reads statusgen already has, so no consumer parses a Runner cell or runs `git blame` itself. For a v1 claim the field also carries `landing_cr`: the change request named, per step 2's landing-form table, by the first-parent commit that landed the citing Evidence line (empty when the landing is in no recognised form). `statusgen witnessdigest --json --brief <path> [--rev <rev> | --worktree] [--row <id>]` prints the canonical v2 content fields (`inputs`, `implicit`, `attributes`, `tree_digest`, and `dirty_paths` with `--worktree`) and, with `--row`, that Verify row's `command_sha256`, `expect_sha256` and `class_sha256` at the same revision, computed by the library; it never prints a releasable record. Cross-module consumers use only these.

**Performance budget.** One `git ls-tree -r`, one `git check-attr --stdin` and one `git log --first-parent --diff-filter=A` over `docs/streams/witnesses/` per distinct revision per run, memoized in the existing `memoRepo`; no per-row git process. The offline `statusgen --root . --coverage` over the corpus stays within 1.10× the wall-clock median of the pinned release before this brief.

## Task

1. Add `witness.go` to `statusgen/evaluation`: the cell parser (v1, v2, unknown), the record parser and id check, the `inputs:` parser, the implicit-input derivation, the digest computation over the `Repo` port, the landing-form table and the step-2 change-request check, the identity, replay, acceptance-text and dirty-path gates, the literal-path lower bound, and the v2 judgment mapped onto `DependencyManifest`. Extend the `Repo` port with batched `ListTree(rev)`, `CheckAttr(rev, paths)` and `HashWorktree(paths)`; implement them in `statusgen/coverage_repo.go` and memoize them in `coverage_repomemo.go`.
2. Route the three `evaluateCoverage` call sites through one function that builds `CoverageInput.Dependencies` from the library, so no call site passes an empty manifest by omission.
3. Move Runner-cell parsing out of `verifyrun.go` into the library; `witnessTreeOf` becomes a thin call. v1 judgment is unchanged, including `witnessLanded`, `forVerifyRow` and the Expect-changed guard. The identity reads in `attribution.go` and `verifiedrunneragree.go` keep taking the token before `@` and need no change for v2 cells; row 16 pins that.
4. Add the `witnessdigest` verb and the versioned `witness` field on `--coverage --json`, and add the literal-path lower bound to `--lint` as a PROBLEM.
5. Write `docs/witness-readers.md` (planned): every non-test file in statusgen and tools/desk that reads a Runner cell, a witness record or a verify-outcome record, each with a disposition (judges through the library, reads informational fields only, or migrates in statusgen/37). `TestWitnessReaderCensus` (planned) fails when a file matches the census grep but is not listed.
6. Add the `R-witness-schema` contract row under S-evidence-coverage in `docs/contracts.md`, the `inputs:` field to `spec/brief-v1.md` and `docs/brief-template.md`, and a pointer from `docs/declared-input-scope.md`.
7. Build `statusgen/testdata/witnessv2/` fixture repositories with a squash-merged witness, a one-byte change, a missing record, a mismatched record id, a `pending` change request, an unknown `[w3]` schema, a `go.sum` change, a pin change, a CRLF checkout, an attribute change on an input, an `+env` row, a v1 witness that main refuses, a record cited from a different row and from a different brief, a citing row whose Result or Date differs from the record, a cell whose identity differs from the record's `runner`, a v2 cell written by the implementer identity, an Expect loosened after the run, a Class changed after the run, a dirty path outside the declared inputs, a `cr` that the landing commit does not name, a rebase-merged landing, and a command that names an undeclared path.
8. Write `statusgen/witnessv2-mutations.json` for `tools/desk/cmd/muhar`: a failing control, and mutants that skip the record-id check, accept `pending`, skip the landing-commit check on `cr`, skip the brief and row replay check, skip the outcome comparison, skip the runner match, skip the Expect digest, drop the implicit set, skip the attribute digest, skip the dirty-path gate, skip the literal-path lower bound, treat an unknown schema as v1, skip the tool pin check, and release an `+env` row; and, for each identity read (`implementerAttributed`, `evidenceHasIndependentRow`, `belowFloorRunner`, `humanRunnerName`, `runnerKey`), a mutant that takes the token after `@` instead of before it. Each must be caught by a named test.
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
| 6 | check +mutation | `d=; trap 'test -n "$d" && git worktree remove --force "$d/old" 2>/dev/null; test -n "$d" && rm -rf "$d"' EXIT; d=$(mktemp -d "${TMPDIR:-/tmp}/sg36-v1.XXXXXX") && git worktree add --detach "$d/old" v1.0.52 && (cd "$d/old/statusgen" && GOWORK=off go build -o "$d/sg-old" .) && (cd statusgen && SG_OLD_BIN="$d/sg-old" GOWORK=off go test -count=1 -v -run '^TestWitnessV1Unchanged$' . > "$d/out" 2>&1) && grep -F -- '--- PASS: TestWitnessV1Unchanged ' "$d/out"` | exit 0; named PASS. The baseline is the release binary built in this command (v1.0.52 at authoring; the implementer re-pins it to the latest tag that does not contain this brief's merge, as in row 5), never the changed code or a golden file the changed code produced. Over a frozen copy of main's corpus at the authoring base (`statusgen/testdata/witnessv2/v1corpus/`, planned), every v1 claim's verdict and reason string from `$SG_OLD_BIN --coverage --json` equal the new binary's, including the Expect-changed refusals, and the v1 witness that main refuses is still refused. The test fails, never skips, when `SG_OLD_BIN` is unset. |
| 7 | check:ci +mutation | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r7.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestWitnessImplicitInputs$' ./evaluation/ > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessImplicitInputs ' "$out")` | exit 0; named PASS. With every declared input unchanged, a `go.sum` change on an input's module path refuses, a changed harness digest or statusgen tag in the pin file refuses, and a changed `eol` attribute on an input refuses. A `go.sum` change in an unrelated module does not. |
| 8 | check:ci +mutation | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r8.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestWitnessEnvRerunOwed$' ./evaluation/ > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessEnvRerunOwed ' "$out")` | exit 0; named PASS. An `+env` row, a host row naming a network tool, a host row whose command cannot be read, and a recorded tool with no pin at R each report `witness-rerun-owed` with unchanged inputs, under the policy the Human decision selects. |
| 9 | check:ci | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r9.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestWitnessDigestPortable$' ./evaluation/ > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessDigestPortable ' "$out")` | exit 0; named PASS. The same content checked out with LF and with CRLF (`core.autocrlf=true`, `eol=crlf` attributes) yields the same digest from the committed listing and from the working tree; a path given with a backslash is refused, not normalised silently. |
| 10 | check:ci +mutation | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r10.XXXXXX") && GOWORK=off go test -count=1 -v -run '^(TestWitnessDigestOneOwner\|TestWitnessOwnerFlagsPlant)$' . > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessDigestOneOwner ' "$out" && grep -F -- '--- PASS: TestWitnessOwnerFlagsPlant ' "$out")` | exit 0; both PASS. Class guard (fail-first): no non-test Go file in statusgen or tools/desk outside `statusgen/evaluation` parses a Runner cell's `@` token, calls `ls-tree`, `hash-object` or `check-attr` for witness purposes, or reads a record's `inputs` map, except entries on a committed allow-list with a reason. The plant fixture (a second parser outside the library) is flagged. The Evidence records the pre-fix red run with a planted site. |
| 11 | check:ci +mutation | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r11.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestCoverageGitCallsBounded$' . > "$out" 2>&1 && grep -F -- '--- PASS: TestCoverageGitCallsBounded ' "$out")` | exit 0; named PASS. Over a fixture of 20 briefs × 10 rows at 3 revisions, git invocations through the `Repo` port are at most 3 per distinct revision plus a constant, independent of row count. |
| 12 | check | `d=; trap 'test -n "$d" && git worktree remove --force "$d/base" 2>/dev/null; test -n "$d" && rm -rf "$d"' EXIT; d=$(mktemp -d "${TMPDIR:-/tmp}/sg36-perf.XXXXXX") && git worktree add --detach "$d/base" v1.0.52 && (cd "$d/base/statusgen" && GOWORK=off go build -o "$d/sg-base" .) && (cd statusgen && GOWORK=off go build -o "$d/sg-new" .) && for b in base new; do for i in 1 2 3 4 5; do s=$(perl -MTime::HiRes=time -e 'print time'); "$d/sg-$b" --root . --coverage > /dev/null 2>&1 \|\| exit 3; perl -MTime::HiRes=time -e "print time-$s, qq(\n)" >> "$d/$b"; done; done; mb=$(sort -n "$d/base" \| sed -n 3p); mn=$(sort -n "$d/new" \| sed -n 3p); echo "base=$mb new=$mn"; perl -e "exit(($mn <= 1.10*$mb) ? 0 : 1)"` | exit 0; prints both medians. The base is a fixed release tag, not a merge-base: v1.0.52 at authoring, re-pinned by the implementer to the latest tag that does not contain this brief's merge and recorded in Evidence, so the row compares two different binaries before and after merge. `--coverage` is offline by construction, so the timing includes no forge or network latency. Every run of both binaries must exit 0 (exit 3 otherwise). The new binary's median wall-clock over five runs on this tree is at most 1.10× the base binary's. |
| 13 | check:ci +dereference | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r13.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestWitnessReaderCensus$' . > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessReaderCensus ' "$out")` | exit 0; named PASS. Every non-test file in statusgen and tools/desk that matches the census grep (Runner-cell parse, witness record, verify-outcome record fields) is listed in `docs/witness-readers.md` (planned) with a disposition, and every listed file exists. |
| 14 | check +mutation | `d=; trap 'test -n "$d" && rm -rf "$d"' EXIT; d=$(mktemp -d "${TMPDIR:-/tmp}/sg36-muhar.XXXXXX") && (cd tools/desk && GOWORK=off go build -o "$d/muhar" ./cmd/muhar) && (cd statusgen && "$d/muhar" -spec witnessv2-mutations.json -j 4 > "$d/out" 2>&1); rc=$?; tail -1 "$d/out"; test "$rc" -eq 0 && grep -qE '^Totals: [1-9][0-9]* caught, 0 NOT CAUGHT, 0 could-not-mutate' "$d/out"` | exit 0. The failing control is caught and every mutant is caught, with no survivors. |
| 15 | check | `d=; trap 'test -n "$d" && rm -rf "$d"' EXIT; d=$(mktemp -d "${TMPDIR:-/tmp}/sg36-lint.XXXXXX") && (cd statusgen && GOWORK=off go build -o "$d/sg" .) && "$d/sg" --root . --lint > "$d/out" 2>&1; rc=$?; tail -1 "$d/out"; exit $rc` | exit 0; `LINT: PASS`. The `inputs:` field parses in the brief template and in every fixture brief. |
| 16 | check:ci +mutation | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r16.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestWitnessV2AttributionUnchanged$' . > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessV2AttributionUnchanged ' "$out")` | exit 0; named PASS. For a table of identities (the implementer, the verifier App, a human on whose behalf an App ran, a below-floor runner), each identity read gives the same answer on a v2 cell as on the equivalent v1 cell: `implementerAttributed`, `evidenceHasIndependentRow`, `belowFloorRunner`, `humanRunnerName` and `runnerKey`. An implementer-run v2 row is never counted as independent. A v2 cell is never read as its runner class. Each read has a mutant in row 14 that takes the token after `@`, and each is caught. `--coverage --json` reports the same `runner` and `independent` values for the v1 and v2 cell of each identity, and `evidence_actor` matches the Evidence-actor check for the fixture's authoring commit. |
| 17 | check:ci +mutation | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r17.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestWitnessV2Replay$' ./evaluation/ > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessV2Replay ' "$out")` | exit 0; named PASS. Negative controls for replay: a valid record cited from another Verify row of the same brief, from another brief with the same Command and inputs, and from its own row with a different Result, output digest or Date each refuse with `witness-record-replayed`; a cell whose identity differs from the record's `runner` refuses with `witness-runner-mismatch`. The unaltered citation releases. |
| 18 | check:ci +mutation | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r18.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestWitnessV2AcceptanceChanged$' ./evaluation/ > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessV2AcceptanceChanged ' "$out")` | exit 0; named PASS. With every input unchanged, an Expect loosened after the run refuses with `witness-expect-changed`, a changed Class with `witness-class-changed`, and a changed Command with `witness-command-changed`. This is v1's Expect-changed guard, kept for v2. |
| 19 | check:ci +mutation +flow | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r19.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestWitnessV2CRVerified$' ./evaluation/ > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessV2CRVerified ' "$out")` | exit 0; named PASS. A record landed by a merge commit or a squash commit that names its `cr` (GitHub forms) and by a GitLab merge commit releases. Each of these refuses: a cell `cr` that differs from the record's, a landing commit naming a different change request, a landing commit whose diff does not change the citing brief, a rebase-merged landing with no recognised form (offline: `witness-cr-unverified`), and a record not yet on R's first-parent history. |
| 20 | check:ci +mutation | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r20.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestWitnessV2DirtyOutsideInputs$' ./evaluation/ > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessV2DirtyOutsideInputs ' "$out")` | exit 0; named PASS. A record whose `dirty_paths` holds a path outside the required set refuses with `witness-dirty-outside-inputs`, even when every input digest matches. A dirty edit to a declared input that later landed byte-identical releases; one that landed differently refuses with `witness-input-changed`. |
| 21 | check:ci +mutation | `(cd statusgen && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg36-r21.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestWitnessLiteralPathBound$' . > "$out" 2>&1 && grep -F -- '--- PASS: TestWitnessLiteralPathBound ' "$out")` | exit 0; named PASS. A Verify row whose Command names a package path or file outside `inputs:` is a `--lint` PROBLEM and refuses with `witness-input-undeclared`; the same row with the path declared passes lint and releases. A command the library cannot read adds no bound and is not itself a failure. |
| 22 | check +dereference | `statusgen --root . --consumers --brief statusgen/36 --base "$(git merge-base refs/remotes/origin/main HEAD)"` | pre-merge only; exit 0. The implementation diff corroborates the routing in `consumers:`. At implementation replace this provisional command with the actual delivery and first-parent pins and exact summary counts, per `docs/consumer-delivery-checks.md`. |

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
- The board slows down with per-row git calls. Caught by rows 11 and 12; row 12's base is a fixed release tag, so it still compares two binaries after merge.
- A v2 cell hides the executing identity, so an implementer-run row counts as independent. Caught by row 16 and its per-read mutants.
- A valid record is replayed onto another row, brief or outcome. Caught by row 17.
- An Expect is loosened after a pass and v2 releases it. Caught by row 18.
- `cr` is a label nobody checked. Caught by row 19.
- A dirty edit outside the declared inputs hides behind matching digests. Caught by row 20.
- A reader of cells or records is missed and keeps its own interpretation. Caught by row 13's census.
- An author under-declares `inputs:` and a stale pass is released on a file the row read but did not declare. Caught only when the command names the omitted path literally (row 21). A transitively read file is not caught by any row here: human review of the declaration is the only control, and the observed-read check is the follow-up that would catch it (see single-point-of-failure).
- A record is forged by someone who can write Evidence. Out of scope by the trust assumption: v2 detects staleness, not forgery; authorship and review of the Evidence change request are the controls, as for v1.
- Rows whose subject is `docs/streams/**` content are judged with bookkeeping excluded, as today. Carried-over limit; review only.

## Evidence

<!-- Independent verifier records the subject revision, environment, command, exit, named result and date. This authoring change executes no implementation check. -->

## Review

Gate: human. The reviewer answers, in the verdict, which single control stands between an under-declared input and a stale release (the declaration and its human review; NONE mechanical for transitive reads), and whether any Verify row proves a lower layer catches a fault with the upper one bypassed (rows 7, 8, 18 and 20 are the candidates, each for its own fault class). Reject any of these:
- a v1 witness that main refuses being released;
- a v2 release that consults commit ancestry;
- a v2 cell or record without the executing identity, or an identity read that gives a different answer for a v2 cell than for the equivalent v1 cell;
- a record released for a row, brief or outcome other than the one it records;
- a changed Expect, Class or Command that does not refuse;
- a `cr` released without the landing check;
- a dirty path outside the required set that does not refuse;
- a missing digest, an unbound change request or an unknown schema that does not refuse;
- a digest computed outside `statusgen/evaluation`;
- an environment-dependent row released by digest.
