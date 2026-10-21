package deskkit

// The fixed discovery mode reuses SearchIssues' admitted, read-only transport.
// It is a candidate index, never a ruling snapshot or a permission grant. Ordinary
// dedupe searches retain their existing bounded behavior and wire shape.

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

const decisionSearchPages = 10 // GitHub search exposes at most 1,000 matches.

var (
	errDecisionSearchCap        = errors.New("decision-marker search result limit reached")
	errDecisionSearchIncomplete = errors.New("decision-marker search response or pagination incomplete")
)

// Closed diagnostic categories preserve the useful cause without leaking a
// provider response body, endpoint, credential or private locator into a row.
func decisionSearchUnknown(repo ForgeRepo, id string, err error) string {
	cause := "transport or response unavailable"
	var api *ForgeAPIError
	switch {
	case errors.Is(err, errDecisionSearchCap):
		cause = "decision-marker result limit reached"
	case errors.Is(err, errDecisionSearchIncomplete):
		cause = "decision-marker response or pagination incomplete"
	case ExitCodeOf(err) == ExitRefused:
		cause = "decision query refused locally"
	case errors.Is(err, errForgeTokenMissing):
		cause = "minted provider token unavailable"
	case IsForgeRateLimited(err):
		cause = "provider rate limit"
	case errors.As(err, &api) && (api.Status == http.StatusUnauthorized || api.Status == http.StatusForbidden):
		cause = "provider authentication or authorization failed"
	}
	return fmt.Sprintf("the search for %s's decision issue on %s could not run (%s)", id, repo.Slug(), cause)
}

func (g *GitHubForge) searchDecisionMarkers(repo ForgeRepo, brief string) ([]IssueSearchResult, error) {
	text := DecisionGateMarkerPrefix
	if brief != "" {
		text += ": " + brief
	}
	q := fmt.Sprintf("repo:%s is:issue %q in:body", repo.Slug(), text)
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
		path := fmt.Sprintf("/search/issues?q=%s&per_page=100&page=%d&sort=created&order=asc", url.QueryEscape(q), page)
		hdr, err := g.doJSONHeader(http.MethodGet, path, nil, &w)
		if err != nil {
			return nil, err
		}
		if w.Total != nil && *w.Total > 1000 {
			return nil, Unverifiable("decision-marker search exceeds 1000 results", errDecisionSearchCap)
		}
		if w.Total == nil || w.Incomplete == nil || w.Items == nil || *w.Total < 0 || *w.Incomplete {
			return nil, Unverifiable("decision-marker search is incomplete", errDecisionSearchIncomplete)
		}
		for _, item := range *w.Items {
			if item.Number <= 0 || item.Body == nil || seen[item.Number] || item.PullRequest != nil {
				return nil, Unverifiable("decision-marker search has invalid or repeated issues", errDecisionSearchIncomplete)
			}
			seen[item.Number] = true
			out = append(out, IssueSearchResult{Number: item.Number, Body: *item.Body})
		}
		if len(out) == *w.Total && !hasNextLink(hdr) {
			return out, nil
		}
		if len(out) > *w.Total || len(*w.Items) == 0 {
			return nil, Unverifiable("decision-marker search changed or ended before its total", errDecisionSearchIncomplete)
		}
	}
	return nil, Unverifiable("decision-marker search exceeded its page bound", errDecisionSearchIncomplete)
}

func (g *GitLabForge) searchDecisionMarkers(repo ForgeRepo, brief string) ([]IssueSearchResult, error) {
	cl, err := g.client()
	if err != nil {
		return nil, err
	}
	var out []IssueSearchResult
	seen := map[int]bool{}
	search := DecisionGateMarkerPrefix
	if brief != "" {
		search += ": " + brief
	}
	for page := 1; page <= gitlabMaxIssuePage; page++ {
		items, resp, err := cl.Issues.ListProjectIssues(repo.Slug(), &gitlab.ListProjectIssuesOptions{
			Search: gitlab.Ptr(search), In: gitlab.Ptr("description"),
			Scope: gitlab.Ptr("all"), State: gitlab.Ptr("all"),
			OrderBy: gitlab.Ptr("created_at"), Sort: gitlab.Ptr("asc"),
			ListOptions: gitlab.ListOptions{PerPage: gitlabPerPage, Page: int64(page)},
		})
		if err != nil {
			return nil, g.mapErr(http.MethodGet, "/projects/"+g.projectPath(repo)+"/issues", err)
		}
		if resp == nil || items == nil {
			return nil, Unverifiable("decision-marker search has no issue page", errDecisionSearchIncomplete)
		}
		for _, item := range items {
			if item == nil || item.IID <= 0 || seen[int(item.IID)] {
				return nil, Unverifiable("decision-marker search has invalid or repeated issues", errDecisionSearchIncomplete)
			}
			seen[int(item.IID)] = true
			out = append(out, IssueSearchResult{Number: int(item.IID), Body: item.Description})
		}
		if resp.NextPage == 0 {
			return out, nil
		}
		if resp.NextPage != int64(page+1) || len(items) == 0 {
			return nil, Unverifiable("decision-marker search has discontinuous pagination", errDecisionSearchIncomplete)
		}
	}
	return nil, Unverifiable("decision-marker search exceeded its page bound", errDecisionSearchCap)
}

func validateDecisionSearch(in SearchIssuesInput) error {
	if strings.TrimSpace(in.Query) != "" {
		return Refused("decision-marker discovery does not accept free-text qualifiers")
	}
	if strings.IndexFunc(in.DecisionGateBrief, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) || r == '"' || r == '\\' }) >= 0 {
		return Refused("decision-marker brief must be one opaque marker token")
	}
	return nil
}
