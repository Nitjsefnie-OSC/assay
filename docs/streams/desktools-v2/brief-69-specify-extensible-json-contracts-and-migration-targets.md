---
brief: assay:assay:desktools-v2:69
title: Specify extensible JSON contracts and migration targets
why: "Whole-record version bumps couple unrelated additions and make supported transitions hard to qualify."
wave: 1
depends: []
unblocks: ["desktools-v2/67"]
effort: M
gate: model
risk: {regulatory: no, customer: no, irreversible: no, sensitive-data: no}
issues: []
schema: brief-v2
version: 1
authored: 2026-10-21 by statusgen newbrief
sources:
  - 'desktools-v2: Specify extensible JSON contracts and migration targets'
  - "freshness-checked 2026-10-21 @ af897c3c112bd426cbbc794839dca6d41df911d6 (origin/main fetched)"
exec-tier: strong
exec-tier-why: "Cross-contract compatibility and consumer closure require design judgment."
domain: complicated
outcome: none
consumers:
  - "docs/proposals/non-ci-architecture/: fixed-here"
  - "docs/performance-attribution-spec.md: fixed-here"
  - "docs/contracts.md: fixed-here"
  - "perfrec/ and tools/desk/cmd/deskperf/: follow-up desktools-v2/67"
  - "producer selection: follow-up desktools-v2/68"
  - "other JSON implementations: out-of-scope (migration targets specified here; implementation belongs to their owning contract revisions)"
---

# Brief 69 — Specify extensible JSON contracts and migration targets

## Context

files: docs/contracts.md; docs/proposals/non-ci-architecture/{11,12,16,20,29}-*.md; docs/performance-attribution-spec.md; docs/streams/desktools-v2/brief-67* and brief-68*.
facts: Source census pinned to af897c3c112bd426cbbc794839dca6d41df911d6 on 2026-10-21. perfrec has v1/v2/v3 whole-record dispatch; runner requests already have optional ext/required capabilities; opmetrics/2 documents additive fields. These are source observations, not deployed inventory. Driver requested perf-v4 and future architecture extension migration on 21 October.

Design-fit: update the existing future architecture, not a parallel spec. This is docs/spec delivery plus authorship of /67 and /68; it delivers neither runtime brief. All risk answers are no because no codec, data collection, credentials, live configuration or authority transition changes in this item.

## Ground rules

Use normal native worker admission, one isolated worktree/PR and role-bound publication.
No workflow triggers, releases, installs, uploads or live infrastructure. Preserve other
owners and completed evidence. Stop at implemented; independent verification owns Evidence.

## Task

1. Amend the controlling target, model, delivery sequence and owner manifest to use stable cores and independently versioned extensions. Preserve revision-bound prior approvals and current-versus-target distinctions.
2. Publish the pinned source census with actual codec witnesses and ownership/transition dispositions, including existing extensible precedents and authority-bearing contracts. State discovery gaps; do not equate grep sites with unique protocols or tool deployment counts.
3. Freeze the performance successor contract in the existing attribution spec: fields, bounds, optional-version handling, privacy, historical reads, mixed reports and reader-first selection. Author /67 and /68 with named discriminating tests and explicit dependencies.
4. Preserve /26 snapshot reuse/bench and /62 integration obligations and all independent /60/64/65 holds. Reconcile future format targets in the architecture without changing completed Verify/Evidence or someone else's active code.

## Verify

| # | Command | Expect |
|---|---|---|
| 1 | `git diff --check` | exit 0 |
| 2 | `statusgen --root . --lint` | exit 0; no new lint failures |
| 3 | `git show af897c3c112bd426cbbc794839dca6d41df911d6:loopadmin/runner/request.go` | exit 0; DecodeRequest separates strict core from ext, and CheckCapabilities refuses unsupported require entries, corroborating the census precedent |
| 4 | `git show af897c3c112bd426cbbc794839dca6d41df911d6:tools/desk/cmd/opmetrics/emit.go` | exit 0; SchemaVersion comment explicitly permits additive fields, disproving a blanket claim that all JSON is closed |
| 5 | `git show af897c3c112bd426cbbc794839dca6d41df911d6:perfrec/v3.go` | exit 0; V3Data matches all mapped payload members, including requestDrops and nullable query, with no invented queryCoverage field; validV3 requires tool role none and workflow role equality, and query observations only apply to client/handler kinds |
| 6 | `git show af897c3c112bd426cbbc794839dca6d41df911d6:tools/desk/internal/deskkit/perfread.go` | exit 0; FilterPerf explicitly excludes v3 from the legacy desk-role filter, corroborating the retained distinction from workflow assigned-role filtering |

Pre-mortem: falsely describing every decoder as closed is caught by rows 3–4;
losing requestDrops or fabricating query coverage is caught by row 5; conflating
legacy desk role with assigned role is caught by rows 5–6. Adequacy of
family grouping, future authority boundaries and consumer migration is review-only:
reviewers must challenge the source witnesses and retained dependency obligations,
not accept section presence as architectural conformance.

## Evidence

<!-- Independent postmerge verifier records actual commands, source and outcomes here. -->

## Review

Independent correctness and security review of the exact head; normal ready/merge gates.
