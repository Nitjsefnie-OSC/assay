---
brief: assay:assay:cell-api:05
title: Package the deskd image and generic Kubernetes service profile
why: "A desk workload image and an operator image do not package the service. A separate public deskd image and generic Kubernetes profile make its read, health and network boundaries inspectable before any deployment."
wave: 14
depends: ["cell-api/03", "desk-containers/27"]
unblocks: ["cell-api/06"]
effort: M
gate: human
risk: {regulatory: no, customer: no, irreversible: no, sensitive-data: yes}
gate-why: "Confirm the service deployment profile, admitted read capability, effective network policy and health/recovery behavior. Operator-selected workload and recovery targets must exist before real-profile exposure."
issues: []
schema: brief-v2
version: 1
id: 682f451b-986d-442c-97d7-acf2f8f4e70b
authored: 2026-10-20 by cell-api authoring
decision-trigger: spec
exec-tier: strong
exec-tier-why: Cross-component compatibility, privacy and authority boundaries require design judgment and independent negative controls.
domain: complicated
sources:
  - docs/streams/cell-api/README.md
  - docs/proposals/non-ci-architecture/11-target-architecture.md
  - docs/proposals/non-ci-architecture/16-delivery-sequence.md
  - docs/proposals/non-ci-architecture/20-amendment-manifest.md
  - docs/proposals/non-ci-architecture/21-cell-operator-surface.md
  - docs/proposals/non-ci-architecture/22-forge-service-and-source-queries.md
  - docs/proposals/non-ci-architecture/23-runtime-contract-alignment.md
  - docs/proposals/non-ci-architecture/28-interactive-kubernetes-desks.md
  - https://github.com/medici-finance/assay/pull/2890
  - https://github.com/medici-finance/assay/issues/3669
  - freshness-checked 2026-10-20 @ 15b60414f3ec134942a0ec3162b550cd95cbbc8e
consumers:
  - "images/deskd/: follow-up cell-api/05"
  - "containers/k8s/components/cell-api/: follow-up cell-api/05"
  - "tools/desk/internal/cellapi/: follow-up cell-api/05"
  - "tools/desk/testdata/cellapi/: follow-up cell-api/05"
  - "docs/cell-api-kubernetes.md: follow-up cell-api/05"
  - "changelog/cell-api-image.md: follow-up cell-api/05"
  - "legacy source consumers: out-of-scope (private source owner retains their exact crosswalk; this brief records public dispositions and does not edit them)"
---

# Brief 05 — Package the deskd image and generic Kubernetes service profile

## Context

files: `images/deskd/`, `containers/k8s/components/cell-api/`, `tools/desk/internal/cellapi/`, `tools/desk/testdata/cellapi/`, `docs/cell-api-kubernetes.md` (planned), `changelog/cell-api-image.md` (planned)

facts:
- pin: public main 15b60414f, inspected 2026-10-20; images/deskd and service component are planned.
- separation: the service, dedicated Linux operator and five role Pods have distinct capabilities and lifecycle.
- storage: desk-containers/27 owns the CNPG profile; no new SQLite runtime or mirror database is introduced.
- limits: no production latency, density, RTO or RPO target is established by authoring.

layering: A service host is justified by read-side lifecycle and custody separation; it composes 03 and existing R6, never agent/operator machinery. Task 1–4 and Verify 1–3 inspect rendered and running disposable fixtures.

single-point-of-failure: the service's claimed mapping from caller and subject to authorized result; independent store/executor enforcement and literal fixture oracles back that mapping. Verify bypasses the upper adapter and checks the owning boundary separately; a transport allowlist alone is insufficient.

design-fit:
  owner: R7 service host packaging; desk-containers owns the operator/desk and database profiles
  contract: none — R7's transport responsibility is registered by cell-api/01; existing semantic owners and contracts remain controlling
  retires: []
  weight: verbs +0; flags +0; refusals +0; rule-text +0 (no new operator vocabulary proposed; record any actual implementation delta for review)
  why-add: Reuse the existing owners and one client seam; remove duplicate service composition only after the last-consumer proof. Package/process additions are bounded by the declared files, not permission to add another controller.

All planned paths and named tests below are implementation deliverables. The source pin
proves the starting point, not that these deliverables exist or were tested. The
[stream refactor map](README.md#bounded-refactor-map) defines the public transfer boundary.

## Read first

Read the [R7 ownership and admission record](README.md#authority-and-admission),
[cell operator contract](../../proposals/non-ci-architecture/21-cell-operator-surface.md),
[forge-service boundary](../../proposals/non-ci-architecture/22-forge-service-and-source-queries.md),
[interactive Kubernetes direction](../../proposals/non-ci-architecture/28-interactive-kubernetes-desks.md),
`spec/cell-op-v1.md`, `docs/contracts.md` and `docs/go-engineering.md`.
Resolve every declared dependency's current contract and independent Evidence before
pickup. The package mapping is a reviewable proposal; a conflicting contract is
NEEDS_CONTEXT, never permission to select new custody or duplicate its owner.

## Human decision

Confirm the service deployment profile, admitted read capability, effective network policy and health/recovery behavior. Operator-selected workload and recovery targets must exist before real-profile exposure.

Options:
1. **Accept the demonstrated bounded design** — accept the evidenced compatibility and containment for its declared scope; provisioning, deployment and source retirement retain their separate operator acts.
2. **Hold for correction** — name the missing proof or incompatible behavior; retain the current supported path until corrected.

Default if no answer: none

The decision becomes concrete from the implementation design and evidence. No answer
keeps acceptance pending; it grants no time-based default or activation authority.

## Ground rules

- Work in an owned branch; no merge, ready flip, release, installation or live infrastructure contact.
- Public source and fixtures must be self-contained and synthetic. Never copy private locators, configuration, credentials, incidents or history into Assay.
- Preserve the single R2 PostgreSQL runtime, R3 isolation, R4 role authority and R6 forge mirror. A service, terminal, restored file or cached result grants no authority.
- Keep this stream parked until admitted. Stop at implemented; independent verification, human acceptance and operational cutover remain distinct.

## Task

1. Build a publication-clean standalone deskd image from the public source and pinned supported build inputs; no copied private build history, host paths or baked credentials. Keep the runtime minimal, non-root and separate from the model/desk/operator images. Declare the artifact pin and supported service/client matrix; publishing an image remains a separate release act.

2. Create the generic offline-renderable Kubernetes service component with explicit cell/source bindings, default-deny network policy and named destinations. No privileged host mounts, Docker socket, wildcard RBAC, ambient kubeconfig or agent/operator credential mounts. A Service may expose the qualified read API only within its admitted profile; no public Ingress/default external exposure. Agent Pods cannot use the service as a route to SQL or publishers.

3. Wire only the existing R6 reader and approved R2 projection/storage access. Reuse the CNPG owner for database profile; no second mirror or adopted legacy SQLite coordination. Distinguish process live, dependencies ready, projection freshness and acceptance: a live process with stale/unavailable authorization or data must report degraded readiness/coverage. Probe responses contain no private counts.

4. Qualify restart, graceful termination, SSE resync and dependency outage with local disposable processes and offline render checks. A Pod restart or reconnect cannot launch agents, renew authority or replay effects. Scope retry/load behavior to inherited forge budget and perf-v3 observations; add no invented production budget. Record a real-profile acceptance procedure using explicit endpoint/TLS-bound owned configuration; this brief does not execute a live cluster.

5. Document image/profile pins, secret-reference ownership, network/health matrix and rollback; regenerate examples with offline tools. Keep actual policy/custody proof for the existing Kubernetes qualification and 07 handover.

## Interface and acceptance contract

Consume cell-api/01's versioned HTTP/JSON and SSE seam and the existing cell-op-v1
result/plan meanings. Authentication, current authorization, provider custody and
acceptance are separate. Required coverage/freshness, exact subject and unknown outcomes
survive every adapter. No agent-facing `desk__answer` or `desk__adopt`; no generic role,
URL, SQL-write, token or apply escape hatch. Human named actions remain closed and
separately qualified. Initial service exposure is read/plan-only; approval-based apply
stays absent until amendment 21's five R7/R3 properties are independently qualified.

Update the declared documentation and generated schema/client/profile examples affected
by the Task. Preserve prior Evidence and source history; unresolved consumer or contract
coverage is a hold. Do not invent deployment performance/recovery thresholds: reuse the
inherited workload and require the operator to select missing targets before exposure.

## Verify

These exact named tests must be delivered and run, with a recorded failing control
before the restored pass. Each uses real production composition and an independently
specified expected result; zero selected tests or a skipped required fixture is failure.
Only explicit disposable fixtures may open local listeners/databases. No production or
ambient kubeconfig is permitted. Real-profile qualification remains a separate operator
act; offline passes do not establish it.

| # | Class | Command | Expect |
|---|---|---|---|
| 1 | check:ci +dereference +mutation | `(cd tools/desk && o=$(mktemp) && GOWORK=off go test -timeout 90s -count=1 -v -run '^TestDeskdProfileRender$' ./internal/cellapi/... > "$o" && grep -F -- '--- PASS: TestDeskdProfileRender ' "$o")` | exit 0; named PASS. Actual image/build inputs and rendered component contain no credentials, wildcard access or operator/agent privileges; planted secret mount or second SQLite runtime fails. |
| 2 | check:ci +flow +mutation | `(cd tools/desk && o=$(mktemp) && GOWORK=off go test -timeout 90s -count=1 -v -run '^TestDeskdHealthStates$' ./internal/cellapi/... > "$o" && grep -F -- '--- PASS: TestDeskdHealthStates ' "$o")` | exit 0; named PASS. Real service health distinguishes alive, ready and stale/unavailable projection; private counts are absent and injected false-ready behavior fails. |
| 3 | check:ci +flow +mutation | `(cd tools/desk && o=$(mktemp) && GOWORK=off go test -timeout 90s -count=1 -v -run '^TestDeskdRestart$' ./internal/cellapi/... > "$o" && grep -F -- '--- PASS: TestDeskdRestart ' "$o")` | exit 0; named PASS. Real disposable service restart/disconnect/resync under dependency failure creates zero dispatch/effect/lease renewal; old cursor and inaccessible cache fail safely. |
| 4 | check:ci +dereference | `statusgen --root . --consumers --brief cell-api/05` | exit 0; routing corroborates the implementation diff; exclusions remain review judgments. |
| 5 | check | `statusgen --root . --lint` | exit 0; no PROBLEM; no generated lifecycle cell or historical Evidence hand-edited. |

## Threat model and detection

| Failure mode | Caught by |
|---|---|
| Service image/profile gains operator powers or private build residue | row 1 |
| Healthy process is reported as healthy data/acceptance | row 2 |
| Restart/reconnect becomes new runtime ownership or repeats work | row 3 |
| A fixture is mistaken for deployment authority or a private inventory is called complete without its owner | No row; review-only: inspect decision provenance, source-owner receipt and explicit operational limits before acceptance. |

## Evidence

<!-- Independent verifier appends one row per Verify item: exact subject/source revision,
command, exit, named result, failing control and restored pass, environment, date and
runner. Empty at authoring; no implementation or deployment result is asserted. -->

## Review

Gate: human. Review the actual mounted/effect boundary and retained ownership, the
compatibility inventory and the independent negative controls. Test success grants
neither deployment nor credential provisioning, source closure or a human decision.
