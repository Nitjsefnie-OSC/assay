package main

// Condition reviewer-approved on the Read + Act install tier (#3062).
//
// `deskapps init --tier team` binds EVERY desk role to one `<prefix>-act` App, so the reviewer
// App is the worker App and cannot APPROVE the PRs it authored. These cases pin both sides:
// on that shape a TRUSTED, ACCOUNTABLE human's APPROVED at the current head satisfies the gate
// (and nothing weaker does); on the full-suite shape nothing changed — a human APPROVED is
// still never a substitute for the reviewer App's.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/medici-finance/assay/tools/desk/internal/deskkit"
)

// readActSlug is the one App the Read + Act tier binds every role to.
const readActSlug = "demo-act"

// readActRoster is fixtureRoster with the bot set replaced by the Read + Act tier's shape:
// all six desk roles bound to ONE App, exactly what `deskapps init --tier team` writes. Every
// other line (humans, bless authority, repos, forges) is the fixture's own, so a verdict here
// differs from the full-suite one ONLY by the shape under test.
func readActRoster(t *testing.T) string {
	t.Helper()
	var out []string
	replaced := false
	for _, line := range strings.Split(fixtureRoster, "\n") {
		if strings.HasPrefix(line, deskkit.EnvTrustedBotSlugs+"=") {
			var roles []string
			for _, r := range []string{"desk", "intake-loop", "issue-loop", "reviewer", "verifier", "worker"} {
				roles = append(roles, r+"="+readActSlug+":300000010")
			}
			line = deskkit.EnvTrustedBotSlugs + "=" + strings.Join(roles, ",")
			replaced = true
		}
		out = append(out, line)
	}
	if !replaced {
		t.Fatal("fixtureRoster carries no bot-slug line to replace")
	}
	return strings.Join(out, "\n")
}

// rosterPath is the roster file the stub's install planted under the test's private HOME.
func rosterPath() string {
	return filepath.Join(os.Getenv("HOME"), ".config", "assay", "roster.env")
}

// installReadAct overwrites the roster the stub planted with the Read + Act one.
func installReadAct(t *testing.T) {
	t.Helper()
	path := rosterPath()
	if err := os.WriteFile(path, []byte(readActRoster(t)), 0o600); err != nil {
		t.Fatalf("planting the Read + Act roster: %v", err)
	}
	deskkit.ReloadConfig()
}

func review(login, state, commit string) reviewInfo {
	r := reviewInfo{State: state, CommitID: commit, Body: "verdict", SubmittedAt: "2026-01-01T00:00:00Z"}
	r.User.Login = login
	return r
}

const olderHead = "0000000011111111222222223333333344444444"

func TestReviewerApprovedByTier(t *testing.T) {
	// actBot is the shared App's review-author rendering, read from the roster at run time so
	// no case carries its own copy of the binding.
	const actBot = "$act"
	const fullBot = "$reviewer"

	cases := []struct {
		name      string
		readAct   bool
		author    string // "" = the stub's default worker App; actBot = the shared App
		reviews   []reviewInfo
		wantRC    int
		wantInMsg []string
	}{
		{
			name:    "full suite: reviewer App APPROVED at head passes",
			reviews: []reviewInfo{review(fullBot, "APPROVED", headSHA)},
			wantRC:  deskkit.ExitOK,
		},
		{
			name:      "full suite: a trusted human's APPROVED alone still refuses",
			reviews:   []reviewInfo{review("ada", "APPROVED", headSHA)},
			wantRC:    deskkit.ExitRefused,
			wantInMsg: []string{condReviewerApproved},
		},
		{
			name:      "full suite: a comment-only reviewer verdict refuses without naming a tier",
			reviews:   []reviewInfo{review(fullBot, "COMMENTED", headSHA)},
			wantRC:    deskkit.ExitRefused,
			wantInMsg: []string{condReviewerApproved},
		},
		{
			name:      "read+act: trusted human APPROVED at head passes",
			readAct:   true,
			author:    actBot,
			reviews:   []reviewInfo{review(actBot, "COMMENTED", headSHA), review("ada", "APPROVED", headSHA)},
			wantRC:    deskkit.ExitOK,
			wantInMsg: []string{"reviewer-approved OK: ada APPROVED"},
		},
		{
			name:      "read+act: trusted human APPROVED at an older head is stale",
			readAct:   true,
			author:    actBot,
			reviews:   []reviewInfo{review("ada", "APPROVED", olderHead)},
			wantRC:    deskkit.ExitRefused,
			wantInMsg: []string{readActTierName, "stale", short(olderHead), "awaiting a human approval at head"},
		},
		{
			name:      "read+act: the App's comment alone refuses, naming the tier",
			readAct:   true,
			author:    actBot,
			reviews:   []reviewInfo{review(actBot, "COMMENTED", headSHA)},
			wantRC:    deskkit.ExitRefused,
			wantInMsg: []string{readActTierName, "awaiting a human approval at head"},
		},
		{
			name:      "read+act: an untrusted human's APPROVED refuses",
			readAct:   true,
			author:    actBot,
			reviews:   []reviewInfo{review("mallory", "APPROVED", headSHA)},
			wantRC:    deskkit.ExitRefused,
			wantInMsg: []string{readActTierName, "mallory", "not counted"},
		},
		{
			name:      "read+act: a trusted but non-accountable (shared) login's APPROVED refuses",
			readAct:   true,
			author:    actBot,
			reviews:   []reviewInfo{review("shared-agent", "APPROVED", headSHA)},
			wantRC:    deskkit.ExitRefused,
			wantInMsg: []string{readActTierName, "shared-agent"},
		},
		{
			name:      "read+act: another App's APPROVED refuses",
			readAct:   true,
			author:    actBot,
			reviews:   []reviewInfo{review("some-other-app[bot]", "APPROVED", headSHA)},
			wantRC:    deskkit.ExitRefused,
			wantInMsg: []string{readActTierName},
		},
		{
			name:      "read+act: the human's own later CHANGES_REQUESTED at head blocks",
			readAct:   true,
			author:    actBot,
			reviews:   []reviewInfo{review("ada", "APPROVED", olderHead), review("ada", "CHANGES_REQUESTED", headSHA)},
			wantRC:    deskkit.ExitRefused,
			wantInMsg: []string{"CHANGES_REQUESTED"},
		},
		{
			name:      "read+act: the App approving its own PR (forge did not block it) refuses",
			readAct:   true,
			author:    actBot,
			reviews:   []reviewInfo{review(actBot, "APPROVED", headSHA)},
			wantRC:    deskkit.ExitRefused,
			wantInMsg: []string{readActTierName, "which it authored"},
		},
		{
			name:      "read+act: the App's standing CHANGES_REQUESTED blocks a human APPROVED",
			readAct:   true,
			author:    "bob",
			reviews:   []reviewInfo{review(actBot, "CHANGES_REQUESTED", headSHA), review("ada", "APPROVED", headSHA)},
			wantRC:    deskkit.ExitRefused,
			wantInMsg: []string{"CHANGES_REQUESTED"},
		},
		{
			name:    "read+act: the App APPROVED on a PR it did not author still passes",
			readAct: true,
			author:  "bob",
			reviews: []reviewInfo{review(actBot, "APPROVED", headSHA)},
			wantRC:  deskkit.ExitOK,
		},
		{
			name:      "read+act: a human approving their own PR refuses",
			readAct:   true,
			author:    "ada",
			reviews:   []reviewInfo{review("ada", "APPROVED", headSHA)},
			wantRC:    deskkit.ExitRefused,
			wantInMsg: []string{readActTierName, "not counted"},
		},
		{
			name:      "read+act: a security-lane pass from the human is not a correctness approval",
			readAct:   true,
			author:    actBot,
			reviews:   []reviewInfo{securityPass("ada", "APPROVED", headSHA)},
			wantRC:    deskkit.ExitRefused,
			wantInMsg: []string{readActTierName, "awaiting a human approval at head"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newStub()
			// A linked change, so an App-authored PR is not risk-classed by the trailer-absent
			// term (#587) — that condition is not under test here.
			s.pr.Body = "## Context\nA change.\n\nIssue: #1\n"
			s.install(t)
			if tc.readAct {
				installReadAct(t)
			}
			reviewer := reviewerBot(t)
			resolve := func(login string) string {
				switch login {
				case actBot, fullBot:
					return reviewer
				}
				return login
			}
			s.prAuthor = resolve(tc.author)
			for _, r := range tc.reviews {
				r.User.Login = resolve(r.User.Login)
				s.reviews = append(s.reviews, r)
			}

			var rc int
			out := captureStderr(t, func() { rc = run([]string{"7", "--repo", privateCIRepo}) })
			if rc != tc.wantRC {
				t.Fatalf("rc = %d, want %d\n%s", rc, tc.wantRC, out)
			}
			for _, want := range tc.wantInMsg {
				if !strings.Contains(out, want) {
					t.Errorf("output does not name %q:\n%s", want, out)
				}
			}
			if !tc.readAct && strings.Contains(out, readActTierName) {
				t.Errorf("a full-suite run named the %s tier:\n%s", readActTierName, out)
			}
			if tc.wantRC == deskkit.ExitOK {
				if !s.flipped() {
					t.Errorf("a passing gate did not flip:\n%s", out)
				}
			} else if m := s.mutated(); len(m) != 0 {
				t.Errorf("a refusal mutated: %v", m)
			}
		})
	}
}

// securityPass is a review carrying the security-lane marker, which rule 1 excludes from the
// correctness lane on either tier.
func securityPass(login, state, commit string) reviewInfo {
	r := review(login, state, commit)
	r.Body = "Security-Review: pass"
	return r
}

// The shape detector itself: the full-suite fixture is NOT the shared shape; the Read + Act
// roster is; and a roster with no worker binding is not either (it falls to the unchanged
// full-suite gate, never to the human lane).
func TestSharedReviewerAppShape(t *testing.T) {
	s := newStub()
	s.install(t)
	if sharedReviewerApp(reviewerBot(t)) {
		t.Fatal("the full-suite fixture roster was read as the Read + Act shape")
	}

	installReadAct(t)
	if !sharedReviewerApp(reviewerBot(t)) {
		t.Fatal("the Read + Act roster was not read as the shared-App shape")
	}

	noWorker := strings.Replace(readActRoster(t), ",worker="+readActSlug+":300000010", "", 1)
	if noWorker == readActRoster(t) {
		t.Fatal("could not drop the worker binding from the Read + Act roster")
	}
	if err := os.WriteFile(rosterPath(), []byte(noWorker), 0o600); err != nil {
		t.Fatal(err)
	}
	deskkit.ReloadConfig()
	if sharedReviewerApp(reviewerBot(t)) {
		t.Fatal("a roster with no worker binding was read as the shared-App shape")
	}
}
