---
brief: assay:assay:statusgen:38
title: "One-time v1-to-v2 witness backfill where the witnessed content is still fetchable"
why: "Thirty verified briefs are stuck on v1 witnesses whose inputs never changed. Where the witnessed commit can still be fetched, by its sha or through the delivering change request's head ref, and the brief declares its inputs, a v2 record can be derived from that commit's tree without re-running anything. Every other brief stays v1 and is re-witnessed."
wave: 3
depends: ["statusgen/37"]
unblocks: []
effort: M
gate: human
risk: {regulatory: no, customer: yes, irreversible: no, sensitive-data: no}
issues: [2026, 3105, 3658]
schema: brief-v2
outcome: none
version: 1
authored: "2026-10-20 by the-desk (design brief for the witness-staleness class)"
gate-why: "A backfilled record asserts what an old run read without re-running it. The driver decides whether that assertion may release a brief or only document it."
decision-trigger: start
sources:
  - "medici-finance/assay#3105 — the stuck verified-to-done report"
  - "medici-finance/assay#3658 — squash-orphaned record shas; the 95 existing records are append-only and are not rewritten"
  - "docs/streams/statusgen/brief-36-content-bound-witness-schema-v2-reader.md and docs/streams/statusgen/brief-37-content-bound-witness-schema-v2-writers.md — the schema, seams and writers"
  - "`statusgen --root . --auto-flip-model --dry-run` at 75cc11f14: 30 REFUSED, 0 flipped"
  - "freshness-checked 2026-10-20 @ 75cc11f1423301ea8ba0b111d48c8d5b8a303537 (origin/main)"
exec-tier: strong
exec-tier-why: "It writes evidence derived from history. Each refusal path (unfetchable, undeclared, dirty, command changed, environment-dependent) must be proven to leave the brief at v1."
domain: complicated
consumers:
  - "tools/desk/cmd/deskevidence/ (the `backfill` subcommand): follow-up statusgen/38"
  - "tools/desk/internal/deskkit/forge.go and its adapters (change-request head ref per forge): follow-up statusgen/38"
  - "docs/streams/statusgen/README.md (rollout section: which briefs were backfilled, which await a re-witness): follow-up statusgen/38"
---

# Brief 38 — One-time v1-to-v2 witness backfill where the witnessed content is still fetchable

## Context

files:
- `tools/desk/cmd/deskevidence/deskevidence.go`
- tools/desk/cmd/deskevidence/backfill.go (planned)
- tools/desk/cmd/deskevidence/backfill_test.go (planned)
- `tools/desk/internal/deskkit/forge.go`
- `tools/desk/internal/deskkit/forge_gitlab.go`
- tools/desk/internal/deskkit/witnessseam.go (planned, statusgen/37)
- tools/desk/cmd/deskevidence/testdata/backfill/ (planned)
- `docs/streams/statusgen/README.md`
- changelog/statusgen-witness-v2-backfill.md (planned)

facts (source and date in frontmatter):
- At 75cc11f14 the dry run refuses 30 verified briefs (build-less-brittle 02, 04, 05, 06, 09, 10, 11; contributor-trust 04; desk-supervision 17, 18, 19, 27; desk-tools 15, 23; desktools-v2 05, 09, 12, 14, 15, 61; graph-execution 09, 20; harness-portability 06; iso-9001 03; quality 19; statusgen 15, 17, 19, 22; windows-port 00). Four claims in build-less-brittle 04, 05, 06 and 09 are refused because the Verify row's command changed after the witness. Eight of the briefs have `+dirty` witnesses.
- A v1 Runner cell names a short commit. A squash merge leaves that commit off main, but the forge keeps the change request's head ref (GitHub `refs/pull/<N>/head`, GitLab `refs/merge-requests/<iid>/head`), so the commit is usually still fetchable by change request.
- A `+dirty` v1 witness ran on content that no commit holds, so its inputs cannot be recovered.
- No v1 run was in a network-less runner; statusgen/36's environment policy decides whether a back-filled record may release.

layering: deskevidence orchestrates (find candidates, fetch through the forge-neutral ref namespace, land an Evidence change). The digests come only from `statusgen witnessdigest --json --rev <commit>`. statusgen/36's reader judges the result. No new judge.

single-point-of-failure: the claim that the fetched commit is the one the v1 witness ran on. Behind it: the full commit must match the cell's short prefix and be the unique match among the change request's commits; the Verify row's command at that commit must equal the witnessed command; and the reader re-checks every digest at the item revision.

design-fit:
  owner: tools/desk deskevidence (orchestration), statusgen/evaluation (digests)
  contract: "R-witness-schema (statusgen/36); a back-filled record carries `origin: backfill`, `source_cell` and the change request it was fetched through"
  retires: ["the stuck-brief backlog in medici-finance/assay#3105 where the witnessed content is still fetchable"]
  weight: "one subcommand (`deskevidence backfill`, dry-run by default); flags +1 (`--apply`); no new module or service"
  why-add: "Re-running thirty briefs' rows by hand is the alternative; a derived record is cheaper and is still judged by the same reader."
  reuse-check: "Reuses statusgen/37's witnessseam and LandedCommit, the forge-neutral change-request ref mapping deskdispatch already uses, and deskevidence's Evidence landing path."

## Human decision

Old verification runs recorded only a commit. For many of them that commit can still be fetched and its files hashed, which shows what the files were when the run happened. It cannot show which tool versions the run used or whether it touched the network.

Options:
1. Recommended: write a back-filled record with tool versions taken from the pins committed at that commit, marked as back-filled, and let the reader's environment policy decide whether it releases.
2. Write back-filled records for audit only: tool versions unknown, so the reader always reports them as needing a re-run.
3. Do not back-fill; re-run every stuck brief instead.

Default if no answer: option 2. Records are written, and nothing is released by a back-fill.

## Read first

Read statusgen/36 and statusgen/37 in full, `witnessseam.go` (planned), the Evidence landing path in `tools/desk/cmd/deskevidence/deskevidence.go`, and the change-request ref mapping in `tools/desk/cmd/deskdispatch/dispatch.go`. Re-run `statusgen --root . --auto-flip-model --dry-run` on fresh main; if statusgen/37 has not merged, report NEEDS_CONTEXT.

## Ground rules

- Feature branch and draft PR only. Never push to main, merge, release or dispatch a workflow. Merge main in, never rebase.
- Stop at `implemented`. The verifier identity runs `--apply` and lands the Evidence change; the implementer does not.
- Fixture repositories and fake forges only in tests.
- Append only. Never rewrite or delete a v1 cell, an Evidence row or a verify-outcome record.
- If anything is unclear or contradicts repo state, report NEEDS_CONTEXT. Do not guess.

## Interface contract

`deskevidence backfill [--brief <stream>/<NN>]... [--apply]`. Dry run by default: it prints one line per Verify row with `backfilled`, or `stays-v1: <reason>`. For each v1 witness:

1. Skip with `stays-v1: dirty` for `+dirty` or `+unknown`; `stays-v1: undeclared` when the brief has no parseable `inputs:`; `stays-v1: command-changed` when the row's command at the item revision differs from the one at the witnessed commit; `stays-v1: env` for an `+env` row.
2. Resolve the full commit: locally, else by fetching the delivering change request's head ref through the forge-neutral ref mapping. The full commit must start with the cell's prefix and be the only such commit reachable from that ref. Otherwise `stays-v1: unfetchable` or `stays-v1: ambiguous`.
3. Get digests from `statusgen witnessdigest --json --rev <commit>` and write a v2 record with `origin: backfill`, `source_cell` (the v1 cell text), `cr`, and `informational.commit`. Tools follow the Human decision.
4. With `--apply`, append a v2 Evidence row beside the v1 row, citing it, with the v1 row's result and date. Then update the README rollout section.

## Task

1. Implement `deskevidence backfill` as above, using only `witnessseam.go` (planned) for digests.
2. Add a `ChangeRequestHeadRef(repo, id)` read to the forge-neutral layer for GitHub and GitLab, reusing deskdispatch's mapping.
3. Add fixtures: a squash-orphaned witness fetchable only by change-request ref; a locally present witness; a `+dirty` witness; a brief without `inputs:`; a changed command; an `+env` row; an unfetchable commit; and an ambiguous short prefix.
4. Add a rollout section to the README listing, after a dry run on main, how many of the 30 briefs back-fill and how many await a re-witness, by reason.
5. Add changelog/statusgen-witness-v2-backfill.md. Replace each `follow-up statusgen/38` entry in `consumers:` with `fixed-here` in the same change.

## Verify

Named tests and fixtures are planned deliverables. A missing selector is a failure.

| # | Class | Command | Expect |
|---|---|---|---|
| 1 | check:ci +flow +mutation | `(cd tools/desk && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg38-r1.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestBackfillOrphanViaCR$' ./cmd/deskevidence/ > "$out" 2>&1 && grep -F -- '--- PASS: TestBackfillOrphanViaCR ' "$out")` | exit 0; named PASS. A squash-orphaned witness is fetched through the change-request head ref on both fake forges, back-filled, and then released by statusgen's dry run on the fixture repository under option 1. |
| 2 | check:ci +mutation | `(cd tools/desk && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg38-r2.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestBackfillStaysV1$' ./cmd/deskevidence/ > "$out" 2>&1 && grep -F -- '--- PASS: TestBackfillStaysV1 ' "$out")` | exit 0; named PASS. Dirty, undeclared, command-changed, `+env`, unfetchable and ambiguous witnesses each print `stays-v1` with that reason and write nothing. |
| 3 | check:ci +mutation | `(cd tools/desk && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg38-r3.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestBackfillAppendOnly$' ./cmd/deskevidence/ > "$out" 2>&1 && grep -F -- '--- PASS: TestBackfillAppendOnly ' "$out")` | exit 0; named PASS. After `--apply`, the brief's previous Evidence section is a byte prefix of the new one and every existing verify-outcome record is byte-unchanged. |
| 4 | check:ci +mutation | `(cd tools/desk && out=; trap 'test -n "$out" && rm -f "$out"' EXIT; out=$(mktemp "${TMPDIR:-/tmp}/sg38-r4.XXXXXX") && GOWORK=off go test -count=1 -v -run '^TestBackfillChangedInput$' ./cmd/deskevidence/ > "$out" 2>&1 && grep -F -- '--- PASS: TestBackfillChangedInput ' "$out")` | exit 0; named PASS. A back-fill whose declared input changed after the witnessed commit is written, but statusgen's dry run refuses it with `witness-input-changed`. Under option 2, an unchanged one is reported as needing a re-run. |
| 5 | check +dereference | `d=; trap 'test -n "$d" && rm -rf "$d"' EXIT; d=$(mktemp -d "${TMPDIR:-/tmp}/sg38-dry.XXXXXX") && (cd tools/desk && GOWORK=off go build -o "$d/de" ./cmd/deskevidence) && "$d/de" backfill > "$d/out" 2>&1; rc=$?; grep -c -e '^backfilled' -e '^stays-v1: ' "$d/out"; exit $rc` | exit 0; prints a positive count. The dry run on main writes nothing (`git status --porcelain` is unchanged) and classifies every row of the stuck briefs. The counts match the README rollout section. |
| 6 | check +dereference | `statusgen --root . --consumers --brief statusgen/38 --base "$(git merge-base refs/remotes/origin/main HEAD)"` | pre-merge only; exit 0. The implementation diff corroborates the routing in `consumers:`. At implementation replace this provisional command with the actual delivery and first-parent pins and exact summary counts, per `docs/consumer-delivery-checks.md`. |

## Pre-mortem and detection

- A back-fill hashes the wrong commit, for example a different commit sharing the short prefix. Caught by row 2's ambiguous case.
- A `+dirty` witness is back-filled from its parent commit, asserting content it never read. Caught by row 2.
- A back-fill rewrites history. Caught by row 3.
- A back-fill releases a witness whose inputs have since changed. Caught by row 4; the reader re-checks at the item revision.
- The rollout section claims counts that do not match the tool. Caught by row 5.
- A back-filled row was environment-dependent but not marked `+env`. Not caught here; it is governed by statusgen/36's environment policy, and option 2 of this brief's decision keeps every back-fill from releasing.

## Evidence

<!-- Independent verifier records the subject revision, environment, command, exit, named result and date. This authoring change executes no implementation check. -->

## Review

Gate: human. Reject any of these:
- a back-fill of a `+dirty` witness;
- a back-filled record whose digests did not come from `statusgen witnessdigest`;
- an existing record or Evidence row changed;
- a back-fill that releases under option 2.
