---
brief: assay:assay:desk-containers:06
title: Kubernetes manifests for the five desks
wave: 3
depends: ["desk-containers/02", "desk-containers/03"]
unblocks: ["desk-containers/12", "desk-containers/13", "desk-containers/14", "desk-containers/27", "desk-containers/28"]
effort: M
gate: model
risk: {regulatory: no, customer: no, irreversible: no, sensitive-data: no}
issues: []
schema: brief-v2
authored: 2026-08-22 by desk-containers scoping session
sources:
  - "medici-finance/assay#63 — the request (secondary aim: launched via k8s)"
  - "docs/streams/desk-containers/spec.md — secondary launch targets"
  - "containers/secrets.md (brief 02) — Secret volume + envFrom recipe"
  - "freshness-checked 2026-08-22 @ b3a2067 — no k8s manifests exist in the tree"
why: >-
  The second half of the request's secondary aim: run the same desk images on a
  cluster, with per-desk persistent volumes and credentials from k8s Secrets, so a desk
  can live on shared infrastructure instead of one person's desktop.
version: 5
id: 9ff1192a-c441-4dd3-b574-85b8cf85954e
exec-tier: strong
exec-tier-why: Per-role persistence, credential profiles and offline validation cross artifacts.
consumers:
  - "containers/k8s/: follow-up desk-containers/06"
  - "docs/docker.md: follow-up desk-containers/06"
  - "docs/streams/desktools-v2/cli-migration.json: fixed-here"
outcome: none

---

# Brief 06 — Kubernetes manifests for separately interactive desks

## Context

files: `containers/k8s/` (planned namespace, five StatefulSets/PVCs, Kustomize base, placeholder-only secret example and `check.sh`), `docs/docker.md`, `docs/streams/desktools-v2/cli-migration.json` (exact test-class routes for the source-only checker and YAML adapter), `changelog/kubernetes-desk-manifests.md` (planned).

facts: Assay main d8687cab6 inspected 20 October 2026; the five desk images exist but
containers/k8s does not. Each configured desk gets its own Pod/container and PVC at /work,
replicas 1, stdin/tty enabled, pinned image digest and explicit namespace/container name.
Storage class and deployment values belong to an adopter overlay. A Running Pod is not an
admitted role owner; the managed launch owner is graph-execution/26–27 plus agent-execution/09.

layering: Kustomize packages existing role images; offline structured render/schema checks
validate packaging. Runtime ownership and credential policy remain in their existing owners.
No controller or process supervisor is introduced by a manifest.

design-fit:
  owner: existing image/profile owners; Kustomize is packaging
  contract: containers/secrets.md compatibility contract; managed profile follows agent-execution/09
  retires: []
  weight: verbs +0, flags +0 runtime, refusals +0 runtime policy, packaging files +7 plus check fixtures
  why-add: Kubernetes needs declarative workload packaging; one shared base and opt-in profile overlays avoid separate launch engines.

## Read first

- [Current target and migration routing](../../proposals/non-ci-architecture/28-interactive-kubernetes-desks.md).
- [Credential contract](../../../containers/secrets.md) and [historical spec](spec.md).

## Ground rules

Branch/draft PR only; no workflow, deployment or live API contact. Render with pinned local
Kustomize and kubeconform plus cached schemas. Never run kubectl apply, including a claimed
client-only dry run, as offline validation. Stop at implemented; independent verifier owns
Evidence. No credential value in manifests, layers or examples.

## Task

1. Package namespace and one StatefulSet/PVC per configured current desk, stdin/tty true,
   pinned digest, declared resources, non-root, no privilege escalation, dropped capabilities
   and explicit service-account token automount disabled for agent workloads. Require
   default-deny ingress and no Service/Ingress exposing desk terminals. Compatibility
   egress is explicitly allowed only to its qualified forge/model endpoints and DNS;
   managed /09 owns its stricter API/metadata/CNPG/direct-forge denial. Preserve role
   names/image compatibility and independently retained work volumes. No hostPath, privileged
   container, Docker socket or hidden dependency on a desktop multiplexer.
2. Document the **compatibility** profile's runtime Secret references and model configuration
   under the existing human-gated /02 contract. Secret examples contain placeholders only and
   are excluded from Kustomize resources; a build produces no credential Secret. Explicitly
   disclose that App-PEM-bearing legacy images do not satisfy the managed target's isolation.
   /09 and /28 own managed boot/overlay changes; do not pretend a manifest removes that custody.
   Keep each independent role's App-PEM and model credential Secret references distinct;
   the checker rejects a shared credential Secret between role workloads. Compatibility
   terminal access and container session logs are transitive key custody and cannot use
   the /28 managed operator. Document equivalent custody restrictions for `pods/log`,
   node log files/shipping and downstream log readers; shared logging cannot widen access.
3. Supply offline render/schema and structured checks plus hostile manifest fixtures. Reject
   one workload hosting all roles, absent/duplicate role, missing PVC/tty, mutable image,
   privileged/host mounts, agent service-account credential or example Secret in the build.
   Checks must fail on each planted defect and make zero API/discovery requests. Extend
   the same checker with `--source-secrets` to scan every committed manifest/example,
   including unrendered Secret examples, using the existing pinned credential scanner.
   Keep a synthetic positive control that is excluded from rendering yet still rejected.
   Reject terminal Service/Ingress exposure, absent ingress deny and shared per-role
   credential Secret references. The repository-wide credential scan remains independent.
4. Update docs/docker.md with the operator-installed apply/attach workflow, explicit Pod and
   container selection, storage/Secret/profile limits and disconnect behavior. Distinguish
   attaching a retained session from creating a process. Link the future in-cluster operator
   /28; no Windows, compose, CronJob or workflow-controller prerequisite for this package.

## Verify

The script and fixtures below are implementation deliverables; checks are entirely offline.

For the consumer row, the independent verifier obtains the actual implementation delivery
from the merged PR, records its full immutable SHA in `ASSAY_DELIVERY_SHA`, and records that
commit's actual first-parent SHA in `ASSAY_DELIVERY_BASE_SHA`. Neither value defaults to
HEAD, origin/main or this authoring PR. Both commits must be available in the verification
checkout; absent/unknown provenance fails closed. Retain exact consumer result counts and
resolve unchecked required consumers separately; zero disproofs alone is not complete flow
coverage. These inputs are future implementation-verification evidence, not claims made here.

| # | Class | Command | Expect |
|---|---|---|---|
| 1 | check +flow +dereference | `sh containers/k8s/check.sh --base` | exit 0; pinned render/cached-schema validation and structured assertions prove five distinct interactive role workloads/PVCs, image pins and explicit identity; no API calls or credential Secret in rendered output |
| 2 | check +mutation | `sh containers/k8s/check.sh --base --self-test` | exit 0; checker reports every hostile fixture rejected; removing a role, tty, PVC, image digest or token isolation fails; fixture containing an example Secret fails |
| 3 | check +neighbour | `sh containers/scripts/layer-secret-scan.test.sh` | exit 0; current image credential-layer protections remain; coordinate any merged scan replacement rather than restoring a retired scanner |
| 4 | check +flow | `test -n "$ASSAY_DELIVERY_SHA" && test -n "$ASSAY_DELIVERY_BASE_SHA" && statusgen --root . --consumers --brief desk-containers/06 --delivery "$ASSAY_DELIVERY_SHA" --base "$ASSAY_DELIVERY_BASE_SHA"` | exit 0 only for the named landed delivery and its first parent; record exact corroborated/disproved/unchecked counts, 0 disproved; unchecked required consumers stay held and omitted consumers remain a review obligation |
| 5 | check +mutation | `sh containers/k8s/check.sh --source-secrets --self-test` | exit 0; all committed manifests and unrendered examples are scanned; excluded-example synthetic credential is rejected independently of rendering |

## Definition of Done

The package and docs implement the declared compatibility scope, all offline rows are
independently witnessed and managed isolation remains explicitly unqualified until /09,/28,/29.
Human acceptance of /02 is not replaced by this model-gated packaging review.

## Evidence

<!-- Independent verifier records commands, exact versions, profile and observations. -->

## Review

Gate: model, unchanged risk answers: generic offline templates only, no custody redesign.
Review all structured negative controls and the honest compatibility/managed distinction.
