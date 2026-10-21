package deskkit

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// regression: medici-finance/assay#3373
func TestDecisionMarkerSearchGitHubPages(t *testing.T) {
	gh := newB33GitHub(t, func(w http.ResponseWriter, r *http.Request, _ string) {
		q := r.URL.Query()
		if r.URL.Path != "/search/issues" || q.Get("q") != `repo:example-org/tracker is:issue "decision-gate" in:body` || q.Get("per_page") != "100" {
			t.Errorf("unexpected scope/query %s", r.URL)
		}
		page := q.Get("page")
		if page == "1" {
			w.Header().Set("Link", `<https://example.invalid/?page=2>; rel="next"`)
		} else if page != "2" {
			t.Errorf("unexpected page %s", page)
		}
		fmt.Fprintf(w, `{"total_count":2,"incomplete_results":false,"items":[{"number":%s,"body":"<!-- decision-gate: s/01 -->"}]}`, page)
	})
	// Use the real wrapper: discovery remains a read, never a new credential path.
	got, err := ReadOnly(gh.forge()).SearchIssues(ForgeRepo{Owner: "example-org", Name: "tracker"}, SearchIssuesInput{DecisionGateMarkers: true})
	if err != nil || len(got) != 2 || got[1].Number != 2 || len(gh.seen) != 2 {
		t.Fatalf("walk: %+v %v requests=%v", got, err, gh.seen)
	}
}

// regression: medici-finance/assay#3373
func TestDecisionMarkerSearchGitHubIncomplete(t *testing.T) {
	for name, body := range map[string]string{
		"partial":          `{"total_count":1,"incomplete_results":true,"items":[{"number":1,"body":"decision-gate: s/01"}]}`,
		"missing_total":    `{"incomplete_results":false,"items":[]}`,
		"missing_coverage": `{"total_count":0,"items":[]}`,
		"null_items":       `{"total_count":0,"incomplete_results":false,"items":null}`,
		"missing_body":     `{"total_count":1,"incomplete_results":false,"items":[{"number":1}]}`,
		"search_ceiling":   `{"total_count":1001,"incomplete_results":false,"items":[]}`,
		"early_empty":      `{"total_count":2,"incomplete_results":false,"items":[]}`,
		"duplicate":        `{"total_count":2,"incomplete_results":false,"items":[{"number":1,"body":"x"},{"number":1,"body":"x"}]}`,
		"pull_request":     `{"total_count":1,"incomplete_results":false,"items":[{"number":1,"body":"x","pull_request":{}}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			gh := newB33GitHub(t, func(w http.ResponseWriter, _ *http.Request, _ string) { fmt.Fprint(w, body) })
			if got, err := gh.forge().SearchIssues(forgeTestRepo, SearchIssuesInput{DecisionGateMarkers: true}); err == nil || got != nil {
				t.Fatalf("partial became absence: %+v %v", got, err)
			}
		})
	}
	// More pages despite a matching total must not be silently dropped.
	gh := newB33GitHub(t, func(w http.ResponseWriter, _ *http.Request, _ string) {
		w.Header().Set("Link", `<https://example.invalid/>; rel="next"`)
		fmt.Fprint(w, `{"total_count":1,"incomplete_results":false,"items":[{"number":1,"body":"x"}]}`)
	})
	if got, err := gh.forge().SearchIssues(forgeTestRepo, SearchIssuesInput{DecisionGateMarkers: true}); err == nil || got != nil {
		t.Fatal("next page ignored", got, err)
	}
}

// regression: medici-finance/assay#3373
func TestDecisionMarkerSearchGitLabPages(t *testing.T) {
	pages := 0
	s := newB33GitHub(t, func(w http.ResponseWriter, r *http.Request, _ string) {
		pages++
		q := r.URL.Query()
		if q.Get("search") != "decision-gate" || q.Get("in") != "description" || q.Get("state") != "all" || q.Get("scope") != "all" || !strings.HasSuffix(r.URL.Path, "/issues") {
			t.Errorf("wrong candidate population: %s", r.URL)
		}
		if pages == 1 {
			w.Header().Set("X-Next-Page", "2")
		}
		fmt.Fprintf(w, `[{"id":100,"iid":%d,"description":"<!-- decision-gate: s/01 -->"}]`, pages)
	})
	f := &GitLabForge{Token: "fixture-token", BaseURL: s.srv.URL}
	got, err := ReadOnly(f).SearchIssues(forgeTestRepo, SearchIssuesInput{DecisionGateMarkers: true})
	if err != nil || len(got) != 2 || pages != 2 || got[1].Number != 2 {
		t.Fatalf("walk: %+v %v pages=%d", got, err, pages)
	}
}

// regression: medici-finance/assay#3373
func TestDecisionMarkerSearchGitLabIncomplete(t *testing.T) {
	for _, mode := range []string{"ceiling", "skip", "duplicate", "null"} {
		t.Run(mode, func(t *testing.T) {
			pages := 0
			s := newB33GitHub(t, func(w http.ResponseWriter, _ *http.Request, _ string) {
				pages++
				if mode == "null" {
					fmt.Fprint(w, `null`)
					return
				}
				n := pages
				switch mode {
				case "ceiling":
					w.Header().Set("X-Next-Page", fmt.Sprint(pages+1))
				case "skip":
					w.Header().Set("X-Next-Page", "3")
				case "duplicate":
					n = 1
					if pages == 1 {
						w.Header().Set("X-Next-Page", "2")
					}
				}
				fmt.Fprintf(w, `[{"id":100,"iid":%d,"description":"x"}]`, n)
			})
			got, err := (&GitLabForge{Token: "fixture-token", BaseURL: s.srv.URL}).SearchIssues(forgeTestRepo, SearchIssuesInput{DecisionGateMarkers: true})
			if err == nil || got != nil {
				t.Fatal("partial became absence", got, err)
			}
			if mode == "ceiling" && pages != gitlabMaxIssuePage {
				t.Fatal("page bound changed", pages)
			}
		})
	}
}

// regression: medici-finance/assay#3373
func TestDecisionMarkerSearchRejectsQualifiers(t *testing.T) {
	gh := newB33GitHub(t, func(http.ResponseWriter, *http.Request, string) { t.Error("invalid mode reached transport") })
	for _, f := range []DecisionForge{gh.forge(), &GitLabForge{Token: "fixture-token", BaseURL: gh.srv.URL}} {
		if _, err := f.SearchIssues(forgeTestRepo, SearchIssuesInput{Query: "repo:other/tracker", DecisionGateMarkers: true}); err == nil {
			t.Fatal("qualifier accepted")
		}
	}
}
