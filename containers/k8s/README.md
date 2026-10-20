# Interactive Kubernetes desk package

This Kustomize base packages the **compatibility profile** of the five existing
role images. It is an offline-reviewed template, not a cluster installation or
an admitted managed cell. See [operator instructions](../../docs/docker.md#kubernetes-compatibility-profile)
and the [current architecture](../../docs/proposals/non-ci-architecture/28-interactive-kubernetes-desks.md).

Each role has one StatefulSet, one interactive container and one independently
provisioned `work-<role>-0` PVC. The base renders nine objects: one Namespace, one
unprivileged ServiceAccount, five StatefulSets and two NetworkPolicies. The five
PVC templates are inside their respective StatefulSets; Kubernetes creates the
claims. There is no Service, Ingress, operator workload or RBAC grant. The
StatefulSet `serviceName` supplies a stable identity field; no headless Service
is created, and this package does not rely on stable Pod DNS.

The image entrypoint still requires its App PEM and model credentials. Terminal
access therefore confers transitive custody of those credentials. This is not
untrusted-candidate isolation, and it cannot be attached through the future
managed operator. Human acceptance of desk-containers/02 is still required;
packaging review does not supply it. Agent-execution/09 owns managed boot and
isolation; desk-containers/28 owns the managed operator; /29 owns real-profile
qualification. A Running Pod does not acquire a role lease or authorize effects.

## Before installation

Supply an adopter-owned overlay and review its complete render.
[`overlays/compatibility-example/`](overlays/compatibility-example/) shows the
opt-in storage-class patch and endpoint policy shape using reserved documentation
addresses. It is never part of the base and is not a deployable configuration;
replace every placeholder in an adopter-owned copy before installation. The base's
external egress is **closed**. It permits only UDP/TCP 53 to selected kube-dns
Pods in kube-system. An operator must qualify the CNI and DNS selection (including
NodeLocal DNS if used), then add an explicit allowlist for **only** the intended
forge/model endpoints. Standard NetworkPolicy uses addresses, not DNS names;
maintain exact endpoint addresses, or use a separately qualified egress gateway
that enforces destination names. Do not replace this with an Internet-wide rule.
The base checks do not qualify an adopter's overlay or effective CNI behavior.

Other operator values are storage class, storage size, resource requests/limits,
namespace if customized, and the ten distinct credential Secret names. The
initial 20Gi, 250m/512Mi requests and 2 CPU/4Gi limits are packaging defaults,
not measured capacity promises. Supply an approved storage class with suitable
retention, backup and access controls. StatefulSet deletion and scale-down
retain PVCs; namespace deletion or explicit PVC deletion can still destroy data.

The process uses numeric UID/GID 1000, with fsGroup 1000 granting write access to
`/work` and group read on the 0440 Secret volume. No privileged ownership-fix init
container is used. Validate CSI fsGroup handling and preflight readability before
admitting a profile. Kubernetes Secret files otherwise default to root ownership;
0400 alone would be unreadable to this non-root process. The mount remains
read-only, and only `app.pem` is projected.

`secrets.example.yaml` contains placeholders for all ten Secrets and is excluded
from the build. Do not apply or populate that tracked file. Provision real Secrets
out of band using the [credential contract](../secrets.md); independently scoped
roles must not share a Secret reference **or actual credential value**. Static
checks can prove distinct references, not distinct key material inside live
Secrets. The model Secret carries the selected runtime model credential, endpoint
and model configuration. The generic examples use API-key mode; file-based WIF or
Vertex profiles need separate, reviewed runtime Secret mounts per that contract.

Pre-provision the retained work volume with the approved pinned agent CLI and
required work inputs. The legacy first-run npm installer is not run by these
manifests, and the forge/model/DNS network floor does not permit arbitrary package
registries. A retained shell does not install or start the agent automatically.

## Offline checks

From the repository root:

```sh
sh containers/k8s/check.sh --base
sh containers/k8s/check.sh --base --self-test
sh containers/k8s/check.sh --source-secrets --self-test
```

Required installed tools: Kustomize **v5.8.3**, kubeconform **v0.8.0**, gitleaks
**8.30.1** (the existing repository scanner pin), Python 3.9+ and the repository's
Go toolchain with its existing `gopkg.in/yaml.v3 v3.0.1` module cached. No tool is
installed by the checker. The small Go YAML-to-JSON adapter runs from the existing
desk module; Python's standard library performs the package assertions. No new
parser or package dependency is introduced. The scanner uses the repository's
existing `.gitleaks.toml`, redacts output, ignores inline allow directives and
also scans decoded Secret.data. Its staged positive controls prove that an
example omitted from Kustomize remains in the source inventory.

The checker sets `GOPROXY=off`, `GOSUMDB=off`, `GOTOOLCHAIN=local` and
`KUBECONFIG=/dev/null`. It preflights the base to admit only the seven local
resource files before calling Kustomize; remote bases, plugins, generators and
Helm are outside this base. Kubeconform receives only the committed, checksummed
local schemas, without a default/HTTP fallback. Missing tools/caches report
`could-not-check` and exit 2; rejected manifests or leaked credentials exit 1;
only completed clean checks exit 0. No kubectl or discovery call runs. These
checks prove packaging; runtime behavior, endpoint reachability, effective policy,
model usage and recovery remain for qualification.

The package source scan is independent of the repository-wide credential scan.
It scans the whole tracked/unignored package, including unrendered examples and
nonstandard filenames. It does not rely on a successful render. `testdata/mutations.py`
is the committed fail-first recipe: every planted invariant defect must be
rejected by the structured check, a wrong Kubernetes field type must be rejected
by kubeconform, and excluded synthetic credentials must fail the real scanner.

Tool/schema dependency assessment: [issue #3725](https://github.com/medici-finance/assay/issues/3725).
Schema source and upgrade procedure: [schemas/README.md](schemas/README.md).

## Image provenance

The five manifests consume the `linux/amd64` immutable role references from
[release v1.0.52](https://github.com/medici-finance/assay/releases/tag/v1.0.52),
`release-images.json`, SHA-256
`ff9237178e5bbf9523012a83e2366b99d2ff3cd0092c6116ad4c0dee08afdd64`.
Published source revision: `2718fc978b7454f0f2aa422e1236fb6ad543e0d2`.
The node selector matches that published platform. To update, verify the new
release asset against its published checksums, update all five references from
that same release and repeat both offline checks. No registry lookup occurs
inside `check.sh`; a well-formed digest is not a fresh image availability proof.
