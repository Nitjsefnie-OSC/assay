package deskkit

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// filingkey_test.go — the exact filing identity (brief example-stream/49): its derivation, the local
// confirmation of search hits, and the typed forge's complete-or-error exact search.

const fkRepo = "example-org/tracker"

func mustKey(t *testing.T, repo, brief, role, caseID string) string {
	t.Helper()
	k, err := FilingKey(repo, brief, role, caseID)
	if err != nil {
		t.Fatalf("FilingKey(%q,%q,%q,%q): %v", repo, brief, role, caseID, err)
	}
	return k
}

// regression: example-stream/49
func TestFilingKeyIdentity(t *testing.T) {
	base := mustKey(t, fkRepo, "example-stream/49", "verifier", "repair/0123456789abcdef")
	// Same identity under every spelling of its parts.
	for _, k := range []string{
		mustKey(t, "Example-Org/Tracker", "example-stream/49", "verifier", "repair/0123456789abcdef"),
		mustKey(t, fkRepo, "assay:assay:example-stream:49", "verifier", "repair/0123456789abcdef"),
		mustKey(t, fkRepo, " example-stream/49 ", "verifier", " repair/0123456789abcdef "),
	} {
		if k != base {
			t.Fatalf("key %s differs from %s for the same identity", k, base)
		}
	}
	// Every component changes the identity.
	for name, k := range map[string]string{
		"repo":   mustKey(t, "example-org/agents", "example-stream/49", "verifier", "repair/0123456789abcdef"),
		"brief":  mustKey(t, fkRepo, "example-stream/50", "verifier", "repair/0123456789abcdef"),
		"role":   mustKey(t, fkRepo, "example-stream/49", "worker", "repair/0123456789abcdef"),
		"case":   mustKey(t, fkRepo, "example-stream/49", "verifier", "repair/fedcba9876543210"),
		"prefix": mustKey(t, fkRepo, "example-stream/49", "verifier", "receipt/0123456789abcdef"),
	} {
		if k == base {
			t.Fatalf("changing the %s did not change the key", name)
		}
	}
	for name, args := range map[string][4]string{
		"bad repo":  {"not a slug", "example-stream/49", "verifier", "repair/0123456789abcdef"},
		"bad brief": {fkRepo, "49", "verifier", "repair/0123456789abcdef"},
		"bad role":  {fkRepo, "example-stream/49", "nobody", "repair/0123456789abcdef"},
		"bad case":  {fkRepo, "example-stream/49", "verifier", "repair/xyz"},
		"empty":     {"", "", "", ""},
	} {
		if k, err := FilingKey(args[0], args[1], args[2], args[3]); err == nil || k != "" {
			t.Fatalf("%s: key %q err %v, want a refusal", name, k, err)
		}
	}
	o := NewRepairObligation(fkRepo, "example-stream/49", "receipt/aaaaaaaaaaaaaaaa", []int{2}, BlockerImplementation, "r", "e", "", "2026-10-22")
	if got, err := o.FilingKey("verifier"); err != nil || got != mustKey(t, fkRepo, "example-stream/49", "verifier", o.ID) {
		t.Fatalf("obligation key %q %v", got, err)
	}
}

func TestHasFilingKeyClaim(t *testing.T) {
	for body, want := range map[string]bool{
		"plain text":           false,
		"filing key unrelated": false,
		"<!-- filing-key: fk1/0123456789abcdef -->": true,
		"FILING-KEY :  fk1/0123456789abcdef":        true,
		"Filing_Key: fk9/0123456789abcdef":          true,
		"<!--filing key: fk1/zz-->":                 true,
	} {
		if got := HasFilingKeyClaim(body); got != want {
			t.Errorf("HasFilingKeyClaim(%q) = %v, want %v", body, got, want)
		}
	}
}

func TestMatchFiledIssues(t *testing.T) {
	key := mustKey(t, fkRepo, "example-stream/49", "verifier", "repair/0123456789abcdef")
	other := mustKey(t, fkRepo, "example-stream/49", "verifier", "repair/fedcba9876543210")
	m, om := FilingKeyMarker(key), FilingKeyMarker(other)
	hits := []IssueSearchResult{
		{Number: 9, State: "open", Labels: []string{"raised-by:verifier"}, Body: "later\n" + m},
		{Number: 3, State: "closed", Labels: []string{"raised-by:verifier"}, Body: m},
		{Number: 4, State: "open", Labels: []string{"raised-by:worker"}, Body: m},
		{Number: 5, State: "open", Labels: []string{"raised-by:verifier"}, Body: "quoted without a marker: filing-key: " + key},
		{Number: 6, State: "open", Labels: []string{"raised-by:verifier"}, Body: om},
		{Number: 7, State: "open", Labels: []string{"bug"}, Body: m}, // unstamped: the stamp may have failed
		{Number: 2, State: "OPEN", Labels: []string{"bug", "raised-by:verifier"}, Body: "x\n" + m},
	}
	got := MatchFiledIssues(hits, key, "verifier")
	if len(got) != 3 || got[0].Number != 2 || got[1].Number != 7 || got[2].Number != 9 {
		t.Fatalf("matched %+v, want #2, #7 (unstamped) and #9 only", got)
	}
	if MatchFiledIssues(hits, "garbage", "verifier") != nil || MatchFiledIssues(hits, key, "nobody") != nil {
		t.Fatal("a malformed key or unknown role must match nothing")
	}
}

// regression: example-stream/49
func TestFilingKeySearchGitHub(t *testing.T) {
	key := mustKey(t, fkRepo, "example-stream/49", "verifier", "repair/0123456789abcdef")
	var queries []string
	gh := newB33GitHub(t, func(w http.ResponseWriter, r *http.Request, _ string) {
		queries = append(queries, r.URL.Query().Get("q"))
		page := r.URL.Query().Get("page")
		if page == "1" {
			w.Header().Set("Link", `<https://example.invalid/?page=2>; rel="next"`)
		}
		fmt.Fprintf(w, `{"total_count":2,"incomplete_results":false,"items":[{"number":%s,"title":"t%s","state":"open","html_url":"https://example.invalid/%s","body":"b","labels":[{"name":"raised-by:verifier"}]}]}`, page, page, page)
	})
	got, err := gh.forge().SearchIssues(forgeTestRepo, SearchIssuesInput{FilingKey: key})
	if err != nil || len(got) != 2 {
		t.Fatalf("walk: %+v %v", got, err)
	}
	if got[1].Number != 2 || got[1].Title != "t2" || got[1].State != "open" || got[1].URL != "https://example.invalid/2" || len(got[1].Labels) != 1 {
		t.Fatalf("the exact search must return the issue it reconciles to: %+v", got[1])
	}
	want := fmt.Sprintf(`repo:%s is:issue is:open "filing-key: %s" in:body`, forgeTestRepo.Slug(), key)
	for _, q := range queries {
		if q != want {
			t.Fatalf("query %q, want %q", q, want)
		}
	}
}

// regression: example-stream/49
func TestFilingKeySearchGitHubIncomplete(t *testing.T) {
	key := mustKey(t, fkRepo, "example-stream/49", "verifier", "repair/0123456789abcdef")
	for name, body := range map[string]string{
		"partial":        `{"total_count":1,"incomplete_results":true,"items":[{"number":1,"body":"x"}]}`,
		"missing_total":  `{"incomplete_results":false,"items":[]}`,
		"search_ceiling": `{"total_count":1001,"incomplete_results":false,"items":[]}`,
		"early_empty":    `{"total_count":2,"incomplete_results":false,"items":[]}`,
		"pull_request":   `{"total_count":1,"incomplete_results":false,"items":[{"number":1,"body":"x","pull_request":{}}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			gh := newB33GitHub(t, func(w http.ResponseWriter, _ *http.Request, _ string) { fmt.Fprint(w, body) })
			if got, err := gh.forge().SearchIssues(forgeTestRepo, SearchIssuesInput{FilingKey: key}); err == nil || got != nil {
				t.Fatalf("a partial read became an answer: %+v %v", got, err)
			}
		})
	}
	// A complete empty read is a definite "not filed", distinct from the unknowns above.
	gh := newB33GitHub(t, func(w http.ResponseWriter, _ *http.Request, _ string) {
		fmt.Fprint(w, `{"total_count":0,"incomplete_results":false,"items":[]}`)
	})
	if got, err := gh.forge().SearchIssues(forgeTestRepo, SearchIssuesInput{FilingKey: key}); err != nil || len(got) != 0 {
		t.Fatalf("empty complete read: %+v %v", got, err)
	}
	// The mode refuses a malformed key and refuses to combine with free text.
	for _, in := range []SearchIssuesInput{
		{FilingKey: "fk1/zz"},
		{FilingKey: key, Query: "stale oracle"},
		{FilingKey: key, DecisionGateMarkers: true},
	} {
		if _, err := gh.forge().SearchIssues(forgeTestRepo, in); err == nil || !strings.Contains(err.Error(), "filing-key") {
			t.Fatalf("%+v must be refused as a filing-key misuse, got %v", in, err)
		}
	}
}
