package deskkit

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// The help-path echo tests (medici-finance/assay#3754). A help screen is not a run: it makes no
// trust decision, so the effective-config echo must not print a resolved roster VALUE on it. The
// roster here is synthetic and every value is distinctive, so a single substring hit is a leak.

// helpRosterValues are the distinctive synthetic values the help path must never print.
var helpRosterValues = []string{"qzhelpbless", "9101", "qzhelptrust", "9102", "qz-help-bot", "9103", "qzorg/qzrepo", "qzknobvalue"}

// helpRoster is a CONFIGURED roster (so the echo carries its values) plus one co-tenant key
// outside the ASSAY_ namespace, which the echo reports as unrecognised.
func helpRoster() map[string]string {
	return map[string]string{
		EnvBlessLogin:      "qzhelpbless:9101",
		EnvTrustedLogins:   "qzhelpbless:9101,qzhelptrust:9102",
		EnvTrustedBotSlugs: "reviewer=qz-help-bot:9103",
		EnvAllowedRepos:    "qzorg/qzrepo:ci:private",
		"QZ_HELP_KNOB":     "qzknobvalue",
	}
}

// withArgs points os.Args at a tool invocation for one test.
func withArgs(t *testing.T, args ...string) {
	t.Helper()
	prev := os.Args
	os.Args = append([]string{"desktool"}, args...)
	t.Cleanup(func() { os.Args = prev })
}

// helpShapes are the invocations HelpInvocation recognises.
var helpShapes = [][]string{
	{"--help"}, {"-h"}, {"-help"}, {"help"}, {"--help", "finish"},
	{"finish", "--help"}, {"create", "-h"}, {"review", "-help"},
}

func echoBoth(t *testing.T) string {
	t.Helper()
	var b bytes.Buffer
	EchoEffectiveConfig(&b)
	EchoEffectiveConfigUnsurfaced(&b)
	return b.String()
}

func assertNoHelpValues(t *testing.T, args []string, out string) {
	t.Helper()
	for _, v := range helpRosterValues {
		if strings.Contains(out, v) {
			t.Errorf("help request %q: the echo printed the roster value %q:\n%s", args, v, out)
		}
	}
}

// regression: #3754
func TestEchoHelpPrintsNoValues(t *testing.T) {
	t.Setenv(EnvConfigEcho, "full") // the loudest form: a help request must still print no value
	withRoster(t, helpRoster())
	if !EffectiveConfig().Configured() {
		t.Fatalf("fixture roster must load configured, or the echo carries no value to leak: %v",
			EffectiveConfig().Problems)
	}
	for _, args := range helpShapes {
		withArgs(t, args...)
		out := echoBoth(t)
		assertNoHelpValues(t, args, out)
		// The diagnostic that remains names the KEY of the unrecognised entry, never its value.
		if !strings.Contains(out, "QZ_HELP_KNOB") {
			t.Errorf("help request %q: the keys-only diagnostic must still name the unrecognised key:\n%s", args, out)
		}
	}
}

// regression: #3754
func TestEchoHelpRefusedKeysOnly(t *testing.T) {
	// A REFUSED roster is never quieter than it was: the help path still says so, by count.
	t.Setenv(EnvConfigEcho, "full")
	vals := helpRoster()
	vals["ASSAY_QZ_HELP_KNOB"] = "qzknobvalue"
	withRoster(t, vals)
	if len(EffectiveConfig().Problems) == 0 {
		t.Fatal("fixture must carry a refused key")
	}
	for _, args := range helpShapes {
		withArgs(t, args...)
		out := echoBoth(t)
		assertNoHelpValues(t, args, out)
		if !strings.Contains(out, "REFUSED") {
			t.Errorf("help request %q: a refused configuration must still be reported:\n%s", args, out)
		}
	}
}

// regression: #3754
func TestEchoNonHelpStillFull(t *testing.T) {
	// The negative control: the help check must not swallow a real run's echo. Each of these
	// is an invocation that ACTS (or may), including the shapes where `--help` is another
	// flag's value or a positional after the terminator.
	t.Setenv(EnvConfigEcho, "full")
	withRoster(t, helpRoster())
	for _, args := range [][]string{
		nil, {"finish"}, {"create", "--title", "--help"}, {"create", "--", "--help"},
		{"create", "--help", "--json"}, {"help", "create"},
	} {
		withArgs(t, args...)
		var b bytes.Buffer
		EchoEffectiveConfig(&b)
		if !isFullEcho(b.String()) {
			t.Errorf("invocation %q is not a help request, so the full block is due:\n%s", args, b.String())
		}
	}
}

// regression: #3754
func TestHelpInvocationShapes(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want bool
	}{
		{nil, false},
		{[]string{"--help"}, true},
		{[]string{"-h", "x"}, true},
		{[]string{"help"}, true},
		{[]string{"help", "create"}, false},
		{[]string{"finish", "--help"}, true},
		{[]string{"finish", "review", "--help"}, false},
		{[]string{"create", "--title", "--help"}, false},
	} {
		if got := HelpInvocation(tc.args); got != tc.want {
			t.Errorf("HelpInvocation(%q) = %t, want %t", tc.args, got, tc.want)
		}
	}
}
