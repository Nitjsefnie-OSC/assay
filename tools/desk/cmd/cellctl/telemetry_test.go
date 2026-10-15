package main

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The one opt-in ASSAY_TELEMETRY reaches a launched desk's tools only as the EXACT value 1,
// resolved cell.env over defaults.env over the launching shell, and an explicit off stays off.
// Every test below observes the CHILD's environment (a real process), never only a composed
// slice, so a launcher that composes the value and then drops it fails here.

const telemetryUnset = "<unset>"

// telemetryChild runs a real child with exactly env and reports what it sees for the key.
func telemetryChild(t *testing.T, name string, env []string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX sh child")
	}
	out, err := exec.Command("/usr/bin/env", append(append([]string{"-i"}, env...),
		"/bin/sh", "-c", `printf %s "${ASSAY_TELEMETRY-`+telemetryUnset+`}"`)...).CombinedOutput()
	if err != nil {
		t.Fatalf("%s: child: %v: %s", name, err, out)
	}
	return string(out)
}

func telemetryEnvLayers(process, defaults, cell string) *Env {
	e := &Env{vals: map[string]string{}, set: map[string]bool{}}
	if process != "" {
		e.putFrom(envAssayTelemetry, strings.TrimPrefix(process, "="), layerProcess)
	}
	if defaults != "" {
		e.putFrom(envAssayTelemetry, strings.TrimPrefix(defaults, "="), layerDefaults)
	}
	if cell != "" {
		e.putFrom(envAssayTelemetry, strings.TrimPrefix(cell, "="), layerCellEnv)
	}
	return e
}

// telemetryCases: layers are "" (layer absent) or "=<value>" (layer sets <value>, which may be
// empty); parent is the launching shell's ASSAY_TELEMETRY as os.Environ() carried it ("" =
// absent); want is what a launched child must see.
var telemetryCases = []struct {
	name, process, defaults, cell, parent, want string
}{
	{"cell opt-in, shell silent", "", "", "=1", "", "1"},
	{"cell opt-in over shell off", "=0", "", "=1", "0", "1"},
	{"cell explicit off over shell on", "=1", "", "=0", "1", "0"},
	{"defaults opt-in, shell silent", "", "=1", "", "", "1"},
	{"defaults opt-in over shell off", "=0", "=1", "", "0", "1"},
	{"cell off over defaults on", "", "=1", "=0", "", "0"},
	{"cell on over defaults off", "", "=0", "=1", "", "1"},
	{"shell opt-in alone is carried", "=1", "", "", "1", "1"},
	{"nothing set stays absent", "", "", "", "", telemetryUnset},
	{"truthy word is not opt-in", "", "", "=true", "", "0"},
	{"padded one is not opt-in", "", "", "= 1", "", "0"},
	{"empty cell value is off", "=1", "", "=", "1", "0"},
}

// regression: #3048
func TestTelemetryResolvedValue(t *testing.T) {
	for _, tc := range telemetryCases {
		c := &Cell{Env: telemetryEnvLayers(tc.process, tc.defaults, tc.cell)}
		want := map[string]string{"1": "1", "0": "0", telemetryUnset: ""}[tc.want]
		if got := c.telemetryValue(); got != want {
			t.Errorf("%s: telemetryValue = %q, want %q", tc.name, got, want)
		}
	}
}

// regression: #3048
func TestTelemetryHouseChildEnv(t *testing.T) {
	for _, tc := range telemetryCases {
		c := &Cell{Env: telemetryEnvLayers(tc.process, tc.defaults, tc.cell)}
		parent := []string{"PATH=/usr/bin:/bin", "GH_TOKEN=never-forwarded"}
		if tc.parent != "" {
			parent = append(parent, envAssayTelemetry+"="+tc.parent)
		}
		got := telemetryChild(t, tc.name, c.telemetryLaunchEnv(parent))
		if got != tc.want {
			t.Errorf("%s: house child sees %q, want %q", tc.name, got, tc.want)
		}
	}
}

// regression: #3048
func TestTelemetryHouseKeepsOthers(t *testing.T) {
	c := &Cell{Env: telemetryEnvLayers("", "", "=1")}
	in := []string{"PATH=/x", "DESK_LOOP=worker-desk"}
	out := c.telemetryLaunchEnv(in)
	if len(in) != 2 || in[0] != "PATH=/x" || in[1] != "DESK_LOOP=worker-desk" {
		t.Fatalf("input slice mutated: %v", in)
	}
	if len(out) != 3 || out[0] != in[0] || out[1] != in[1] || out[2] != "ASSAY_TELEMETRY=1" {
		t.Fatalf("only the one key may be added: %v", out)
	}
}

// regression: #3048
func TestTelemetryCodexCommandEnv(t *testing.T) {
	for _, tc := range telemetryCases {
		c := codexEnvironmentCell(t)
		c.Env = mergeEnv(c.Env, telemetryEnvLayers(tc.process, tc.defaults, tc.cell))
		parent := []string{"HOME=" + c.Env.Get("HOME"), "GH_TOKEN=never-on-argv", "CUSTOM_SECRET=never-on-argv"}
		if tc.parent != "" {
			parent = append(parent, envAssayTelemetry+"="+tc.parent)
		}
		args, err := c.codexEnvironmentArgs(parent)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		got := telemetryUnset
		for _, arg := range args {
			if strings.Contains(arg, "never-on-argv") {
				t.Fatalf("%s: credential copied onto argv", tc.name)
			}
			if rest, ok := strings.CutPrefix(arg, "shell_environment_policy.set."+envAssayTelemetry+"="); ok {
				if err := json.Unmarshal([]byte(rest), &got); err != nil {
					t.Fatal(err)
				}
			}
		}
		if got != tc.want {
			t.Errorf("%s: Codex command env = %q, want %q", tc.name, got, tc.want)
		}
	}
}

// regression: #3048
func TestTelemetryScrubbedChildEnv(t *testing.T) {
	for _, tc := range telemetryCases {
		c := scrubbedFixture(t)
		c.Env = mergeEnv(c.Env, telemetryEnvLayers(tc.process, tc.defaults, tc.cell))
		for _, harness := range []string{"claude", "codex"} {
			composed := c.scrubbedComposeEnv("worker-desk", harness, "s")
			// The scrubbed launch never inherits: a shell value is only the Env's own layer.
			got := telemetryChild(t, tc.name, composed.Pairs)
			if got != tc.want {
				t.Errorf("%s/%s: scrubbed child sees %q, want %q", tc.name, harness, got, tc.want)
			}
			if !valueIn(envAssayTelemetry, scrubbedEnvKeys) {
				t.Fatalf("%s is not in the scrubbed allowlist", envAssayTelemetry)
			}
		}
	}
}

// regression: #3048
func TestTelemetryScrubbedAddsOnlyKey(t *testing.T) {
	c := scrubbedFixture(t)
	before := c.scrubbedComposeEnv("worker-desk", "claude", "s").Sorted
	c.Env.Put("GH_TOKEN", "leaked-token-value")
	c.Env.Put(envAssayTelemetry, "1")
	after := c.scrubbedComposeEnv("worker-desk", "claude", "s").Sorted
	if len(after) != len(before)+1 || !valueIn("ASSAY_TELEMETRY=1", after) {
		t.Fatalf("scrubbed env must gain exactly ASSAY_TELEMETRY=1:\nbefore %v\nafter  %v", before, after)
	}
	for _, kv := range after {
		if strings.Contains(kv, "leaked") {
			t.Fatalf("parent credential reached the scrubbed env: %s", kv)
		}
	}
	_ = filepath.Separator
}

// mergeEnv overlays b onto a copy of a (b's keys win, keeping b's layer).
func mergeEnv(a, b *Env) *Env {
	out := &Env{vals: map[string]string{}, set: map[string]bool{}}
	for _, src := range []*Env{a, b} {
		for k := range src.set {
			out.putFrom(k, src.vals[k], src.Source(k))
		}
	}
	return out
}
