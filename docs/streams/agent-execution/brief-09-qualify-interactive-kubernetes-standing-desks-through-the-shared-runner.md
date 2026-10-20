---
brief: assay:assay:agent-execution:09
title: Qualify interactive Kubernetes standing desks through the shared runner
why: Operators need separately interactive Kubernetes desks without a second scheduler or moving privileged credentials into agent workloads.
wave: 9
depends: ["agent-execution/03", "graph-execution/27"]
unblocks: ["desk-containers/28", "agent-execution/08"]
effort: M
gate: human
risk: {regulatory: no, customer: no, irreversible: yes, sensitive-data: yes}
issues: []
schema: brief-v2
version: 1
authored: 2026-10-20 by Assay architecture authoring session
sources:
  - https://github.com/medici-finance/assay/issues/3669
  - docs/proposals/non-ci-architecture/28-interactive-kubernetes-desks.md
  - freshness-checked 2026-10-20 @ d8687cab6 (fetched origin/main; code and open PRs inspected)

exec-tier: strong
exec-tier-why: Cross-process ownership, custody and recovery require design judgment.
domain: complicated
decision-trigger: spec
gate-why: Operator and workload authority cross a process boundary; the human confirms effective credential isolation, current-owner fencing and the independently witnessed profile before adoption.
id: cb59f755-2881-4a8f-8354-ee422db4f3ea
consumers:
  - "loopadmin/adapters/kubernetes/: follow-up agent-execution/09"
  - "loopadmin/supervisor/: follow-up agent-execution/09"
  - "containers/entrypoint.sh: follow-up agent-execution/09"
  - "containers/secrets.md: follow-up agent-execution/09"
  - "loopadmin/adapters/kubernetes verifier profile: follow-up agent-execution/08"
  - "tools/desk/cmd/cellctl/: follow-up desk-containers/28"

---

# Brief 09 — Interactive Kubernetes standing desks through the shared runner

## Context

files: `loopadmin/adapters/kubernetes/` (planned shared API transport and standing profile),
`loopadmin/testdata/kubernetes/` (planned), `loopadmin/supervisor/` (profile binding),
`containers/entrypoint.sh`, `containers/secrets.md`, `docs/kubernetes-desk-profile.md` (planned), `changelog/kubernetes-standing-desks.md` (planned).

facts: Inspected Assay main d8687cab6, 20 October 2026. The runner protocol exists;
the Kubernetes adapter, shared supervisor and manifests remain planned. The current image
preflight requires an App PEM and model credentials. /08 is a distinct candidate Job
profile requiring the workflow controller/publication bridge. Standing roles must not
inherit that requirement or its execution identity. Consume /03 readiness/profile and
/27 standing-desk outputs as actual importable APIs, never cmd/main or tools/desk/internal
from this external module. The execution stream remains parked pending admission.

single-point-of-failure: adapter maps the wrong Pod or credential profile; independent
supervisor generation/stop checks and the existing executor's current-authority checks
reject effects even when transport is bypassed. Effective Pod/network/RBAC isolation is
separately qualified before activation.

layering: One loopadmin Kubernetes API-effect adapter implements runner translation;
shared pure identity/profile checks and API transport are reused by /08. Desk policy,
PostgreSQL journal, effect custody and acceptance retain their existing owners. No
second lifecycle controller, queue, credential broker or scheduler. Verify 1–4 exercise the seam.

design-fit:
  owner: loopadmin runner/supervisor; Kubernetes adapter owns transport translation only
  contract: none — new profile extends spec/loop-admin-runner-v1.md through its existing negotiation
  retires: []
  weight: verbs +0, flags +0 public verbs, refusal predicates +4 (binding, credential profile, old owner, unknown transport); actual text delta recorded at implementation
  why-add: A host-process adapter cannot address a Kubernetes Pod identity. One adapter shares transport with the planned Job profile rather than adding a parallel remote runner.

## Read first

- [Interactive target](../../proposals/non-ci-architecture/28-interactive-kubernetes-desks.md).
- [Runner contract](../../../spec/loop-admin-runner-v1.md).
- [Standing clients](../graph-execution/brief-27-standing-desk-loop-admin.md).
- [Readiness owner](brief-03-local-readiness-enforcement.md).
- [Distinct verifier profile](brief-08-qualify-a-kubernetes-verification-runner-through-the-existing-supervisor.md).

## Human decision

At pickup, present the concrete Pod, credential, transport and network boundary with its
negative-path evidence. Recommended: separate preinstalled role Pods, a trusted transport
identity and credential-free agent/candidate containers. Alternative: retain the explicitly
labelled compatibility images until the managed boundary can be proven. The decision must
name what terminal access and model authentication permit, including their limits.
No response means no adoption. Implementation fixtures do not select a production cluster.

## Ground rules

Isolated branch and draft PR only. No deployment, workflow dispatch, live cluster or model
contact. Fake Kubernetes API/streams and disposable local storage only. Preserve the parked
stream's admission gate. Independent verification owns Evidence. Existing budgets, effect
and human gates remain in force. Refresh claims/active PRs before implementation.

## Task

1. Add the standing profile to the existing runner/supervisor API. Bind cell, role, namespace,
   Pod UID, container, owner incarnation/generation, qualified image digest, profile and
   preparation receipt. Installation pre-creates the declared workloads; normal launch
   cannot submit arbitrary Pod/service-account/mount specifications. Inject an explicit
   scoped API credential into the trusted adapter only; missing target/opt-in/credential
   causes zero API calls. Print the non-secret target before first contact. No ambient
   kubeconfig or in-cluster credential discovery fallback.
2. Implement start/reconcile, inspect, attach/reconnect and control using the existing
   shared journal and role leases. Typed terminal streams preserve input/output, resize,
   detach and bounded reconnect. Attachment never starts the model or becomes a lease.
   Repeated up, concurrent operators and a timed-out request reconcile the same role
   generation. Pod replacement and partitioned-node old execution hold until the old
   execution group is confirmed stopped/fenced; DELETE acceptance or TTL alone is insufficient.
   Resume does not renew an exhausted budget or clear an unrelated stop/decision.
3. Add a **named managed boot profile** to the image preflight, consuming /03's validated
   binding. Retain legacy PEM-required behavior for compatibility consumers. Managed agent
   Pods receive no operator/cluster or forge-publisher credentials, disable service-account
   automount, and cannot obtain the adapter's credential or call operator lifecycle APIs.
   Bind model authentication through the already qualified custody owner; if unavailable,
   hold rather than invent a broker or mount an App PEM. Missing/unknown profile refuses.
   Document actual enforcement and residual trust; never infer isolation from containers.
4. Factor Kubernetes identity, API observation and unknown-outcome transport primitives so
   /08 consumes them from this package. Keep Job and standing profile allowlists distinct:
   a verifier cannot acquire attach/exec or a standing credential, and a standing role
   cannot acquire publication authority or arbitrary Job admission. Preserve the stricter
   /08 credential/network floor and its controller dependencies.
5. Exercise real production adapter and supervisor code against hostile fake API/stream
   fixtures: lost reply, duplicate delivery, API unavailable, stale UID/generation, wrong
   role, concurrent reconnect, unknown cancellation and permission denial. Carry current
   perf-v3 parent/child identities and preserve unknown usage. Idle and reconnect cause zero
   extra model calls. Regenerate the owning profile documentation and image compatibility
   matrix. Add path triggers for every module/container file the checks read.

## Interface contract

Standing mode carries a desk binding, not fabricated workflow/node fields. Immutable
invocation and execution identity enter; sanitized lifecycle/stream receipts leave. Transport
liveness, process failure, retained files, session continuity and accepted work are separate.
Unknown observations hold. Existing conformance and extension negotiation remain mandatory.

## Verify

Tests below are implementation deliverables, not existing PASS claims. Each named test must
run (zero skips) and fail under its stated mutation. Run from repository root.

| # | Class | Command | Expect |
|---|---|---|---|
| 1 | check +flow +mutation | `(cd loopadmin && o=$(mktemp) && GOWORK=off go test -count=1 -v ./adapters/kubernetes/... -run '^TestStandingIdentity$' > "$o" && grep -F -- '--- PASS: TestStandingIdentity ' "$o")` | exit 0, named PASS; actual supervisor/adapter maps all five roles; duplicate up/attach makes one owner and zero extra model calls; accepting another Pod UID makes it fail |
| 2 | check +flow +mutation | `(cd loopadmin && o=$(mktemp) && GOWORK=off go test -count=1 -v ./adapters/kubernetes/... -run '^TestStandingRecovery$' > "$o" && grep -F -- '--- PASS: TestStandingRecovery ' "$o")` | exit 0, named PASS; lost reply/partition/unknown stop holds replacement; paused state, effects and budgets survive transport/operator restart; treating delete submission as stopped fails |
| 3 | check +flow +mutation | `(cd loopadmin && o=$(mktemp) && GOWORK=off go test -count=1 -v ./adapters/kubernetes/... -run '^TestKubeProfileIsolation$' > "$o" && grep -F -- '--- PASS: TestKubeProfileIsolation ' "$o")` | exit 0, named PASS; wrong credential/profile and missing opt-in make zero calls, agent cannot reach operator/executor custody with upper caller bypassed, verifier never gains terminal authority; swapping profile capability fails |
| 4 | check +flow +neighbour | `(cd loopadmin && GOWORK=off go test -count=1 ./runner/conformance/... ./adapters/kubernetes/...)` | exit 0, no skipped required profile cases; shared runner negotiation, existing local behavior and legacy image preflight preserved; actual image compatibility checks included by adapter tests |

## Pre-mortem

Wrong Pod or duplicate role owner: row 1. Lost stop mistaken for cancellation: row 2.
Cross-profile authority or ambient credential fallback: row 3. Compatibility regression:
row 4. Effective cluster containment is review plus /29's separate operator-run qualification;
fake API success alone does not establish it.

## Evidence

<!-- Independent verifier records commands, revision, profile and explicit limits. -->

## Review

Human gate. Review the independent controls, import boundaries and actual credential/model
profile. Offline fixture success does not authorize deployment or unpark the stream.
