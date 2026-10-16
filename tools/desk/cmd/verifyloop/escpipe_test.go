package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/medici-finance/assay/tools/desk/internal/loopengine"
)

// TestRunnerGetsWholePipeline — #1953's class, its second instance: the deterministic
// runner's Verify parser split rows on every `|`, escaped or not, so a row whose Command is a
// pipeline reached the executor cut at the first `\|`. The runner must hand the executor the
// un-escaped pipeline.
//
// regression: #1953
func TestRunnerGetsWholePipeline(t *testing.T) {
	root := t.TempDir()
	brief := "# B\n\n## Verify\n\n| # | Class | Command | Expect |\n|---|---|---|---|\n" +
		"| 1 | check | `grep -c a f.txt \\| wc -l \\|\\| true` | exit 0 |\n"
	if err := os.WriteFile(filepath.Join(root, "b.md"), []byte(brief), 0o644); err != nil {
		t.Fatal(err)
	}
	var ran []string
	exec := func(_, command string) (int, string) { ran = append(ran, command); return 0, "" }
	runBriefRows(root, loopengine.Item{BriefPath: "b.md"}, exec)
	if want := "grep -c a f.txt | wc -l || true"; len(ran) != 1 || ran[0] != want {
		t.Fatalf("executor ran %q, want exactly [%q]", ran, want)
	}
}
