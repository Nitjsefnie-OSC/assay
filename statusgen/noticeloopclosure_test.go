package main

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// noticeloopclosure_test.go — statusgen/32: a board-drift NOTICE must reach the owner
// that can act on it, on evidence strong enough to act, and no further.
//
// TestNoticeLoopClosure is one matrix, one sub-test per routing commitment. Each
// sub-test pairs the routed case with a NEGATIVE CONTROL (the look-alike that must
// not route) and a MUTATION (the routed case with the fix undone must go red), so a
// green run cannot be a reader that never fires.

// Closes statusgen/32: board-drift notices (phantom, merged-unflipped, Verified
// attribution, generated-view links) were observed but routed to nobody.

func nlcHash(t *testing.T, path string) [32]byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return sha256.Sum256(b)
}

// nlcStream builds an active stream on disk with the given brief files and in-memory
// todo rows.
func nlcStream(t *testing.T, root, name string, files map[string]string, rows ...Brief) *Stream {
	t.Helper()
	s := writeStream(t, root, name, "Active stream.\n", "", "")
	for num, body := range files {
		if err := os.WriteFile(filepath.Join(s.Dir, "brief-"+num+"-x.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	s.Status, s.Priority, s.Briefs = "active", "P1", rows
	s.LastTouch = day(0)
	return s
}

func nlcPickIDs(nu NextUp) []string {
	var ids []string
	for _, p := range nu.Picks {
		ids = append(ids, p.Stream.Name+"/"+p.Brief.Num)
	}
	sort.Strings(ids)
	return ids
}

func TestNoticeLoopClosure(t *testing.T) {
	t.Run("foreign-owned or already-delivered rows cannot dispatch", func(t *testing.T) {
		build := func() (*Stream, []mergedPR) {
			root := t.TempDir()
			s := nlcStream(t, root, "gtm", map[string]string{
				"01": "## Context\nDelivered by a merged implementation PR.\n",
				// NEGATIVE CONTROL: a brief ABOUT de-housing. The heuristic matches
				// it, but a discussion of ownership is not an ownership declaration.
				"03": "## Context\nThis brief designs how work is de-housed to another repo; it is local work.\n",
				"04": "## Context\nNamed only in a merge subject; nothing says it was delivered.\n",
				"05": "## Context\nA delivery PR landed part of it.\n",
				"06": "## Context\nAn ordinary brief.\n",
			},
				Brief{Num: "01", Status: "todo"},
				Brief{Num: "02", Status: "todo", RawCell: "Old row [homed→acme/widgets]"}, // pointer row: no file
				Brief{Num: "03", Status: "todo"},
				Brief{Num: "04", Status: "todo"},
				Brief{Num: "05", Status: "todo", Delivery: []DeliveryClaim{{In: "sib#1", Covers: "partial", Note: "task 2 owed"}}},
				Brief{Num: "06", Status: "todo"},
			)
			merged := []mergedPR{
				{Number: 11, Subject: "feat: gtm/01 (#11)", Briefs: []string{"gtm/01"}, Delivers: []string{"gtm/01"}},
				{Number: 12, Subject: "Merge pull request #12 from o/brief/gtm-04-x", Briefs: []string{"gtm/04"}},
				{Number: 13, Subject: "feat: gtm/05 (#13)", Briefs: []string{"gtm/05"}, Delivers: []string{"gtm/05"}},
			}
			return s, merged
		}

		s, merged := build()
		applyPhantomHolds([]*Stream{s}, merged, nil)
		nu := nextUp([]*Stream{s}, ClaimView{}, nil)
		got := nlcPickIDs(nu)
		if want := "gtm/03 gtm/04 gtm/05 gtm/06"; strings.Join(got, " ") != want {
			t.Fatalf("dispatchable rows = %v, want %s (01 delivered and 02 re-homed pointer are held)", got, want)
		}
		if len(nu.PhantomElsewhere) != 2 || nu.PhantomElsewhere["gtm/01"] == "" || nu.PhantomElsewhere["gtm/02"] == "" {
			t.Errorf("the held rows must be NAMED, not dropped; PhantomElsewhere = %v", nu.PhantomElsewhere)
		}
		// The negative controls still SURFACE (the notice owner is unchanged), they are
		// just not excluded.
		notices := strings.Join(boardHonestyNotices([]*Stream{s}, merged, nil), "\n")
		if !strings.Contains(notices, "(dehoused): gtm/03") {
			t.Errorf("the de-housing mention must still surface as a NOTICE; got:\n%s", notices)
		}
		// The same classifier and the same holds reach the other emitters' shared
		// eligibility function; there is no second allowlist to drift.
		if eligibleBase([]*Stream{s}, s, s.Briefs[0], map[string]bool{}, nil) {
			t.Error("eligibleBase must refuse the held row")
		}

		// MUTATION: without the application step the delivered row dispatches again.
		s2, _ := build()
		if ids := nlcPickIDs(nextUp([]*Stream{s2}, ClaimView{}, nil)); !strings.Contains(strings.Join(ids, " "), "gtm/01") {
			t.Fatalf("mutation control: with no hold applied gtm/01 must be dispatchable; got %v", ids)
		}
	})

	t.Run("an authoring-only PR does not advance status", func(t *testing.T) {
		root := t.TempDir()
		s := nlcStream(t, root, "gtm", map[string]string{"07": "## Context\nAuthored by a docs PR.\n"},
			Brief{Num: "07", Status: "todo"})
		authoring := []mergedPR{{Number: 20, Subject: "docs: add brief gtm/07 (#20)", Briefs: []string{"gtm/07"}, Authors: []string{"gtm/07"}}}

		if got := mergedPRStatusNotices([]*Stream{s}, authoring, nil); len(got) != 0 {
			t.Errorf("an Authors-only PR must not read as delivery; got %v", got)
		}
		if got := boardHonestyNotices([]*Stream{s}, authoring, nil); len(got) != 0 {
			t.Errorf("an Authors-only PR must not raise already-merged-unflipped; got %v", got)
		}
		applyPhantomHolds([]*Stream{s}, authoring, nil)
		if s.Briefs[0].PhantomHeld != "" {
			t.Errorf("an authoring PR must not hold the row: %q", s.Briefs[0].PhantomHeld)
		}

		// The same PR with a Brief: trailer IS a delivery (the control that proves the
		// authoring branch is what silenced it).
		delivering := []mergedPR{{Number: 20, Subject: "docs: add brief gtm/07 (#20)", Briefs: []string{"gtm/07"}, Delivers: []string{"gtm/07"}}}
		if got := mergedPRStatusNotices([]*Stream{s}, delivering, nil); len(got) != 1 || strings.Contains(got[0], "UNCONFIRMED") {
			t.Errorf("a Brief:-trailer merge must fire a confirmed NOTICE; got %v", got)
		}

		// The reader: trailers come out of the commit body, fence-aware.
		log := "aaa\x1fdocs: add brief gtm/07 (#20)\x1fAuthors: gtm:07\n\x1e" +
			"bbb\x1ffeat: gtm/08 (#21)\x1fBrief: gtm:08\n\x1e" +
			"ccc\x1ffeat: gtm/09 (#22)\x1f```\nBrief: gtm:09\n```\n\x1e" +
			"ddd\x1fMerge pull request #23 from o/brief/gtm-10-x\x1f\n\x1e"
		prs := parseMergedLog(log, func(sha string) string {
			if sha == "ddd" {
				return "feat: part\n\nBrief: gtm:10\n"
			}
			return ""
		})
		byNum := map[int]mergedPR{}
		for _, p := range prs {
			byNum[p.Number] = p
		}
		if byNum[20].deliveryOf("gtm/07") != mergeAuthoring {
			t.Errorf("PR 20 must read as authoring: %+v", byNum[20])
		}
		if byNum[21].deliveryOf("gtm/08") != mergeDelivered {
			t.Errorf("PR 21 must read as delivered: %+v", byNum[21])
		}
		if byNum[22].deliveryOf("gtm/09") != mergeUnconfirmed {
			t.Errorf("a trailer inside a code fence is documentation, not delivery: %+v", byNum[22])
		}
		if byNum[23].deliveryOf("gtm/10") != mergeDelivered {
			t.Errorf("a merge commit takes its trailers from the branch commits: %+v", byNum[23])
		}
	})

	t.Run("absent forge state stays unknown", func(t *testing.T) {
		root := t.TempDir()
		s := nlcStream(t, root, "gtm", map[string]string{"01": "## Context\nx\n"}, Brief{Num: "01", Status: "todo"})
		delivered := []mergedPR{{Number: 11, Subject: "feat: gtm/01 (#11)", Briefs: []string{"gtm/01"}, Delivers: []string{"gtm/01"}}}

		// An unreadable history holds nothing, even when a stale slice came back with it.
		applyPhantomHolds([]*Stream{s}, delivered, errNoGitForMerge)
		if s.Briefs[0].PhantomHeld != "" {
			t.Errorf("a failed merge read must not hold a row: %q", s.Briefs[0].PhantomHeld)
		}
		notices := strings.Join(boardHonestyNotices([]*Stream{s}, delivered, errNoGitForMerge), "\n")
		if !strings.Contains(notices, "could-not-check") || strings.Contains(notices, "NON-DISPATCHABLE (already-merged-unflipped)") {
			t.Errorf("a failed merge read must say could-not-check and conclude nothing; got:\n%s", notices)
		}

		// A tree with no .git is the same state through the emitter-side entry point.
		holdPhantomRowsFromGit([]*Stream{s}, root)
		if s.Briefs[0].PhantomHeld != "" {
			t.Errorf("a tree with no git history must hold nothing: %q", s.Briefs[0].PhantomHeld)
		}

		// A merge that names the brief but whose delivery is unconfirmed is a lead, not a hold.
		unconfirmed := []mergedPR{{Number: 11, Subject: "feat: gtm/01 (#11)", Briefs: []string{"gtm/01"}}}
		applyPhantomHolds([]*Stream{s}, unconfirmed, nil)
		if s.Briefs[0].PhantomHeld != "" {
			t.Errorf("a subject-only merge must not hold a row: %q", s.Briefs[0].PhantomHeld)
		}
		if got := mergedPRStatusNotices([]*Stream{s}, unconfirmed, nil); len(got) != 1 || !strings.Contains(got[0], "UNCONFIRMED") {
			t.Errorf("an unconfirmed merge must say so; got %v", got)
		}
	})

	t.Run("the Verified actor follows the accepted run", func(t *testing.T) {
		const first, second = "assay-worker-app[bot]", vfRunner
		run := func(runner, sha, date string) string {
			return witnessTableFor(
				"| 1 | `true` | pass exit=0 | sha256:6f1a0b3c9d22 | "+date+" | "+runner+" @ "+sha+" |",
				"| 2 | `false` | pass exit=1 | sha256:aa11bb22cc33 | "+date+" | "+runner+" @ "+sha+" |") +
				"\n\n" + vfPass + "\n\n"
		}
		evidence := run(first, "aaaaaaaaaaaa", "2026-08-10") + run(second, vfSHA, "2026-08-13")

		stamp, _, _, _, err := flipStampFromEvidence(evidence, vfSHA, second)
		if err != nil || !strings.Contains(stamp, second) || strings.Contains(stamp, first) {
			t.Fatalf("the stamp must name the LATEST accepted run's actor: stamp=%q err=%v", stamp, err)
		}
		// The caller cannot substitute an earlier run's actor (or its own identity).
		if _, _, _, _, err := flipStampFromEvidence(evidence, vfSHA, first); err == nil {
			t.Error("the first run's actor must be refused once a later run is accepted")
		}
		if hint := acceptedRunHint(evidence); !strings.Contains(hint, second) {
			t.Errorf("the attribution notice must name the cell the accepted run derives; got %q", hint)
		}
		// Historical records without a strict PASS stay as they were: no hint, nothing derived.
		if hint := acceptedRunHint(witnessTableFor("| 1 | `true` | pass | sha256:6f1a0b3c9d22 | 2026-08-10 | someone |")); hint != "" {
			t.Errorf("a record with no accepted run must derive nothing; got %q", hint)
		}
		// The stale-cell fixture's notice is unchanged for a historical record.
		if !hasProblem(rerunNotices(t), "rerun/brief-01", "k3-verifier", "opus-verifier", "2 of 3") {
			t.Error("the historical attribution notice must still fire")
		}
	})

	t.Run("a unique register link repairs once", func(t *testing.T) {
		root, brief := nlcRegisterTree(t)
		before := nlcHash(t, filepath.Join(filepath.Dir(brief), "brief-02-frozen.md"))

		rep, err := backfillRegisterRefsReport(root)
		if err != nil || rep.Repaired != 1 {
			t.Fatalf("exactly the one unique prose link must be repaired: rep=%+v err=%v", rep, err)
		}
		raw, _ := os.ReadFile(brief)
		got := string(raw)
		if !strings.Contains(got, "[F-15](../findings/2026-07-09-one.md)") {
			t.Errorf("the unique link must point at its durable entry; got:\n%s", got)
		}
		// Frozen text stays byte-identical: the fence, the code span, the Verify command
		// and the recorded Evidence still carry the view link.
		if n := strings.Count(got, "[F-15](../FINDINGS.md)"); n != 4 {
			t.Errorf("fence, code span and Evidence links must be untouched (want 4 view links left, have %d):\n%s", n, got)
		}
		if nlcHash(t, filepath.Join(filepath.Dir(brief), "brief-02-frozen.md")) != before {
			t.Error("a file with only frozen links must be byte-identical after the pass")
		}

		// Once: the second pass finds nothing to do and writes nothing.
		again := nlcHash(t, brief)
		rep2, err := backfillRegisterRefsReport(root)
		if err != nil || rep2.Repaired != 0 || nlcHash(t, brief) != again {
			t.Errorf("the repair must be idempotent: rep=%+v err=%v", rep2, err)
		}
	})

	t.Run("an ambiguous or unmapped link does not change", func(t *testing.T) {
		root, _ := nlcRegisterTree(t)
		amb := filepath.Join(root, "docs", "streams", "gtm", "brief-03-ambiguous.md")
		before := nlcHash(t, amb)

		rep, err := backfillRegisterRefsReport(root)
		if err != nil {
			t.Fatal(err)
		}
		if nlcHash(t, amb) != before {
			t.Error("a file whose only view links are ambiguous or unmapped must be byte-identical")
		}
		joined := strings.Join(rep.Unresolved, "\n")
		if !strings.Contains(joined, "[F-20](../FINDINGS.md) — ambiguous: 2 entries") ||
			!strings.Contains(joined, "[F-99](../FINDINGS.md) — no durable entry") {
			t.Errorf("both unresolved links must be LISTED as author-owed; got:\n%s", joined)
		}
		// Unresolved links are findings of the lint, not silently dropped.
		_, notices := registerRefProblems(root, []string{amb})
		if len(notices) != 2 {
			t.Errorf("the lint must keep reporting both links; got %v", notices)
		}
	})
}

// nlcRegisterTree lays a register with one unique entry (F-15) and one ambiguous id
// (F-20), and three brief files: one with a repairable prose link plus frozen copies,
// one holding only frozen copies, one holding only ambiguous and unmapped links.
func nlcRegisterTree(t *testing.T) (root, brief string) {
	t.Helper()
	root = t.TempDir()
	fDir := filepath.Join(root, "docs", "streams", "findings")
	gDir := filepath.Join(root, "docs", "streams", "gtm")
	for _, d := range []string{fDir, gDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	entry := func(id string) string {
		return "---\nid: " + id + "\ndate: \"2026-07-09\"\ntitle: T\naffects: []\nresolved: false\n---\nBody.\n"
	}
	writeTemp(t, fDir, "2026-07-09-one.md", entry("F-15"))
	writeTemp(t, fDir, "2026-07-09-dup-a.md", entry("F-20"))
	writeTemp(t, fDir, "2026-07-09-dup-b.md", entry("F-20"))

	frozen := "```\nsee [F-15](../FINDINGS.md)\n```\n" +
		"Code: `[F-15](../FINDINGS.md)`.\n" +
		"\n## Verify\n| # | Command | Expect |\n|---|---------|--------|\n| 1 | `grep -F '[F-15](../FINDINGS.md)' x` | exit 0 |\n" +
		"\n## Evidence\nRecorded: [F-15](../FINDINGS.md) on 2026-07-10.\n"
	brief = writeTemp(t, gDir, "brief-01-repair.md", "# One\n\nSee [F-15](../FINDINGS.md) for the cause.\n\n"+frozen)
	writeTemp(t, gDir, "brief-02-frozen.md", "# Two\n\n"+frozen)
	writeTemp(t, gDir, "brief-03-ambiguous.md", "# Three\n\nSee [F-20](../FINDINGS.md) and [F-99](../FINDINGS.md).\n")
	return root, brief
}
