package gitcore

import (
	"github.com/medici-finance/assay/tools/desk/internal/gittest"
	"strings"
	"testing"
)

// Identity must not silently become a transport rewrite or push destination.
func TestConfiguredRemoteURLMatchesGit(t *testing.T) {
	f := gittest.NewFixture(t)
	for _, args := range [][]string{
		{"config", "--add", "remote.origin.url", "https://github.com/example-org/first.git"},
		{"config", "--add", "remote.origin.url", "https://github.com/example-org/last.git"},
		{"config", "remote.origin.pushurl", "https://github.com/example-org/push.git"},
		{"config", "url.https://example.invalid/.insteadOf", "https://github.com/"},
	} {
		if _, err := f.Git(args...); err != nil {
			t.Fatal(err)
		}
	}
	repo, err := Open(f.Dir)
	if err != nil {
		t.Fatal(err)
	}
	want, err := f.Git("config", "--get", "remote.origin.url")
	if err != nil {
		t.Fatal(err)
	}
	got, err := repo.ConfiguredRemoteURL("origin")
	if err != nil || got != strings.TrimSpace(want) {
		t.Fatalf("got %q %v want %q", got, err, want)
	}
	if _, err := repo.ConfiguredRemoteURL("absent"); err == nil {
		t.Fatal("absent remote accepted")
	}
	if _, err := f.Git("config", "--unset-all", "remote.origin.url"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ConfiguredRemoteURL("origin"); err == nil {
		t.Fatal("push-only remote supplied identity")
	}
}
