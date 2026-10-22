package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// noticeFixtureBrief writes one brief that trips every included emitter family:
// a vacuous go test -run row and an ERE literal pipe (Family A), an unresolved
// identifier (id-dereference), a risk-path file under all-"no" risk answers
// (risk-files-crossread), no why and exec-tier strong without a rationale
// (checkBriefFiles), and gate-shaped prose with no typed edge (ordering gate).
func noticeFixtureBrief(num, slug string, humanGate bool) string {
	gate, risk := "model", "{regulatory: no, customer: no, irreversible: no, sensitive-data: no}"
	if humanGate {
		gate, risk = "human", "{regulatory: yes, customer: no, irreversible: no, sensitive-data: no}"
	}
	body := fmt.Sprintf(`---
brief: demo/%s
title: fixture brief %s for the notice report
wave: 0
depends: []
unblocks: []
effort: S
gate: %s
risk: %s
issues: [1]
schema: brief-v1
authored: 2026-09-23 by fixture
exec-tier: strong
sources: ["fixture: statusgen/28"]
---

# Fixture %s

## Context
files: secrets/key.txt

This work is blocked on demo/01 landing first.
It names §TestPhantomNoticeName§, which exists in no file.

## Verify (executable — no prose-only DoD items)
| # | Command | Expect |
|---|---------|--------|
| 1 | §go test ./x/... -run TestFixtureBare§ | PASS |
| 2 | §grep -E 'a\\|b' f.txt§ | exit 0 |

## Evidence
Recorded by fixture.

## Review
Gate: %s. Reviewed 2026-09-23 human:fixture.
`, num, slug, gate, risk, slug, gate)
	return strings.ReplaceAll(body, "§", "`")
}

// noticeFixtureRoot builds a one-stream tree whose briefs cover every standing:
// 01 active, 02 done and 03 verified (both already closed at the base), 04 done
// but closed on this branch, 05 a brief file with no README row.
func noticeFixtureRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "streams", "demo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	readme := `---
stream: demo
status: active
priority: P2
track: platform
---

# demo

| # | Brief | Wave | Effort | Status | Verified | Reviewed |
|---|-------|------|--------|--------|----------|----------|
| 01 | [Active](./brief-01-active.md) | 0 | S | todo | — | — |
| 02 | [PreDone](./brief-02-predone.md) | 0 | S | done | 2026-09-23 human:fixture | 2026-09-23 human:fixture |
| 03 | [PreVerified](./brief-03-preverified.md) | 0 | S | verified | 2026-09-23 human:fixture | — |
| 04 | [NewlyClosed](./brief-04-newlyclosed.md) | 0 | S | done | 2026-10-20 human:fixture | 2026-10-20 human:fixture |
`
	files := map[string]string{
		"README.md":                 readme,
		"brief-01-active.md":        noticeFixtureBrief("01", "active", false),
		"brief-02-predone.md":       noticeFixtureBrief("02", "predone", true),
		"brief-03-preverified.md":   noticeFixtureBrief("03", "preverified", false),
		"brief-04-newlyclosed.md":   noticeFixtureBrief("04", "newlyclosed", false),
		"brief-05-unknownstatus.md": noticeFixtureBrief("05", "unknownstatus", false),
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// resetNoticeReportState restores the package-level report switches around a
// test, so one test's --all / --json mode never leaks into the next.
func resetNoticeReportState(t *testing.T) {
	t.Helper()
	d, j, a, p := lintDetailAll, lintJSONMode, activeRecordIndex, pendingNoticeReport
	t.Cleanup(func() { lintDetailAll, lintJSONMode, activeRecordIndex, pendingNoticeReport = d, j, a, p })
	lintDetailAll, lintJSONMode, activeRecordIndex, pendingNoticeReport = false, false, nil, nil
}

// lintFixture runs the real lint entrypoint over a fixture root and returns the
// exit code, stdout and stderr.
func lintFixture(t *testing.T, root string) (code int, stdout, stderr string) {
	t.Helper()
	stderr = captureStderr(t, func() {
		stdout = captureStdout(t, func() { code = run(root, "lint", nil, nil, "") })
	})
	return code, stdout, stderr
}

var historicalSummaryRe = regexp.MustCompile(`^NOTICE: \[historical-notices\] ([a-z0-9-]+): (\d+) finding\(s\) on (\d+) closed record\(s\)`)

// notice28Rules are the rules the fixture trips through their REAL emitters: one
// Family A (structural Verify row), the Family B strength / dereference /
// metadata checks, the risk cross-read and the ordering gate.
var notice28Rules = []string{
	"gotest-run-vacuous", "no-output-assertion", "table-touches-no-files",
	"id-dereference", "risk-files-crossread", "why-absent", "exec-tier-why-absent",
	"ordering-gate",
}

// TestHistoricalNoticeReport is the statusgen/28 verification row. Matrix over
// the standing of the record a finding lands on (active, pre-existing done,
// pre-existing verified, unknown status, newly closed) crossed with every
// included emitter, driven for real through `--lint`, plus text / JSON /
// --all agreement and the closure / witness exit codes.
//
// regression: F-historical-notices-lint-noise — before the shared report the
// default --lint printed every finding on a closed brief as its own NOTICE
// (2179 lines on the real tree) beside three private per-emitter aggregates.
func TestHistoricalNoticeReport(t *testing.T) {
	root := noticeFixtureRoot(t)
	withBase(t, true, "demo/02", "demo/03")

	t.Run("default text: actionable lines + bounded per-rule totals", func(t *testing.T) {
		resetNoticeReportState(t)
		code, _, stderr := lintFixture(t, root)
		if code == 0 {
			t.Fatalf("the newly closed brief's witness-absence closure must keep --lint red, got exit 0:\n%s", stderr)
		}
		for _, rule := range notice28Rules {
			var summary string
			for _, l := range strings.Split(stderr, "\n") {
				if m := historicalSummaryRe.FindStringSubmatch(l); m != nil && m[1] == rule {
					summary = l
				}
			}
			if summary == "" {
				t.Errorf("rule %q: no [historical-notices] total — the emitter is not recognised by the shared report (inventory drift):\n%s", rule, stderr)
				continue
			}
			// The closed (done + verified) briefs' own findings must not print
			// individually by default.
			for _, closed := range []string{"brief-02-predone.md", "brief-03-preverified.md"} {
				for _, l := range strings.Split(stderr, "\n") {
					if strings.HasPrefix(l, "NOTICE: ") && noticeRuleOf(strings.TrimPrefix(l, "NOTICE: ")) == rule && strings.Contains(l, closed) {
						t.Errorf("rule %q: pre-existing closed record still printed individually by default: %.200s", rule, l)
					}
				}
			}
		}
		// Standings that must STAY actionable, one line each per rule they trip:
		// active (01), newly closed (04), unknown / no README row (05).
		for _, rule := range []string{"gotest-run-vacuous", "id-dereference", "ordering-gate", "why-absent"} {
			for _, brief := range []string{"brief-01-active.md", "brief-04-newlyclosed.md", "brief-05-unknownstatus.md"} {
				found := false
				for _, l := range strings.Split(stderr, "\n") {
					if strings.HasPrefix(l, "NOTICE: ") && noticeRuleOf(strings.TrimPrefix(l, "NOTICE: ")) == rule && strings.Contains(l, brief) {
						found = true
					}
				}
				if !found {
					t.Errorf("rule %q on %s must remain an individual actionable NOTICE (active / newly-closed / unknown are never historical)", rule, brief)
				}
			}
		}
		// Output is bounded by the number of rules, not the size of the history.
		var summaries int
		for _, l := range strings.Split(stderr, "\n") {
			if historicalSummaryRe.MatchString(l) {
				summaries++
			}
		}
		if summaries > len(noticeRuleInventory) {
			t.Errorf("%d summary lines exceed the %d-rule inventory", summaries, len(noticeRuleInventory))
		}
		if !strings.Contains(stderr, "statusgen: notices: ") {
			t.Errorf("missing the plain totals line:\n%s", stderr)
		}
	})

	t.Run("PROBLEMs and exit code are untouched by the report", func(t *testing.T) {
		resetNoticeReportState(t)
		_, _, plain := lintFixture(t, root)
		resetNoticeReportState(t)
		lintDetailAll = true
		code, _, all := lintFixture(t, root)
		count := func(s, prefix string) int {
			n := 0
			for _, l := range strings.Split(s, "\n") {
				if strings.HasPrefix(l, prefix) {
					n++
				}
			}
			return n
		}
		if p1, p2 := count(plain, "PROBLEM:"), count(all, "PROBLEM:"); p1 != p2 || p1 == 0 {
			t.Errorf("PROBLEM lines differ between default (%d) and --all (%d) or are absent", p1, p2)
		}
		if code == 0 {
			t.Errorf("exit code must stay nonzero while PROBLEMs exist")
		}
		// closure / witness failures remain, named, on the branch's own closure.
		for _, want := range []string{"cannot close as done", "brief-04-newlyclosed.md"} {
			if !strings.Contains(all, want) {
				t.Errorf("closure/witness failure %q lost from --all output", want)
			}
		}
	})

	t.Run("--all restores the individual historical lines", func(t *testing.T) {
		resetNoticeReportState(t)
		_, _, def := lintFixture(t, root)
		resetNoticeReportState(t)
		lintDetailAll = true
		_, _, all := lintFixture(t, root)
		var hist int
		for _, l := range strings.Split(def, "\n") {
			if m := historicalSummaryRe.FindStringSubmatch(l); m != nil {
				var n int
				fmt.Sscanf(m[2], "%d", &n)
				hist += n
			}
		}
		if hist == 0 {
			t.Fatalf("fixture produced no historical findings:\n%s", def)
		}
		var restored int
		for _, l := range strings.Split(all, "\n") {
			if strings.HasPrefix(l, "NOTICE: ") && (strings.Contains(l, "brief-02-predone.md") || strings.Contains(l, "brief-03-preverified.md")) {
				if noticeRuleOf(strings.TrimPrefix(l, "NOTICE: ")) != "" {
					restored++
				}
			}
			if strings.Contains(l, "[historical-notices]") {
				t.Errorf("--all must print individual lines, not summaries: %.160s", l)
			}
		}
		if restored != hist {
			t.Errorf("--all restored %d individual historical lines; the default totals said %d", restored, hist)
		}
	})

	t.Run("JSON totals agree with text totals and are deterministic", func(t *testing.T) {
		resetNoticeReportState(t)
		_, _, text := lintFixture(t, root)
		var ta, th, tn int
		for _, l := range strings.Split(text, "\n") {
			if _, err := fmt.Sscanf(l, "statusgen: notices: %d actionable, %d historical (%d total)", &ta, &th, &tn); err == nil {
				break
			}
		}
		run1 := func() (int, string, string) {
			resetNoticeReportState(t)
			lintJSONMode = true
			return lintFixture(t, root)
		}
		code, out1, _ := run1()
		_, out2, _ := run1()
		if out1 != out2 {
			t.Errorf("--lint --json is not deterministic across runs")
		}
		var rep noticeReport
		if err := json.Unmarshal([]byte(out1), &rep); err != nil {
			t.Fatalf("stdout is not the JSON report: %v\n%s", err, out1)
		}
		if rep.Schema != noticeReportSchema {
			t.Errorf("schema = %q", rep.Schema)
		}
		if rep.Totals.Actionable != ta || rep.Totals.Historical != th || rep.Totals.Notices != tn || tn == 0 {
			t.Errorf("JSON totals %+v disagree with text totals %d/%d/%d", rep.Totals, ta, th, tn)
		}
		if rep.Verdict != "FAIL" || rep.Problems == 0 || code == 0 {
			t.Errorf("JSON must not hide PROBLEMs or change the exit code: verdict=%s problems=%d code=%d", rep.Verdict, rep.Problems, code)
		}
		var sumA, sumH int
		for _, rc := range rep.ByRule {
			sumA += rc.Actionable
			sumH += rc.Historical
		}
		if sumA != rep.Totals.Actionable || sumH != rep.Totals.Historical {
			t.Errorf("per-rule counts (%d/%d) do not sum to totals (%d/%d)", sumA, sumH, rep.Totals.Actionable, rep.Totals.Historical)
		}
		var stA, stH int
		got := map[string]statusCount{}
		for _, sc := range rep.ByStatus {
			stA += sc.Actionable
			stH += sc.Historical
			got[sc.Status] = sc
		}
		if stA != rep.Totals.Actionable || stH != rep.Totals.Historical {
			t.Errorf("per-status counts (%d/%d) do not sum to totals", stA, stH)
		}
		for _, st := range []string{"active", "done-pre-existing", "verified-pre-existing", "closed-this-branch", "unknown"} {
			if _, ok := got[st]; !ok {
				t.Errorf("by_status lacks the %q standing: %+v", st, rep.ByStatus)
			}
		}
		// An included rule on a pre-existing closed record is never actionable;
		// a notice outside the inventory stays actionable wherever it lands.
		for _, f := range rep.Actionable {
			if (f.Status == "done-pre-existing" || f.Status == "verified-pre-existing") && f.Rule != noticeRuleUnclassified {
				t.Errorf("included rule %q on a pre-existing closed record must be historical: %.160s", f.Rule, f.Line)
			}
		}
		if got["closed-this-branch"].Historical != 0 || got["active"].Historical != 0 || got["unknown"].Historical != 0 {
			t.Errorf("active / newly-closed / unknown must never be historical: %+v", got)
		}
		// The no-.git fixture tree cannot resolve git history for several
		// checks; the report must carry the third state, not a clean.
		if !bytes.Contains([]byte(out1), []byte("could_not_check")) {
			t.Errorf("JSON lacks a could_not_check section")
		}
	})
}

// TestHistoricalNoticeBaseUnresolved: with no resolvable merge-base every closed
// record is grandfathered. The report must SAY so (a could-not-check state),
// never present the demotion as a verified fact.
func TestHistoricalNoticeBaseUnresolved(t *testing.T) {
	root := noticeFixtureRoot(t)
	withBase(t, false)
	resetNoticeReportState(t)
	lintJSONMode = true
	_, out, _ := lintFixture(t, root)
	var rep noticeReport
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
	found := false
	for _, c := range rep.CouldNotCheck {
		if c.Scope == "closure-baseline" {
			found = true
		}
	}
	if !found {
		t.Errorf("unresolved merge-base must surface a closure-baseline could-not-check: %+v", rep.CouldNotCheck)
	}
	// nil index: every finding stays actionable, with a record-standing state.
	rep2 := buildNoticeReport([]string{"[gotest-run-vacuous] demo/brief-02-predone.md: x"}, nil)
	if rep2.Totals.Historical != 0 || rep2.Totals.Actionable != 1 {
		t.Errorf("without standings nothing may be demoted: %+v", rep2.Totals)
	}
	var sawStanding bool
	for _, c := range rep2.CouldNotCheck {
		if c.Scope == "record-standing" {
			sawStanding = true
		}
	}
	if !sawStanding {
		t.Errorf("nil index must report record-standing could-not-check")
	}
}

// TestNoticeInventoryMatchesEmitters is the drift guard: every phrase-matched
// rule must still be spelled by its production emitter, and every tag-matched
// rule must be recognised from its own bracket tag. A renamed emitter message
// would otherwise silently fall into the (unclassified) bucket.
func TestNoticeInventoryMatchesEmitters(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("cannot list production sources: %v", err)
	}
	var src strings.Builder
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") || f == "notice_report.go" {
			continue
		}
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		src.Write(b)
	}
	for _, fr := range noticeFragmentRules {
		if !strings.Contains(src.String(), fr.fragment) {
			t.Errorf("rule %q: no production emitter spells %q any more — update noticeFragmentRules with the emitter", fr.rule, fr.fragment)
		}
		if _, ok := noticeRuleFamily[fr.rule]; !ok {
			t.Errorf("fragment rule %q is not in noticeRuleInventory", fr.rule)
		}
	}
	for _, r := range noticeRuleInventory {
		if got := noticeRuleOf("[" + r.ID + "] x"); got != r.ID {
			// Phrase-matched rules have no bracket form; they are covered above.
			isFragment := false
			for _, fr := range noticeFragmentRules {
				if fr.rule == r.ID {
					isFragment = true
				}
			}
			if !isFragment {
				t.Errorf("inventory rule %q is neither tag- nor phrase-recognised", r.ID)
			}
		}
	}
	// The ID-dereference constant is the one tag linkcheck.go emits.
	if ruleIDDereference != "id-dereference" {
		t.Errorf("ruleIDDereference drifted: %q", ruleIDDereference)
	}
}

// TestSecondSiteHistoricalPartition plants a SECOND, independently-shaped
// instance of the defect class (a different rule on a different closed record)
// to show the partition is the shared classifier and not a fixture-specific
// path match.
func TestSecondSiteHistoricalPartition(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "alpha")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"brief-07-a.md", "brief-08-b.md", "brief-09-c.md"} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	streams := []*Stream{{Name: "alpha", Dir: dir, Briefs: []Brief{{Num: "07", Status: "verified"}, {Num: "08", Status: "in-progress"}, {Num: "09", Status: "done"}}}}
	idx := buildRecordIndex(streams, map[string]bool{"alpha/07": true}, true)
	lines := []string{
		"/x/docs/streams/alpha/brief-07-a.md: Verify row 3 [verify-row-portability] bad",
		"/x/docs/streams/alpha/brief-08-b.md: Verify row 3 [verify-row-portability] bad",
		"/x/docs/streams/alpha/brief-09-c.md: Verify row 3 [verify-row-portability] bad",
		"[id-dereference] docs/streams/alpha/brief-07-a.md: named identifier `X` resolves against no file",
		"unrelated aggregate notice with no rule",
	}
	rep := buildNoticeReport(lines, idx)
	if rep.Totals.Historical != 2 || rep.Totals.Actionable != 3 {
		t.Fatalf("want 2 historical (07 x2) and 3 actionable (08 active, 09 closed-this-branch, unclassified): %+v", rep.Totals)
	}
	for _, f := range rep.Historical {
		if !strings.Contains(f.Line, "brief-07") {
			t.Errorf("only the pre-existing closed record may be historical: %s", f.Line)
		}
	}
}

// TestArchiveExcludedFromLifecycleOnly pins task 4: docs/archive is out of the
// live lifecycle classification and nowhere else.
func TestArchiveExcludedFromLifecycleOnly(t *testing.T) {
	root := t.TempDir()
	for _, p := range []string{"docs/archive/old.md", "docs/live.md"} {
		full := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("**Status:** approved\n\nbody\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cands, _ := lifecycleCandidateDocs(root)
	var sawLive, sawArchive bool
	for _, c := range cands {
		switch filepath.ToSlash(strings.TrimPrefix(c, root+"/")) {
		case "docs/live.md":
			sawLive = true
		case "docs/archive/old.md":
			sawArchive = true
		}
	}
	if !sawLive || sawArchive {
		t.Errorf("lifecycle candidates: live=%v archive=%v (want true,false): %v", sawLive, sawArchive, cands)
	}
	docs, _ := docFiles(root)
	var linkSeesArchive bool
	for _, d := range docs {
		if strings.Contains(filepath.ToSlash(d), "docs/archive/old.md") {
			linkSeesArchive = true
		}
	}
	if !linkSeesArchive {
		t.Errorf("docs/archive must stay in the link-checked set; only lifecycle classification excludes it")
	}
}
