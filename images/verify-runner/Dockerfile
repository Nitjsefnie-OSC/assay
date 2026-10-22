# The release's own statusgen bytes plus the fleet's digest-pinned toolchain.
FROM golang:1.27.2-bookworm@sha256:5cf287a799e6b94384bad13d16b14904c531f51ba65792237e122ce42b392f61
# Dependency admission and update policy: issue #3469. Install only at build time.
# Exact direct versions fail closed if the archive changes; review a pin update.
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      python3-yaml=6.0-3+b2 \
      python3-jsonschema=4.10.3-1 \
 && mkdir -p /usr/local/share/verify-runner \
 && dpkg-query -W -f='${binary:Package}=${Version}\n' \
      > /usr/local/share/verify-runner/packages.txt \
 && rm -rf /var/lib/apt/lists/*

# Rows run with no network and a read-only root, so everything they resolve is
# provisioned here, at build time only.
#
# Offline packaging-check tools, at the exact versions containers/k8s/check.py
# requires. Tool assessment: issue #3725. Each archive is compared with the
# sha256 pinned below before it is unpacked; a mismatch, an unpinned platform or
# a version other than the pinned one fails the build.
ARG TARGETARCH
RUN set -eu; \
    case "${TARGETARCH:-}" in \
      amd64) gl_arch=x64 \
             kustomize_sha=cb9e31198d3f63b44848bd0afc4c8efc56e6cfef9c93a4a39a211940e6decd61 \
             kubeconform_sha=9bc2bffbf71f261128533edaf912153948b7ff238f9a531ae6d34466ec287883 \
             gitleaks_sha=551f6fc83ea457d62a0d98237cbad105af8d557003051f41f3e7ca7b3f2470eb ;; \
      arm64) gl_arch=arm64 \
             kustomize_sha=9867b76482cfc6d25b546abc1d2b5a32bd137505aba640649b4b378f64f6d68f \
             kubeconform_sha=1f53fc8e81258197a35e8603054162a5af1de8c5af13746c71ab680d9534ed87 \
             gitleaks_sha=e4a487ee7ccd7d3a7f7ec08657610aa3606637dab924210b3aee62570fb4b080 ;; \
      *) echo "verify-runner: no pinned tool checksums for '${TARGETARCH:-}'" >&2; exit 1 ;; \
    esac; \
    fetch() { \
      curl -fsSL --proto '=https' --tlsv1.2 -o /tmp/tool.tar.gz "$1"; \
      echo "$2  /tmp/tool.tar.gz" | sha256sum -c -; \
      tar -xzf /tmp/tool.tar.gz -C /usr/local/bin --no-same-owner "$3"; \
      chmod 0755 "/usr/local/bin/$3"; \
      rm -f /tmp/tool.tar.gz; \
    }; \
    fetch "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv5.8.3/kustomize_v5.8.3_linux_${TARGETARCH}.tar.gz" \
      "$kustomize_sha" kustomize; \
    fetch "https://github.com/yannh/kubeconform/releases/download/v0.8.0/kubeconform-linux-${TARGETARCH}.tar.gz" \
      "$kubeconform_sha" kubeconform; \
    fetch "https://github.com/gitleaks/gitleaks/releases/download/v8.30.1/gitleaks_8.30.1_linux_${gl_arch}.tar.gz" \
      "$gitleaks_sha" gitleaks; \
    test "$(kustomize version)" = v5.8.3; \
    test "$(kubeconform -v)" = v0.8.0; \
    test "$(gitleaks version)" = 8.30.1

# Go module cache for tools/desk (which also holds the gopkg.in/yaml.v3 module
# the containers/k8s YAML adapter runs on). Only module files are copied; the
# downloads are checked against tools/desk/go.sum. Rows need no proxy, and the
# cache stays read-only and world-readable for a non-root row user.
ENV GOMODCACHE=/go/pkg/mod
COPY tools/desk/go.mod tools/desk/go.sum /tmp/modsrc/tools/desk/
COPY comms/go.mod /tmp/modsrc/comms/
COPY deskcore/go.mod /tmp/modsrc/deskcore/
COPY forgeread/go.mod /tmp/modsrc/forgeread/
COPY gitquiet/go.mod /tmp/modsrc/gitquiet/
COPY perfrec/go.mod /tmp/modsrc/perfrec/
RUN cd /tmp/modsrc/tools/desk \
 && GOWORK=off GOFLAGS=-mod=readonly GOTOOLCHAIN=local go mod download \
 && GOWORK=off GOFLAGS=-mod=readonly GOTOOLCHAIN=local go mod download gopkg.in/yaml.v3 \
 && cd / && rm -rf /tmp/modsrc /root/.cache

COPY .release-images/harness/statusgen /usr/local/bin/statusgen
