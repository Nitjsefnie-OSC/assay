package main

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/medici-finance/assay/perfrec"
	"github.com/medici-finance/assay/tools/desk/internal/deskkit"
)

type perfWire func(*http.Request) (*http.Response, error)

func (f perfWire) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// regression: class issue-projection-v3-composition
func TestDiscoveryV3ActualRecorder(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ASSAY_TELEMETRY", "1")
	t.Setenv("DESK_TOOLS_DISABLED", "")
	oldMint, oldRole, oldBase, oldTransport := mintTokenFn, sessionRoleFn, forgeAPIBase, http.DefaultTransport
	defer func() {
		mintTokenFn, sessionRoleFn, forgeAPIBase, http.DefaultTransport = oldMint, oldRole, oldBase, oldTransport
	}()
	mintTokenFn = func(string, string) (string, string, error) { return "fixture-token", "", nil }
	sessionRoleFn = func(string) (string, string, error) { return "worker", "worker-desk", nil }
	forgeAPIBase = ""
	attempts := 0
	empty := false
	http.DefaultTransport = perfWire(func(r *http.Request) (*http.Response, error) {
		attempts++
		body := "[]"
		switch {
		case strings.HasSuffix(r.URL.Path, "/comments"):
			body = `[{"user":{"login":"assay-desk-app[bot]"},"body":"## Context\nfixture detail"}]`
		case strings.HasSuffix(r.URL.Path, "/issues/1"):
			body = `{"number":1,"body":"## Context\nfixture body","state":"open","user":{"login":"external","id":99}}`
		default:
			body = `[{"number":1,"title":"question-one","state":"open","created_at":"2026-01-01T00:00:00Z","user":{"login":"external","id":99},"labels":[{"name":"Question"}]}]`
		}
		if empty {
			body = "[]"
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})
	now := time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)
	args := []string{"discovery", "example-org/tracker"}
	var baseline, baseErr bytes.Buffer
	if code := execute(args, &baseline, &baseErr, now, nil); code != 0 {
		t.Fatal(code, baseErr.String())
	}
	attempts = 0
	var out, errs bytes.Buffer
	if code := measured(append(args, "--perf-profile=workflow", "--perf-role=review"), &out, &errs, now, nil); code != 0 || out.String() != baseline.String() || attempts != 1 {
		t.Fatal("native discovery parity", code, attempts, out.String(), errs.String())
	}
	history, err := deskkit.LoadPerfHistory(1)
	if err != nil || len(history.Records) != 1 {
		t.Fatal(err, len(history.Records))
	}
	sum, err := perfrec.ReduceV3(history.Records, perfrec.ProfileWorkflow, "tool", perfrec.V3Filter{})
	if err != nil {
		t.Fatal(err)
	}
	var reported int64
	for _, r := range sum.Requests {
		reported += r.Attempts
	}
	if reported != 1 || sum.RequestCoverage[perfrec.CoverageMeasured] != 1 || history.Records[0].V3.Attribution.Role != "review" {
		t.Fatalf("actual=%d reported=%d coverage=%v", attempts, reported, sum.RequestCoverage)
	}
	// Raw detail comments retain their original reader and pagination; its sends
	// are NOT falsely covered by native GetIssue observation.
	out.Reset()
	errs.Reset()
	attempts = 0
	if code := measured([]string{"walk", "example-org/tracker", "--perf-profile=workflow", "--perf-role=implement"}, &out, &errs, now, nil); code != 0 {
		t.Fatal(code, errs.String())
	}
	if attempts != 3 {
		t.Fatal("raw detail traffic changed", attempts)
	}
	history, err = deskkit.LoadPerfHistory(1)
	if err != nil || len(history.Records) != 2 {
		t.Fatal(err, len(history.Records))
	}
	rec := history.Records[1]
	var own int64
	for _, r := range rec.V3.Requests {
		own += r.Attempts
	}
	if own != 2 || rec.V3.RequestCoverage == perfrec.CoverageMeasured || rec.V3.Attribution.Role != "implement" {
		t.Fatalf("raw gap lost or invocation leaked: %+v", rec.V3)
	}
	// An empty discovery is still one measured enumeration send; it is not
	// proof that the raw detail reader was inventoried or that other queues are idle.
	empty = true
	out.Reset()
	errs.Reset()
	attempts = 0
	if code := measured(append(args, "--perf-profile=workflow"), &out, &errs, now, nil); code != 0 || attempts != 1 || strings.TrimSpace(out.String()) != "[]" {
		t.Fatal("empty discovery", code, attempts, out.String())
	}
	history, err = deskkit.LoadPerfHistory(1)
	if err != nil || len(history.Records) != 3 || history.Records[2].V3.RequestCoverage != perfrec.CoverageMeasured {
		t.Fatal("empty enumeration lost completeness", err)
	}
	empty = false
	// Consent-off invocation has the same native sends/output, no observer left over.
	t.Setenv("ASSAY_TELEMETRY", "")
	out.Reset()
	errs.Reset()
	attempts = 0
	if code := measured(append(args, "--perf-profile=workflow"), &out, &errs, now, nil); code != 0 || out.String() != baseline.String() || attempts != 1 {
		t.Fatal("consent off parity", code, attempts)
	}
	history, err = deskkit.LoadPerfHistory(1)
	if err != nil || len(history.Records) != 3 {
		t.Fatal("consent off wrote", err)
	}
	out.Reset()
	errs.Reset()
	attempts = 0
	if code := measured([]string{"--help", "--perf-profile=workflow", "--perf-session-id=fixture-private-context"}, &out, &errs, now, nil); code != 0 || attempts != 0 || strings.Contains(out.String()+errs.String(), "fixture-private-context") {
		t.Fatal("help performed work or echoed metadata")
	}
	files, _ := filepath.Glob(filepath.Join(os.Getenv("HOME"), ".config", "assay", "perf", "*.jsonl"))
	raw, _ := os.ReadFile(files[0])
	for _, private := range []string{"fixture-token", "fixture body", "fixture detail", "question-one", "api.github.com"} {
		if bytes.Contains(raw, []byte(private)) {
			t.Fatal("operational text leaked", private)
		}
	}
}
