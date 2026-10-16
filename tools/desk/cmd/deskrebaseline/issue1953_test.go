package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Fail-first tests for #1953. Each one goes through parseVerifyRows (and, for the probe tests,
// gatherRowFacts and Classify) so it compiles against the pre-fix package and reds there.

const verifyHead = "# B\n\n## Verify\n\n| # | Class | Command | Expect |\n|---|---|---|---|\n"

// TestEscapedPipeStaysInCommand — defect 1. A `\|` in the Command cell is cell content, so
// the probe runs the whole pipeline, not the text before the first escaped pipe.
//
// regression: #1953
func TestEscapedPipeStaysInCommand(t *testing.T) {
	rows := parseVerifyRows(verifyHead +
		"| 1 | check | `grep -c a f.txt \\| wc -l \\|\\| true` | exit 0 |\n")
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %+v", rows)
	}
	if want := "grep -c a f.txt | wc -l || true"; rows[0].Command != want {
		t.Fatalf("Command = %q, want %q", rows[0].Command, want)
	}
	if rows[0].Expect != "exit 0" {
		t.Fatalf("Expect = %q, want the cell after the escaped pipes", rows[0].Expect)
	}
}

// TestRenameRerunsAgainstTarget — defect 2. The pinned file moved by one rename hop, but the
// renamed file no longer carries the row's pattern: re-run against the hop, the row fails, so
// it is refused and never re-baselined safe:rename.
//
// regression: #1953
func TestRenameRerunsAgainstTarget(t *testing.T) {
	root := gitRepo(t)
	body := strings.Repeat("// filler line for rename detection\n", 30)
	writeFile(t, root, "tools/old/w.go", "package old\n// needle\n"+body)
	gitCommit(t, root, "add w")
	if err := os.MkdirAll(filepath.Join(root, "tools/new"), 0o755); err != nil {
		t.Fatal(err)
	}
	mv := exec.Command("git", "mv", "tools/old/w.go", "tools/new/w.go")
	mv.Dir = root
	if out, err := mv.CombinedOutput(); err != nil {
		t.Fatalf("git mv: %v\n%s", err, out)
	}
	writeFile(t, root, "tools/new/w.go", "package old\n"+body) // the pattern is gone
	gitCommit(t, root, "move w, drop needle")

	rows := parseVerifyRows(verifyHead + "| 1 | check | `grep -q needle tools/old/w.go` | exit 0 |\n")
	f := gatherRowFacts(root, rows[0], false, "")
	if f.RenameHop != "tools/new/w.go" {
		t.Fatalf("positive control: rename hop = %q, want tools/new/w.go (facts %+v)", f.RenameHop, f)
	}
	if got := Classify(f); got.Verdict.IsSafe() {
		t.Fatalf("renamed file lacks the pattern: got %s (%s), want a refusal", got.Verdict, got.Reason)
	}
}

// TestExpectExit1NotChanged — defect 3. A row whose Expect IS `exit 1` and whose command still
// exits 1 passes; the probe must judge rc against Expect, not read any non-zero as changed.
//
// regression: #1953
func TestExpectExit1NotChanged(t *testing.T) {
	root := gitRepo(t)
	writeFile(t, root, "pkg/x.txt", "one\n")
	gitCommit(t, root, "add x")

	rows := parseVerifyRows(verifyHead + "| 1 | check | `test -f pkg/x.txt && exit 1` | exit 1 (refuses) |\n")
	f := gatherRowFacts(root, rows[0], false, "")
	if !f.CommandProbed || !f.CommandRan {
		t.Fatalf("positive control: the row should be probed and run: %+v", f)
	}
	if got := Classify(f); got.Verdict == RefusedBehaviourChanged {
		t.Fatalf("rc 1 matches Expect `exit 1`: got %s (%s)", got.Verdict, got.Reason)
	}
	// Perturb: the same row now exits 0, which its Expect does not allow.
	rows = parseVerifyRows(verifyHead + "| 1 | check | `test -f pkg/x.txt` | exit 1 (refuses) |\n")
	if got := Classify(gatherRowFacts(root, rows[0], false, "")); got.Verdict != RefusedBehaviourChanged {
		t.Fatalf("rc 0 against Expect `exit 1`: got %s, want %s", got.Verdict, RefusedBehaviourChanged)
	}
}
