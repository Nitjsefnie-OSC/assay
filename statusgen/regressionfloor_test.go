package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/medici-finance/assay/statusgen/domain"
	"github.com/medici-finance/assay/statusgen/evaluation"
	"gopkg.in/yaml.v3"
)

const floorDir = "testdata/regressionfloor"

type floorEntry struct {
	File, Test, Contract, Source, Disposition, Reason string
	BodySHA256                                        string `yaml:"body_sha256"`
	Cases                                             []string
}
type floorInventory struct {
	Entries []floorEntry
	Added   []string `yaml:"added_cases"`
}
type floorResult struct {
	Released    bool     `json:"released"`
	Claim       string   `json:"claim"`
	Eligibility string   `json:"eligibility"`
	Holds       []string `json:"hold_states" yaml:"hold_states"`
	Notices     []string `json:"notice_states" yaml:"notice_states"`
}
type floorCase struct {
	ID, Verify, Evidence, Edge string
	Expected                   floorResult
}
type floorDifference struct {
	Case, Reason string
	Old, New     floorResult
}

func floorRead(path string, into any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	d := yaml.NewDecoder(bytes.NewReader(b))
	d.KnownFields(true)
	return d.Decode(into)
}
func floorLoad(dir string) (floorInventory, []floorCase, []floorDifference, error) {
	var inv floorInventory
	var diffs []floorDifference
	if err := floorRead(filepath.Join(dir, "inventory.yaml"), &inv); err != nil {
		return inv, nil, nil, err
	}
	if err := floorRead(filepath.Join(dir, "differences.yaml"), &diffs); err != nil {
		return inv, nil, nil, err
	}
	paths, err := filepath.Glob(filepath.Join(dir, "cases", "*.json"))
	if err != nil {
		return inv, nil, nil, err
	}
	selected := map[string]bool{}
	for _, e := range inv.Entries {
		for _, id := range e.Cases {
			selected[id] = true
		}
	}
	for _, id := range inv.Added {
		selected[id] = true
	}
	var cases []floorCase
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			return inv, nil, nil, err
		}
		var c floorCase
		d := json.NewDecoder(bytes.NewReader(b))
		d.DisallowUnknownFields()
		if err = d.Decode(&c); err != nil {
			return inv, nil, nil, err
		}
		if c.ID == "" || filepath.Base(p) != c.ID+".json" {
			return inv, nil, nil, fmt.Errorf("invalid case identity: %s", p)
		}
		if !selected[c.ID] {
			return inv, nil, nil, fmt.Errorf("unselected case %s", c.ID)
		}
		delete(selected, c.ID)
		cases = append(cases, c)
	}
	if len(cases) == 0 {
		return inv, nil, nil, fmt.Errorf("zero selected cases")
	}
	if len(selected) > 0 {
		return inv, nil, nil, fmt.Errorf("selected case missing: %v", selected)
	}
	return inv, cases, diffs, nil
}
func floorResolve(inv floorInventory) error {
	contracts, err := os.ReadFile("../docs/contracts.md")
	if err != nil {
		return err
	}
	if len(inv.Entries) == 0 {
		return fmt.Errorf("empty inventory")
	}
	seen := map[string]bool{}
	for _, e := range inv.Entries {
		if !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(e.Source) {
			return fmt.Errorf("unqualified source for %s", e.Test)
		}
		if !strings.Contains(string(contracts), "| "+e.Contract+" |") {
			return fmt.Errorf("unknown contract %s", e.Contract)
		}
		switch e.Disposition {
		case "adapt harness", "reuse unchanged", "replace", "add":
		default:
			return fmt.Errorf("invalid disposition %s", e.Disposition)
		}
		if e.Reason == "" {
			return fmt.Errorf("missing disposition reason %s", e.Test)
		}
		if filepath.IsAbs(e.File) || strings.Contains(e.File, "..") {
			return fmt.Errorf("invalid inventory path")
		}
		key := e.File + ":" + e.Test
		if seen[key] {
			return fmt.Errorf("duplicate entry %s", key)
		}
		seen[key] = true
		f, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", e.File), nil, 0)
		if err != nil {
			return err
		}
		found := false
		for _, d := range f.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Recv == nil && fn.Name.Name == e.Test && fn.Type.Params.NumFields() == 1 {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("inventory test does not resolve: %s", key)
		}
		body := func(file *ast.File) string {
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil && fn.Name.Name == e.Test {
					var out bytes.Buffer
					if err := format.Node(&out, token.NewFileSet(), fn.Body); err == nil {
						return out.String()
					}
				}
			}
			return ""
		}
		if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(e.BodySHA256) || fmt.Sprintf("%x", sha256.Sum256([]byte(body(f)))) != e.BodySHA256 {
			return fmt.Errorf("retained baseline test changed: %s", key)
		}
	}
	return nil
}

// A new historical read panics through the nil port rather than inventing facts.
type floorRepo struct {
	evaluation.Repo
	Tree string
}

func (r floorRepo) InRepository() bool { return false }
func (r floorRepo) VerifyExpectAt(rev, path, row string) (string, bool) {
	if rev == r.Tree && path == "docs/streams/build/brief-07.md" && row == "1" {
		return "exit 0", true
	}
	return "", false
}
func floorProject(c evaluation.Coverage, e evaluation.Eligibility) floorResult {
	out := floorResult{Released: c.Released, Eligibility: string(e.Verdict), Holds: []string{}, Notices: []string{}}
	if len(c.Claims) != 1 {
		out.Claim = fmt.Sprintf("claim-count:%d", len(c.Claims))
	} else {
		out.Claim = c.Claims[0].Result
	}
	for _, h := range e.Holds {
		out.Holds = append(out.Holds, string(h.State))
		if h.Why == "" {
			out.Holds = append(out.Holds, "missing-explanation")
		}
	}
	for _, n := range e.Notices {
		out.Notices = append(out.Notices, string(n.State))
		if n.Why == "" {
			out.Notices = append(out.Notices, "missing-explanation")
		}
	}
	return out
}
func floorAdapters(t *testing.T, c floorCase) (floorResult, floorResult) {
	t.Helper()
	var r domain.Record
	b, err := os.ReadFile(filepath.Join(floorDir, "record.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &r); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "streams", "build")
	if err = os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	writeCoverageBrief(t, dir, "07", "build", c.Verify, c.Evidence)
	path := filepath.Join(dir, "brief-07.md")
	bf, ok, err := parseBriefFile(path)
	if err != nil || !ok {
		t.Fatalf("parse frozen brief: %v %v", ok, err)
	}
	obs := observeBrief(root, path, "build/07", bf)
	brief := &Brief{Num: "07", Status: "implemented"}
	edge := GraphEdge{Ref: "unpub:example-a/01", Type: "ordering-gate", Reason: "required frozen target"}
	if c.Edge == "gate" {
		brief.Gates = []GraphEdge{edge}
	} else if c.Edge == "feather" {
		brief.Feathers = []GraphEdge{edge}
	}
	streams := []*Stream{{Name: "build", Briefs: []Brief{*brief}}}
	reg := &graphRepos{Aliases: map[string]graphRepoEntry{"unpub": {Unpublished: true}}}
	input := evaluation.CoverageInput{Revision: r.Candidate.Tree, Briefs: []evaluation.BriefObservation{obs}}
	repo := floorRepo{Tree: r.Candidate.Tree}
	// Both existing observation adapters and the domain reuse the one evaluator owner.
	old := floorProject(evaluation.EvaluateCoverage(input, repo)["build/07"], evaluateEligibility(streams, reg, root)["build/07"])
	a, err := domain.Assess(r, domain.Observations{Work: r.Work, Candidate: r.Candidate, Generation: r.Generation.Recorded, AcceptanceVersion: r.Acceptance.Version, AcceptanceDigest: r.Acceptance.Digest, Eligibility: evaluation.EligibilityInput{Streams: evaluationStreams(streams), Registry: evaluationRegistry(reg)}, Coverage: input, Repo: repo})
	if err != nil {
		t.Fatal(err)
	}
	if a.Coverage == nil || a.Eligibility == nil {
		t.Fatalf("domain omitted complete observation: %+v", a)
	}
	newer := floorProject(*a.Coverage, *a.Eligibility)
	if (a.State == "supported") != (newer.Released && newer.Eligibility != "held") {
		t.Fatalf("domain state disagrees: %+v", a)
	}
	return old, newer
}
func floorCompare(id string, old, newer, expected floorResult, diffs []floorDifference) error {
	var match *floorDifference
	for i := range diffs {
		d := &diffs[i]
		if d.Case == id {
			if match != nil {
				return fmt.Errorf("duplicate difference %s", id)
			}
			match = d
		}
	}
	equal := reflect.DeepEqual(old, newer)
	if equal && match != nil {
		return fmt.Errorf("stale difference %s", id)
	}
	if !equal {
		if match == nil {
			return fmt.Errorf("unlisted difference %s", id)
		}
		if match.Reason == "" || !reflect.DeepEqual(old, match.Old) || !reflect.DeepEqual(newer, match.New) {
			return fmt.Errorf("difference outcomes/reason mismatch %s", id)
		}
		if !reflect.DeepEqual(old, expected) {
			return fmt.Errorf("old expectation mismatch %s: %+v", id, old)
		}
		return nil
	}
	if !reflect.DeepEqual(old, expected) {
		return fmt.Errorf("expectation mismatch %s: got %+v want %+v", id, old, expected)
	}
	return nil
}

// regression: graph-execution/29
func TestRegressionFloorInventoryResolves(t *testing.T) {
	inv, _, _, err := floorLoad(floorDir)
	if err != nil {
		t.Fatal(err)
	}
	if err = floorResolve(inv); err != nil {
		t.Fatal(err)
	}
	bad := inv
	bad.Entries = append([]floorEntry(nil), inv.Entries...)
	for _, digest := range []string{"", "not-a-digest", strings.Repeat("0", 64)} {
		bad.Entries[0].BodySHA256 = digest
		if floorResolve(bad) == nil {
			t.Fatal("missing, invalid or mismatched baseline digest accepted")
		}
	}
	inv.Entries[0].Test += "Misspelled"
	if floorResolve(inv) == nil {
		t.Fatal("misspelled inventory accepted")
	}
}

// regression: graph-execution/29
func TestRegressionFloorOldNewAgree(t *testing.T) {
	inv, cases, diffs, err := floorLoad(floorDir)
	if err != nil {
		t.Fatal(err)
	}
	if err = floorResolve(inv); err != nil {
		t.Fatal(err)
	}
	head, err := coverageGit("..", "rev-parse", "HEAD").Output()
	if err != nil {
		head = []byte("unavailable (source export)")
	}
	t.Logf("checkout-head=%s (adapters execute this working source, baseline denotes inventory provenance)", strings.TrimSpace(string(head)))
	known := map[string]bool{}
	for _, c := range cases {
		known[c.ID] = true
	}
	for _, d := range diffs {
		if !known[d.Case] {
			t.Fatalf("difference names unselected case %s", d.Case)
		}
	}
	for _, c := range cases {
		t.Run(c.ID, func(t *testing.T) {
			old, newer := floorAdapters(t, c)
			for name, result := range map[string]floorResult{"package-main-observation-adapter": old, "domain.Assess": newer} {
				t.Logf("implementation=%s evaluator=%+v case=%s baseline=%s result=%+v", name, evaluation.CurrentIdentity(), c.ID, inv.Entries[0].Source, result)
			}
			if err := floorCompare(c.ID, old, newer, c.Expected, diffs); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// regression: graph-execution/29
func TestRegressionFloorUnlistedDifferenceFails(t *testing.T) {
	_, cases, _, err := floorLoad(floorDir)
	if err != nil {
		t.Fatal(err)
	}
	a := cases[0].Expected
	b := a
	b.Released = !b.Released
	d := floorDifference{Case: "control", Reason: "synthetic approved control", Old: a, New: b}
	if floorCompare("control", a, b, a, nil) == nil {
		t.Fatal("unlisted difference accepted")
	}
	if err := floorCompare("control", a, b, a, []floorDifference{d}); err != nil {
		t.Fatal(err)
	}
	if floorCompare("control", a, a, a, []floorDifference{d}) == nil {
		t.Fatal("stale difference accepted")
	}
	d.New = a
	if floorCompare("control", a, b, a, []floorDifference{d}) == nil {
		t.Fatal("different outcomes accepted")
	}
}

// regression: graph-execution/29
func TestRegressionFloorZeroSelectedFails(t *testing.T) {
	for _, mode := range []string{"empty-selection", "empty-directory"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			inv := floorInventory{}
			if mode == "empty-directory" {
				inv.Added = []string{"missing"}
			}
			b, err := yaml.Marshal(inv)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(dir, "inventory.yaml"), b, 0600); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(dir, "differences.yaml"), []byte("[]\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if _, _, _, err = floorLoad(dir); err == nil {
				t.Fatal("zero selected cases accepted")
			}
		})
	}
}

// regression: graph-execution/29
func TestFloorReused(t *testing.T) {
	t.Run("offline", TestOfflineReaderNeverAnswersEmpty)
	t.Run("graph", TestGraphJSONLDeterministic)
}
