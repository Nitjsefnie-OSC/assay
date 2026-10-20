package main

import (
	"strings"
	"testing"
	"time"

	"github.com/medici-finance/assay/tools/desk/internal/deskkit"
)

const aliasRepo = "medici-finance/assay"

// One PR claimed under two label spellings is ONE claim (medici-finance/assay#3700): a live
// claim under another spelling refuses and creates nothing.
func TestAcquire_refuses_while_alias_live(t *testing.T) {
	f := newStore()
	f.seedClaim("a--pr-108", "sess-A", "dispatched", "feat/x", 10*time.Minute)
	run, so, _ := harness(t, f)

	rc := run("acquire", "b--pr-108", "--repo", aliasRepo, "--owner", "sess-B")
	if rc != exitRefused {
		t.Fatalf("rc = %d, want 5; out=%s", rc, so.String())
	}
	if _, made := f.claims["b--pr-108"]; made {
		t.Error("a second claim was created for the same PR")
	}
	if !strings.Contains(so.String(), "a--pr-108") {
		t.Errorf("the live holder is not named: %s", so.String())
	}
}

// An alias past its TTL with no live branch is dead and does not block.
func TestAcquireIgnoresAnExpiredAlias(t *testing.T) {
	f := newStore()
	f.seedClaim("a--pr-108", "sess-A", "dispatched", "feat/x", 121*time.Minute)
	run, so, _ := harness(t, f)

	if rc := run("acquire", "b--pr-108", "--repo", aliasRepo, "--owner", "sess-B"); rc != exitOK {
		t.Fatalf("rc = %d, want 0; out=%s", rc, so.String())
	}
	if _, made := f.claims["b--pr-108"]; !made {
		t.Error("no claim was created next to an expired alias")
	}
}

// Branch-as-claim: an alias past its TTL whose branch is still on the remote is live.
func TestAcquire_refuses_alias_with_live_branch(t *testing.T) {
	f := newStore()
	f.seedClaim("a--pr-108", "sess-A", "dispatched", "feat/x", 300*time.Minute)
	f.branches["feat/x"] = true
	run, so, _ := harness(t, f)

	if rc := run("acquire", "b--pr-108", "--repo", aliasRepo, "--owner", "sess-B"); rc != exitRefused {
		t.Fatalf("rc = %d, want 5; out=%s", rc, so.String())
	}
}

// The correctness and security lanes stay separate; spellings inside one lane still collide.
func TestAliasLanesStaySeparate(t *testing.T) {
	f := newStore()
	f.seedClaim("a--pr-108", "sess-A", "dispatched", "feat/x", 5*time.Minute)
	run, so, _ := harness(t, f)

	if rc := run("acquire", "b--pr-108--security", "--repo", aliasRepo, "--owner", "sess-B"); rc != exitOK {
		t.Fatalf("security lane blocked by the correctness lane: rc = %d; out=%s", rc, so.String())
	}
	if rc := run("acquire", "c--pr-108--security", "--repo", aliasRepo, "--owner", "sess-C"); rc != exitRefused {
		t.Fatalf("a second spelling inside the security lane: rc = %d, want 5; out=%s", rc, so.String())
	}
}

// pr-108 and pr-1081 are different PRs; non-PR keys have no alias at all.
func TestAlias_ignores_other_PRs_and_non_PR_keys(t *testing.T) {
	f := newStore()
	f.seedClaim("a--pr-1081", "sess-A", "dispatched", "", 5*time.Minute)
	f.seedClaim("a--stream--07", "sess-A", "dispatched", "", 5*time.Minute)
	run, so, _ := harness(t, f)

	if rc := run("acquire", "b--pr-108", "--repo", aliasRepo, "--owner", "sess-B"); rc != exitOK {
		t.Fatalf("pr-108 blocked by pr-1081: rc = %d; out=%s", rc, so.String())
	}
	if rc := run("acquire", "b--stream--07", "--repo", aliasRepo, "--owner", "sess-B"); rc != exitOK {
		t.Fatalf("a non-PR key was treated as an alias: rc = %d; out=%s", rc, so.String())
	}
}

// A claim set that cannot be listed is could-not-check (exit 6), never "no alias".
func TestAcquire_fails_closed_on_unreadable_set(t *testing.T) {
	f := newStore()
	f.readFails = true
	run, so, se := harness(t, f)

	rc := run("acquire", "b--pr-108", "--repo", aliasRepo, "--owner", "sess-B")
	if rc != exitUnverifiable {
		t.Fatalf("rc = %d, want 6; out=%s err=%s", rc, so.String(), se.String())
	}
	if _, made := f.claims["b--pr-108"]; made {
		t.Error("a claim was created while the alias check could not run")
	}
}

// raceStore seeds a competing spelling right after our create applies, the way a concurrent
// acquire under another label lands between our pre-check and our post-check.
type raceStore struct {
	*fakeStore
	rival     string
	rivalDate string
}

func (r *raceStore) CreateIfAbsent(id, msg string) deskkit.ClaimWriteOutcome {
	out := r.fakeStore.CreateIfAbsent(id, msg)
	if out == deskkit.ClaimWriteApplied {
		r.claims[r.rival] = fakeClaim{sha: r.nextSHA(), msg: claimMessage(r.rival, "sess-R", "dispatched", "", ""), date: r.rivalDate}
	}
	return out
}

func raceHarness(t *testing.T, r *raceStore) func(args ...string) int {
	t.Helper()
	oldBuild, oldOut, oldErr := buildStore, out, errOut
	var so, se strings.Builder
	buildStore = func(_, _ string) (deskkit.ClaimStore, error) { return r, nil }
	out, errOut = &so, &se
	t.Cleanup(func() { buildStore, out, errOut = oldBuild, oldOut, oldErr })
	return func(args ...string) int { return run(args) }
}

// Same instant: the larger id loses, removes its own ref, and refuses; the smaller id wins.
func TestAliasRaceTieGoesToTheSmallerID(t *testing.T) {
	cases := []struct {
		name, mine, rival string
		wantRC            int
	}{
		{"larger id loses", "b--pr-108", "a--pr-108", exitRefused},
		{"smaller id wins", "a--pr-108", "b--pr-108", exitOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newStore()
			r := &raceStore{fakeStore: f, rival: tc.rival, rivalDate: f.now.Format(time.RFC3339)}
			rc := raceHarness(t, r)("acquire", tc.mine, "--repo", aliasRepo, "--owner", "sess-M")
			if rc != tc.wantRC {
				t.Fatalf("rc = %d, want %d", rc, tc.wantRC)
			}
			_, kept := f.claims[tc.mine]
			if kept != (tc.wantRC == exitOK) {
				t.Errorf("own ref present = %v after rc %d", kept, rc)
			}
			if _, ok := f.claims[tc.rival]; !ok {
				t.Error("the rival's claim was disturbed")
			}
		})
	}
}

// The earlier forge-stamped date wins regardless of id order.
func TestAliasRaceEarlierDateWins(t *testing.T) {
	f := newStore()
	r := &raceStore{fakeStore: f, rival: "z--pr-108", rivalDate: f.now.Add(-time.Second).Format(time.RFC3339)}
	if rc := raceHarness(t, r)("acquire", "a--pr-108", "--repo", aliasRepo, "--owner", "sess-M"); rc != exitRefused {
		t.Fatalf("an earlier rival did not win: rc = %d", rc)
	}
	if _, kept := f.claims["a--pr-108"]; kept {
		t.Error("the loser kept its ref")
	}

	f2 := newStore()
	r2 := &raceStore{fakeStore: f2, rival: "a--pr-108", rivalDate: f2.now.Add(time.Second).Format(time.RFC3339)}
	if rc := raceHarness(t, r2)("acquire", "z--pr-108", "--repo", aliasRepo, "--owner", "sess-M"); rc != exitOK {
		t.Fatalf("an earlier claim lost to a later rival: rc = %d", rc)
	}
}

// show prints ALIAS-HELD after FREE when another spelling is live, and nothing when none is.
func TestShowReportsAliasHeld(t *testing.T) {
	f := newStore()
	f.seedClaim("a--pr-108", "sess-A", "dispatched", "feat/x", 10*time.Minute)
	run, so, _ := harness(t, f)

	if rc := run("show", "b--pr-108", "--repo", aliasRepo); rc != exitOK {
		t.Fatalf("show rc = %d; out=%s", rc, so.String())
	}
	free := strings.Index(so.String(), "FREE b--pr-108")
	alias := strings.Index(so.String(), "ALIAS-HELD a--pr-108")
	if free < 0 || alias < free {
		t.Errorf("want FREE then ALIAS-HELD, got: %s", so.String())
	}

	if rc := run("show", "b--stream--07", "--repo", aliasRepo); rc != exitOK || strings.Contains(so.String(), "ALIAS-HELD") {
		t.Errorf("a non-PR key printed an alias line: %s", so.String())
	}
}

// list does not print ALIAS-HELD: heldClaimID parses HELD lines of list output.
func TestListIsUnchangedByAliases(t *testing.T) {
	f := newStore()
	f.seedClaim("a--pr-108", "sess-A", "dispatched", "", 10*time.Minute)
	f.seedClaim("b--pr-108", "sess-B", "dispatched", "", 10*time.Minute)
	run, so, _ := harness(t, f)

	if rc := run("list", "--repo", aliasRepo); rc != exitOK {
		t.Fatalf("list rc = %d", rc)
	}
	if strings.Contains(so.String(), "ALIAS-HELD") {
		t.Errorf("list printed an alias line: %s", so.String())
	}
}

// A listing that errors is unverifiable even when single-claim reads still work.
type listFailStore struct{ *fakeStore }

func (listFailStore) List() ([]string, deskkit.ClaimReadStatus) {
	return nil, deskkit.ClaimReadUnverifiable
}

func listFailHarness(t *testing.T, s deskkit.ClaimStore) func(args ...string) int {
	t.Helper()
	oldBuild, oldOut, oldErr := buildStore, out, errOut
	var so, se strings.Builder
	buildStore = func(_, _ string) (deskkit.ClaimStore, error) { return s, nil }
	out, errOut = &so, &se
	t.Cleanup(func() { buildStore, out, errOut = oldBuild, oldOut, oldErr })
	return func(args ...string) int { return run(args) }
}

func TestAcquire_listing_error_fails_closed(t *testing.T) {
	f := newStore()
	if rc := listFailHarness(t, listFailStore{f})("acquire", "b--pr-9", "--repo", aliasRepo, "--owner", "s"); rc != exitUnverifiable {
		t.Fatalf("rc = %d, want 6", rc)
	}
	if _, made := f.claims["b--pr-9"]; made {
		t.Error("a claim was created while the alias check could not run")
	}
}

// Leading zeros: pr-05 and pr-5 are one PR; pr-50 is not.
func TestAlias_leading_zeros(t *testing.T) {
	for _, tc := range []struct {
		held, want string
		rc         int
	}{
		{"a--pr-5", "b--pr-05", exitRefused},
		{"a--pr-005", "b--pr-5", exitRefused},
		{"a--pr-00", "b--pr-0", exitRefused},
		{"a--pr-5", "b--pr-50", exitOK},
	} {
		f := newStore()
		f.seedClaim(tc.held, "s-A", "dispatched", "", 5*time.Minute)
		if rc, _, _ := harnessRun(t, f, "acquire", tc.want, "--repo", aliasRepo, "--owner", "s-B"); rc != tc.rc {
			t.Errorf("%s vs %s: rc = %d, want %d", tc.held, tc.want, rc, tc.rc)
		}
	}
}

func harnessRun(t *testing.T, f *fakeStore, args ...string) (int, *strings.Builder, *strings.Builder) {
	t.Helper()
	oldBuild, oldOut, oldErr := buildStore, out, errOut
	so, se := &strings.Builder{}, &strings.Builder{}
	buildStore = func(_, _ string) (deskkit.ClaimStore, error) { return f, nil }
	out, errOut = so, se
	defer func() { buildStore, out, errOut = oldBuild, oldOut, oldErr }()
	return run(args), so, se
}

// steal (and the reclaim that reaches it) and progress honour a live alias; nothing is touched.
func TestSteal_refused_by_live_alias(t *testing.T) {
	f := newStore()
	f.seedClaim("a--pr-7", "s-A", "dispatched", "", 5*time.Minute)
	f.seedClaim("b--pr-7", "s-B", "claimed", "", 200*time.Minute)
	before := f.claims["b--pr-7"]
	rc, _, _ := harnessRun(t, f, "steal", "b--pr-7", "--repo", aliasRepo, "--owner", "s-C", "--reason", "takeover")
	if rc != exitRefused {
		t.Fatalf("rc = %d, want 5", rc)
	}
	if f.claims["b--pr-7"] != before {
		t.Error("the refused steal changed the existing claim")
	}
}

func TestSteal_without_alias_proceeds(t *testing.T) {
	f := newStore()
	f.seedClaim("b--pr-7", "s-B", "claimed", "", 200*time.Minute)
	if rc, _, _ := harnessRun(t, f, "steal", "b--pr-7", "--repo", aliasRepo, "--owner", "s-C", "--reason", "takeover"); rc != exitOK {
		t.Fatalf("rc = %d, want 0", rc)
	}
}

func TestReclaim_refused_by_live_alias(t *testing.T) {
	f := newStore()
	f.seedClaim("a--pr-7", "s-A", "dispatched", "", 5*time.Minute)
	f.seedClaim("b--pr-7", "s-B", "claimed", "", 200*time.Minute)
	// The alias is seeded after acquire's own pre-check by hiding it from the first listing.
	h := &hideOnceStore{fakeStore: f, hide: "a--pr-7"}
	rc := listFailHarness(t, h)("acquire", "b--pr-7", "--repo", aliasRepo, "--owner", "s-C")
	if rc != exitRefused {
		t.Fatalf("rc = %d, want 5", rc)
	}
	if f.claims["b--pr-7"].msg == claimMessage("b--pr-7", "s-C", "claimed", "", "") {
		t.Error("the stale claim was taken over despite a live alias")
	}
}

type hideOnceStore struct {
	*fakeStore
	hide string
	done bool
}

func (h *hideOnceStore) List() ([]string, deskkit.ClaimReadStatus) {
	ids, st := h.fakeStore.List()
	if h.done {
		return ids, st
	}
	h.done = true
	var kept []string
	for _, id := range ids {
		if id != h.hide {
			kept = append(kept, id)
		}
	}
	return kept, st
}

func TestSteal_race_loser_removes_ref(t *testing.T) {
	f := newStore()
	r := &raceStore{fakeStore: f, rival: "a--pr-7", rivalDate: f.now.Format(time.RFC3339)}
	rc := raceHarness(t, r)("steal", "b--pr-7", "--repo", aliasRepo, "--owner", "s-C", "--reason", "takeover")
	if rc != exitRefused {
		t.Fatalf("rc = %d, want 5", rc)
	}
	if _, kept := f.claims["b--pr-7"]; kept {
		t.Error("the steal loser kept its ref")
	}
}

func TestProgress_refused_by_live_alias(t *testing.T) {
	f := newStore()
	f.seedClaim("a--pr-7", "s-A", "dispatched", "", 5*time.Minute)
	f.seedClaim("b--pr-7", "s-B", "claimed", "", 5*time.Minute)
	before := f.claims["b--pr-7"]
	rc, _, _ := harnessRun(t, f, "progress", "b--pr-7", "--repo", aliasRepo, "--owner", "s-B", "--branch", "feat/y")
	if rc != exitRefused {
		t.Fatalf("rc = %d, want 5", rc)
	}
	if f.claims["b--pr-7"] != before {
		t.Error("the refused progress advanced the claim")
	}
}

func TestProgress_without_alias_advances(t *testing.T) {
	f := newStore()
	f.seedClaim("b--pr-7", "s-B", "claimed", "", 5*time.Minute)
	if rc, _, _ := harnessRun(t, f, "progress", "b--pr-7", "--repo", aliasRepo, "--owner", "s-B", "--branch", "feat/y"); rc != exitOK {
		t.Fatalf("rc = %d, want 0", rc)
	}
}

// A post-create check that cannot be answered removes the ref it just created.
type postFailStore struct {
	*fakeStore
	created bool
}

func (p *postFailStore) CreateIfAbsent(id, msg string) deskkit.ClaimWriteOutcome {
	o := p.fakeStore.CreateIfAbsent(id, msg)
	if o == deskkit.ClaimWriteApplied {
		p.created = true
	}
	return o
}

func (p *postFailStore) List() ([]string, deskkit.ClaimReadStatus) {
	if p.created {
		return nil, deskkit.ClaimReadUnverifiable
	}
	return p.fakeStore.List()
}

func TestPostCheck_unverifiable_releases_ref(t *testing.T) {
	f := newStore()
	p := &postFailStore{fakeStore: f}
	if rc := listFailHarness(t, p)("acquire", "b--pr-7", "--repo", aliasRepo, "--owner", "s"); rc != exitUnverifiable {
		t.Fatalf("rc = %d, want 6", rc)
	}
	if _, kept := f.claims["b--pr-7"]; kept {
		t.Error("an unconfirmed claim was left standing")
	}
}

// show: an alias leg that cannot be read exits 6 but still prints the claim's own line.
func TestShow_unreadable_alias_leg_exits_6(t *testing.T) {
	f := newStore()
	f.seedClaim("b--pr-7", "s-B", "claimed", "", 5*time.Minute)
	oldBuild, oldOut, oldErr := buildStore, out, errOut
	so, se := &strings.Builder{}, &strings.Builder{}
	buildStore = func(_, _ string) (deskkit.ClaimStore, error) { return listFailStore{f}, nil }
	out, errOut = so, se
	defer func() { buildStore, out, errOut = oldBuild, oldOut, oldErr }()
	if rc := run([]string{"show", "b--pr-7", "--repo", aliasRepo}); rc != exitUnverifiable {
		t.Fatalf("rc = %d, want 6", rc)
	}
	if !strings.Contains(so.String(), "HELD b--pr-7") || !strings.Contains(se.String(), "unverifiable") {
		t.Errorf("out=%q err=%q", so.String(), se.String())
	}
}
