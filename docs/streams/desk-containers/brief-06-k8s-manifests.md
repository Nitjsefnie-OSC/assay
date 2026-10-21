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

Verified against merged main 3cea6aa621ec0364fb7a6cbd1196141829ba6c4c (detached verifier worktree), darwin/arm64, python 3.15.0, go 1.27.2, statusgen v1.0.54; the checker itself asserted kustomize v5.8.3, kubeconform v0.8.0 and gitleaks 8.30.1 before rendering. KUBECONFIG=/dev/null for every row; no cluster contact. Delivery: merge commit 099de97604b0075229ff96b697534ccee07f0477 (PR 3729), first parent 2fe90aa47808766df70f0b75161a0920de50d209; no later main commit touches the k8s package or docs/docker.md.
Verification-Attestation: medici-finance/assay#3844 run=987389ee03deca992750fa0122cc04d3485b9ab3fe97781e5bdd31996f0735c4 source=3cea6aa621ec0364fb7a6cbd1196141829ba6c4c brief=docs/streams/desk-containers/brief-06-k8s-manifests.md model=claude-opus-5-5 tier=strong

| # | Command | Expect | Observed | Date | Runner |
|---|---|---|---|---|---|
| 1 | `sh containers/k8s/check.sh --base` | exit 0; five distinct interactive role workloads/PVCs, pins, identity; no API calls or credential Secret | exit 0; "checked-clean: 9 schemas; five distinct interactive role workloads/PVCs" | 2026-10-21 | assay-verifier-app[bot] @ 3cea6aa621ec0364fb7a6cbd1196141829ba6c4c (claude-opus-5-5 verifier) (on-behalf-of human:ian) |
| 2 | `sh containers/k8s/check.sh --base --self-test` | exit 0; every hostile fixture rejected (role, tty, PVC, digest, token isolation, example Secret) | exit 0; 51 "rejected:" lines incl. missing-role, duplicate-role, no-tty, no-pvc, mutable-image, agent-sa-token, missing-token-isolation, example-secret-rendered, one-workload-all-roles, terminal-service, absent-ingress-deny, shared-pem-secret; final "checked-clean: 51 hostile fixtures rejected" | 2026-10-21 | assay-verifier-app[bot] @ 3cea6aa621ec0364fb7a6cbd1196141829ba6c4c (claude-opus-5-5 verifier) (on-behalf-of human:ian) |
| 3 | `sh containers/scripts/layer-secret-scan.test.sh` | exit 0; image credential-layer protections remain | exit 0; 53 PASS lines, 0 FAIL lines; last "PASS labels: separate file after benign match" | 2026-10-21 | assay-verifier-app[bot] @ 3cea6aa621ec0364fb7a6cbd1196141829ba6c4c (claude-opus-5-5 verifier) (on-behalf-of human:ian) |
| 4 | `test -n "$ASSAY_DELIVERY_SHA" && test -n "$ASSAY_DELIVERY_BASE_SHA" && statusgen --root . --consumers --brief desk-containers/06 --delivery "$ASSAY_DELIVERY_SHA" --base "$ASSAY_DELIVERY_BASE_SHA"` with ASSAY_DELIVERY_SHA=099de97604b0075229ff96b697534ccee07f0477, ASSAY_DELIVERY_BASE_SHA=2fe90aa47808766df70f0b75161a0920de50d209 | exit 0; exact counts, 0 disproved | exit 0; "summary: 3 corroborated, 0 disproved, 0 unchecked, 0 brief(s) claiming nothing" (containers/k8s/, docs/docker.md, cli-migration.json all CORROBORATED); omitted-consumer review: the delivery also changed the changelog fragment, proposal 28 and brief 29, which are documentation of this item, not consumers | 2026-10-21 | assay-verifier-app[bot] @ 3cea6aa621ec0364fb7a6cbd1196141829ba6c4c (claude-opus-5-5 verifier) (on-behalf-of human:ian) |
| 5 | `sh containers/k8s/check.sh --source-secrets --self-test` | exit 0; all manifests and unrendered examples scanned; excluded-example synthetic credential rejected | exit 0; "checked-clean: 25 package sources scanned, including unrendered examples"; "rejected: excluded-example-credential"; "rejected: excluded-example-base64" | 2026-10-21 | assay-verifier-app[bot] @ 3cea6aa621ec0364fb7a6cbd1196141829ba6c4c (claude-opus-5-5 verifier) (on-behalf-of human:ian) |

Rows: 5 of 5 passed; every row executed.

Risk-bearing values (gate model; risk answers all "no", not irreversible). Enumerated over the delivered package: replicas 1, persistentVolumeClaimRetentionPolicy whenDeleted/whenScaled Retain, runAsUser/runAsGroup/fsGroup 999, image tag v1.0.52 plus sha256 digest per role, secret defaultMode 0440, requests cpu 250m / memory 512Mi, limits cpu 2 / memory 4Gi, PVC storage 20Gi (each at the same line in all five role manifests: lines 7, 11-12, 30-32, 38, 72, 57-61, 83); DNS egress port 53 in the base network policy line 33; checker pins kustomize v5.8.3, kubeconform v0.8.0, gitleaks 8.30.1 (check.py lines 18-19), schema kubernetes-version 1.34.1 (check.py line 135), subprocess timeout 90 (check.py line 38). Ranked: retention policy and secret mode first (work-volume loss and credential readability are not undone by a redeploy), then image digest and UID, then the resource/storage/timeout knobs (reversible, no derivation needed).

- RISK-VALUE: DERIVED — whenDeleted = Retain (and whenScaled = Retain) @ containers/k8s/base/verify-desk.yaml:11-12 (same lines in all five) — the brief requires independently retained work volumes; Delete would destroy a desk's /work PVC when its StatefulSet is removed or scaled down, which no redeploy restores, so Retain is the only value that meets the contract.
- RISK-VALUE: DERIVED — defaultMode = 0440 @ containers/k8s/base/verify-desk.yaml:72 (same line in all five) — the PEM is read by the runAsUser 999 process with fsGroup 999; 0440 gives owner and group read only, no write or execute and no other-read, the narrowest mode that still lets the desk process read the key.
- RISK-VALUE: DERIVED — replicas = 1 @ containers/k8s/base/verify-desk.yaml:7 (same line in all five) — the brief fixes one interactive owner per role with its own RWO work volume; a second replica would be a second, unadmitted owner of the same role.
- RISK-VALUE: NAMED, NOT DERIVED — runAsUser/runAsGroup/fsGroup = 999 @ containers/k8s/base/verify-desk.yaml:30-32 (same in all five) — the base image Dockerfile creates the desk user with useradd --system and no fixed UID, so 999 depends on allocation at build time; proving it for the pinned v1.0.52 digests needs those images, which are not in the local image store and would need a registry pull. Supplemental probe only (no acceptance credit): a locally cached v1.0.46 verify-desk image run with no network reports uid=999(desk) gid=999(desk). A fixed UID in the Dockerfile would remove this dependency.
- RISK-VALUE: NAMED, NOT DERIVED — image = verify-desk:v1.0.52@sha256:906d3487ae40858ca527af023b1f7443c6fc52d3c3afa771f9c6ba3c2e15423c @ containers/k8s/base/verify-desk.yaml:38 (and the four other role digests at line 38) — no record in the tree binds these digests to the v1.0.52 release build; confirming them needs a registry manifest lookup, which is outside this offline envelope.
- RISK-VALUE question filed: #3857 (desk UID/GID and role image digest provenance; the brief advances, the derivations stay open there).

**VERIFY: PASS** — 5/5 rows passed against 3cea6aa621ec0364fb7a6cbd1196141829ba6c4c; UID and image-digest values named, not derived.

## Review

Gate: model, unchanged risk answers: generic offline templates only, no custody redesign.
Review all structured negative controls and the honest compatibility/managed distinction.
