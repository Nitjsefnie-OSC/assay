#!/bin/bash
# shellcheck shell=bash
# witness.sh: the one fixed entrypoint of the witness Job, for its init container `seed` and its
# container `witness`. Baked into the verify-runner image at /usr/local/bin/witness.sh,
# root-owned and mode 0755. Neither container sets a command: each one's args are this path and
# exactly one word.
#
#   witness.sh seed              init container: copy the clone from /cache/clone into /work
#   witness.sh seed-probe        init container of the probe variant: an empty /work
#   witness.sh <stream>/<NN>     the whole run of the one brief /work/docs/streams/<stream>/
#                                brief-<NN>-<slug>.md (exactly one such regular file, else exit 2)
#   witness.sh probe             the probe measurements
#
# With no argument, or any other argument list, it prints its usage on stderr and exits 2
# before it reads or writes anything.
#
# The whole run enters the empty network namespace as ONE process tree: this script records
# its own namespace id, then execs
#   unshare --user --map-root-user --net statusgen __verifyrun-netns <outer-id> \
#     statusgen verifyrun --sandbox container-netns --brief <path> --dry-run --root /work
# statusgen's own helper raises lo inside the new namespace (a new namespace starts with lo
# DOWN) and proves lo is up, no other interface is up and the namespace differs from <outer-id>,
# and only then execs verifyrun. Every boundary failure is terminal: there is no unsandboxed
# retry and no lo-less retry, and no row command runs before the helper has approved.
set -u

usage() {
  printf 'usage: witness.sh seed | seed-probe | probe | <stream>/<NN>\n' >&2
  exit 2
}
die() { printf 'witness: %s\n' "$*" >&2; exit 2; }

# The argument list is checked before anything is read or written.
KEY_RE='^[a-z0-9-]+/[0-9]{2,}$'
[ $# -eq 1 ] || usage
case $1 in
  seed|seed-probe|probe) ;;
  *) [[ $1 =~ $KEY_RE ]] || usage ;;
esac

# VR_CACHE_DIR and VR_WORK_DIR are test hooks; the Job's env allowlist never sets them
CACHE=${VR_CACHE_DIR:-/cache}
WORK=${VR_WORK_DIR:-/work}
NETNS_RE='^net:\[[0-9]+\]$'
BASE_RE='^brief-[0-9]{2,}-[A-Za-z0-9._-]+[.]md$'

# whole_run <brief> <root>: replaces this process
whole_run() {
  local brief=$1 root=$2 outer
  outer=$(readlink /proc/self/ns/net) || die "cannot read this process's network namespace"
  [[ $outer =~ $NETNS_RE ]] || die "unreadable network namespace id: $outer"
  exec unshare --user --map-root-user --net \
    statusgen __verifyrun-netns "$outer" \
    statusgen verifyrun --sandbox container-netns --brief "$brief" --dry-run --root "$root"
}

# --- the probe's fixture: one row of each class, each recording what the row itself sees -----
FX=/tmp/vr-fx
write_fixture() {
  mkdir -p "$FX/docs/streams/fx" "$FX/out" || return 1
  cat > "$FX/probe.go" <<'GO'
package main

import (
	"fmt"
	"net"
)

func main() {
	iface, lo := "loopback-only", "down"
	ifs, _ := net.Interfaces()
	for _, i := range ifs {
		up := i.Flags&net.FlagUp != 0
		if i.Flags&net.FlagLoopback != 0 {
			if up {
				lo = "up"
			}
			continue
		}
		if up {
			iface = "other"
		}
	}
	lis := "refused"
	if l, err := net.Listen("tcp", "127.0.0.1:0"); err == nil {
		go func() {
			if c, err := l.Accept(); err == nil {
				c.Close()
			}
		}()
		if c, err := net.Dial("tcp", l.Addr().String()); err == nil {
			c.Close()
			lis = "ok"
		}
		l.Close()
	}
	fmt.Printf("iface=%s lo=%s listener=%s\n", iface, lo, lis)
}
GO
  cat > "$FX/row.sh" <<'ROW'
#!/bin/sh
# row.sh <class>: record what a row of that class sees inside the whole-run namespace
outer=$(cat /tmp/vr-fx/outer 2>/dev/null)
me=$(readlink /proc/self/ns/net)
ns=differs
[ -n "$outer" ] && [ "$me" != "$outer" ] || ns=same
res=$(go run /tmp/vr-fx/probe.go 2>/dev/null) || res="iface=other lo=down listener=refused"
echo "ns=$ns $res" > "/tmp/vr-fx/out/$1.txt"
exit 0
ROW
  cat > "$FX/docs/streams/fx/brief-01-fixture.md" <<'BRIEF'
---
brief: assay:at:fx:01
title: fixture
version: 2
---

# Fixture

## Verify (executable)

| # | Command | Expect | Class |
|---|---------|--------|-------|
| 1 | `sh row.sh ordinary` | exit 0 | check |
| 2 | `sh row.sh ci` | exit 0 | check:ci |

## Evidence

## Review
BRIEF
  ( cd "$FX" && git init -q . && git add -A && git -c commit.gpgsign=false commit -q -m fixture ) || return 1
}

# field <file> <key>: the value of key= in the one-line record, or the empty string
field() { tr ' ' '\n' < "$1" 2>/dev/null | sed -n "s/^$2=//p" | head -n 1; }

# class_state <class>: sets C_NS (loopback-only|other), C_LO (up|down), C_LIS (ok|refused)
class_state() {
  local f=$FX/out/$1.txt ns iface
  C_NS=other; C_LO=down; C_LIS=refused
  [ -f "$f" ] || return 0
  ns=$(field "$f" ns); iface=$(field "$f" iface)
  [ "$ns" = differs ] && [ "$iface" = loopback-only ] && C_NS=loopback-only
  [ "$(field "$f" lo)" = up ] && C_LO=up
  [ "$(field "$f" listener)" = ok ] && C_LIS=ok
  return 0
}

probe() {
  local dns egress api token unsh seccomp apparmor first
  # the OUTER pod, before the inner wrapper
  if timeout 5 getent hosts proxy.golang.org >/dev/null 2>&1; then dns=ok; else dns=refused; fi
  if timeout 5 bash -c 'exec 3<>/dev/tcp/1.1.1.1/443' >/dev/null 2>&1; then egress=open; else egress=refused; fi
  if timeout 5 bash -c 'exec 3<>/dev/tcp/10.96.0.1/443' >/dev/null 2>&1; then api=open; else api=refused; fi
  if [ -e /var/run/secrets/kubernetes.io/serviceaccount/token ]; then token=present; else token=absent; fi
  if unshare --user --map-root-user --net true >/dev/null 2>&1; then unsh=ok; else unsh=refused; fi
  seccomp=$(awk '$1 == "Seccomp:" { print $2; exit }' /proc/self/status 2>/dev/null)
  apparmor=
  if read -r first _ < /proc/self/attr/current 2>/dev/null; then apparmor=$first; fi
  printf 'probe dns=%s egress=%s apiserver=%s token=%s unshare=%s seccomp=%s apparmor=%s\n' \
    "$dns" "$egress" "$api" "$token" "$unsh" "${seccomp:-unknown}" "${apparmor:-unknown}"

  # the WHOLE-RUN entrypoint, with one fixture row of each class
  write_fixture >/dev/null 2>&1 || die "could not write the probe fixture"
  readlink /proc/self/ns/net > "$FX/outer" || die "cannot read this process's network namespace"
  ( whole_run "$FX/docs/streams/fx/brief-01-fixture.md" "$FX" ) > "$FX/run.log" 2>&1
  class_state ordinary; o_ns=$C_NS; o_lo=$C_LO; o_lis=$C_LIS
  class_state ci; c_ns=$C_NS; c_lo=$C_LO; c_lis=$C_LIS
  w_ns=other; [ "$o_ns" = loopback-only ] && [ "$c_ns" = loopback-only ] && w_ns=loopback-only
  w_lo=down; [ "$o_lo" = up ] && [ "$c_lo" = up ] && w_lo=up
  w_lis=refused; [ "$o_lis" = ok ] && [ "$c_lis" = ok ] && w_lis=ok
  printf 'whole-netns=%s lo=%s listener=%s ordinary-check-netns=%s ordinary-check-listener=%s ci-netns=%s ci-listener=%s\n' \
    "$w_ns" "$w_lo" "$w_lis" "$o_ns" "$o_lis" "$c_ns" "$c_lis"
}

# resolve_key <stream>/<NN>: sets BRIEF to the one regular brief file of that key under $WORK
resolve_key() {
  local stream=${1%/*} nn=${1#*/} f n=0
  BRIEF=
  for f in "$WORK/docs/streams/$stream/brief-$nn-"*.md; do
    [ -f "$f" ] && [ ! -L "$f" ] || continue
    [[ ${f##*/} =~ $BASE_RE ]] || continue
    n=$((n + 1)); BRIEF=$f
  done
  [ "$n" = 1 ] || die "$n brief files match $1 (need exactly one regular file docs/streams/$stream/brief-$nn-<slug>.md)"
}

# --- dispatch ---------------------------------------------------------------------------------
case $1 in
  seed)
    [ -d "$CACHE/clone/.git" ] || die "no clone on the PVC"
    cp -a "$CACHE/clone/." "$WORK/" || die "could not copy the clone"
    ;;
  seed-probe)
    ;;
  probe)
    probe
    ;;
  *)
    resolve_key "$1"
    whole_run "$BRIEF" "$WORK"
    ;;
esac
