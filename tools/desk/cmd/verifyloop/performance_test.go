package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/medici-finance/assay/perfrec"
	"github.com/medici-finance/assay/tools/desk/internal/deskkit"
)

// Use the production wrapper, command callback, writer and reader with existing
// offline preflight/forge seams; no real credentials or providers are involved.
func measuredVerify(t *testing.T, args []string, action func(verifyPerformance) int) (int, []perfrec.Record) {
	t.Helper()
	ctx := perfrec.ResolveContext(perfrec.ContextInput{})
	code := deskkit.PerfRunMeasured("verifyloop", verifyPerfVerbs, args, func(_ []string, s *perfrec.Session) (int, perfrec.Context) {
		if s != nil {
			s.SelectContext(ctx)
		}
		return action(verifyPerformance{session: s, context: ctx}), ctx
	}, verifyPerfPhases...)
	history, err := deskkit.LoadPerfHistory(1)
	if err != nil {
		t.Fatal(err)
	}
	if history.Skipped != 0 {
		t.Fatalf("invalid performance output: %d skipped", history.Skipped)
	}
	return code, history.Records
}
func phaseCount(r perfrec.Record, name string) int64 {
	for _, p := range r.Phases {
		if p.Name == name {
			return p.N
		}
	}
	return 0
}
func TestVerifyPerformanceOperationsAndParity(t *testing.T) {
	root := planFixtureRoot(t, map[string]string{"01": planBrief("model", "no", "no", "no", "no")})
	old := preflightRoot
	preflightRoot = func(string) error { return nil }
	t.Cleanup(func() { preflightRoot = old })
	for _, tc := range []struct {
		name string
		args []string
		verb string
		exit int
	}{
		{"plan", []string{"plan", "--root", root, "--json", "--no-forge"}, "plan", 0},
		{"verdict-refusal", []string{"verdict", "--unknown-option"}, "verdict", 5},
		{"dry-run", []string{"--dry-run", "--unknown-option"}, "verdict", 5},
		{"single-dash-dry-run", []string{"-dry-run", "--unknown-option"}, "verdict", 5},
		{"unknown", []string{"private-subject-canary"}, "other", 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var outputs []string
			for _, on := range []string{"0", "1"} {
				plantAllowedRoster(t, "example-org/tracker")
				t.Setenv("ASSAY_TELEMETRY", on)
				t.Setenv("ASSAY_PERF_PROFILE", "")
				t.Setenv("DESK_ROOTS", "")
				t.Setenv("DESK_LOOP", "verify-desk")
				var code int
				var records []perfrec.Record
				out := captureStdout(t, func() {
					code, records = measuredVerify(t, tc.args, func(p verifyPerformance) int { return runMeasured(tc.args, p) })
				})
				outputs = append(outputs, out)
				if code != tc.exit {
					t.Fatalf("recording=%s exit=%d want=%d", on, code, tc.exit)
				}
				if on == "0" {
					if len(records) != 0 {
						t.Fatal("disabled recording wrote a record")
					}
					if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".config", "assay", "perf")); !os.IsNotExist(err) {
						t.Fatalf("disabled recording created state: %v", err)
					}
					continue
				}
				if len(records) != 1 || records[0].Verb != tc.verb || records[0].Schema != perfrec.SchemaV2 {
					t.Fatalf("wrong operation/schema: %+v", records)
				}
				if phaseCount(records[0], "guard") != 1 {
					t.Fatal("missing guard phase")
				}
				if tc.verb == "plan" {
					var doc planDoc
					if err := json.Unmarshal([]byte(out), &doc); err != nil {
						t.Fatalf("JSON stdout polluted: %v", err)
					}
					for _, phase := range []string{"preflight", "queue-scan", "classify", "render"} {
						if phaseCount(records[0], phase) != 1 {
							t.Fatalf("missing %s phase", phase)
						}
					}
				}
				raw, _ := json.Marshal(records)
				if strings.Contains(string(raw), "private-subject-canary") || strings.Contains(string(raw), root) {
					t.Fatal("argv or path leaked")
				}
			}
			if outputs[0] != outputs[1] {
				t.Fatal("recording changed stdout")
			}
		})
	}
}

type verifyRoundTripper func(*http.Request) (*http.Response, error)

func (f verifyRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// Compose the real scheduling source/backend/counted transport with a fake leaf.
// Removing the invocation observer loses the request/cost assertions below.
func TestVerifyPerformanceSchedulingRequests(t *testing.T) {
	root := planFixtureRoot(t, map[string]string{"01": planBrief("model", "no", "no", "no", "no")})
	gitT(t, root, "init", "-q")
	gitT(t, root, "remote", "add", "origin", "https://github.com/example-org/tracker.git")
	old := issueSourceFn
	t.Cleanup(func() { issueSourceFn = old })
	for _, profile := range []string{"", "tool", "workflow"} {
		t.Run("profile-"+profile, func(t *testing.T) {
			plantAllowedRoster(t, "example-org/tracker")
			t.Setenv("ASSAY_TELEMETRY", "1")
			t.Setenv("ASSAY_PERF_PROFILE", profile)
			t.Setenv("DESK_ROOTS", "")
			sends := 0
			metadata := false
			issueSourceFn = func() deskkit.IssueStateSource {
				source := &forgeIssueSource{}
				source.forgeFor = func(deskkit.ForgeRepo) (deskkit.Forge, error) {
					return &deskkit.GitHubForge{Token: "fixture", Observer: source.observer, Client: &http.Client{Transport: verifyRoundTripper(func(r *http.Request) (*http.Response, error) {
						sends++
						body := `{"default_branch":"main"}`
						if r.URL.Path == "/graphql" {
							raw, _ := io.ReadAll(r.Body)
							metadata = strings.Contains(string(raw), "_assayPerf")
							body = `{"data":{"repository":{"pullRequests":{"pageInfo":{"hasNextPage":false},"nodes":[]}},"_assayPerf":{"cost":2}}}`
						}
						return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body))}, nil
					})}}, nil
				}
				return source
			}
			var records []perfrec.Record
			var code int
			captureStdout(t, func() {
				code, records = measuredVerify(t, []string{"plan"}, func(p verifyPerformance) int {
					return deskkit.ExitCodeOf(cmdPlanMeasured([]string{"--root", root, "--json"}, p))
				})
			})
			if code != 0 || sends != 2 || len(records) != 1 {
				t.Fatalf("exit=%d sends=%d records=%d", code, sends, len(records))
			}
			r := records[0]
			if phaseCount(r, "pending-discovery") != 1 || phaseCount(r, "pending-evidence") != 1 {
				t.Fatal("scheduling phases missing")
			}
			if r.Counters[perfrec.CounterHTTP] != 2 {
				t.Fatal("actual attempts not counted")
			}
			if profile == "" {
				if r.V3 != nil || metadata {
					t.Fatal("legacy consent expanded")
				}
				return
			}
			if r.V3 == nil || !metadata || r.V3.RequestCoverage != perfrec.CoverageMissing {
				t.Fatal("observer absent or full inventory falsely claimed")
			}
			var attempts, points int64
			for _, q := range r.V3.Requests {
				attempts += q.Attempts
				if q.Points.Coverage == perfrec.CoverageMeasured {
					points += *q.Points.Value
				}
			}
			if attempts != 2 || points != 2 {
				t.Fatalf("attempts=%d points=%d", attempts, points)
			}
			sum, err := perfrec.ReduceV3(records, perfrec.ProfileTool, "tool", perfrec.V3Filter{})
			if err != nil || len(sum.Requests) != 2 {
				t.Fatalf("record/report flow: %+v %v", sum, err)
			}
		})
	}
}
func TestVerifyPerformancePendingAdmissionParity(t *testing.T) {
	v, f, _, _ := pendingSetup(t)
	t.Setenv("ASSAY_TELEMETRY", "1")
	t.Setenv("ASSAY_PERF_PROFILE", "tool")
	baseline, err := v.SelectQueue()
	if err != nil {
		t.Fatal(err)
	}
	source := &forgeIssueSource{forgeFor: func(deskkit.ForgeRepo) (deskkit.Forge, error) { return f, nil }}
	v.Issues = source
	code, records := measuredVerify(t, []string{"plan"}, func(p verifyPerformance) int {
		v.performance, source.performance = p, p
		items, e := v.SelectQueue()
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(items, baseline) {
			t.Fatal("measurement changed pending admission")
		}
		return 0
	})
	if code != 0 || len(records) != 1 {
		t.Fatalf("%d %+v", code, records)
	}
	for phase, n := range map[string]int64{"queue-scan": 1, "pending-discovery": 1, "pending-evidence": 2, "attestation": 1} {
		if phaseCount(records[0], phase) != n {
			t.Fatalf("%s count=%d want=%d", phase, phaseCount(records[0], phase), n)
		}
	}
	if !records[0].PhaseOverlap {
		t.Fatal("inclusive nested phases not marked")
	}
}
func TestVerifyPerformanceVerdictPhases(t *testing.T) {
	plantAllowedRoster(t, "example-org/tracker")
	t.Setenv("ASSAY_TELEMETRY", "1")
	t.Setenv("ASSAY_PERF_PROFILE", "")
	pem, _ := writeTestKey(t)
	cfg := verdictRunConfig{root: demoRoot(t), pem: pem, dryRun: true, exec: fakeExec, now: fixedNow, out: io.Discard}
	code, records := measuredVerify(t, []string{"verdict"}, func(p verifyPerformance) int {
		cfg.performance = p
		if err := runVerdict(cfg); err != nil {
			t.Fatal(err)
		}
		return 0
	})
	if code != 0 || len(records) != 1 {
		t.Fatal("verdict measurement missing")
	}
	for _, phase := range []string{"queue-scan", "rows", "sign"} {
		if phaseCount(records[0], phase) == 0 {
			t.Fatalf("missing %s", phase)
		}
	}
}
