package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/medici-finance/assay/tools/desk/internal/deskkit"
	"github.com/medici-finance/assay/tools/desk/internal/reviewtest"
)

// Real command and GitHubForge request encoder/decoder; only HTTP is substituted.
func TestQueueBridgeProbe(t *testing.T) {
	q := reviewtest.Read(t)
	q.Case = strings.TrimPrefix(q.Case, "queue-")
	reviewtest.Home(t, q, `ASSAY_BLESS_LOGIN=ada:2001
ASSAY_TRUSTED_LOGINS=ada:2001,shared-agent:2002
ASSAY_TRUSTED_BOT_SLUGS=reviewer=github:reviewer:41987965
ASSAY_ALLOWED_REPOS=example-org/tracker:ci:private
ASSAY_REPO_FORGES=example-org/tracker=github
`)
	deskkit.ReloadConfig()
	t.Cleanup(deskkit.ReloadConfig)
	t.Setenv("DESKBOARD_GATE_SCORES_JSON", "[]")
	mintTokenFn = func(string, string) (string, string, error) { return "fixture-token", "", nil }
	sessionRoleFn = func(string) (string, string, error) { return "worker", "fixture", nil }
	forgeAPIBase = "https://fixture.invalid"
	transport := &reviewtest.Transport{Answer: func(r *http.Request) (int, string) {
		if r.Method == "POST" && r.URL.Path == "/graphql" {
			if q.Case == "denied" {
				return 403, `{"message":"Forbidden"}`
			}
			var nodes []any
			for i := 1; i <= q.Size; i++ {
				head := strings.Repeat("a", 40)
				if q.Case == "missing-head" {
					head = ""
				}
				reviews := map[string]any{"pageInfo": map[string]any{"hasNextPage": false}, "nodes": []any{}}
				if q.Case == "lanes" {
					review := func(id int, state, body string) map[string]any {
						return map[string]any{"databaseId": id, "author": map[string]any{"login": "reviewer", "__typename": "Bot", "databaseId": 41987965}, "state": state, "commit": map[string]any{"oid": head}, "body": body, "submittedAt": "2026-10-02T00:00:00Z"}
					}
					switch i {
					case 1:
						reviews["nodes"] = []any{review(1, "APPROVED", "Security-Review: pass")}
					case 2:
						reviews["nodes"] = []any{review(2, "CHANGES_REQUESTED", "Verdict: request-changes"), review(3, "APPROVED", "Security-Review: pass")}
					case 3:
						reviews["nodes"] = []any{review(4, "APPROVED", "Verdict: approve")}
					}
				}
				if q.Case == "partial" {
					reviews["pageInfo"] = map[string]any{"hasNextPage": true}
				}
				nodes = append(nodes, map[string]any{"number": i, "title": fmt.Sprintf("change %d", i), "body": "", "state": "OPEN", "isDraft": true, "createdAt": "2026-10-01T00:00:00Z", "updatedAt": "2026-10-22T00:00:00Z", "author": map[string]any{"login": "shared-agent", "__typename": "User"}, "mergeStateStatus": "CLEAN", "headRefOid": head, "headRefName": "feat/x", "baseRefName": "main", "labels": map[string]any{"nodes": []any{}}, "reviews": reviews, "commits": map[string]any{"nodes": []any{map[string]any{"commit": map[string]any{"oid": head, "committedDate": "2026-10-01T00:00:00Z", "statusCheckRollup": map[string]any{"contexts": map[string]any{"nodes": []any{map[string]any{"__typename": "CheckRun", "name": "build", "databaseId": i, "status": "COMPLETED", "conclusion": "SUCCESS"}}}}}}}}})
			}
			b, _ := json.Marshal(map[string]any{"data": map[string]any{"repository": map[string]any{"pullRequests": map[string]any{"nodes": nodes}}}})
			return 200, string(b)
		}
		if strings.HasSuffix(r.URL.Path, "/files") {
			return 200, `[{"filename":"example.go","status":"modified"}]`
		}
		if q.Case != "partial" && r.Method == "GET" && strings.Contains(r.URL.Path, "/pulls/") {
			return 200, `{"number":1,"state":"open","changed_files":1,"head":{"sha":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}`
		}
		return 500, `{"message":"unexpected or unavailable enrichment"}`
	}}
	old := http.DefaultTransport
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = old })
	boardGuard = func() error { return nil }
	out := filepath.Join(q.Root, "observation.json")
	var stdout, stderr bytes.Buffer
	exit := run([]string{"review-queue", "--out", out, "--rate-floor", "0"}, &stdout, &stderr)
	b, _ := os.ReadFile(out)
	reviewtest.Write(t, reviewtest.Response{Exit: exit, Output: stdout.String() + stderr.String(), Snapshot: b, Attempts: transport.Snapshot(), HTTPComplete: true})
}
