---
brief: assay:assay:desk-tools:32
title: "golangci-lint joins the Go CI — a pinned binary, a curated green set (errcheck, staticcheck, unused, govet), and a human-applied workflow hunk"
why: >-
  The repo's deterministic Go correctness layer is `go vet` (cross-compiled in ci.yml's
  compile job) plus the test suites. Every other defect class a static analyzer catches for
  free — unchecked errors (errcheck), dead code (unused, whose U1000 is not staticcheck's),
  a long tail of correctness smells (staticcheck SA) — is caught by nobody, or only
  by the model review lanes, which have no completeness contract over machine-checkable
  classes. This is the `enforce: tool` tier done properly: cheap, deterministic, and pinned.
  The review lanes stay for what needs judgment; the machine gets the classes the machine
  proves. One leg is structurally human: GitHub blocks a role App without the `workflows`
  scope from pushing any `.github/workflows/` change, so the ci.yml hunk lands by a human
  runsheet the worker prepares — everything else (config, measurement, code fixes) is the
  worker's.
wave: 1
depends: []
unblocks: []
effort: M
gate: model
risk: {regulatory: no, customer: no, irreversible: no, sensitive-data: no}
issues: []
schema: brief-v2
authored: 2026-10-11 by a house authoring session at the driver's direction
sources:
  - ".github/workflows/ci.yml — the compile job's cross-compiled `go vet ./...` for tools/desk and the build-test job's per-module `go test` for tools/desk, tools/harnessgen, tools/harnesslint, tools/plugindrift; tools/skillslint runs via `go run .` in its own gating job. No analyzer beyond vet runs anywhere in this repo's CI today."
  - "Module set (driver ruling 2026-10-11, relayed to the worker): EVERY module `git ls-files '*go.mod'` finds — 20 at authoring — mirroring ci.yml build-test's own discovery, so a new module is linted without editing the runsheet. Not six, not a hard-coded list."
  - "docs/distribution.md (the .assay-versions pin doctrine) — shared tools are sha256-pinned binaries acquired over plain HTTPS, never floating; golangci-lint joins the CI under the same discipline (pinned version + recorded sha256), in-tree, not via the .assay-versions channel."
  - "freshness-checked 2026-10-11 @ 323d7f080 (origin/main) — no .golangci.yml exists at the repo root; no golangci-lint reference exists in any workflow; brief-32 is the next free number after brief-31 (PR #2533)."
exec-tier: strong
exec-tier-why: >-
  (b) correctness is a cross-artifact argument: one curated set stated in the config,
  measured green across every module the discovery loop finds, with every exclusion carrying its own inline reason —
  and a runsheet hunk that must match what was measured.
consumers:
  - ".golangci.yml (repo root, new): fixed-here (the curated linter set, the version pin comment, per-entry exclusions with reasons)"
  - "the Go sources of every module `git ls-files '*go.mod'` finds (20 at authoring, tools/* and the five root modules alike): fixed-here (only where the measurement found a real finding — no drive-by refactors)"
  - "docs/runsheets/golangci-ci-hunk.md (planned, new): fixed-here (the exact ci.yml diff and the one-line apply command for the human leg)"
  - ".github/workflows/ci.yml: out-of-scope for the worker (the App lacks the workflows scope; the hunk lands by the runsheet — named here so the reviewer checks the runsheet matches the config)"
version: 1
id: 5c8d1f30-92a4-4b6e-8f3c-77d2e9a4b1c8
---

# Brief 32 — golangci-lint joins the Go CI

## Context

What runs today on this repo's Go: `go vet` (cross-compiled, compile job), per-module `go
test`, and skillslint for the skill bodies. What does not run: any of the analyzer families
between vet and human review — errcheck's unchecked errors, staticcheck's SA correctness
checks, unused's dead code. Those classes are exactly the ones a machine proves cheaper than
a reviewer, and the review lanes gain nothing from re-deriving them by attention.

**Discipline this brief holds (so the lint is a control, not theater):**

1. **Pinned, not floating.** A recorded golangci-lint version, sha256-verified at acquisition
   (same discipline as the statusgen binary channel). The version appears in the config's
   header comment; bumping it is a deliberate act.
2. **Curated, green from day one.** The set is errcheck + staticcheck + unused + govet, run
   across every module `git ls-files '*go.mod'` finds. The worker MEASURES FIRST; real findings get fixed in-PR when
   small and mechanical; anything not fixed carries a per-entry exclusion with an inline
   reason (a dated comment naming the class and why), never a blanket skip, never a bare
   `//nolint` without a rationale.
3. **One human leg, and where it lands.** The ci.yml hunk (a single new job that acquires the
   pinned binary and runs the config per discovered module) is prepared as a runsheet file
   with the exact diff. The human applies it AS A COMMIT ON THIS BRANCH, authored by the
   human identity, AFTER the worker's final commit and BEFORE merge — so row 6's
   worker-scoped workflow check (over the worker's own commits only) still holds, and merged
   main carries the job the moment the PR lands. Until that commit exists, CI shows no
   golangci job — the PR's evidence is the worker's recorded local runs, which the
   runsheet's commands reproduce.
4. **No drive-by refactors.** Fixes touch only what a finding names. A finding whose fix is
   structural (an API change, a behavior question) is excluded with its reason and named as
   a follow-up, not fixed inside this brief.

## Task

**Task 1 — measure.** Acquire the pinned golangci-lint (record version + sha256 in the PR
evidence), then for EVERY directory containing a go.mod — discovered the way ci.yml build-test
discovers them, `git ls-files '*go.mod'`, never a hand list — run
`golangci-lint run --config <repo-root>/.golangci.yml ./...` from that directory. Record
per-module finding counts by linter in the PR body. The runsheet job iterates the same
discovery, so a new module joins the lint with no one editing anything.

**Task 2 — the config.** `.golangci.yml` at the repo root, written for golangci-lint
**v2's** config schema (`version: "2"`, linters under `linters.default`, the v2
exclusion keys — gosimple/stylecheck are folded into staticcheck in v2; the worker pins
ONE exact vX.Y.Z version and writes the config for THAT version's schema, never a mix):
the four linters enabled explicitly (no `enable-all`), default exclusions OFF except
golangci's own generated-code skip, the version-pin header comment, and whatever per-entry
exclusions Task 1 justified — each with its dated inline reason.

**Task 3 — fixes.** Small mechanical findings fixed in-PR (unchecked errors that genuinely
want checking, unused symbols deleted). Structural findings excluded + named as follow-ups.

**Task 4 — the runsheet.** docs/runsheets/golangci-ci-hunk.md (planned): the exact unified diff for
ci.yml's new `golangci` job (acquire pinned binary + sha256 check + run per module) and the
one-line apply command. The worker does NOT touch ci.yml.

**Task 5 — changelog fragment.** changelog/desk-tools-32.md (planned), one bullet.

files:
- `.golangci.yml` (new) — Tasks 1–2
- every discovered module's Go sources (`git ls-files '*go.mod'`, all 20) — Task 3 (only where a finding names them)
- docs/runsheets/golangci-ci-hunk.md (planned) — Task 4
- changelog/desk-tools-32.md (planned) — Task 5

## Verify (executable — no prose-only DoD items)

Rows 1–3, 6 are discriminating: each FAILS against today's tree (at 323d7f080) and passes
only once the Task lands. Rows 4, 5, 7 pass today and must still pass. Row 8 is
merge-gated: it cannot pass on the branch (the ci.yml hunk is the human's commit) and MUST
be run on merged main — it is the row that proves the brief's title outcome actually reached
CI. Run from the repo root.

| # | Class | Command | Expect |
|---|-------|---------|--------|
| 1 | check | `test -f .golangci.yml && grep -q 'errcheck' .golangci.yml && grep -q 'staticcheck' .golangci.yml && grep -q 'unused' .golangci.yml && grep -q 'govet' .golangci.yml && ! grep -q 'enable-all' .golangci.yml && grep -qE '^# golangci-lint v[0-9]+\.[0-9]+\.[0-9]+' .golangci.yml && echo ok` | prints `ok`: the curated set is explicit, enable-all is absent, and the header comment pins an exact version (`# golangci-lint vX.Y.Z`), which row 2's run must match. FAILS today: no .golangci.yml exists. |
| 2 | check +flow | `v=$(grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' .golangci.yml \| head -1) && [ "$(golangci-lint version --short)" = "${v#v}" ] && cd tools/desk && golangci-lint run --config ../../.golangci.yml ./... && cd ../../statusgen && golangci-lint run --config ../.golangci.yml ./... && echo ok` | exit 0: the binary on PATH IS the pinned version (not whatever floats), then the two largest modules pass with the landed config — the run the runsheet reproduces for every discovered module. FAILS today: no config exists. |
| 3 | check | `for m in $(git ls-files '*go.mod' \| xargs -n1 dirname); do grep -rn 'nolint' --include='*.go' "$m"; done \| grep -vE "nolint:[a-zA-Z0-9_,-]+ +// +[^ ]" \| wc -l > "${TMPDIR:-/tmp}/dt32-nl.txt"; [ "$(cat "${TMPDIR:-/tmp}/dt32-nl.txt")" = "0" ] && echo ok` | prints `ok`: every `nolint` in every module carries a rationale in the tree's own `//nolint:<linter> // <reason>` convention (the form golangci's nolintlint require-explanation expects), none is bare. FAILS today: 3 bare directives exist — tools/desk/internal/weight/weight_test.go:387,391,394 — the worker fixes exactly those three in-PR (each gains its reason) and touches no other pre-existing directive. |
| 4 | check +neighbour | `cd tools/desk && GOOS=windows GOARCH=amd64 go vet ./... && GOOS=darwin GOARCH=arm64 go vet ./... && echo ok` | exit 0: the existing cross-compile vet is unaffected. Passes today and must still pass. |
| 5 | check +neighbour | `cd tools/desk && go test -count=1 ./internal/deskkit/ ./cmd/deskdispatch/ > /dev/null && echo ok` | exit 0: representative suites unaffected by Task 3 fixes. Passes today and must still pass. |
| 6 | check | `test -f docs/runsheets/golangci-ci-hunk.md && grep -q '^diff --git a/.github/workflows/ci.yml' docs/runsheets/golangci-ci-hunk.md && grep -q 'sha256' docs/runsheets/golangci-ci-hunk.md && git log --format= --name-only --author='assay-worker-app' refs/remotes/origin/main..HEAD \| grep -c '^\.github/workflows/' > "${TMPDIR:-/tmp}/dt32-wf.txt"; [ "$(cat "${TMPDIR:-/tmp}/dt32-wf.txt")" = "0" ] && echo ok` | prints `ok`: the runsheet carries the exact ci.yml diff with the sha256 acquisition, and the WORKER's own commits touch zero workflow files — the check is worker-scoped because the human's runsheet commit (Discipline 3) lands on this branch after the worker's final commit and is expected to touch ci.yml. FAILS today: no runsheet exists. |
| 7 | check | `cd statusgen && go build -o "${TMPDIR:-/tmp}/dt32-sg" . && "${TMPDIR:-/tmp}/dt32-sg" --root .. --lint 2>&1 \| grep -q 'LINT: PASS' && echo ok` | prints `ok`: the tree lints clean. The changelog fragment is deliberately NOT tested — the release workflow clears `changelog/` when it aggregates, so a fragment-existence check cannot survive a release. Passes today and must still pass. |
| 8 | check +flow | `v=$(grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' .golangci.yml \| head -1) && grep -qE '^  golangci:' .github/workflows/ci.yml && grep -qF "$v" .github/workflows/ci.yml && grep -q sha256 .github/workflows/ci.yml && echo ok` — run on MERGED MAIN after the human's runsheet commit lands. A CI run of the job at a merged head is the additional evidence. | The one row that proves golangci-lint actually runs in CI — without it the brief could reach done with every other row green and the lint never wired. Could-not-check at authoring: the hunk is the human's commit, not yet on any branch. |

## Evidence
<!-- appended at implementation time by a NON-implementer: one row per Verify item
     (command, exit code, output line(s) or hash, date, runner). -->

| # | Command | Result | Output | Date | Runner |
|---|---------|--------|--------|------|--------|

## Review
Gate: model (from frontmatter). The reviewer answers, in the verdict: (1) is every exclusion
a per-entry, dated, reasoned exclusion — no blanket skips? (2) do the in-PR fixes stay
mechanical, or has anything structural slipped in? (3) does the runsheet's diff match the
landed config (same linters, same pin) and carry a sha256-verified acquisition? (4) is the
worker's diff genuinely free of workflow-file changes? (5) after the human applies the
runsheet hunk, would the CI job run exactly what row 2 ran?
