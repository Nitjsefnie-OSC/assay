package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The BUILT cellctl launches a stub harness; the stub prints what its own environment (and, for
// Codex, its command-environment overrides) carries for ASSAY_TELEMETRY. Nothing contacts a
// model endpoint or a forge: the stub is a shell script and the repository is a local fixture.

const claudeTelemetryStub = `#!/bin/sh
case "$1" in
 --version) echo '2.1.278 (stub)'; exit 0;;
 plugin) exit 0;;
esac
printf 'TELEMETRY=%s\n' "${ASSAY_TELEMETRY-<unset>}"
`

const codexTelemetryStub = `#!/bin/sh
printf 'TELEMETRY=%s\n' "${ASSAY_TELEMETRY-<unset>}"
for a in "$@"; do printf 'ARG=%s\n' "$a"; done
`

var telemetryLaunchCases = []struct {
	name, cell, defaults, parent, want string
}{
	{"cell opt-in", "ASSAY_TELEMETRY=1\n", "", "", "1"},
	{"cell opt-in over shell off", "ASSAY_TELEMETRY=1\n", "", "0", "1"},
	{"cell off over shell on", "ASSAY_TELEMETRY=0\n", "", "1", "0"},
	{"defaults opt-in", "", "ASSAY_TELEMETRY=1\n", "", "1"},
	{"cell off over defaults on", "ASSAY_TELEMETRY=0\n", "ASSAY_TELEMETRY=1\n", "", "0"},
	{"missing everywhere", "", "", "", telemetryUnset},
}

func telemetryLaunch(t *testing.T, harness, cell, defaults, parent string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX sh harness stub")
	}
	f := catalogFixture(t)
	f.prepareLocalLaunch(t)
	path := filepath.Join(f.cellDir, "cell.env")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, []byte(cell)...), 0600); err != nil {
		t.Fatal(err)
	}
	if defaults != "" {
		if err := os.WriteFile(filepath.Join(f.cellsRoot, "defaults.env"), []byte(defaults), 0600); err != nil {
			t.Fatal(err)
		}
	}
	extra := []string{"CELLCTL_DESKWT=0"}
	if parent != "" {
		extra = append(extra, "ASSAY_TELEMETRY="+parent)
	}
	args := []string{"desk", "example", "pr-review-desk", "--provider", "anthropic"}
	stub, name := claudeTelemetryStub, "claude"
	if harness == "codex" {
		stub, name = codexTelemetryStub, "codex"
		writeCatalogJSON(t, filepath.Join(f.cellDir, "providers.json"), map[string]any{"roles": map[string]string{"worker-desk": "codex"}})
		args = []string{"desk", "example", "worker-desk"}
	}
	if err := os.WriteFile(filepath.Join(f.binDir, name), []byte(stub), 0755); err != nil {
		t.Fatal(err)
	}
	r := f.run(t, extra, args...)
	if r.code != 0 {
		t.Fatalf("launch: %+v", r)
	}
	return r.stdout
}

// regression: #3048
func TestTelemetryClaudeLaunch(t *testing.T) {
	for _, tc := range telemetryLaunchCases {
		t.Run(tc.name, func(t *testing.T) {
			out := telemetryLaunch(t, "claude", tc.cell, tc.defaults, tc.parent)
			if !strings.Contains(out, "TELEMETRY="+tc.want+"\n") {
				t.Errorf("claude child: want TELEMETRY=%s in:\n%s", tc.want, out)
			}
		})
	}
}

// regression: #3048
func TestTelemetryCodexLaunch(t *testing.T) {
	for _, tc := range telemetryLaunchCases {
		t.Run(tc.name, func(t *testing.T) {
			out := telemetryLaunch(t, "codex", tc.cell, tc.defaults, tc.parent)
			if !strings.Contains(out, "TELEMETRY="+tc.want+"\n") {
				t.Errorf("codex process env: want TELEMETRY=%s in:\n%s", tc.want, out)
			}
			cmdEnv := `ARG=shell_environment_policy.set.ASSAY_TELEMETRY="` + tc.want + `"` + "\n"
			if has := strings.Contains(out, "ARG=shell_environment_policy.set.ASSAY_TELEMETRY="); has != (tc.want != telemetryUnset) {
				t.Errorf("codex command env override present=%v for want=%s:\n%s", has, tc.want, out)
			} else if has && !strings.Contains(out, cmdEnv) {
				t.Errorf("codex command env: want %q in:\n%s", cmdEnv, out)
			}
		})
	}
}
