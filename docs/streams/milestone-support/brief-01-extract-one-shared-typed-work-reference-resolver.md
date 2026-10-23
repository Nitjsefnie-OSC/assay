---
brief: assay:assay:milestone-support:01
title: "Extract one shared typed work-reference resolver"
why: "Separate parsers currently assign meaning to the same work references. One public pure implementation lets both binaries agree without moving credentials or execution authority."
wave: 0
depends: []
unblocks: ["milestone-support/02", "milestone-support/03", "milestone-support/04", "milestone-support/08"]
effort: M
gate: human
risk: {regulatory: no, customer: no, irreversible: no, sensitive-data: yes}
gate-why: "Refactors identity matching consumed by human rulings and claim ownership; preservation of existing strict profiles and exact work binding requires human design confirmation."
decision-trigger: spec
issues: []
schema: brief-v2
version: 1
id: eb74beee-5f67-4490-83db-2fd68255cc41
authored: 2026-10-22 by assay-worker-app[bot]
exec-tier: strong
exec-tier-why: Cross-module identity, version, evidence and compatibility behavior must agree at independent boundaries.
domain: complicated
outcome: none
sources:
  - docs/streams/milestone-support/README.md
  - spec/milestone-v1.md
  - spec/stream-v1.md
  - spec/brief-v1.md
  - docs/library-first.md
  - "freshness-checked 2026-10-22 @ a631d47ea71b3270ab33ba08af47d3dde8c9e7d5"
consumers:
  - "statusgen/workref/: follow-up milestone-support/01 (this delivery; replace with fixed-here in implementation)"
  - "statusgen/contexthandoff/ids.go: follow-up milestone-support/01 (this delivery; replace with fixed-here in implementation)"
  - "statusgen/contexthandoff/ids_test.go: follow-up milestone-support/01 (this delivery; replace with fixed-here in implementation)"
  - "tools/desk/internal/deskkit/trailer.go: follow-up milestone-support/01 (this delivery; replace with fixed-here in implementation)"
  - "tools/desk/internal/deskkit/trailer_test.go: follow-up milestone-support/01 (this delivery; replace with fixed-here in implementation)"
  - "tools/desk/go.mod: follow-up milestone-support/01 (this delivery; replace with fixed-here in implementation)"
  - "tools/desk/go.sum: follow-up milestone-support/01 (this delivery; replace with fixed-here in implementation)"
  - "docs/contracts.md: follow-up milestone-support/01 (this delivery; replace with fixed-here in implementation)"
  - "docs/library-first.md: follow-up milestone-support/01 (this delivery; replace with fixed-here in implementation)"
  - "docs/streams/milestone-support/brief-02-negotiate-reader-capabilities-through-one-shared-contract.md: follow-up milestone-support/02 (consumes this delivery; exact runtime readers enumerated at pickup)"
  - "docs/streams/milestone-support/brief-03-load-and-validate-versioned-milestone-records.md: follow-up milestone-support/03 (consumes this delivery; exact runtime readers enumerated at pickup)"
  - "docs/streams/milestone-support/brief-04-add-negotiated-brief-schema-and-prefixed-authoring.md: follow-up milestone-support/04 (consumes this delivery; exact runtime readers enumerated at pickup)"
  - "docs/streams/milestone-support/brief-08-normalize-pr-trailers-and-claims-through-the-shared-resolver.md: follow-up milestone-support/08 (consumes this delivery; exact runtime readers enumerated at pickup)"
---

# Brief 01 — Extract one shared typed work-reference resolver

## Context

files: `statusgen/workref/` (planned), `statusgen/contexthandoff/ids.go`, `statusgen/contexthandoff/ids_test.go` (planned), `tools/desk/internal/deskkit/trailer.go`, `tools/desk/internal/deskkit/trailer_test.go`, `tools/desk/go.mod`, `tools/desk/go.sum`, `docs/contracts.md`, `docs/library-first.md`, `changelog/milestone-support-01.md` (planned), `statusgen/testdata/milestone-support/01-mutations.json` (planned). New paths and tests are planned implementation deliverables, not current artifacts.

facts:
- Source baseline: merged main `a631d47ea71b3270ab33ba08af47d3dde8c9e7d5`, inspected 22 October 2026.
- Existing reference owners: `statusgen/contexthandoff/ids.go`, `statusgen/briefv2.go` and `tools/desk/internal/deskkit/trailer.go` currently implement different legacy grammars; do not erase those boundary contracts.
- Module boundary: `statusgen` and `tools/desk` are separate Go modules. Public `statusgen/workref` and `statusgen/readercaps` are planned; neither exists at this baseline. `tools/desk/internal` is not importable by statusgen.
- Support baseline: brief schemas v1-v3 are advertised; milestone, prefixed-reference and negotiated reader-capability support is not advertised. The installed plugin or a permissive YAML parser is not evidence of support.
- Availability: `statusgen/evaluation`, `statusgen/domain`, `statusgen/contexthandoff`, `statusgen/streamview` and `loopadmin` exist; the root `workflow` store/controller module does not. Preserve its future integration obligations without inventing a working store.

layering: statusgen/workref owns this slice. Use in-process public Go library calls for shared parsing, reference resolution and capability assessment. No new binary, daemon or subprocess-RPC seam; host adapters own filesystem, transport, authentication and effects. Existing online forge reads remain behind their current authorized deskread boundary. Task and Verify exercise the actual consumer seam.

design-fit:
  owner: "statusgen/workref"
  contract: none — existing index does not yet name this exact typed record/projection; update its existing owner row in the implementation, not a second registry.
  retires: ["duplicated interpretation or implicit state inference at the named existing consumer boundary"]
  weight: "New binaries 0; new services 0; required pure API/schema/fixture changes only. Measure actual rule/flag/refusal deltas at implementation review; no current weight reduction is claimed."
  why-add: "Reuse the existing semantic owner and public module. This records a missing shared meaning; moving it into a helper process or copying another parser would add authority or drift without benefit."

single-point-of-failure: incorrect shared identity or capability interpretation; retain independent current-subject, authority and consumer checks and plant faults with the outer check bypassed where Verify names that boundary.

## Read first

- [Implementation scope and consumer census](README.md#implementation-scope).
- [Milestone contract](../../../spec/milestone-v1.md).
- [Brief contract and compatibility](../../../spec/brief-v1.md).
- [Stream contract](../../../spec/stream-v1.md).
- [Library-first direction](../../library-first.md).
- [Existing semantic owners](../../contracts.md).
- [Dependency waves and adoption holds](README.md).

## Ground rules

- Own isolated branch and draft implementation PR. Do not merge, flip status/spec/stream/PR readiness, dispatch work, cut a tag, change installed pins or contact a cluster.
- Scope approval and stream activation take effect on the maintainer's merge of PR #4182; this brief has no completion evidence. A non-implementer owns post-merge verification; never fill Evidence from authorship or a fixture expectation.
- Refresh main, active claims/PRs, module imports, source contracts and consumer inventory at pickup. Record brief/source revisions, actual public API availability, coverage/unknowns and continue/adapt/hold in the implementation handoff. No obsolete active claim is overwritten.
- Preserve existing credentials, independent human/risk gates, UUIDs, archived evidence and exact-subject meanings. No ambient-auth fallback or automatic bulk migration.
- Shared resolution and capability assessment MUST be direct library calls, not subprocess calls to statusgen, desk-tools or a new helper executable. External I/O remains in existing adapters.

## Human decision

Review the concrete identity/capability boundary and preserved existing authority checks through the existing decision process before implementation. Options: approve the bounded design, request correction, or hold. No answer supplies no approval. The authoring request does not grant implementation sign-off.

## Task

1. Add public stdlib-only package github.com/medici-finance/assay/statusgen/workref in the existing statusgen module, not a new module, daemon or tools/desk/internal import. Expose typed Parse/Resolve operations over explicit declaring context and an immutable supplied alias registry. Return kind, full identity, optional exact revision selector and compatibility spelling; do not read files, environment, network or tokens.
2. Implement final-segment-only b-/m- classification, colon elisions, all legacy aliases in brief §3.8, exact positive revision syntax, namespace collisions and missing/ambiguous context. Keep stable logical UUID/key distinct from spelling, source revision and version selector. Prefix-shaped names elsewhere in the ID must not select the kind. Overflow, whitespace, malformed and unknown required forms refuse/are unknown under the contracts.
3. Refactor existing contexthandoff.ResolveQualifiedID and deskkit trailer parsing to delegate to the common implementation through explicit legacy profiles preserving their DIFFERENT existing acceptance sets and outputs. Do not broaden old APIs to accept b-/m-/@ implicitly. Both production consumers must actually call the package; retaining independent regular expressions for the same semantics is not completion.
4. Wire tools/desk to the public statusgen module using the repository's require/replace pattern. Prove GOWORK=off compilation from both modules and a temporary external-consumer fixture; local replace success is source-build proof, not published-release proof. Record exact consumer/module closure; update semantic-owner and library-first documentation. No credential or online forge owner moves.

## Interface and completion contract

Consume only delivered prerequisite APIs at reviewed revisions. Shared core functions accept explicit values and return typed results; they do not fetch, claim, approve, launch or mutate the environment. Unknown coverage and unsupported required semantics remain explicit. Update affected canonical docs and the per-delivery changelog. Enumerate newly discovered readers with fixed-here/follow-up/out-of-scope dispositions before changing shared values. No new stateful store, task queue or acceptance authority is introduced.

## Verify

The following named tests, fixtures and executable mutation maps are FUTURE DELIVERABLES. They are not authoring-time results. Each test must exercise the production entrypoint/API named in Task and fail under its stated wrong-but-plausible mutation. A named PASS guard prevents an absent `-run` target from appearing green. Record concrete mutation commands in this brief's declared mutation manifest; no test may prove completion by file existence alone.

| # | Class | Shell | Command | Expect |
|---|---|---|---|---|
| 1 | check:ci +flow +mutation | sh | `(cd statusgen && milestone_verify_out=$(mktemp "${TMPDIR:-/tmp}/milestone-verify.XXXXXX") && trap 'rm -f "$milestone_verify_out"' 0 && GOWORK=off go test -count=1 -v -run '^TestWorkrefGrammarAndAliases$' ./workref/... > "$milestone_verify_out" && grep -q -- '--- PASS: TestWorkrefGrammarAndAliases ' "$milestone_verify_out")` | exit 0 and named PASS; mixed colon/slash aliases yield one identity, exact revision survives, wrong kind/whitespace/unknown root/overflow fail; remove alias context or classify a prefix-shaped stream as the kind and the test fails. |
| 2 | check:ci +flow +mutation | sh | `(cd statusgen && milestone_verify_out=$(mktemp "${TMPDIR:-/tmp}/milestone-verify.XXXXXX") && trap 'rm -f "$milestone_verify_out"' 0 && GOWORK=off go test -count=1 -v -run '^TestLegacyQIDParity$' ./contexthandoff/... > "$milestone_verify_out" && grep -q -- '--- PASS: TestLegacyQIDParity ' "$milestone_verify_out")` | exit 0 and named PASS; actual legacy API still refuses prefixes/slashes/revision selectors while valid old QIDs preserve meaning; broadening the legacy profile fails. |
| 3 | check:ci +flow +mutation | sh | `(cd tools/desk && milestone_verify_out=$(mktemp "${TMPDIR:-/tmp}/milestone-verify.XXXXXX") && trap 'rm -f "$milestone_verify_out"' 0 && GOWORK=off go test -count=1 -v -run '^TestSharedWorkrefTrailerParity$' ./internal/deskkit/... > "$milestone_verify_out" && grep -q -- '--- PASS: TestSharedWorkrefTrailerParity ' "$milestone_verify_out")` | exit 0 and named PASS; actual trailer parser delegates and preserves existing outputs; a divergent alias interpretation or dropped revision in the new explicit adapter fails. |
| 4 | check:ci +flow +mutation | sh | `(cd statusgen && milestone_verify_out=$(mktemp "${TMPDIR:-/tmp}/milestone-verify.XXXXXX") && trap 'rm -f "$milestone_verify_out"' 0 && GOWORK=off go test -count=1 -v -run '^TestWorkrefPublicImportBoundary$' ./workref/... > "$milestone_verify_out" && grep -q -- '--- PASS: TestWorkrefPublicImportBoundary ' "$milestone_verify_out")` | exit 0 and named PASS; both module-consumer fixtures import the public package with GOWORK=off; planted tools/desk/internal, net/http, os/exec or credential dependency fails. |

## Pre-mortem and dispatch check

The wrong-but-plausible delivery is the mutation named in each Verify row. Each must turn that production-path test red and the correct implementation green; do not merely inspect imports or fixture shapes. The facts, public-module visibility, derived waves, risk answers, sizing and consumers were inspected while authoring. API adequacy, completeness of the reader census and human acceptance remain review obligations. Implementation pickup rechecks them.

## Evidence

<!-- Empty by design. Independent verifier records actual commit, environment, command/exit, named test and mutation result, coverage/unknowns and evidence location. -->

## Review

Gate: human, derived from the declared risks. Independent review checks compatibility and consumer-flow evidence. No reviewer/worker stamp, release permission or acceptance receipt is created by this authoring change.
