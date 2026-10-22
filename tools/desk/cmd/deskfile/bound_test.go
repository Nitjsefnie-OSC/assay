package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/medici-finance/assay/tools/desk/internal/deskkit"
)

// bound_test.go — workflow-bound filings (`deskfile new --for-pr ... --brief ...`).

// prState is the change-request side of the fake forge: the PRs, their file lists and the
// repository files the binding check reads. installBoundFake resets it per test.
type prState struct {
	prs     map[int]*deskkit.PullRequest
	files   map[int][]deskkit.ChangedFile
	readErr map[int]error
	content map[string]string // path -> file content (any ref)
	// commentErr fails the back-link comment only (the issue create still succeeds).
	commentErr error
}

var prFake *prState

func (f *dfForge) GetPullRequest(repo deskkit.ForgeRepo, n int) (*deskkit.PullRequest, error) {
	if err := prFake.readErr[n]; err != nil {
		return nil, err
	}
	pr, ok := prFake.prs[n]
	if !ok {
		return nil, fmt.Errorf("HTTP 404: %s#%d not found", repo.Slug(), n)
	}
	c := *pr
	return &c, nil
}

func (f *dfForge) ListChangedFiles(repo deskkit.ForgeRepo, n int) ([]deskkit.ChangedFile, error) {
	return append([]deskkit.ChangedFile(nil), prFake.files[n]...), nil
}

func (f *dfForge) ReadFile(repo deskkit.ForgeRepo, in deskkit.ReadFileInput) (*deskkit.FileContent, error) {
	c, ok := prFake.content[in.File]
	if !ok {
		return &deskkit.FileContent{Exists: false}, nil
	}
	return &deskkit.FileContent{Content: []byte(c), Exists: true}, nil
}

func (f *dfForge) RepoDefaultBranch(repo deskkit.ForgeRepo) (string, error) { return "main", nil }

const (
	bfBrief07 = "docs/streams/demo/brief-07-thing.md"
	bfBrief08 = "docs/streams/demo/brief-08-other.md"
)

// installBoundFake plants: #42 open carrying `Brief: demo/07`; #46 the same; #43 open with no
// brief; #44 closed; #45 unreadable; #47 open, no trailer, its diff touching brief 07.
func installBoundFake(t *testing.T) {
	t.Helper()
	open := func(n int, body string, files ...string) {
		cf := make([]deskkit.ChangedFile, 0, len(files))
		for _, p := range files {
			cf = append(cf, deskkit.ChangedFile{Filename: p, Status: "modified"})
		}
		prFake.prs[n] = &deskkit.PullRequest{Number: n, State: "open", Body: body, HeadSHA: "abc123",
			BaseRef: "main", ChangedFiles: len(cf)}
		prFake.files[n] = cf
	}
	prFake = &prState{prs: map[int]*deskkit.PullRequest{}, files: map[int][]deskkit.ChangedFile{},
		readErr: map[int]error{}, content: map[string]string{
			"docs/streams/demo/README.md": "| [07](brief-07-thing.md) |\n| [08](brief-08-other.md) |\n",
			bfBrief07:                     "---\nbrief: example:demo:07\n---\n# 07\n",
			bfBrief08:                     "---\nbrief: example:demo:08\n---\n# 08\n",
		}}
	open(42, "Change.\n\nBrief: demo/07\n", "tools/a.go")
	open(46, "Change.\n\nBrief: demo/07\n", "tools/b.go")
	open(43, "No brief.\n", "tools/c.go")
	open(47, "Authoring.\n", bfBrief07)
	open(44, "Brief: demo/07\n", "tools/d.go")
	prFake.prs[44].State = "closed"
	prFake.readErr[45] = deskkit.Unverifiable("HTTP 502", nil)
	t.Cleanup(func() { prFake = nil })
}

func boundEnv(t *testing.T) *[][]string {
	t.Helper()
	calls := withEnv(t)
	t.Setenv("FAKEGH_SEARCH_HITS", "[]")
	for _, k := range []string{"ASSAY_DESK_ISSUE_FILE_BOUND_PER_PR", "ASSAY_DESK_ISSUE_FILE_BOUND_PER_PR__WORKER",
		"ASSAY_DESK_ISSUE_FILE_BOUND_PER_PR__REVIEWER"} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}
	installBoundFake(t)
	return calls
}

var bfSeq int

// fileBound runs one `deskfile new` with a fresh title and the extra flags.
func fileBound(t *testing.T, extra ...string) (int, string) {
	t.Helper()
	bfSeq++
	body := bodyFileWith(t, "a blocker the review of this change requires")
	args := append([]string{"new", "-R", allowedRepo, "--title",
		fmt.Sprintf("bound blocker number %d for review", bfSeq), "--body-file", body}, extra...)
	return runCapture(args)
}

func boundEntries(t *testing.T) []deskkit.Entry {
	t.Helper()
	var out []deskkit.Entry
	for _, e := range readAudit(t) {
		if e.Tool == "deskfile" && e.Verb == boundVerb {
			out = append(out, e)
		}
	}
	return out
}

// TestBoundFiling: the binding, its refusals and its exemption from the per-session rate.
func TestBoundFiling(t *testing.T) {
	t.Run("spent budget, bound files", func(t *testing.T) {
		calls := boundEnv(t)
		seedNewAudit(t, defaultNewRate, allowedRepo, "test")
		// The unbound filing is refused by the spent rate (unchanged)...
		if rc, out := fileBound(t); rc != deskkit.ExitRateLimited {
			t.Fatalf("unbound rc = %d, want 4\n%s", rc, out)
		}
		assertNoIssueCreate(t, *calls)
		// ...and the bound one files.
		rc, out := fileBound(t, "--for-pr", allowedRepo+"#42", "--brief", "demo/07")
		if rc != deskkit.ExitOK {
			t.Fatalf("bound rc = %d, want 0\n%s", rc, out)
		}
		if curForge.filed == nil {
			t.Fatal("bound filing did not create an issue")
		}
		es := boundEntries(t)
		if len(es) != 1 || es[0].Result != deskkit.ResultOK || es[0].PR == nil || *es[0].PR != 42 {
			t.Fatalf("bound audit entry = %+v, want one ok entry with PR 42", es)
		}
		want := "for-pr " + allowedRepo + "#42 brief example:demo:07 role worker | " + createSentMarker
		if !strings.HasPrefix(es[0].Detail, want) {
			t.Fatalf("Detail = %q, want prefix %q", es[0].Detail, want)
		}
		// The back-link: one comment, on the change request, naming the issue.
		if len(curForge.comments) != 1 || curForge.comments[0] != 42 ||
			curForge.commentKinds[0] != deskkit.TargetChange {
			t.Fatalf("back-link comments = %v kinds %v, want one on change 42", curForge.comments, curForge.commentKinds)
		}
		// The per-session count is untouched: an unbound filing is still refused.
		if rc, _ := fileBound(t); rc != deskkit.ExitRateLimited {
			t.Fatalf("after bound, unbound rc = %d, want 4", rc)
		}
	})

	refuse := func(name string, wantRC int, extra ...string) {
		t.Run(name, func(t *testing.T) {
			calls := boundEnv(t)
			rc, out := fileBound(t, extra...)
			if rc != wantRC {
				t.Fatalf("rc = %d, want %d\n%s", rc, wantRC, out)
			}
			assertNoIssueCreate(t, *calls)
			if curForge.filed != nil || len(curForge.comments) != 0 {
				t.Fatal("a refused binding wrote to the forge")
			}
			for _, e := range readAudit(t) {
				if e.Tool == "deskfile" && chargedNewEntry(e) {
					t.Fatalf("a refused binding was charged: %+v", e)
				}
			}
		})
	}
	refuse("brief alone refused", deskkit.ExitRefused, "--brief", "demo/07")
	refuse("for-pr alone, PR has brief", deskkit.ExitRefused, "--for-pr", allowedRepo+"#42")
	refuse("for-pr alone, diff has brief", deskkit.ExitRefused, "--for-pr", allowedRepo+"#47")
	refuse("closed PR refused", deskkit.ExitRefused, "--for-pr", allowedRepo+"#44", "--brief", "demo/07")
	refuse("not the PR's brief", deskkit.ExitRefused, "--for-pr", allowedRepo+"#42", "--brief", "demo/08")
	refuse("brief resolves to no file", deskkit.ExitRefused, "--for-pr", allowedRepo+"#42", "--brief", "demo/09")
	refuse("qualified prefix mismatch", deskkit.ExitRefused, "--for-pr", allowedRepo+"#42", "--brief", "other:demo:07")
	refuse("unreadable PR holds", deskkit.ExitUnverifiable, "--for-pr", allowedRepo+"#45", "--brief", "demo/07")
	refuse("malformed for-pr", deskkit.ExitRefused, "--for-pr", "42", "--brief", "demo/07")
	refuse("force-file with for-pr", deskkit.ExitRefused, "--for-pr", allowedRepo+"#42", "--brief", "demo/07",
		"--force-file", "--reason", "urgent")

	t.Run("for-pr alone, PR has none", func(t *testing.T) {
		boundEnv(t)
		rc, out := fileBound(t, "--for-pr", allowedRepo+"#43")
		if rc != deskkit.ExitOK {
			t.Fatalf("rc = %d, want 0\n%s", rc, out)
		}
		es := boundEntries(t)
		if len(es) != 1 || !strings.HasPrefix(es[0].Detail, "for-pr "+allowedRepo+"#43 brief none role worker | ") {
			t.Fatalf("entries = %+v", es)
		}
	})

	t.Run("diff touches the brief", func(t *testing.T) {
		boundEnv(t)
		rc, out := fileBound(t, "--for-pr", allowedRepo+"#47", "--brief", "demo:07")
		if rc != deskkit.ExitOK {
			t.Fatalf("rc = %d, want 0\n%s", rc, out)
		}
	})

	t.Run("back-link failure holds", func(t *testing.T) {
		boundEnv(t)
		prFake.commentErr = fmt.Errorf("HTTP 500")
		rc, out := fileBound(t, "--for-pr", allowedRepo+"#42", "--brief", "demo/07")
		if rc != deskkit.ExitUnverifiable {
			t.Fatalf("rc = %d, want 6\n%s", rc, out)
		}
		if curForge.filed == nil || !strings.Contains(out, "issues/101") {
			t.Fatalf("the issue must stay filed and be named; out:\n%s", out)
		}
		// The create was sent, so the line charges the bound count.
		es := boundEntries(t)
		if len(es) != 1 || !chargedNewEntry(deskkit.Entry{Result: es[0].Result, Detail: strings.SplitN(es[0].Detail, " | ", 2)[1]}) {
			t.Fatalf("back-link failure entry = %+v, want a charged line", es)
		}
	})
}

// TestBoundFilingCap: the per-role, per-change cap and how its role is chosen.
func TestBoundFilingCap(t *testing.T) {
	boundEnv(t)
	t.Setenv("ASSAY_DESK_ISSUE_FILE_BOUND_PER_PR__WORKER", "2")
	bind42 := []string{"--for-pr", allowedRepo + "#42", "--brief", "demo/07"}
	for i := 1; i <= 2; i++ {
		if rc, out := fileBound(t, bind42...); rc != deskkit.ExitOK {
			t.Fatalf("filing %d rc = %d, want 0\n%s", i, rc, out)
		}
	}
	filedBefore := len(boundEntries(t))
	rc, out := fileBound(t, bind42...)
	if rc != deskkit.ExitRateLimited || !strings.Contains(out, "retry-after:") {
		t.Fatalf("third filing rc = %d, want 4 with retry-after\n%s", rc, out)
	}
	for _, e := range boundEntries(t) {
		if !strings.Contains(e.Detail, " role worker | ") {
			t.Fatalf("Detail does not record the session-token role: %q", e.Detail)
		}
	}

	// --raised-by another role, and a per-role override for another role, change nothing.
	t.Setenv("ASSAY_DESK_ISSUE_FILE_BOUND_PER_PR__REVIEWER", "100")
	if rc, out := fileBound(t, append(bind42, "--raised-by", "reviewer")...); rc != deskkit.ExitRateLimited {
		t.Fatalf("--raised-by reviewer rc = %d, want 4\n%s", rc, out)
	}

	// A different change request is a different count.
	if rc, out := fileBound(t, "--for-pr", allowedRepo+"#46", "--brief", "demo/07"); rc != deskkit.ExitOK {
		t.Fatalf("other PR rc = %d, want 0\n%s", rc, out)
	}

	// A different role (the session's own loop, not a flag) is counted separately.
	t.Setenv("ASSAY_DESK_ISSUE_FILE_BOUND_PER_PR__REVIEWER", "2")
	t.Setenv("DESK_LOOP", "pr-review-desk")
	if rc, out := fileBound(t, bind42...); rc != deskkit.ExitOK {
		t.Fatalf("reviewer on PR 42 rc = %d, want 0\n%s", rc, out)
	}
	last := boundEntries(t)
	if d := last[len(last)-1].Detail; !strings.Contains(d, " role reviewer | ") {
		t.Fatalf("reviewer Detail = %q", d)
	}
	t.Setenv("DESK_LOOP", "worker-desk")
	if rc, _ := fileBound(t, bind42...); rc != deskkit.ExitRateLimited {
		t.Fatalf("worker still capped after reviewer filed: rc = %d, want 4", rc)
	}
	_ = filedBefore

	// An unresolvable role holds before anything is created. cmdNew is called directly: run()
	// refuses an unset loop at its own gate first, which is not the path under test.
	for _, loop := range []string{"", "no-such-loop", "conference-scan"} {
		t.Run("role "+loop, func(t *testing.T) {
			created := curForge.filed
			curForge.filed = nil
			t.Setenv("DESK_LOOP", loop)
			body := bodyFileWith(t, "a blocker the review of this change requires")
			err := cmdNew([]string{"-R", allowedRepo, "--title", "unresolvable role filing", "--body-file", body,
				"--for-pr", allowedRepo + "#43"})
			if deskkit.ExitCodeOf(err) != deskkit.ExitUnverifiable {
				t.Fatalf("DESK_LOOP=%q: rc = %d (%v), want 6", loop, deskkit.ExitCodeOf(err), err)
			}
			if curForge.filed != nil {
				t.Fatal("an issue was created under an unresolvable role")
			}
			curForge.filed = created
		})
	}
}

// TestBoundFilingStillChargesRepoCeiling: with the cap off, the repo-wide deskfile write
// ceiling still refuses a bound filing.
func TestBoundFilingStillChargesRepoCeiling(t *testing.T) {
	calls := boundEnv(t)
	seedRepoWrites(t, deskkit.RateLimitPerPRPerHour)
	rc, out := fileBound(t, "--for-pr", allowedRepo+"#42", "--brief", "demo/07")
	if rc != deskkit.ExitRateLimited {
		t.Fatalf("rc = %d, want 4 (repo-wide ceiling)\n%s", rc, out)
	}
	assertNoIssueCreate(t, *calls)
}

// seedRepoWrites appends n charged deskfile attach writes on allowedRepo, each on its own
// issue, dated now — counted by the repo-wide tier, not by any one change's.
func seedRepoWrites(t *testing.T, n int) {
	t.Helper()
	dir := filepath.Join(os.Getenv("HOME"), ".config", "assay")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for i := 0; i < n; i++ {
		num := 500 + i
		line, _ := json.Marshal(deskkit.Entry{TS: time.Now().UTC().Format(time.RFC3339), Tool: "deskfile",
			Verb: "attach", Repo: allowedRepo, PR: &num, Result: deskkit.ResultOK, Detail: "seed", SessionTag: "test"})
		b.Write(append(line, '\n'))
	}
	f, err := os.OpenFile(filepath.Join(dir, "audit.jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(b.String()); err != nil {
		t.Fatal(err)
	}
}
