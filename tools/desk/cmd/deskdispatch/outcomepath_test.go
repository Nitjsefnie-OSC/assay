package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/medici-finance/assay/tools/desk/internal/deskkit"
)

// outcomeDirLines returns the backticked directory of every "Outcome directory" line.
func outcomeDirLines(t *testing.T, prompt string) []string {
	t.Helper()
	var dirs []string
	for _, line := range strings.Split(prompt, "\n") {
		const lead = "Outcome directory (dispatcher-named, absolute, outside the home): `"
		if i := strings.Index(line, lead); i >= 0 {
			rest := line[i+len(lead):]
			if j := strings.Index(rest, "`"); j >= 0 {
				dirs = append(dirs, rest[:j])
			}
		}
	}
	return dirs
}

// regression: verifier outcome records written inside the home refuse admission
//
// TestVerifierOutcomePath: the verifier prompt names ONE absolute outcome directory outside
// the home and the dispatcher root, with the exact `deskevidence --outcome-record … --root`
// landing command; an allocation that resolves inside the home or the root (directly or
// through a symlink alias) is refused, never moved elsewhere; the cache form is owner-only
// and week-pruned; and no other kit's prompt carries the line.
func TestVerifierOutcomePath(t *testing.T) {
	t.Run("verifier prompt names one absolute outside path and the landing command", func(t *testing.T) {
		home, root, cache := t.TempDir(), t.TempDir(), t.TempDir()
		old := packetCacheDir
		packetCacheDir = func() (string, error) { return cache, nil }
		t.Cleanup(func() { packetCacheDir = old })
		o := dispatchOpts{kit: "verifier", item: "example-stream/07", root: root, brief: filepath.Join(root, vpBriefRel)}
		plan := dispatchPlan{repo: "example/repo", claimKey: "repo--example-stream--07"}
		dir, err := allocateOutcomeDir(o, plan, home)
		if err != nil {
			t.Fatalf("allocateOutcomeDir: %v", err)
		}
		plan.outcomePath = dir
		prompt, err := assemblePrompt(o, plan, home)
		if err != nil {
			t.Fatal(err)
		}
		got := outcomeDirLines(t, prompt)
		if len(got) != 1 {
			t.Fatalf("outcome directory lines = %q, want exactly one:\n%s", got, prompt)
		}
		if !filepath.IsAbs(got[0]) || vpWithin(home, got[0]) || vpWithin(root, got[0]) {
			t.Errorf("outcome directory %q is not an absolute path outside home %s and root %s", got[0], home, root)
		}
		cmd := "deskevidence example/repo main --outcome-record " + filepath.Join(dir, "<file>.json") + " --root " + home
		if strings.Count(prompt, cmd) != 1 {
			t.Errorf("prompt lacks the exact landing command %q:\n%s", cmd, prompt)
		}
		if !strings.Contains(prompt, "never write the record inside the home") {
			t.Error("prompt does not say never to write the record inside the home")
		}
	})

	t.Run("dry-run verifier prompt carries the line beside --prompt-file", func(t *testing.T) {
		asg := assignmentSection(t, dispatchPrompt(t, "verifier", "verify-item-9"))
		got := outcomeDirLines(t, asg)
		if len(got) != 1 || !filepath.IsAbs(got[0]) || filepath.Base(filepath.Dir(got[0])) != outcomeDirName ||
			filepath.Base(got[0]) != "assay--verify-item-9" {
			t.Fatalf("dry-run verifier outcome lines = %q, want one absolute <prompt dir>/%s/<claim key>:\n%s", got, outcomeDirName, asg)
		}
		if !strings.Contains(asg, "--outcome-record ") || !strings.Contains(asg, " --root ") {
			t.Errorf("dry-run verifier assignment lacks the --outcome-record … --root command:\n%s", asg)
		}
		if _, err := os.Stat(got[0]); !os.IsNotExist(err) {
			t.Errorf("a dry run created the outcome directory %s (stat err %v); it must touch nothing", got[0], err)
		}
	})

	t.Run("allocation inside the home or the root is refused, never moved", func(t *testing.T) {
		home, root := t.TempDir(), t.TempDir()
		alias := filepath.Join(t.TempDir(), "alias")
		if err := os.Symlink(home, alias); err != nil {
			if runtime.GOOS == "windows" {
				t.Skip("symlinks unavailable")
			}
			t.Fatal(err)
		}
		cases := map[string]string{
			"prompt file in the home":            filepath.Join(home, "prompt.md"),
			"prompt file deep in the home":       filepath.Join(home, "docs", "streams", "prompt.md"),
			"prompt file in the root":            filepath.Join(root, "prompt.md"),
			"prompt file through a home symlink": filepath.Join(alias, "prompt.md"),
		}
		for name, pf := range cases {
			t.Run(name, func(t *testing.T) {
				o := dispatchOpts{kit: "verifier", item: "example-stream/07", root: root, promptFile: pf}
				plan := dispatchPlan{repo: "example/repo", claimKey: "repo--example-stream--07"}
				dir, err := allocateOutcomeDir(o, plan, home)
				if err == nil || !deskkit.IsRefused(err) {
					t.Fatalf("allocateOutcomeDir(%s) = %q, %v; want a refusal", pf, dir, err)
				}
				if !strings.Contains(err.Error(), "OUTSIDE") {
					t.Errorf("refusal does not name the outside-the-home rule: %v", err)
				}
				if _, serr := os.Lstat(filepath.Join(filepath.Dir(pf), outcomeDirName)); !os.IsNotExist(serr) {
					t.Errorf("a refused allocation still created a directory (stat err %v)", serr)
				}
			})
		}
		t.Run("cache directory inside the home", func(t *testing.T) {
			old := packetCacheDir
			packetCacheDir = func() (string, error) { return filepath.Join(alias, ".cache"), nil }
			t.Cleanup(func() { packetCacheDir = old })
			o := dispatchOpts{kit: "verifier", item: "example-stream/07", root: root}
			if _, err := allocateOutcomeDir(o, dispatchPlan{claimKey: "repo--example-stream--07"}, home); err == nil || !deskkit.IsRefused(err) {
				t.Fatalf("cache allocation inside the home: err = %v, want a refusal", err)
			}
			if _, err := os.Stat(filepath.Join(home, ".cache")); !os.IsNotExist(err) {
				t.Errorf("a refused cache allocation wrote into the home (stat err %v)", err)
			}
		})
		t.Run("dry run refuses the same way", func(t *testing.T) {
			o := dispatchOpts{kit: "verifier", item: "example-stream/07", root: root, promptFile: filepath.Join(root, "p.md")}
			if _, err := outcomeDirFor(o, dispatchPlan{claimKey: "repo--example-stream--07"}, ""); err == nil || !deskkit.IsRefused(err) {
				t.Fatalf("outcomeDirFor with the prompt in the root and no home yet: err = %v, want a refusal", err)
			}
		})
	})

	t.Run("cache form is owner-only, per claim key and week-pruned", func(t *testing.T) {
		cache := t.TempDir()
		old, oldNow := packetCacheDir, packetNow
		now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
		packetCacheDir = func() (string, error) { return cache, nil }
		packetNow = func() time.Time { return now }
		t.Cleanup(func() { packetCacheDir, packetNow = old, oldNow })
		base := filepath.Join(cache, "assay", "outcomes")
		stale, fresh, file := filepath.Join(base, "stale-key"), filepath.Join(base, "fresh-key"), filepath.Join(base, "stray.json")
		for _, d := range []string{stale, fresh} {
			if err := os.MkdirAll(d, 0o700); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(file, []byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}
		for _, p := range []string{stale, file} {
			if err := os.Chtimes(p, now.Add(-8*24*time.Hour), now.Add(-8*24*time.Hour)); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.Chtimes(fresh, now.Add(-time.Hour), now.Add(-time.Hour)); err != nil {
			t.Fatal(err)
		}
		home, root := t.TempDir(), t.TempDir()
		o := dispatchOpts{kit: "verifier", item: "example-stream/07", root: root}
		dir, err := allocateOutcomeDir(o, dispatchPlan{claimKey: "repo--example-stream--07"}, home)
		if err != nil {
			t.Fatal(err)
		}
		if want := filepath.Join(base, "repo--example-stream--07"); dir != want {
			t.Errorf("cache outcome dir = %s, want %s", dir, want)
		}
		fi, err := os.Stat(dir)
		if err != nil || !fi.IsDir() {
			t.Fatalf("outcome dir not created: %v", err)
		}
		if runtime.GOOS != "windows" && fi.Mode().Perm() != 0o700 {
			t.Errorf("outcome dir mode = %v, want owner-only 0700", fi.Mode().Perm())
		}
		if _, err := os.Stat(stale); !os.IsNotExist(err) {
			t.Errorf("a week-old outcome directory was not pruned (stat err %v)", err)
		}
		for _, keep := range []string{fresh, file} {
			if _, err := os.Stat(keep); err != nil {
				t.Errorf("pruning removed %s, which it must keep: %v", keep, err)
			}
		}
	})

	t.Run("two claims with prompt files in one directory get different outcome directories", func(t *testing.T) {
		home, root, side := t.TempDir(), t.TempDir(), t.TempDir()
		oa := dispatchOpts{kit: "verifier", item: "example-stream/07", root: root, promptFile: filepath.Join(side, "prompt-07.md")}
		ob := dispatchOpts{kit: "verifier", item: "example-stream/08", root: root, promptFile: filepath.Join(side, "prompt-08.md")}
		a, err := allocateOutcomeDir(oa, dispatchPlan{claimKey: "repo--example-stream--07"}, home)
		if err != nil {
			t.Fatal(err)
		}
		b, err := allocateOutcomeDir(ob, dispatchPlan{claimKey: "repo--example-stream--08"}, home)
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(side, outcomeDirName, "repo--example-stream--07")
		if a != want || a == b || filepath.Dir(a) != filepath.Dir(b) {
			t.Errorf("beside-prompt outcome dirs = %s, %s; want per-claim subdirectories like %s", a, b, want)
		}
	})

	t.Run("traversal-shaped claim keys are refused on both branches", func(t *testing.T) {
		home, root, side, cache := t.TempDir(), t.TempDir(), t.TempDir(), t.TempDir()
		old := packetCacheDir
		packetCacheDir = func() (string, error) { return cache, nil }
		t.Cleanup(func() { packetCacheDir = old })
		branches := map[string]string{
			"beside --prompt-file": filepath.Join(side, "prompt.md"),
			"user cache":           "",
		}
		keys := []string{"../escape", "a/../../b", `a\..\b`, "..", "a..b", "/abs", "a/b", " repo--x", ""}
		for branch, pf := range branches {
			for _, key := range keys {
				t.Run(branch+"/"+key, func(t *testing.T) {
					o := dispatchOpts{kit: "verifier", item: "example-stream/07", root: root, promptFile: pf}
					dir, err := allocateOutcomeDir(o, dispatchPlan{claimKey: key}, home)
					if err == nil || !deskkit.IsRefused(err) {
						t.Fatalf("claim key %q on the %s branch = %q, %v; want a refusal", key, branch, dir, err)
					}
				})
			}
			parent := filepath.Join(side, outcomeDirName)
			if pf == "" {
				parent = filepath.Join(cache, "assay", "outcomes")
			}
			if entries, err := os.ReadDir(parent); err == nil && len(entries) != 0 {
				t.Errorf("%s branch: refused keys still created %d entr(ies) under %s", branch, len(entries), parent)
			}
		}
		for _, outside := range []string{filepath.Join(side, "escape"), filepath.Join(cache, "assay", "escape"), filepath.Join(side, "b")} {
			if _, err := os.Lstat(outside); !os.IsNotExist(err) {
				t.Errorf("a traversal-shaped key reached %s (stat err %v)", outside, err)
			}
		}
	})

	t.Run("non-verifier prompts carry no outcome line", func(t *testing.T) {
		for _, p := range []struct{ kit, item string }{{"worker", "example-stream--07"}, {"review", "review-item-9"}} {
			prompt := dispatchPrompt(t, p.kit, p.item)
			if strings.Contains(prompt, "Outcome directory") || strings.Contains(prompt, "--outcome-record") {
				t.Errorf("%s prompt carries the verifier outcome line:\n%s", p.kit, assignmentSection(t, prompt))
			}
		}
	})
}
