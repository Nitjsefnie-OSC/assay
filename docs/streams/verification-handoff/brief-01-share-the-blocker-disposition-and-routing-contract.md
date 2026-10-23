---
brief: assay:assay:verification-handoff:01
title: "Share the blocker disposition and routing contract"
why: "The scanner and queue readers disagree about addressed issues, and a triage comment does not establish accountable repair work. One shared definition prevents callers assigning different meanings to the same blocker."
wave: 0
depends: []
unblocks: ["verification-handoff/02", "verification-handoff/03"]
effort: M
gate: human
risk: {regulatory: no, customer: no, irreversible: no, sensitive-data: yes}
gate-why: "Changes the deskcore/issueflow boundary used to schedule, preserve or resolve verification work; human review must confirm that trust, ownership, retention and independent acceptance cannot be inferred from routing."
decision-trigger: start
issues: []
schema: brief-v2
version: 1
id: 30004285-8520-4c89-aa26-c7a0be53d3a5
authored: "2026-10-23 by assay-worker-app[bot]"
exec-tier: strong
exec-tier-why: "Cross-module shared policy, partial writes and independent actor/source binding require coordinated design and negative controls."
domain: complicated
sources:
  - "docs/streams/verification-handoff/README.md — controlling implementation scope"
  - "spec/intake-exit-v1.md; docs/streams/desk-supervision/repair-obligation-v1.md; docs/verify-wake.md"
  - "freshness-checked 2026-10-23 @ fbe0331e2fc9f5c822c4edeafb8a18865c916828"
consumers:
  - "deskcore/issueflow/: follow-up verification-handoff/01 (this delivery; existing owner or planned extension)"
  - "tools/desk/internal/issueflow/: follow-up verification-handoff/01 (this delivery; existing owner or planned extension)"
  - "statusgen/go.mod: follow-up verification-handoff/01 (this delivery; existing owner or planned extension)"
  - "statusgen/go.sum: follow-up verification-handoff/01 (this delivery; existing owner or planned extension)"
  - "statusgen/scanissues.go: follow-up verification-handoff/01 (this delivery; existing owner or planned extension)"
  - "tools/desk/go.mod: follow-up verification-handoff/01 (this delivery; existing owner or planned extension)"
  - "spec/intake-exit-v1.md: follow-up verification-handoff/01 (this delivery; existing owner or planned extension)"
  - "spec/verification-handoff-v1.md: follow-up verification-handoff/01 (this delivery; existing owner or planned extension)"
  - "docs/contracts.md: follow-up verification-handoff/01 (this delivery; existing owner or planned extension)"
  - "statusgen/testdata/verification-handoff/01-mutations.json: follow-up verification-handoff/01 (this delivery; existing owner or planned extension)"
  - "changelog/verification-handoff-01.md: follow-up verification-handoff/01 (this delivery; existing owner or planned extension)"
---

# Brief 01 — Share the blocker disposition and routing contract

## Context

files: deskcore/issueflow/ (planned), tools/desk/internal/issueflow/, statusgen/go.mod, statusgen/go.sum, statusgen/scanissues.go, tools/desk/go.mod, spec/intake-exit-v1.md, spec/verification-handoff-v1.md (planned), docs/contracts.md, statusgen/testdata/verification-handoff/01-mutations.json (planned), changelog/verification-handoff-01.md (planned)

facts:
- Baseline main fbe0331e2fc9f5c822c4edeafb8a18865c916828, inspected 23 October 2026; source was draft and stream was parked at authoring. Scope approval and stream activation take effect on the maintainer's merge of PR #4206; per-brief gates remain separate.
- Intake discovery sees ordinary verifier bugs; 14 of a bounded 28 recent verifier-filed issues carried explicit intake triage, not proof of claims or repairs.
- The default loop filing/repair sinks print dry-run results. The live skill uses sanctioned deskfile/deskevidence paths; a printed URL is not durable work.
- Existing intake-exit-v1 has two writers and exact-key conformance tests. Its local log is not a fleet-wide source of assignment authority.
- Current scanner/readers differ on addressed handling. Default fanout STATUS.md input has no issue labels; the worker inbox needs an actual read.
- deskcore is public and imported by desk-tools; statusgen does not currently import it. Shared policy packages are planned, not shipped.
- Brief schemas v1-v3 are supported; no milestone or b-prefixed operational records are introduced by this authoring.

layering: deskcore/issueflow owns this slice. Shared decision logic is an in-process public Go library call; no separate binary, daemon or subprocess protocol. Existing typed forge, custody, claims and evidence adapters own effects. No private import across modules and no new lifecycle database.

design-fit:
  owner: "deskcore/issueflow"
  contract: none — reuse the existing issue projection, intake-exit, repair-obligation and wake owners; update their canonical index on implementation.
  retires: ["implicit routing/assignment inference and the duplicate predicate or partial-success ambiguity named in Task"]
  weight: "New binaries 0; new services 0; bounded contract/library/adapter changes. Record actual rule, field, refusal and flag deltas at implementation review."
  why-add: "Reuse existing semantic owners; a shared meaning is needed across existing callers. A new process, queue or acceptance authority would add failure modes without closing the handoff."

single-point-of-failure: stale disposition or successor identity; independent trust/admission/claim and exact-source acceptance checks remain at effect boundaries.

## Read first

- [Controlling scope, coverage and owner decisions](README.md#implementation-scope).
- [Intake exit contract](../../../spec/intake-exit-v1.md).
- [Existing repair-obligation contract](../desk-supervision/repair-obligation-v1.md).
- [Wake and retry contract](../../verify-wake.md).
- [Library-first direction](../../library-first.md).
- [Semantic owners](../../contracts.md).

## Human decision

Confirm that the concrete implementation preserves independent verification and existing human-only decisions while changing how failed work is retained and assigned. A routing record must never grant acceptance or wider execution permissions.

Options:
1. Approve the bounded implementation after its independent review and verification.
2. Request corrections while retaining the existing holds.
3. Hold the proposed change.

Default if no answer: none — blocks until answered.

## Ground rules

- Isolated feature branch and draft PR; no merge, source/stream/brief/PR-ready flip, release, installed pin change, workflow dispatch or live infrastructure effect in this brief.
- Recheck current main, claims, active PR overlap, controlling sources, semantic owners and module imports at pickup/resume; record observed revisions/coverage and continue, adapt or hold. A future API or draft spec is not shipped support.
- Preserve actor/authority separation, legacy identities and immutable Evidence/outcomes. Do not fabricate intake acknowledgement, assignment, historical execution or verified/done from labels or issue closure.
- New shared parsing/policy must be direct library calls, never a new helper executable. Preserve existing online read/custody boundaries.

## Task

1. Adopt the existing intake-exit, repair-obligation, claim and outcome contracts as inputs. Define a versioned extension for accountable handoff: source repo/brief and version/revision, receipt/case and held row identities, blocker ref, disposition and its deciding role, successor typed ref or external owner role/action, observation freshness and incomplete/unknown states. Reuse existing canonical case identity; do not equate a whole brief with one defect. Forge issue/PR records and existing dispatch claims remain authority; local intake exit logs and committed per-file projections are evidence/projections, never a competing lifecycle store.
2. Define the disposition semantics: needs-scoping/brief-authoring, directly implementable repair, external prerequisite, duplicate existing owner, and explicit unresolved/unknown. A mechanical placeholder or investigated comment alone is not an accepted disposition. Existing five intake exits remain distinct from repair scheduling state; extend or version their contract rather than silently repurpose bug/placeholder. Settle version negotiation for both writers/readers before dependent work emits required fields; old observations are unknown, not retroactively approved or discarded.
3. Move the existing pure candidate population selection into a public deskcore/issueflow library behind the existing internal compatibility facade. Extend that owner with pure handoff assessment over caller-supplied facts. Statusgen scanner, issueboard, intake and worker consumers must import and call the same functions in-process. Statusgen currently has no deskcore dependency: add the real module require/replace and prove GOWORK=off imports from both modules and a standalone consumer. Keep trust decisions, effect authorization, provider I/O, clocks and credentials in existing adapters; no helper executable, daemon or process RPC.
4. Specify a consistent addressed policy: ordinary unaddressed verifier bugs remain discoverable by intake; deliberately addressed issues stay in their receiving inbox and cannot leak into the ordinary scanner/work-ready lane. Preserve independently enforced trust, label-history, human and eligibility gates. Update the existing semantic owner index; retire duplicate population predicates only after parity fixtures prove old supported behavior.

## Interface and completion contract

DoD: deliver the production behavior, regression and fault-injection coverage named below; route every affected caller to the canonical owner; update the stated contract/operator guidance and changelog. A contract declaration or function import alone is insufficient. Consumer routes are prospective authoring obligations, not already-correlated implementation evidence. Keep unknown/unreadable distinct from empty/ready, and retain explicit missing source, external owner and activation holds.

## Verify

These named tests and mutation manifests are FUTURE DELIVERABLES, not authoring-time evidence. Each exercises the actual entrypoint/adapter, with the wrong-but-plausible mutation named in Expect. Implement executable mutation commands in the declared manifest and record fail-first/red/green observations. The named PASS assertion prevents an absent test selector from passing.

| # | Class | Shell | Command | Expect |
|---|---|---|---|---|
| 1 | check:ci +flow +mutation | sh | `(cd deskcore && handoff_out=$(mktemp "${TMPDIR:-/tmp}/handoff-verify.XXXXXX") && trap 'rm -f "$handoff_out"' 0 && GOWORK=off go test -count=1 -v -run '^TestHandoffPolicyParity$' ./issueflow/... > "$handoff_out" && grep -q -- '--- PASS: TestHandoffPolicyParity ' "$handoff_out")` | exit 0 and named PASS; scanner/reader consumer fixtures agree on ordinary trusted repair, addressed role, absent disposition, human hold and unreadable facts; bypassing the shared addressed predicate fails. |
| 2 | check:ci +flow +mutation | sh | `(cd deskcore && handoff_out=$(mktemp "${TMPDIR:-/tmp}/handoff-verify.XXXXXX") && trap 'rm -f "$handoff_out"' 0 && GOWORK=off go test -count=1 -v -run '^TestHandoffModuleImports$' ./issueflow/... > "$handoff_out" && grep -q -- '--- PASS: TestHandoffModuleImports ' "$handoff_out")` | exit 0 and named PASS; actual statusgen and desk-tools module consumer builds with GOWORK=off call the same public library; an internal import, duplicated predicate or hidden process/env dependency fails. |

## Evidence

<!-- Empty by design. A non-implementer records actual command/exit/output, Date and Runner, admitted source, evidence links and remaining coverage. No current test or acceptance result is claimed. -->
**VERIFY: BLOCKED** — both required rows could-not-run; base proof unproven.

| # | Command | Result | Output | Date | Runner | Base |
|---|---------|--------|--------|------|--------|------|
| 1 | `(cd deskcore && handoff_out=$(mktemp "${TMPDIR:-/tmp}/handoff-verify.XXXXXX") && trap 'rm -f "$handoff_out"' 0 && GOWORK=off go test -count=1 -v -run '^TestHandoffPolicyParity$' ./issueflow/... > "$handoff_out" && grep -q -- '--- PASS: TestHandoffPolicyParity ' "$handoff_out")` | could-not-run exit=- sandbox=unshare — check:ci hermetic execution requires a network-off sandbox, unavailable on this host: the network sandbox uses `unshare --net`, a Linux facility, and this host is darwin. check:ci rows are re-executed network-off by design (verdict-lane/02, R-6 c.6) — run on a Linux runner that provides `unshare --net` | sha256:e3b0c44298fc | 2026-10-23 | assay-verifier-app[bot] @ 4a4b8e4f3248 (on-behalf-of human:ian) (forge-identity) | base=254a8fb3d5b1 unproven · head=4a4b8e4f3248 |
| 2 | `(cd deskcore && handoff_out=$(mktemp "${TMPDIR:-/tmp}/handoff-verify.XXXXXX") && trap 'rm -f "$handoff_out"' 0 && GOWORK=off go test -count=1 -v -run '^TestHandoffModuleImports$' ./issueflow/... > "$handoff_out" && grep -q -- '--- PASS: TestHandoffModuleImports ' "$handoff_out")` | could-not-run exit=- sandbox=unshare — check:ci hermetic execution requires a network-off sandbox, unavailable on this host: the network sandbox uses `unshare --net`, a Linux facility, and this host is darwin. check:ci rows are re-executed network-off by design (verdict-lane/02, R-6 c.6) — run on a Linux runner that provides `unshare --net` | sha256:e3b0c44298fc | 2026-10-23 | assay-verifier-app[bot] @ 4a4b8e4f3248 (on-behalf-of human:ian) (forge-identity) | base=254a8fb3d5b1 unproven · head=4a4b8e4f3248 |
Verification-Attestation: medici-finance/assay#4231 run=ca6c0ea6b39dff629c3ce40ed424debeeaa6824ebfdc8d3136b3a8e373713acf source=4a4b8e4f324834ba75aa2f211c14655dda4d4bfd brief=docs/streams/verification-handoff/brief-01-share-the-blocker-disposition-and-routing-contract.md model=gpt-6-astra tier=strong


Independent runner: gpt-6-astra, strong, admitted by #4231. Observed source `4a4b8e4f324834ba75aa2f211c14655dda4d4bfd`; delivery #4225 merged as `9d9f603ef2d23e0eec09c9800a8603f46579b7a9`; fail-first base is its first parent `254a8fb3d5b13187f1fa35bfb3ed158c16fcd899`.

Required rows: 2; executed: 0; passed: 0; failed: 0; could-not-run: 2. No acceptance obligation was observed failing. No closed container witness or base-red/head-green proof exists.

The supported network-off container attempt used released statusgen v1.0.57 and `ghcr.io/medici-finance/verify-runner:v1.0.57@sha256:1b0f3a9a32a452b618fcf6761571af35f33239416482b7f748210f7b772566f3`. It refused before any row with exit 2: `could-not-attribute — no executing identity is available`. This reproduces existing blocker #4208. Host admission and bracket opening succeeded; the receipt did not carry the runner fields needed by the container. No identity was injected and no sandbox was disabled.

Disposition: waiting for release/installation/qualification of the existing runner repair, not another implementation of the handoff policy. Source repair #4214 is merged; source presence is not installed-runner qualification. The companion image Git-floor repair #4210 is also merged and belongs to the existing #4207 release work; the observed container refusal in this run is attribution, not a claimed Git-floor failure.

Next owner/action: the existing runner release and qualification owners carry #4208 through a containing published release, aligned consumer pins, actual installed host/image identity checks, and fresh independent qualification. Then the verify desk reruns both unchanged rows and the delivering-parent fail-first proof through a supported isolated execution path. Wake on that qualified installation, not merely on an issue closing or a source merge. The brief's existing human gate remains unchanged.

## Review

Gate: human, derived from risk. Independent review checks source/consumer contracts, operation compatibility and negative controls; this plan authorizes neither acceptance nor activation.
