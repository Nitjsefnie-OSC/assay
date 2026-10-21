package deskkit

// The fixed discovery mode reuses SearchIssues' admitted, read-only transport.
// It is a candidate index, never a ruling snapshot or a permission grant. Ordinary
// dedupe searches retain their existing bounded behavior and wire shape.

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

const decisionSearchPages = 10 // GitHub search exposes at most 1,000 matches.

func (g *GitHubForge) searchDecisionMarkers(repo ForgeRepo) ([]IssueSearchResult, error) {
	q := fmt.Sprintf("repo:%s is:issue %q in:body", repo.Slug(), DecisionGateMarkerPrefix)
	var out []IssueSearchResult
	seen := map[int]bool{}
	for page := 1; page <= decisionSearchPages; page++ {
		var w struct {
			Total      *int  `json:"total_count"`
			Incomplete *bool `json:"incomplete_results"`
			Items      *[]struct {
				Number      int       `json:"number"`
				Body        *string   `json:"body"`
				PullRequest *struct{} `json:"pull_request"`
			} `json:"items"`
		}
		path := fmt.Sprintf("/search/issues?q=%s&per_page=100&page=%d", url.QueryEscape(q), page)
		hdr, err := g.doJSONHeader(http.MethodGet, path, nil, &w)
		if err != nil {
			return nil, err
		}
		if w.Total == nil || w.Incomplete == nil || w.Items == nil || *w.Total < 0 || *w.Incomplete || *w.Total > 1000 {
			return nil, Unverifiable("decision-marker search is incomplete", nil)
		}
		for _, item := range *w.Items {
			if item.Number <= 0 || item.Body == nil || seen[item.Number] || item.PullRequest != nil {
				return nil, Unverifiable("decision-marker search has invalid or repeated issues", nil)
			}
			seen[item.Number] = true
			out = append(out, IssueSearchResult{Number: item.Number, Body: *item.Body})
		}
		if len(out) == *w.Total && !hasNextLink(hdr) {
			return out, nil
		}
		if len(out) > *w.Total || len(*w.Items) == 0 {
			return nil, Unverifiable("decision-marker search changed or ended before its total", nil)
		}
	}
	return nil, Unverifiable("decision-marker search exceeded its page bound", nil)
}

func (g *GitLabForge) searchDecisionMarkers(repo ForgeRepo) ([]IssueSearchResult, error) {
	cl, err := g.client()
	if err != nil {
		return nil, err
	}
	var out []IssueSearchResult
	seen := map[int]bool{}
	for page := 1; page <= gitlabMaxIssuePage; page++ {
		items, resp, err := cl.Issues.ListProjectIssues(repo.Slug(), &gitlab.ListProjectIssuesOptions{
			Search: gitlab.Ptr(DecisionGateMarkerPrefix), In: gitlab.Ptr("description"),
			Scope: gitlab.Ptr("all"), State: gitlab.Ptr("all"),
			OrderBy: gitlab.Ptr("created_at"), Sort: gitlab.Ptr("asc"),
			ListOptions: gitlab.ListOptions{PerPage: gitlabPerPage, Page: int64(page)},
		})
		if err != nil {
			return nil, g.mapErr(http.MethodGet, "/projects/"+g.projectPath(repo)+"/issues", err)
		}
		if resp == nil || items == nil {
			return nil, Unverifiable("decision-marker search has no issue page", nil)
		}
		for _, item := range items {
			if item == nil || item.IID <= 0 || seen[int(item.IID)] {
				return nil, Unverifiable("decision-marker search has invalid or repeated issues", nil)
			}
			seen[int(item.IID)] = true
			out = append(out, IssueSearchResult{Number: int(item.IID), Body: item.Description})
		}
		if resp.NextPage == 0 {
			return out, nil
		}
		if resp.NextPage != int64(page+1) || len(items) == 0 {
			return nil, Unverifiable("decision-marker search has discontinuous pagination", nil)
		}
	}
	return nil, Unverifiable("decision-marker search exceeded its page bound", nil)
}

func validateDecisionSearch(in SearchIssuesInput) error {
	if strings.TrimSpace(in.Query) != "" {
		return Refused("decision-marker discovery does not accept free-text qualifiers")
	}
	return nil
}
