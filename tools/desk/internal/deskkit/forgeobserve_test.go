package deskkit

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/medici-finance/assay/perfrec"
)

type observeRT func(*http.Request) (*http.Response, error)

func (f observeRT) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func writeForgeRecord(t *testing.T, s *perfrec.Session) perfrec.V3Summary {
	t.Helper()
	state, err := StateDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := (&perfrec.Writer{}).Write(state, s.Finish("deskboard", "query", "", 0, "")); err != nil {
		t.Fatal(err)
	}
	h, err := LoadPerfHistory(1)
	if err != nil || h.Skipped != 0 {
		t.Fatalf("history: %v skipped %d", err, h.Skipped)
	}
	sum, err := perfrec.ReduceV3(h.Records, perfrec.ProfileTool, "tool", perfrec.V3Filter{})
	if err != nil {
		t.Fatal(err)
	}
	return sum
}

// regression: class perf-v3-forge-leaf
func TestPerfV3ForgeAttemptTimingFlow(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ASSAY_TELEMETRY", "1")
	s := flowSession(perfrec.ObservationProvider)
	observer := RequestObserver{Sink: s, Origin: perfrec.OriginDirect}
	sends := 0
	leaf := observeRT(func(r *http.Request) (*http.Response, error) {
		sends++
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"ok":true}`))}, nil
	})
	g := &GitHubForge{Token: "fixture", BaseURL: "https://api.example.test", Observer: observer, Client: &http.Client{Transport: countRoundTrips(countRoundTrips(leaf))}}
	var out struct {
		OK bool `json:"ok"`
	}
	if err := g.doJSON("GET", "/repos/example/project", nil, &out); err != nil || !out.OK {
		t.Fatal(err, out)
	}
	// A retry is another actual send, even when the business operation is the same.
	if err := g.doJSON("GET", "/repos/example/project", nil, &out); err != nil {
		t.Fatal(err)
	}
	refused := observedTransport(countedTransport{next: observeRT(func(*http.Request) (*http.Response, error) { return nil, errors.New("admission refused") })}, observer, perfrec.RequestProvider, false)
	req, _ := http.NewRequest("GET", "https://api.example.test/private/path?secret=value", nil)
	if _, err := refused.RoundTrip(req); err == nil {
		t.Fatal("refusal vanished")
	}
	s.ObserveRequests()
	sum := writeForgeRecord(t, s)
	var attempts int64
	for _, r := range sum.Requests {
		attempts += r.Attempts
		if r.DurationCoverage[perfrec.CoverageMeasured] != r.Attempts || r.BytesSamples != r.Attempts {
			t.Fatal("lifetime coverage", r)
		}
	}
	if sends != 2 || attempts != 2 {
		t.Fatalf("actual=%d reported=%d", sends, attempts)
	}
	if len(sum.Requests) != 1 {
		t.Fatal("unbounded URL labels", sum.Requests)
	}
}

type observeSink struct{ items []perfrec.RequestObservation }

func (s *observeSink) ObserveRequest(o perfrec.RequestObservation) bool {
	s.items = append(s.items, o)
	return true
}

type slowBody struct {
	closed int
	reads  int
}

func (s *slowBody) Read(p []byte) (int, error) {
	time.Sleep(3 * time.Millisecond)
	s.reads++
	copy(p, "x")
	return 1, io.EOF
}
func (s *slowBody) Close() error { s.closed++; return nil }

// regression: class perf-v3-forge-body-parity
func TestPerfV3ForgeObservationParity(t *testing.T) {
	sink := &observeSink{}
	body := &slowBody{}
	base := observeRT(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 304, Body: body, Header: http.Header{}}, nil
	})
	req, _ := http.NewRequest("GET", "https://api.example.test/resource", nil)
	rt := observedTransport(countRoundTrips(base), RequestObserver{sink, perfrec.OriginPoll}, perfrec.RequestProvider, false)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(sink.items) != 0 || body.reads != 0 {
		t.Fatal("eager read or early timing")
	}
	p := make([]byte, 1)
	n, readErr := resp.Body.Read(p)
	if n != 1 || readErr != io.EOF || string(p) != "x" {
		t.Fatal("read contract changed")
	}
	_ = resp.Body.Close()
	_ = resp.Body.Close()
	if len(sink.items) != 1 || body.closed != 2 {
		t.Fatal("duplicate observation or changed Close")
	}
	o := sink.items[0]
	if o.Outcome != "304" || o.DurationMs == nil || *o.DurationMs < 3 || *o.Bytes != 1 {
		t.Fatal("body excluded", o)
	}
	// Authentication work must stay separate from provider work.
	a := ObserveAuthClient(&http.Client{Transport: base}, RequestObserver{sink, perfrec.OriginDirect})
	resp, err = a.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if len(sink.items) != 2 || sink.items[1].Class != perfrec.RequestAuthentication {
		t.Fatal("auth conflated")
	}
}
