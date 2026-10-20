package main

// perfphases_test.go — statusgen's perf record is perf-v2, the format the desk tools write:
// per-stage phase timings in pipeline order, the `http` forge counter as a measured counter,
// and a run whose stdout, stderr, exit code and written files are the same with the switch on
// as with it off.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/medici-finance/assay/perfrec"
)

// perfFixtureRun copies the goodrepo fixture into a fresh root, writes the lint roster into a
// fresh HOME and runs the built binary once with the switch on or off. args may name the root
// as "{root}". It returns the run, the root and the HOME.
func perfFixtureRun(t *testing.T, bin string, on bool, args ...string) (perfRun, string, string) {
	t.Helper()
	root := t.TempDir()
	if err := os.CopyFS(root, os.DirFS("testdata/goodrepo")); err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "assay")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "roster.env"), []byte(fixtureRoster+darFixtureExtra()), 0o600); err != nil {
		t.Fatal(err)
	}
	argv := make([]string, len(args))
	for i, a := range args {
		argv[i] = strings.ReplaceAll(a, "{root}", root)
	}
	r := runPerfBinary(t, bin, home, on, "1", argv...)
	// The root and HOME paths are per-run; fold them so two runs' output compares on substance.
	fold := strings.NewReplacer(root, "{root}", home, "{home}")
	r.out = fold.Replace(r.out)
	r.errOut = fold.Replace(r.errOut)
	return r, root, home
}

// Statusgen's record was perf-v1 with no phases, so its wall time had no breakdown.
// regression: #3101
func TestPerfV2PhasesInOrder(t *testing.T) {
	bin := buildPerfBinary(t)
	for _, c := range []struct {
		name   string
		args   []string
		exit   int
		phases []string
	}{
		{"lint", []string{"--lint", "--root", "{root}"}, 0,
			[]string{"parse", "lint", "briefs", "ordering", "links", "claims", "view", "render"}},
		{"regen", []string{"--root", "{root}"}, 0,
			[]string{"parse", "lint", "briefs", "ordering", "links", "claims", "view", "render", "write"}},
		// No git remote in the scratch working directory: corroboration stops before any
		// forge read, and the stage is still timed.
		{"corroborate", []string{"--corroborate", "--pr", "1"}, 1,
			[]string{"corroborate"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			run, _, home := perfFixtureRun(t, bin, true, c.args...)
			if run.code != c.exit {
				t.Fatalf("exit %d, want %d\nstdout: %s\nstderr: %s", run.code, c.exit, run.out, run.errOut)
			}
			recs := perfRecords(t, home)
			if len(recs) != 1 {
				t.Fatalf("one run wrote %d records, want exactly 1", len(recs))
			}
			rec := recs[0]
			if rec["schema"] != "perf-v2" {
				t.Fatalf("schema = %v, want perf-v2", rec["schema"])
			}
			if rec["boundary"] != perfrec.BoundaryProcess {
				t.Errorf("boundary = %v, want %s", rec["boundary"], perfrec.BoundaryProcess)
			}
			var got []string
			phases, _ := rec["phases"].([]any)
			for _, p := range phases {
				m, _ := p.(map[string]any)
				name, _ := m["name"].(string)
				got = append(got, name)
				if n, _ := m["n"].(float64); n < 1 {
					t.Errorf("phase %s n = %v, want >= 1", name, m["n"])
				}
			}
			if strings.Join(got, ",") != strings.Join(c.phases, ",") {
				t.Errorf("phases = %v, want %v (pipeline order)", got, c.phases)
			}
			if rec["phaseOverlap"] != false {
				t.Errorf("phaseOverlap = %v, want false: stages never nest", rec["phaseOverlap"])
			}
			counters, _ := rec["counters"].(map[string]any)
			if v, ok := counters["http"]; !ok || v != float64(0) {
				t.Errorf("counters.http = %v (present %v), want a measured 0 for an offline run", v, ok)
			}
			unavailable, _ := rec["unavailable"].([]any)
			for _, u := range unavailable {
				if u == "http" {
					t.Errorf("http listed unavailable; the forge seam is counted")
				}
			}
		})
	}
}

// The perf-v2 session must not change what a run prints, exits with or writes.
// regression: #3101
func TestPerfV2OutputUnchanged(t *testing.T) {
	bin := buildPerfBinary(t)
	for _, c := range []struct {
		name string
		args []string
	}{
		{"lint", []string{"--lint", "--root", "{root}"}},
		{"regen", []string{"--root", "{root}"}},
		{"check", []string{"--check", "--root", "{root}"}},
		{"corroborate", []string{"--corroborate", "--pr", "1"}},
		{"usage error", []string{"--no-such-flag"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			off, offRoot, offHome := perfFixtureRun(t, bin, false, c.args...)
			on, onRoot, onHome := perfFixtureRun(t, bin, true, c.args...)
			if n := len(perfRecords(t, offHome)); n != 0 {
				t.Fatalf("switch off wrote %d records", n)
			}
			if n := len(perfRecords(t, onHome)); n != 1 {
				t.Fatalf("switch on wrote %d records, want 1", n)
			}
			if on.code != off.code {
				t.Errorf("exit: on %d, off %d", on.code, off.code)
			}
			if on.out != off.out {
				t.Errorf("stdout differs:\n on  %q\n off %q", on.out, off.out)
			}
			if on.errOut != off.errOut {
				t.Errorf("stderr differs:\n on  %q\n off %q", on.errOut, off.errOut)
			}
			for _, f := range []string{"STATUS.md", "docs/streams/INTAKE.md", "docs/streams/FINDINGS.md"} {
				a, aerr := os.ReadFile(filepath.Join(onRoot, f))
				b, berr := os.ReadFile(filepath.Join(offRoot, f))
				if (aerr == nil) != (berr == nil) || string(a) != string(b) {
					t.Errorf("%s differs between on and off runs", f)
				}
			}
		})
	}
}
