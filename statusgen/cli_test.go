package main

// The three CLI contract tests every migrated command declares (desktools-v2
// cli-contract.md): help and version answer offline, the declared settings resolve
// through the declared sources only, and every legacy invocation form callers use keeps
// reaching the handler unchanged. The fixture tables below are copied unchanged from the
// desk module's internal/clicontract/fixtures.go: statusgen is its own module and does not
// import desk internals.

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ---- fixtures (copied from tools/desk/internal/clicontract/fixtures.go) ----

func Opt(s string) *string { return &s }

var PathValues = []struct{ Name, Value string }{
	{"posix-abs", "/var/lib/assay/cell.env"},
	{"posix-space", "/tmp/a dir/with space.txt"},
	{"posix-rel", "./rel/../x"},
	{"win-drive", `C:\Users\op\cell.env`},
	{"win-drive-space", `C:\Program Files\assay\x y.txt`},
	{"win-unc", `\\srv\share\dir\f.md`},
	{"win-rel", `..\rel\x`},
	{"win-forward", "C:/mixed/sep/f"},
	{"dash-leading", "-not-a-flag"},
	{"equals-comma", "a=b,c=d"},
	{"unicode", "dossier/übersicht-ß.md"},
}

type OptionForm struct {
	Name       string
	GoFlagOnly bool
	Args       func(flag, value string) []string
}

var OptionForms = []OptionForm{
	{Name: "long-equals", Args: func(f, v string) []string { return []string{"--" + f + "=" + v} }},
	{Name: "long-space", Args: func(f, v string) []string { return []string{"--" + f, v} }},
	{Name: "go-equals", GoFlagOnly: true, Args: func(f, v string) []string { return []string{"-" + f + "=" + v} }},
	{Name: "go-space", GoFlagOnly: true, Args: func(f, v string) []string { return []string{"-" + f, v} }},
}

type PrecedenceCase struct {
	Name              string
	Flag, Env, Config *string
	Want, WantSource  string
}

const PrecedenceDefault = "dflt"

var Precedence = []PrecedenceCase{
	{Name: "nothing-set", Want: PrecedenceDefault, WantSource: "default"},
	{Name: "config-only", Config: Opt("cfg"), Want: "cfg", WantSource: "config"},
	{Name: "env-over-config", Env: Opt("env"), Config: Opt("cfg"), Want: "env", WantSource: "env"},
	{Name: "flag-over-all", Flag: Opt("flg"), Env: Opt("env"), Config: Opt("cfg"), Want: "flg", WantSource: "flag"},
	{Name: "empty-config-beats-default", Config: Opt(""), Want: "", WantSource: "config"},
	{Name: "empty-env-beats-config", Env: Opt(""), Config: Opt("cfg"), Want: "", WantSource: "env"},
	{Name: "empty-flag-beats-env", Flag: Opt(""), Env: Opt("env"), Want: "", WantSource: "flag"},
	{Name: "unset-env-falls-through", Config: Opt("cfg"), Want: "cfg", WantSource: "config"},
}

func HelpForms(sub string) [][]string {
	return [][]string{
		{"-h"}, {"--help"}, {"help"}, {"help", sub},
		{sub, "-h"}, {sub, "--help"},
	}
}

func GoFlagHelpForms(sub string) [][]string {
	return [][]string{{"-help"}, {sub, "-help"}}
}

var VersionForms = [][]string{{"--version"}}

var GoFlagVersionForms = [][]string{{"-version"}}

var MalformedConfigs = []struct{ Name, Content string }{
	{"absent", ""},
	{"not-key-value", "this is not a config\x00\n"},
	{"unterminated-quote", "KEY=\"open\n"},
}

func EnvFoldsCase(goos string) bool { return goos == "windows" }

// wantUsageExit is the contract's exit code for a parse error, the code every statusgen
// verb's Go-flag parser used. A literal, so a change to the adapter's constant reddens.
const wantUsageExit = 2

// ---- helpers ----------------------------------------------------------------

// commandPaths is every command of the statusgen tree below the root, as the words that
// reach it ("verifyrun", "outcomes split", …).
func commandPaths() [][]string {
	var out [][]string
	var walk func(c *cobra.Command, prefix []string)
	walk = func(c *cobra.Command, prefix []string) {
		for _, sub := range c.Commands() {
			p := append(append([]string{}, prefix...), sub.Name())
			out = append(out, p)
			walk(sub, p)
		}
	}
	walk(newStatusgenRoot(cliIO{}), nil)
	return out
}

// expandSub replaces the placeholder word "\x00" of a fixture form with a command path.
func expandSub(form, path []string) []string {
	var out []string
	for _, w := range form {
		if w == "\x00" {
			out = append(out, path...)
			continue
		}
		out = append(out, w)
	}
	return out
}

// snapshotTree is every regular file under dir with its content, so a test can prove an
// invocation wrote, removed and changed nothing there.
func snapshotTree(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		if d.IsDir() {
			out[rel+"/"] = ""
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		out[rel] = string(b)
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", dir, err)
	}
	return out
}

// captured is what the handler of the command under test received.
type captured struct {
	ran      bool
	args     []string
	values   map[string]string
	changed  map[string]bool
	settings map[string]setting
}

// cliCapture runs argv through the real statusgen tree and parser with the handler of
// the command at path replaced by a recorder: the parse, the Go-flag normalisation and the
// settings resolution are production code; only the handler body is not run.
func cliCapture(t *testing.T, path, argv []string) (int, *captured, string) {
	t.Helper()
	c := &captured{}
	build := func(rio cliIO) *cobra.Command {
		root := newStatusgenRoot(rio)
		target := root
		if len(path) > 0 {
			found, _, err := root.Find(path)
			if err != nil || found == root {
				t.Fatalf("no command %q in the tree", strings.Join(path, " "))
			}
			target = found
		}
		target.RunE = func(cmd *cobra.Command, args []string) error {
			c.ran = true
			c.args = append([]string{}, args...)
			c.values = map[string]string{}
			c.changed = map[string]bool{}
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				c.values[f.Name] = f.Value.String()
				c.changed[f.Name] = f.Changed
			})
			c.settings = resolvedSettings(cmd)
			return nil
		}
		return root
	}
	var errb bytes.Buffer
	code := runCLI(build, argv, cliIO{Out: io.Discard, Err: &errb})
	return code, c, errb.String()
}

// envSpellings are the environment names a generic binding could read for a key: the
// bare upper-case name and a STATUSGEN_ prefix, each with '-' kept and with '_'.
func envSpellings(key string) []string {
	up := strings.ToUpper(key)
	under := strings.ReplaceAll(up, "-", "_")
	names := []string{up, under, "STATUSGEN_" + up, "STATUSGEN_" + under}
	if !EnvFoldsCase(runtime.GOOS) {
		// On a case-sensitive platform the lower-case spelling is a distinct name a
		// binding could also read.
		names = append(names, key, strings.ReplaceAll(key, "-", "_"))
	}
	return names
}

// writeConfigEverywhere writes a config file assigning value to every key under every
// name and place a config search could look: the working directory, $HOME and
// $HOME/.config/statusgen, as statusgen.{yaml,json,env} and config.yaml.
func writeConfigEverywhere(t *testing.T, cwd, home string, keys []string, value string) {
	t.Helper()
	var yml, env strings.Builder
	jsonParts := []string{}
	for _, k := range keys {
		yml.WriteString(k + ": \"" + value + "\"\n")
		env.WriteString(k + "=" + value + "\n")
		jsonParts = append(jsonParts, "\""+k+"\": \""+value+"\"")
	}
	js := "{" + strings.Join(jsonParts, ", ") + "}\n"
	for _, dir := range []string{cwd, home, filepath.Join(home, ".config", "statusgen")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		for name, content := range map[string]string{
			"statusgen.yaml": yml.String(), ".statusgen.yaml": yml.String(), "config.yaml": yml.String(),
			"statusgen.json": js, "statusgen.env": env.String(),
		} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}
}

// ---- TestCLIHelpOffline ------------------------------------------------------

// regression: #2111
//
// TestCLIHelpOffline runs every help and version form for the root and every command of
// the built binary, under an absent and each malformed roster, from a scratch tree. Each
// must exit 0 with its output on stdout, print nothing on stderr (the roster echo, the
// signal of a configuration read, is "assay-config:" on stderr) and leave the tree and
// $HOME exactly as they were.
func TestCLIHelpOffline(t *testing.T) {
	bin := buildStatusgenBinary(t)
	paths := commandPaths()
	if len(paths) < 20 {
		t.Fatalf("command tree has %d commands, want every statusgen verb (>= 20)", len(paths))
	}

	type form struct {
		argv    []string
		version bool
	}
	var forms []form
	seen := map[string]bool{}
	add := func(argv []string, version bool) {
		k := strings.Join(argv, "\x1f")
		if !seen[k] {
			seen[k] = true
			forms = append(forms, form{argv, version})
		}
	}
	for _, p := range paths {
		for _, f := range append(HelpForms("\x00"), GoFlagHelpForms("\x00")...) {
			add(expandSub(f, p), false)
		}
	}
	for _, f := range append(append(append([][]string{}, VersionForms...), GoFlagVersionForms...), []string{"version"}) {
		add(f, true)
	}

	for _, mc := range MalformedConfigs {
		home := t.TempDir()
		if mc.Content != "" {
			dir := filepath.Join(home, ".config", "assay")
			if err := os.MkdirAll(dir, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "roster.env"), []byte(mc.Content), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		cwd := t.TempDir()
		if err := os.MkdirAll(filepath.Join(cwd, "docs", "streams", "s"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cwd, "docs", "streams", "s", "README.md"), []byte("# s\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		tmp := t.TempDir()
		beforeHome, beforeCwd, beforeTmp := snapshotTree(t, home), snapshotTree(t, cwd), snapshotTree(t, tmp)

		for _, f := range forms {
			cmd := exec.Command(bin, f.argv...)
			cmd.Dir = cwd
			cmd.Env = []string{"HOME=" + home, "PATH=" + os.Getenv("PATH"), "TMPDIR=" + tmp, "KUBECONFIG=/dev/null"}
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			err := cmd.Run()
			label := mc.Name + ": statusgen " + strings.Join(f.argv, " ")
			if err != nil {
				t.Errorf("%s: %v (stderr %q)", label, err, stderr.String())
				continue
			}
			if strings.Contains(stderr.String(), "assay-config:") {
				t.Errorf("%s: read the roster before answering (stderr %q)", label, stderr.String())
			} else if stderr.Len() != 0 {
				t.Errorf("%s: wrote to stderr: %q", label, stderr.String())
			}
			switch {
			case f.version && stdout.String() != statusgenVersion+"\n":
				t.Errorf("%s: stdout %q, want the bare release tag %q", label, stdout.String(), statusgenVersion)
			case !f.version && !strings.Contains(stdout.String(), "Usage:"):
				t.Errorf("%s: stdout carries no generated help: %q", label, stdout.String())
			}
		}

		if after := snapshotTree(t, home); !reflect.DeepEqual(beforeHome, after) {
			t.Errorf("%s: help or version changed $HOME", mc.Name)
		}
		if after := snapshotTree(t, cwd); !reflect.DeepEqual(beforeCwd, after) {
			t.Errorf("%s: help or version changed the working tree", mc.Name)
		}
		if after := snapshotTree(t, tmp); !reflect.DeepEqual(beforeTmp, after) {
			t.Errorf("%s: help or version wrote to TMPDIR", mc.Name)
		}
	}
	t.Logf("%d help and version forms x %d roster states", len(forms), len(MalformedConfigs))
}

// ---- TestCLIConfigFlow -------------------------------------------------------

// regression: #2111
//
// TestCLIConfigFlow checks statusgen's declared precedence matrix (statusgen/README.md):
// every setting resolves from the flag, when it appears on the line (an explicit empty
// value included), else from its default — through a fresh command-local Viper instance.
// No environment variable and no config file is bound by the command-line layer, so the
// fixture matrix's env and config layers must never reach a handler. The few domain-owned
// environment reads keep their own combination rules, checked last.
func TestCLIConfigFlow(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("HOME", t.TempDir())

	// 1. The fixture matrix through the adapter, on a probe key declaring the fixture default.
	probeTree := func(c *captured) func(cliIO) *cobra.Command {
		return func(rio cliIO) *cobra.Command {
			root := newCommand("statusgen", "", "")
			probe := newCommand("probe", "", "")
			probe.Flags().String("key", PrecedenceDefault, "")
			probe.RunE = func(cmd *cobra.Command, _ []string) error {
				c.ran = true
				c.settings = resolvedSettings(cmd)
				c.values = map[string]string{"key": cmd.Flags().Lookup("key").Value.String()}
				return nil
			}
			root.AddCommand(probe)
			return root
		}
	}
	for _, row := range Precedence {
		cwd, home := t.TempDir(), t.TempDir()
		t.Chdir(cwd)
		t.Setenv("HOME", home)
		for _, name := range append(envSpellings("key"), "PROBE_KEY") {
			if row.Env != nil {
				t.Setenv(name, *row.Env)
			} else {
				os.Unsetenv(name)
			}
		}
		if row.Config != nil {
			writeConfigEverywhere(t, cwd, home, []string{"key"}, *row.Config)
		}
		argv := []string{"probe"}
		if row.Flag != nil {
			argv = append(argv, "--key="+*row.Flag)
		}
		want, wantSrc := PrecedenceDefault, "default"
		if row.Flag != nil {
			want, wantSrc = *row.Flag, "flag"
		}
		c := &captured{}
		if code := runCLI(probeTree(c), argv, cliIO{}); code != 0 || !c.ran {
			t.Errorf("%s: exit %d, handler ran %t", row.Name, code, c.ran)
			continue
		}
		got := c.settings["key"]
		if got.Value != want || got.Source != wantSrc || c.values["key"] != want {
			t.Errorf("%s: resolved %q from %q (handler read %q), want %q from %q — statusgen binds flag and default only (fixture's generic answer: %q from %q)",
				row.Name, got.Value, got.Source, c.values["key"], want, wantSrc, row.Want, row.WantSource)
		}
	}

	// 2. Every declared setting of every real command: with every env spelling and a
	// config file set for every key, an invocation without flags resolves every setting
	// to its default; an explicit empty string flag resolves to "" from the flag.
	for _, path := range append([][]string{nil}, commandPaths()...) {
		label := "statusgen " + strings.Join(path, " ")
		code, base, stderr := cliCapture(t, path, append([]string{}, path...))
		if !base.ran {
			// Only a command with subcommands and no handler of its own (outcomes) may
			// refuse a bare invocation; every handler-bearing command must reach it.
			if c, _, _ := newStatusgenRoot(cliIO{}).Find(path); c.RunE != nil {
				t.Errorf("%s: bare invocation exit %d, handler not reached (%s)", label, code, stderr)
			}
			continue
		}
		var keys []string
		for k := range base.settings {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for k, s := range base.settings {
			if s.Source != "default" {
				t.Errorf("%s: --%s resolved from %q with nothing on the line, want default", label, k, s.Source)
			}
		}
		cwd, home := t.TempDir(), t.TempDir()
		t.Chdir(cwd)
		t.Setenv("HOME", home)
		writeConfigEverywhere(t, cwd, home, keys, "cfg-x")
		for _, k := range keys {
			for _, name := range envSpellings(k) {
				t.Setenv(name, "env-x")
			}
		}
		code, got, stderr := cliCapture(t, path, append([]string{}, path...))
		if code != 0 || !got.ran {
			t.Errorf("%s: with env and config set for every key: exit %d, handler ran %t (%s)", label, code, got.ran, stderr)
		} else if !reflect.DeepEqual(base.settings, got.settings) || !reflect.DeepEqual(base.values, got.values) {
			for _, k := range keys {
				if !reflect.DeepEqual(base.settings[k], got.settings[k]) || base.values[k] != got.values[k] {
					t.Errorf("%s: --%s became %v (handler read %q) from an env or config source, want the default %v",
						label, k, got.settings[k], got.values[k], base.settings[k])
				}
			}
		}
		for _, k := range keys {
			for _, name := range envSpellings(k) {
				os.Unsetenv(name)
			}
		}

		// Explicit empty versus unset, on every plain string flag.
		for _, k := range keys {
			_, plain := reflectFlagValue(path, k)
			if !plain {
				continue
			}
			argv := append(append([]string{}, path...), "--"+k+"=")
			if base.values[k] == "" {
				// The default is already empty: prove a non-empty value is replaced by "".
				argv = append(append([]string{}, path...), "--"+k+"=x", "--"+k+"=")
			}
			code, got, stderr := cliCapture(t, path, argv)
			if code != 0 || !got.ran {
				t.Errorf("%s --%s=: exit %d, handler ran %t (%s)", label, k, code, got.ran, stderr)
				continue
			}
			if s := got.settings[k]; s.Value != "" || s.Source != "flag" || got.values[k] != "" {
				t.Errorf("%s --%s=: resolved %v (handler read %q), want \"\" from flag", label, k, s, got.values[k])
			}
		}
	}

	// 3. The domain-owned environment reads keep their own combination rules.
	for _, tc := range []struct {
		flag bool
		env  *string
		want bool
	}{
		{false, nil, false}, {true, nil, true}, {false, Opt("1"), true}, {false, Opt("true"), false}, {true, Opt("0"), true},
	} {
		if tc.env != nil {
			t.Setenv(siblingMergeOptInEnv, *tc.env)
		} else {
			os.Unsetenv(siblingMergeOptInEnv)
		}
		argv := []string{}
		if tc.flag {
			argv = []string{"--sibling-merge"}
		}
		if _, c, _ := cliCapture(t, nil, argv); !c.ran || siblingMergeEnabled() != tc.want {
			t.Errorf("sibling-merge flag=%t env=%v: enabled %t, want %t (flag OR env==1)", tc.flag, tc.env, siblingMergeEnabled(), tc.want)
		}
	}
	siblingMergeFlagValue = false
	os.Unsetenv(siblingMergeOptInEnv)
	for _, tc := range []struct {
		flag bool
		env  *string
		want bool
	}{
		{false, Opt("1"), false}, {true, nil, false}, {true, Opt("1"), true}, {true, Opt(""), false},
	} {
		if tc.env != nil {
			t.Setenv(telemetryEnvVar, *tc.env)
		} else {
			os.Unsetenv(telemetryEnvVar)
		}
		if got := telemetryArmed(tc.flag); got != tc.want {
			t.Errorf("telemetry flag=%t env=%v: armed %t, want %t (flag AND env==1)", tc.flag, tc.env, got, tc.want)
		}
	}
	os.Unsetenv(telemetryEnvVar)
	t.Setenv(siblingRootsEnv, "o/r=/env-r,o/s=/env-s")
	_, c, _ := cliCapture(t, nil, []string{"--sibling-root", "o/r=/flag-r"})
	var specs []string
	if c.ran && c.values["sibling-root"] != "" {
		specs = strings.Split(c.values["sibling-root"], ", ")
	}
	if got, want := effectiveSiblingRootOverrides(specs), map[string]string{"o/r": "/flag-r", "o/s": "/env-s"}; !reflect.DeepEqual(got, want) {
		t.Errorf("sibling-root over DESK_ROOTS: %v, want %v (both read, flag wins per repo)", got, want)
	}
}

// reflectFlagValue reports whether the flag k of the command at path is a plain pflag
// string (no custom Set that appends or validates).
func reflectFlagValue(path []string, k string) (*pflag.Flag, bool) {
	root := newStatusgenRoot(cliIO{})
	cmd := root
	if len(path) > 0 {
		cmd, _, _ = root.Find(path)
	}
	f := cmd.Flags().Lookup(k)
	if f == nil {
		return nil, false
	}
	return f, reflect.TypeOf(f.Value).String() == "*pflag.stringValue"
}

// ---- TestCLILegacyForms ------------------------------------------------------

// regression: #2111
//
// TestCLILegacyForms replays the invocation forms statusgen's callers use: every option
// spelling of every string flag of every command with every path fixture, the Go-flag
// single-dash spellings, repeated flags, "--" termination, the root's stop at its first
// positional (which --export-evidence's -o/-generated scan depends on), and the refusals:
// an unknown flag, a bad value and an unknown subcommand exit 2 before any handler runs.
func TestCLILegacyForms(t *testing.T) {
	// A refusal that regressed would fall through to the default regenerate, a write:
	// keep any such write inside a scratch directory.
	t.Chdir(t.TempDir())
	// 1. Every path value, every option form, every accepting string flag, every command.
	cases := 0
	for _, path := range append([][]string{nil}, commandPaths()...) {
		root := newStatusgenRoot(cliIO{})
		cmd := root
		if len(path) > 0 {
			cmd, _, _ = root.Find(path)
		}
		if cmd.RunE == nil {
			continue
		}
		var names []string
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Value.Type() == "string" {
				names = append(names, f.Name)
			}
		})
		for _, name := range names {
			for _, pv := range PathValues {
				// A flag whose own Set validates its value's shape (name=glob, a brief
				// dependency) refuses some fixtures by design; those are domain refusals.
				// A flag whose String is a summary rather than the value (shardcheck's
				// --shard prints shard names) is checked by its own tests.
				f, _ := reflectFlagValue(path, name)
				if f.Value.Set(pv.Value) != nil || f.Value.String() != pv.Value {
					continue
				}
				for _, of := range OptionForms {
					argv := append(append([]string{}, path...), of.Args(name, pv.Value)...)
					label := "statusgen " + strings.Join(argv, " ")
					code, got, stderr := cliCapture(t, path, argv)
					cases++
					if code != 0 || !got.ran {
						t.Errorf("%s (%s/%s): exit %d, handler ran %t (%s)", label, of.Name, pv.Name, code, got.ran, stderr)
						continue
					}
					if got.values[name] != pv.Value || !got.changed[name] {
						t.Errorf("%s (%s/%s): handler read %q, want %q byte-identical", label, of.Name, pv.Name, got.values[name], pv.Value)
					}
					if s := got.settings[name]; s.Value != pv.Value || s.Source != "flag" {
						t.Errorf("%s (%s/%s): resolved %v, want %q from flag", label, of.Name, pv.Name, s, pv.Value)
					}
					if len(got.args) != 0 {
						t.Errorf("%s (%s/%s): value leaked into positionals %q", label, of.Name, pv.Name, got.args)
					}
				}
			}
		}
	}
	if cases < 1000 {
		t.Errorf("only %d path/form cases ran; the string flags of the tree were not enumerated", cases)
	}
	t.Logf("%d path-value x option-form cases reached their handler", cases)

	// 2. The forms the consumers spell, through the real tree.
	type want struct {
		values map[string]string
		args   []string
	}
	for _, tc := range []struct {
		name string
		path []string
		argv []string
		want want
	}{
		{"go-flag dora", nil, []string{"-root", "r", "-dora", "-since", "2026-01-01"},
			want{map[string]string{"root": "r", "dora": "true", "since": "2026-01-01"}, nil}},
		{"go-flag trend", nil, []string{"-trend"}, want{map[string]string{"trend": "true"}, nil}},
		{"go-flag bool explicit", nil, []string{"-lint=false", "-json=true"}, want{map[string]string{"lint": "false", "json": "true"}, nil}},
		{"long-equals root", nil, []string{"--root=r", "--lint"}, want{map[string]string{"root": "r", "lint": "true"}, nil}},
		{"long-space root", nil, []string{"--root", "r", "--consumers", "--base", "origin/main"},
			want{map[string]string{"root": "r", "consumers": "true", "base": "origin/main"}, nil}},
		{"repeated list flag", nil, []string{"--root", "a", "--root", "b", "-root", "c"}, want{map[string]string{"root": "a, b, c"}, nil}},
		{"repeated scalar keeps last", nil, []string{"--diff-base", "a", "-diff-base", "b"}, want{map[string]string{"diff-base": "b"}, nil}},
		{"dash-leading separate value", nil, []string{"--diff-base", "-x"}, want{map[string]string{"diff-base": "-x"}, nil}},
		{"export-evidence positional scan", nil,
			[]string{"--root", "r", "--export-evidence", "2026-01-01", "2026-02-01", "-o", "out", "-generated", "2026-02-01T00:00:00Z"},
			want{map[string]string{"export-evidence": "true", "o": "", "generated": ""},
				[]string{"2026-01-01", "2026-02-01", "-o", "out", "-generated", "2026-02-01T00:00:00Z"}}},
		{"export-audit-pack go-flag -o", nil,
			[]string{"--export-audit-pack", "-release", "v1", "-o", "pack.tgz", "-generated", "2026-02-01T00:00:00Z"},
			want{map[string]string{"export-audit-pack": "true", "release": "v1", "o": "pack.tgz", "generated": "2026-02-01T00:00:00Z"}, nil}},
		{"double-dash terminates the root", nil, []string{"--root", "r", "--", "--lint"},
			want{map[string]string{"root": "r", "lint": "false"}, []string{"--lint"}}},
		{"verb go-flag", []string{"mergecheck"}, []string{"mergecheck", "-root", "r"}, want{map[string]string{"root": "r"}, nil}},
		{"verb flag after positional", []string{"migrate"}, []string{"migrate", "target", "--root", "r"},
			want{map[string]string{"root": "r"}, []string{"target"}}},
		{"verb double-dash", []string{"migrate"}, []string{"migrate", "--", "-dash-target"},
			want{map[string]string{}, []string{"-dash-target"}}},
		{"nested verb go-flag", []string{"outcomes", "split"}, []string{"outcomes", "split", "-root", "r"},
			want{map[string]string{"root": "r"}, nil}},
	} {
		code, got, stderr := cliCapture(t, tc.path, tc.argv)
		if code != 0 || !got.ran {
			t.Errorf("%s: statusgen %s: exit %d, handler ran %t (%s)", tc.name, strings.Join(tc.argv, " "), code, got.ran, stderr)
			continue
		}
		for k, v := range tc.want.values {
			if got.values[k] != v {
				t.Errorf("%s: statusgen %s: --%s = %q, want %q", tc.name, strings.Join(tc.argv, " "), k, got.values[k], v)
			}
		}
		if !reflect.DeepEqual(nonNil(got.args), nonNil(tc.want.args)) {
			t.Errorf("%s: statusgen %s: positionals %q, want %q", tc.name, strings.Join(tc.argv, " "), got.args, tc.want.args)
		}
	}

	// 3. Refusals: exit 2 with one error line and a --help pointer, before any handler.
	for _, tc := range []struct {
		name, msg string
		path      []string
		argv      []string
	}{
		{"unknown root flag", "unknown flag: --bogus", nil, []string{"--bogus"}},
		{"unknown go-flag", "unknown shorthand flag", nil, []string{"-bogus"}},
		{"unknown verb flag", "unknown flag: --bogus", []string{"mergecheck"}, []string{"mergecheck", "--bogus"}},
		{"bad int value", "invalid argument", nil, []string{"--dora-limit", "many"}},
		{"bad bool value", "invalid argument", nil, []string{"--lint=maybe"}},
		{"missing value", "needs an argument", nil, []string{"--diff-base"}},
		{"version takes no args", "unknown command", []string{"version"}, []string{"version", "x"}},
		{"completion request", "unknown command", nil, []string{"__complete", "statusgen", ""}},
	} {
		code, got, stderr := cliCapture(t, tc.path, tc.argv)
		if code != wantUsageExit || got.ran {
			t.Errorf("%s: statusgen %s: exit %d, handler ran %t; want exit %d before any handler", tc.name, strings.Join(tc.argv, " "), code, got.ran, wantUsageExit)
		}
		if !strings.Contains(stderr, tc.msg) || !strings.Contains(stderr, "--help' for usage.") {
			t.Errorf("%s: statusgen %s: stderr %q, want %q and a --help pointer", tc.name, strings.Join(tc.argv, " "), stderr, tc.msg)
		}
	}

	// An unknown subcommand keeps its legacy diagnosis (#1075): the real root handler
	// refuses it before the roster echo or any write.
	var errb, outb bytes.Buffer
	if code := runCLI(newStatusgenRoot, []string{"mergcheck"}, cliIO{Out: &outb, Err: &errb}); code != wantUsageExit {
		t.Errorf("statusgen mergcheck: exit %d, want %d", code, wantUsageExit)
	}
	if !strings.HasPrefix(errb.String(), "statusgen: unknown subcommand \"mergcheck\"\n") || strings.Contains(errb.String(), "assay-config:") {
		t.Errorf("statusgen mergcheck: stderr %q, want the unknown-subcommand diagnosis and no roster read", errb.String())
	}
	errb.Reset()
	if code := runCLI(newStatusgenRoot, []string{"--version", "--lint"}, cliIO{Out: &outb, Err: &errb}); code != wantUsageExit ||
		!strings.Contains(errb.String(), "--version takes no other flag or argument") {
		t.Errorf("statusgen --version --lint: exit %d stderr %q, want exit %d and the sole-argument refusal", code, errb.String(), wantUsageExit)
	}
}
