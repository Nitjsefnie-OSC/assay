package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/medici-finance/assay/tools/desk/internal/deskkit"
	"github.com/medici-finance/assay/tools/desk/internal/packet"
	rf "github.com/medici-finance/assay/tools/desk/internal/reviewflow"
	"github.com/medici-finance/assay/tools/desk/internal/reviewtest"
)

type prepareFixture struct {
	head, base, root string
	transport        *reviewtest.Transport
	moved            string
	prReads          int
}

func prepareTestFixture(t *testing.T) *prepareFixture {
	t.Helper()
	f := &prepareFixture{root: t.TempDir()}
	git := func(args ...string) string {
		c := exec.Command("git", args...)
		c.Dir = f.root
		c.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull)
		b, e := c.CombinedOutput()
		if e != nil {
			t.Fatal(e, string(b))
		}
		return strings.TrimSpace(string(b))
	}
	git("init", "-q")
	commit := func(body string) string {
		if e := os.WriteFile(filepath.Join(f.root, "a.go"), []byte(body), 0600); e != nil {
			t.Fatal(e)
		}
		git("add", "a.go")
		git("-c", "user.name=fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "fixture")
		return git("rev-parse", "HEAD")
	}
	f.base = commit("base bytes\n")
	f.head = commit("assigned bytes\n")
	commit("wrong allocated head\n")
	reviewtest.Home(t, reviewtest.Request{Home: t.TempDir()}, `ASSAY_BLESS_LOGIN=ada:2001
ASSAY_TRUSTED_LOGINS=ada:2001
ASSAY_TRUSTED_BOT_SLUGS=reviewer=github:reviewer:41987965
ASSAY_ALLOWED_REPOS=example-org/tracker:ci:private
ASSAY_REPO_FORGES=example-org/tracker=github
`)
	deskkit.ReloadConfig()
	t.Cleanup(deskkit.ReloadConfig)
	deskkit.SetGitHubCustodyMinter(func(string, deskkit.ForgeRepo) (string, string, error) { return "fixture-token", "", nil })
	t.Cleanup(func() { deskkit.SetGitHubCustodyMinter(nil) })
	oldReviewer := reviewPacketReviewerFn
	reviewPacketReviewerFn = func() (string, bool) { return "reviewer", true }
	t.Cleanup(func() { reviewPacketReviewerFn = oldReviewer })
	f.transport = &reviewtest.Transport{Answer: func(r *http.Request) (int, string) {
		path := r.URL.Path
		if r.Method != "GET" {
			t.Fatalf("preparation wrote %s %s", r.Method, path)
		}
		if strings.HasSuffix(path, "/pulls/77") && !strings.Contains(r.Header.Get("Accept"), "diff") {
			f.prReads++
			if f.prReads > 1 && f.moved == "access" {
				return 403, `{"message":"denied"}`
			}
			head, base, body := f.head, f.base, "description"
			if f.prReads > 1 {
				switch f.moved {
				case "head":
					head = strings.Repeat("e", 40)
				case "base":
					base = strings.Repeat("e", 40)
				case "body":
					body = "edited"
				}
			}
			return 200, fmt.Sprintf(`{"number":77,"state":"open","body":%q,"changed_files":1,"head":{"sha":%q,"ref":"feat/x"},"base":{"sha":%q,"ref":"main"}}`, body, head, base)
		}
		if strings.HasSuffix(path, "/timeline") {
			return 200, `[]`
		}
		if strings.HasSuffix(path, "/reviews") {
			if f.prReads > 1 && f.moved == "reviews" {
				return 200, `[]`
			}
			return 200, fmt.Sprintf(`[{"id":1,"state":"CHANGES_REQUESTED","body":"Verdict: request-changes\nF-1: fix the bytes","commit_id":%q,"user":{"login":"reviewer"},"submitted_at":"2026-10-01T00:00:00Z"}]`, f.base)
		}
		if strings.HasSuffix(path, "/protection/required_status_checks") {
			if f.prReads > 1 && f.moved == "policy" {
				return 200, `{"contexts":["changed"]}`
			}
			return 200, `{"contexts":["build"]}`
		}
		if strings.HasSuffix(path, "/check-runs") {
			conclusion := "success"
			if f.prReads > 1 && f.moved == "checks" {
				conclusion = "failure"
			}
			return 200, fmt.Sprintf(`{"total_count":1,"check_runs":[{"name":"build","status":"completed","conclusion":%q}]}`, conclusion)
		}
		if strings.HasSuffix(path, "/status") {
			return 200, `{"state":"success","total_count":0,"statuses":[]}`
		}
		if strings.HasSuffix(path, "/files") {
			return 200, `[{"filename":"a.go","status":"modified","patch":"@@ -1 +1 @@\n-base\n+assigned"}]`
		}
		if strings.Contains(path, "/compare/") {
			return 200, fmt.Sprintf(`{"total_commits":1,"commits":[{"sha":%q,"parents":[{"sha":%q}]}],"files":[{"filename":"a.go"}]}`, f.head, f.base)
		}
		if strings.Contains(path, "/commits/") {
			return 200, fmt.Sprintf(`{"sha":%q,"parents":[{"sha":%q}],"files":[{"filename":"a.go"}]}`, f.head, f.base)
		}
		if strings.Contains(path, "/contents/") {
			return 200, fmt.Sprintf(`{"type":"file","sha":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","encoding":"base64","content":%q}`, base64.StdEncoding.EncodeToString([]byte("assigned bytes\n")))
		}
		if strings.Contains(r.Header.Get("Accept"), "diff") {
			return 200, "diff --git a/a.go b/a.go\n-base bytes\n+assigned bytes\n"
		}
		t.Fatalf("unexpected preparation read %s", path)
		return 500, ""
	}}
	old := http.DefaultTransport
	http.DefaultTransport = f.transport
	t.Cleanup(func() { http.DefaultTransport = old })
	oldRound := reviewRoundForgeFn
	reviewRoundForgeFn = liveReviewRoundForge
	t.Cleanup(func() { reviewRoundForgeFn = oldRound })
	return f
}
func prepareFlow(t *testing.T, f *prepareFixture) (*preparedReview, packet.Spec) {
	t.Helper()
	o := dispatchOpts{kit: "review", pr: 77, root: f.root}
	p, e := newPreparedReview(o, rrRepo)
	if e != nil || p == nil {
		t.Fatal("context not prepared", e)
	}
	t.Cleanup(p.cancel)
	plan := dispatchPlan{claimKey: rrKeyC, reviewContext: p}
	round, e := prepareReviewRound(o, plan, rrRepo)
	if e != nil || round.head != f.head {
		t.Fatal("round lost assigned head", e, round)
	}
	spec, e := reviewPacket(packetInput{o: o, plan: plan, repo: rrRepo, home: f.root})
	if e != nil {
		t.Fatal(e)
	}
	return p, spec
}
func TestReviewPrepareAdaptersFlow(t *testing.T) {
	legacy := prepareTestFixture(t)
	o := dispatchOpts{kit: "review", pr: 77, root: legacy.root}
	plan := dispatchPlan{claimKey: rrKeyC}
	if _, e := prepareReviewRound(o, plan, rrRepo); e != nil {
		t.Fatal(e)
	}
	oldPacket := reviewPacketForgeFn
	reviewPacketForgeFn = func(repo string) (reviewPacketForge, deskkit.ForgeRepo, error) { return liveReviewRoundPacket(repo) }
	t.Cleanup(func() { reviewPacketForgeFn = oldPacket })
	before, e := reviewPacket(packetInput{o: o, plan: plan, repo: rrRepo, home: legacy.root})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = packet.Build(before); e != nil {
		t.Fatal(e)
	}
	beforeCalls := len(legacy.transport.Snapshot())
	f := prepareTestFixture(t)
	p, spec := prepareFlow(t, f)
	built, e := packet.Build(spec)
	if e != nil {
		t.Fatal("prepared packet failed", e)
	}
	if !strings.Contains(string(built.Text), "assigned bytes") || strings.Contains(string(built.Text), "wrong allocated head") {
		t.Fatal("packet did not use assigned-head bytes")
	}
	if f.prReads != 2 {
		t.Fatalf("shared preparation reread PR %d times; want initial + terminal", f.prReads)
	}
	counts := map[string]int{}
	for _, a := range f.transport.Snapshot() {
		counts[a.Path]++
	}
	if counts["/repos/"+rrRepo+"/pulls/77/files"] != 1 {
		t.Fatal("shared changed files were fetched twice", counts)
	}
	if counts["/repos/"+rrRepo+"/pulls/77/reviews"] != 2 {
		t.Fatal("reviews not shared before terminal check", counts)
	}
	for _, a := range f.transport.Snapshot() {
		if strings.Contains(a.Path, "/contents/") {
			t.Fatal("local exact blob incurred a provider read")
		}
	}
	for i := 0; i < 2; i++ {
		v, e := p.ReadFile(p.repo, deskkit.ReadFileInput{File: "a.go", Ref: f.head})
		if e != nil || string(v.Content) != "assigned bytes\n" {
			t.Fatal(v, e)
		}
	}
	if len(p.content) != 1 {
		t.Fatal("immutable content was not memoized once")
	}
	if beforeCalls != 17 || len(f.transport.Snapshot()) != 14 {
		t.Fatalf("unexpected cold cost before=%d after=%d", beforeCalls, len(f.transport.Snapshot()))
	}
	if len(f.transport.Snapshot()) >= beforeCalls {
		t.Fatalf("preparation did not reduce provider attempts: before=%d after=%d", beforeCalls, len(f.transport.Snapshot()))
	}
	t.Logf("legacy=%d full round+packet attempts=%d; PR=2 reviews=2 files=1 content=0; agent immutable refetch=0 under packet contract", beforeCalls, len(f.transport.Snapshot()))
}
func TestReviewPrepareAdaptersInvalidation(t *testing.T) {
	for _, change := range []string{"head", "base", "body", "reviews", "checks", "access", "policy"} {
		t.Run(change, func(t *testing.T) {
			f := prepareTestFixture(t)
			_, spec := prepareFlow(t, f)
			f.moved = change
			if _, e := packet.Build(spec); e == nil {
				t.Fatal("mutable preparation accepted changed " + change)
			}
		})
	}
	t.Run("wrong-kit", func(t *testing.T) {
		_, e := validateCallerPreconditions(dispatchOpts{kit: "worker", observation: "saved.json", pr: 77})
		if e == nil || !strings.Contains(e.Error(), "--observation requires") {
			t.Fatal("observation bypassed kit admission")
		}
	})
}

func liveReviewRoundPacket(repo string) (reviewPacketForge, deskkit.ForgeRepo, error) {
	fg, fr, e := liveReviewRoundForge(repo)
	if e != nil {
		return nil, fr, e
	}
	return fg.(reviewPacketForge), fr, nil
}

// Full entrypoint: the normal claim/workspace/stamp seams remain in control;
// only their external effects are replaced by the existing dispatch harness.
func TestReviewPrepareDispatchWiring(t *testing.T) {
	s := &stub{}
	s.install(t)
	realTransport := http.DefaultTransport
	f := prepareTestFixture(t)
	plantScripts(t, f.root)
	s.replies = happyReplies(filepath.Join(t.TempDir(), "review-home"))
	t.Setenv("DESK_LOOP", "pr-review-desk")
	fg, fr, e := liveReviewRoundForge(rrRepo)
	if e != nil {
		t.Fatal(e)
	}
	reviewRoundForgeFn = func(string) (reviewRoundForge, deskkit.ForgeRepo, error) { return fg, fr, nil }
	installGHStamp(t)
	http.DefaultTransport = prepareRoutingTransport{preparation: f.transport, effects: realTransport}
	stubQueueLabel(t, nil)
	usePacketBuilder(t)
	args := stampArgs(t, f.root, "--kit", "review", "--quiet", "--repo", rrRepo, "--observation", savedPrepareObservation(t, f, "valid"))
	args[0] = rrKeyC
	rc, stderr := runCapturingStderr(t, args)
	if rc != 0 {
		t.Fatalf("full preparation dispatch failed %d: %s", rc, stderr)
	}
	if f.prReads != 2 {
		t.Fatalf("dispatcher did not read preparation boundaries: PR reads=%d", f.prReads)
	}
	raw, e := os.ReadFile(emittedPathFor(promptFileOf(t, args)))
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(string(raw), "Packet:") || !strings.Contains(string(raw), f.head) {
		t.Fatal("dispatch lost exact-source packet/round")
	}
}

func savedPrepareObservation(t *testing.T, f *prepareFixture, change string) string {
	t.Helper()
	now := time.Now().Add(-time.Second)
	repo := rf.Repository{Provider: "github", Host: "github.com", Repository: rrRepo}
	id := rf.ChangeIdentity{Repository: repo, Number: 77, Head: f.head, Base: f.base, BaseUsed: true}
	if change == "head" {
		id.Head = strings.Repeat("e", 40)
	}
	if change == "base" {
		id.Base = strings.Repeat("e", 40)
	}
	if change == "subject" {
		id.Number = 78
	}
	prov := rf.InputProvenance{Source: "fixture", Version: "v1", Revision: "set1", ObservedAt: now, Freshness: "fresh", Complete: true}
	ap := rf.Applicability{Head: id.Head, Base: id.Base, Body: prov.Revision, Reviews: prov.Revision, Checks: prov.Revision}
	in := rf.ObservationInputs{Evaluator: rf.ObservationEvaluator, Policy: "untrusted-advisory-policy", ReviewerDisplay: "reviewer", Requested: []rf.Repository{repo}, Admitted: []rf.Repository{repo}, Collection: rf.CollectionInterval{Start: now, End: now}, Population: []rf.PopulationCoverage{{Repository: repo, Input: prov, Count: 1}}, Changes: []rf.ObservedChange{{Identity: id, AppliesTo: ap, Inputs: rf.ReviewInputs{Subject: prov, Body: prov, Reviews: prov, Checks: prov, Enrichment: prov}, Facts: rf.ClassificationFacts{Change: rf.ChangeFact{Labels: []string{}, Checks: []rf.CheckFact{}}}, Lanes: []rf.LaneConclusion{{Name: "correctness", State: "missing", AppliesTo: ap}, {Name: "security", State: "missing", AppliesTo: ap}}}}}
	o, e := rf.BuildObservation(in, rf.ObservationPolicy{Now: now, Oldest: now, Policy: in.Policy})
	if e != nil {
		t.Fatal(e)
	}
	if change == "partial" {
		o.Inputs.Population[0].Input.Complete = false
	}
	if change == "schema" {
		o.Schema = "unknown"
	}
	b, e := json.Marshal(o)
	if e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(t.TempDir(), "observation.json")
	if e = os.WriteFile(path, b, 0600); e != nil {
		t.Fatal(e)
	}
	return path
}
func TestReviewPrepareObservation(t *testing.T) {
	for _, change := range []string{"valid", "head", "base", "subject", "partial", "schema"} {
		t.Run(change, func(t *testing.T) {
			f := prepareTestFixture(t)
			path := savedPrepareObservation(t, f, change)
			p, e := newPreparedReview(dispatchOpts{kit: "review", pr: 77, root: f.root, observation: path}, rrRepo)
			if change == "valid" {
				if e != nil || p == nil {
					t.Fatal("valid advisory observation refused", e)
				}
				defer p.cancel()
				f.moved = "body"
				if p.recheck() == nil {
					t.Fatal("saved observation bypassed current body recheck")
				}
			} else {
				if e == nil {
					if p != nil {
						p.cancel()
					}
					t.Fatal("invalid advisory observation accepted " + change)
				}
			}
		})
	}
}

type prepareRoutingTransport struct{ preparation, effects http.RoundTripper }

func (t prepareRoutingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.URL.Hostname() == "api.github.com" {
		return t.preparation.RoundTrip(r)
	}
	return t.effects.RoundTrip(r)
}

func TestReviewPrepareKeys(t *testing.T) {
	f := prepareTestFixture(t)
	p, _ := prepareFlow(t, f)
	p.local = nil
	for i := 0; i < 2; i++ {
		if _, e := p.ReadFile(p.repo, deskkit.ReadFileInput{File: "a.go", Ref: f.head}); e != nil {
			t.Fatal(e)
		}
	}
	calls := func(part string) int {
		n := 0
		for _, a := range f.transport.Snapshot() {
			if strings.Contains(a.Path, part) {
				n++
			}
		}
		return n
	}
	if calls("/contents/") != 1 {
		t.Fatal("immutable provider content read repeated")
	}
	if _, e := p.ReadFile(p.repo, deskkit.ReadFileInput{File: "a.go", Ref: f.base}); e == nil {
		t.Fatal("another source head reused assigned bytes")
	}
	if _, e := p.ReadFile(p.repo, deskkit.ReadFileInput{File: "../a.go", Ref: f.head}); e == nil {
		t.Fatal("unsafe path reused source bytes")
	}
	before := calls("/compare/")
	for i := 0; i < 2; i++ {
		if _, e := p.CompareRefs(p.repo, f.base, f.head); e != nil {
			t.Fatal(e)
		}
	}
	if calls("/compare/") != before {
		t.Fatal("exact endpoint comparison from round was not reused")
	}
	for i := 0; i < 2; i++ {
		if _, e := p.CompareRefs(p.repo, "main", f.head); e != nil {
			t.Fatal(e)
		}
	}
	if calls("/compare/") != before+2 {
		t.Fatal("mutable branch comparison was cached")
	}
	if _, e := p.CompareRefs(p.repo, strings.Repeat("d", 40), f.head); e != nil {
		t.Fatal(e)
	}
	if calls("/compare/") != before+3 {
		t.Fatal("comparison reused another base endpoint")
	}
}
