package main

// admissionhold_test.go — prune holds a worktree an admitted verifier run lives in.
//
// The defect: a verifier home is detached at the main commit it was dispatched at. Once
// main moves on it is clean and merged, so the ordinary sweep removed it — and the
// dead-session arm reaped it — while its Evidence still had to land from it. Nothing read
// the admission record the dispatch left in the home's admin dir.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/medici-finance/assay/tools/desk/internal/deskkit"
)

// admitHome writes an admission record into target's admin dir, where the verifier
// dispatch writes it, dated `at`.
func admitHome(t *testing.T, target string, at time.Time) string {
	t.Helper()
	rec := mustGit(t, target, "rev-parse", "--path-format=absolute", "--git-path", deskkit.VerifierAttestationRecord)
	writeFile(t, rec, "{}\n")
	if err := os.Chtimes(rec, at, at); err != nil {
		t.Fatalf("date the admission record: %v", err)
	}
	return rec
}

// verifierHome builds the incident's shape: a clean worktree detached at a main commit
// that main has since moved past — merged, tracked-clean, below the tip.
func verifierHome(t *testing.T, work, name string) string {
	t.Helper()
	target := addWorktree(t, name)
	mustGit(t, target, "checkout", "--quiet", "--detach")
	advanceMain(t, work, name)
	return target
}

func pruneOK(t *testing.T, args ...string) string {
	t.Helper()
	rc, errout := runCapErr(t, append([]string{"prune"}, args...))
	if rc != deskkit.ExitOK {
		t.Fatalf("prune %v rc = %d, want 0; stderr:\n%s", args, rc, errout)
	}
	return errout
}

// regression: #2830
func TestPruneHoldsAdmittedHome(t *testing.T) {
	work := newRepo(t)
	withEnv(t, work)
	home := verifierHome(t, work, "vhome")
	admitHome(t, home, time.Now().Add(-2*time.Hour)) // past the 60m at-tip window
	control := verifierHome(t, work, "vcontrol")     // same shape, never admitted

	errout := pruneOK(t)
	assertExists(t, home)
	if !strings.Contains(errout, "active verifier admission") {
		t.Fatalf("the hold reason is not named; stderr:\n%s", errout)
	}
	if _, err := os.Stat(control); !os.IsNotExist(err) {
		t.Fatalf("the un-admitted control %s was not removed (err=%v); stderr:\n%s", control, err, errout)
	}
}

// regression: #2830
func TestReapHoldsAdmittedHome(t *testing.T) {
	work := newRepo(t)
	withEnv(t, work)
	// The reap arm's headline shape (TestReapLockedByDeadSession reaps it), now admitted.
	target, branch := openPRWorktree(t, work, "vdead")
	lockAs(t, work, target, "dead-v")
	plantBeacon(t, "dead-v", time.Now().Add(-6*time.Hour))
	admitHome(t, target, time.Now().Add(-2*time.Hour))

	plan := pruneOK(t, "--reap-dead-sessions", "--dry-run")
	if !strings.Contains(plan, "KEEP "+resolvePath(target)) || !strings.Contains(plan, "active verifier admission") {
		t.Fatalf("the plan does not name the admission hold; stderr:\n%s", plan)
	}
	errout := pruneOK(t, "--reap-dead-sessions")
	assertExists(t, target)
	if !localBranchExists(t, work, branch) {
		t.Fatalf("an admitted home's branch %s was deleted; stderr:\n%s", branch, errout)
	}
	if !strings.Contains(errout, "dead-session-reaped 0") {
		t.Fatalf("an admitted home was reaped; stderr:\n%s", errout)
	}
}

// The hold is bounded: an admission past the window, or a released one, drains as usual.
func TestPruneDrainsStaleAdmission(t *testing.T) {
	work := newRepo(t)
	withEnv(t, work)
	expired := verifierHome(t, work, "vexpired")
	admitHome(t, expired, time.Now().Add(-verifierAdmissionHold-time.Hour))
	released := verifierHome(t, work, "vreleased")
	if err := os.Remove(admitHome(t, released, time.Now())); err != nil {
		t.Fatal(err)
	}

	errout := pruneOK(t)
	for _, p := range []string{expired, released} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Fatalf("%s should drain (err=%v); stderr:\n%s", p, err, errout)
		}
	}
}

// regression: #2830
func TestPruneHoldsFutureAdmission(t *testing.T) {
	work := newRepo(t)
	withEnv(t, work)
	home := verifierHome(t, work, "vfuture")
	admitHome(t, home, time.Now().Add(48*time.Hour))

	errout := pruneOK(t)
	assertExists(t, home)
	if !strings.Contains(errout, "dated in the future") {
		t.Fatalf("a future-dated admission was not held by name; stderr:\n%s", errout)
	}
}

// A record that exists but cannot be read holds its worktree, could-not-check.
//
// regression: #2830
func TestPruneHoldsUnreadableAdmission(t *testing.T) {
	work := newRepo(t)
	withEnv(t, work)
	home := verifierHome(t, work, "vloop")
	rec := mustGit(t, home, "rev-parse", "--path-format=absolute", "--git-path", deskkit.VerifierAttestationRecord)
	if err := os.Symlink(rec, rec); err != nil { // a symlink loop: stat fails, not NotExist
		t.Fatal(err)
	}

	errout := pruneOK(t)
	assertExists(t, home)
	if !strings.Contains(errout, "verifier admission record unreadable") {
		t.Fatalf("an unreadable admission did not hold by name; stderr:\n%s", errout)
	}
}

// A record no worktree can be tied to holds every candidate in the sweep. The orphan entry
// is locked so git's own bookkeeping prune keeps it, as it would a locked husk.
//
// regression: #2830
func TestPruneAdmissionUnattributed(t *testing.T) {
	work := newRepo(t)
	withEnv(t, work)
	home := verifierHome(t, work, "vplain")
	common := mustGit(t, work, "rev-parse", "--path-format=absolute", "--git-common-dir")
	orphan := filepath.Join(common, "worktrees", "orphan")
	if err := os.MkdirAll(orphan, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(orphan, "locked"), "")
	writeFile(t, filepath.Join(orphan, deskkit.VerifierAttestationRecord), "{}\n")

	errout := pruneOK(t)
	assertExists(t, home)
	if !strings.Contains(errout, "could-not-check: verifier admissions") {
		t.Fatalf("an unattributable admission did not hold the sweep; stderr:\n%s", errout)
	}
}

// The registry read fails CLOSED: a record no worktree can be tied to is an error, so the
// sweep holds every candidate; an absent registry is simply empty.
func TestAdmissionRegistryFailsClosed(t *testing.T) {
	common := t.TempDir()
	g := &pathGuard{commonDir: common}
	if adm, err := g.verifierAdmissions(); err != nil || len(adm) != 0 {
		t.Fatalf("absent registry: got %v, %v; want empty, nil", adm, err)
	}
	orphan := filepath.Join(common, "worktrees", "orphan")
	if err := os.MkdirAll(orphan, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(orphan, deskkit.VerifierAttestationRecord), "{}\n")
	if _, err := g.verifierAdmissions(); err == nil {
		t.Fatal("a record with no gitdir pointer was not refused")
	}
	if _, err := (&pathGuard{}).verifierAdmissions(); err == nil {
		t.Fatal("an unresolved common dir was not refused")
	}
}

func TestAdmissionHoldReasonWindow(t *testing.T) {
	now := time.Now()
	adm := map[string]admissionEvidence{
		"/in":     {mtime: now.Add(-verifierAdmissionHold + time.Minute)},
		"/edge":   {mtime: now.Add(-verifierAdmissionHold)},
		"/future": {mtime: now.Add(time.Hour)},
		"/broken": {err: os.ErrPermission},
	}
	for rt, held := range map[string]bool{"/in": true, "/edge": false, "/future": true, "/broken": true, "/none": false} {
		if got := admissionHoldReason(adm, rt, now) != ""; got != held {
			t.Errorf("%s: held=%v, want %v", rt, got, held)
		}
	}
}
