package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Tests for `--corroborate --ruling-record` (rulingrecord.go, #2079): one decision
// record corroborated outside any PR, through the same faked forge the ruling-lane
// tests use (rlForge). Fixture identities are example-* only.

// rrSetup writes the canonical record under a temp root, pins the repository and
// the forge client, and installs the example roster.
func rrSetup(t *testing.T, f *rlForge) string {
	t.Helper()
	rlWithRoster(t)
	root := t.TempDir()
	rlWriteRecord(t, root, rlFile, rlRecordID)
	prevRepo, prevClient := repoFromOriginAtFn, rulingRecordClientFn
	t.Cleanup(func() { repoFromOriginAtFn, rulingRecordClientFn = prevRepo, prevClient })
	repoFromOriginAtFn = func(string) string { return rlRepo }
	c := f.client(t)
	rulingRecordClientFn = func() *ghClient { return c }
	return root
}

// regression: #2079
func TestRulingRecord_Corroborates(t *testing.T) {
	root := rrSetup(t, defaultRLForge())
	if got := runCorroborateRecord(root, rlFile); got != 0 {
		t.Fatalf("a record whose ruling link resolves to a mapped human's unedited comment must corroborate (exit 0), got %d", got)
	}
}

// regression: #2079
func TestRulingRecord_RefusesWrongAuthor(t *testing.T) {
	f := defaultRLForge()
	f.commentBody = rlCommentJSON("example-stranger", "User", rlRepo, rlIssue)
	root := rrSetup(t, f)
	if got := runCorroborateRecord(root, rlFile); got != 1 {
		t.Fatalf("a ruling comment by an unmapped login must refuse (exit 1), got %d", got)
	}
}

// regression: #2079
func TestRulingRecord_RefusesBotAuthor(t *testing.T) {
	f := defaultRLForge()
	f.commentBody = rlCommentJSON("example-app[bot]", "Bot", rlRepo, rlIssue)
	root := rrSetup(t, f)
	if got := runCorroborateRecord(root, rlFile); got != 1 {
		t.Fatalf("a bot-authored ruling comment must refuse (exit 1), got %d", got)
	}
}

// regression: #2079
func TestRulingRecord_RefusesNoRulingLink(t *testing.T) {
	root := rrSetup(t, defaultRLForge())
	p := filepath.Join(root, filepath.FromSlash(rlFile))
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var kept []string
	for _, l := range strings.Split(string(raw), "\n") {
		if !strings.HasPrefix(l, "ruling:") {
			kept = append(kept, l)
		}
	}
	if err := os.WriteFile(p, []byte(strings.Join(kept, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := runCorroborateRecord(root, rlFile); got != 1 {
		t.Fatalf("outside a PR a record with no ruling: link has no anchor and must refuse (exit 1), got %d", got)
	}
}

// regression: #2079
func TestRulingRecord_RefusesMissingRecord(t *testing.T) {
	root := rrSetup(t, defaultRLForge())
	if got := runCorroborateRecord(root, "docs/streams/decisions/DR-example-absent.md"); got != 1 {
		t.Fatalf("an absent record is unreadable and must refuse (exit 1), got %d", got)
	}
}

func TestRulingRecord_ForgeErrorCouldNotCheck(t *testing.T) {
	f := defaultRLForge()
	f.commentStatus = 500
	f.commentBody = `{"message":"boom"}`
	root := rrSetup(t, f)
	if got := runCorroborateRecord(root, rlFile); got != 2 {
		t.Fatalf("a forge read that failed is could-not-check (exit 2), got %d", got)
	}
}

func TestRulingRecord_UsageAndRepo(t *testing.T) {
	root := rrSetup(t, defaultRLForge())
	for _, rel := range []string{
		"docs/streams/decisions/README.md",
		"docs/streams/decisions/../x.md",
		"docs/streams/other/DR-x.md",
		"",
	} {
		if got := runCorroborateRecord(root, rel); got != 2 {
			t.Errorf("--ruling-record %q is not a decision record path: want exit 2, got %d", rel, got)
		}
	}
	repoFromOriginAtFn = func(string) string { return "" }
	if got := runCorroborateRecord(root, rlFile); got != 2 {
		t.Fatalf("an undeterminable repository is could-not-check (exit 2), got %d", got)
	}
}
