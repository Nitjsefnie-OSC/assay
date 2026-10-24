#!/bin/bash
# shellcheck shell=bash
# stage-cache.sh: the stage Job's container `cache`, after the init container `source` has
# exited. Baked into the verify-runner image at /usr/local/bin/stage-cache.sh, root-owned and
# mode 0755; the Job runs it as its single-element command. It mounts NO Secret, has no
# credential helper and inherits no token: it reads the clone the source phase left on the PVC
# (the whole claim is mounted at /cache) and fills the module cache at /cache/mod from the public
# proxy. No target code is executed: `go mod download` runs none.
#
#   stage-cache.sh --repo=<owner/repo> --sha=<40-hex sha>   fill /cache/mod for /cache/clone
#   stage-cache.sh --probe                                  make the probe directories, print
#                                                           this process's confinement
#
# With no argument, or any other argument list, it prints its usage on stderr and exits 2
# before it reads or writes anything.
#
# Prints `stage-done <sha>` only after the clone is verified AND every module is cached.
# Anything it cannot fetch from the public proxy under the approved public set is
# could-not-check (`stage-cnc ...`, exit 3), never a reason to look for a credential.
set -u

usage() {
  printf 'usage: stage-cache.sh --repo=<owner/repo> --sha=<40-hex sha>\n       stage-cache.sh --probe\n' >&2
  exit 2
}
die() { printf 'stage-cache: %s\n' "$*" >&2; exit 1; }
cnc() { printf 'stage-cnc %s\n' "$*"; exit 3; }

# The argument list is checked before anything is read or written.
MODE=
case $# in
  1) [ "$1" = --probe ] || usage; MODE=probe ;;
  2)
    case $1 in --repo=*) ;; *) usage ;; esac
    case $2 in --sha=*) ;; *) usage ;; esac
    MODE=run; REPO=${1#--repo=}; SHA=${2#--sha=}
    [[ $REPO =~ ^[A-Za-z0-9._-]+/[A-Za-z0-9._-]+$ ]] || usage
    [[ $SHA =~ ^[0-9a-f]{40}$ ]] || usage
    ;;
  *) usage ;;
esac

# VR_CACHE_DIR is a test hook; the Job's env allowlist never sets it
CACHE=${VR_CACHE_DIR:-/cache}
CLONE=$CACHE/clone

if [ "$MODE" = probe ]; then
  # the probe variant: the directories the witness probe's sub-path mounts need, and one line
  mkdir -p "$CLONE" "$CACHE/mod" 2>/dev/null
  seccomp=$(awk '$1 == "Seccomp:" { print $2; exit }' /proc/self/status 2>/dev/null)
  apparmor=
  if read -r first _ < /proc/self/attr/current 2>/dev/null; then apparmor=$first; fi
  printf 'stage seccomp=%s apparmor=%s\n' "${seccomp:-unknown}" "${apparmor:-unknown}"
  exit 0
fi
OWNER=${REPO%%/*}

# --- the same Go environment the Job template pins, restated so a missing value is loud ------
for v in GOVCS GOPROXY GOSUMDB GOTOOLCHAIN GOMODCACHE GOCACHE; do
  [ -n "${!v:-}" ] || die "$v is not set"
done
[ "$GOVCS" = '*:off' ] || die "GOVCS is not *:off"
[ "$GOPROXY" = https://proxy.golang.org ] || die "GOPROXY is not the public proxy alone"
[ "$GOSUMDB" = sum.golang.org ] || die "GOSUMDB is not sum.golang.org"
[ "$GOTOOLCHAIN" = local ] || die "GOTOOLCHAIN is not local"
for v in GOPRIVATE GONOPROXY GONOSUMDB GONOSUMCHECK GOINSECURE GOFLAGS GIT_ASKPASS; do
  [ -z "${!v:-}" ] || die "$v must be empty here"
done
export GIT_CONFIG_NOSYSTEM=1 GIT_CONFIG_GLOBAL=/dev/null GIT_TERMINAL_PROMPT=0
unset GIT_ASKPASS SSH_ASKPASS GIT_CONFIG_COUNT GIT_CONFIG_PARAMETERS

# --- the clone must be the exact commit, with the pinned origin/main -------------------------
[ -d "$CLONE/.git" ] || die "no clone at $CLONE"
HEAD_SHA=$(git -c safe.directory="$CLONE" -C "$CLONE" rev-parse --verify 'HEAD^{commit}' 2>/dev/null) || die "the clone has no HEAD"
PIN_SHA=$(git -c safe.directory="$CLONE" -C "$CLONE" rev-parse --verify 'refs/remotes/origin/main^{commit}' 2>/dev/null) || die "refs/remotes/origin/main is not pinned"
[ "$HEAD_SHA" = "$SHA" ] || die "the clone is at a different commit"
[ "$PIN_SHA" = "$SHA" ] || die "refs/remotes/origin/main is not the selected commit"

# --- the module set: every declared module path, before any proxy request -------------------
# Approved: the module is declared in this clone (a workspace-local module), or its path is on
# a public vanity host from the list below, or is a github.com path outside the target repo's
# own owner. Anything else, or any path under the owner of the target repo (private by
# default), is could-not-check and is never sent to the public proxy.
APPROVED='^(golang[.]org/x|google[.]golang[.]org|gopkg[.]in|go[.]uber[.]org|go[.]opentelemetry[.]io|go[.]etcd[.]io|go[.]yaml[.]in|honnef[.]co|gotest[.]tools|sigs[.]k8s[.]io|k8s[.]io|cloud[.]google[.]com|modernc[.]org|rsc[.]io|lukechampine[.]com|filippo[.]io)/'
GHPUB='^github[.]com/[A-Za-z0-9_.-]+/'
OWNER_LC=$(printf '%s' "$OWNER" | tr '[:upper:]' '[:lower:]')

GOMODS=$(find "$CLONE" \( -name .git -o -name vendor -o -name testdata \) -prune -o -name go.mod -type f -print | sort)
[ -n "$GOMODS" ] || cnc "no go.mod in the clone: nothing to prefetch"

OWN=$(printf '%s\n' "$GOMODS" | while IFS= read -r m; do
  awk '$1 == "module" { gsub(/"/, "", $2); print $2; exit }' "$m"
done)
DECLARED=$(printf '%s\n' "$GOMODS" | while IFS= read -r m; do
  awk '
    /^[[:space:]]*\/\// { next }
    $1 == "require" && $2 == "(" { inreq = 1; next }
    $1 == "replace" && $2 == "(" { inrep = 1; next }
    $1 == ")" { inreq = 0; inrep = 0; next }
    inreq { print $1; next }
    $1 == "require" { print $2; next }
    inrep || $1 == "replace" {
      for (i = 1; i <= NF; i++) if ($i == "=>") { if ($(i + 1) !~ /^[.]{1,2}\//) print $(i + 1) }
      next
    }
  ' "$m"
done | sort -u)

BAD=
for p in $DECLARED; do
  if printf '%s\n' "$OWN" | grep -q -x -F -e "$p"; then continue; fi
  lc=$(printf '%s' "$p" | tr '[:upper:]' '[:lower:]')
  case $lc in "github.com/$OWNER_LC/"*) BAD="$BAD $p"; continue ;; esac
  if printf '%s' "$p" | grep -q -E -e "$APPROVED"; then continue; fi
  if printf '%s' "$p" | grep -q -E -e "$GHPUB"; then continue; fi
  BAD="$BAD $p"
done
[ -z "$BAD" ] || cnc "module paths outside the approved public set, nothing was sent to the proxy:$BAD"

# --- prefetch: the module cache on the PVC, no target code -----------------------------------
mkdir -p "$GOMODCACHE" "$GOCACHE" || die "no module cache directory"
printf '%s\n' "$GOMODS" | while IFS= read -r m; do
  d=$(dirname "$m")
  ( cd "$d" && go mod download ) || exit 1
done || cnc "go mod download did not complete in a module of this clone"

printf 'stage-done %s\n' "$SHA"
