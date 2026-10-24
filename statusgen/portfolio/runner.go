package portfolio

import (
	"encoding/json"
	"fmt"
)

// RunnerVersion is the loop-admin runner contract this translation reads. The runner
// contract is owned by loopadmin (schemas/loop-admin-runner-v1.json); this file reads only
// the members it needs and never restates or widens that contract.
const RunnerVersion = "loop-admin-runner/v1"

// Attribution is what a runner report does not carry and an attribution source must supply:
// the stable segment identity and revision, the profile revision, the measuring source, the
// subject and, when one was measured, the active interval. Elapsed is a separately labeled
// proxy; it is never read as active time.
type Attribution struct {
	Segment         SegmentRef
	ProfileRevision string
	Source          string
	Subject         *Subject
	Active          *Interval
	Elapsed         *int64
}

// AttributionSource resolves one invocation's attribution. It reports false when it holds
// none for that invocation.
type AttributionSource func(Invocation) (Attribution, bool)

type runnerRequest struct {
	Version string `json:"version"`
	Caller  string `json:"caller"`
	ID      string `json:"id"`
	Profile struct {
		ID    string `json:"id"`
		Model string `json:"model"`
	} `json:"profile"`
	Authority struct {
		Generation int64 `json:"generation"`
	} `json:"authority"`
}

type runnerObservation struct {
	State       string `json:"state"`
	ActualModel string `json:"actual_model"`
	Usage       struct {
		CostMicros *int64 `json:"cost_micros"`
	} `json:"usage"`
	Result *struct {
		Caller     string `json:"caller"`
		ID         string `json:"id"`
		Generation int64  `json:"generation"`
	} `json:"result"`
}

// FromRunner translates one runner request and its terminal observation, both as the bytes
// the runner contract defines, into a ledger segment. The attribution source is required:
// the runner contract carries no active interval, so without one the segment's active time
// stays unknown rather than being inferred from anything else.
//
// The pinned profile model is kept as the profile's model and never stands in for the
// actual model: an observation that reports none leaves the actual model unknown. A negative
// cost is unknown, as the runner contract defines.
func FromRunner(request, observation []byte, attribute AttributionSource) (Segment, error) {
	if attribute == nil {
		return Segment{}, fmt.Errorf("portfolio: runner translation needs an attribution source")
	}
	var req runnerRequest
	if err := json.Unmarshal(request, &req); err != nil {
		return Segment{}, fmt.Errorf("portfolio: runner request: %w", err)
	}
	if req.Version != RunnerVersion {
		return Segment{}, fmt.Errorf("portfolio: runner request version %q (supported: %s)", req.Version, RunnerVersion)
	}
	var obs runnerObservation
	if err := json.Unmarshal(observation, &obs); err != nil {
		return Segment{}, fmt.Errorf("portfolio: runner observation: %w", err)
	}
	if !member(obs.State, []string{"finished", "failed", "stopped"}) || obs.Result == nil {
		return Segment{}, fmt.Errorf("portfolio: runner observation is not a terminal report with a result")
	}
	if obs.Result.Caller != req.Caller || obs.Result.ID != req.ID || obs.Result.Generation != req.Authority.Generation {
		return Segment{}, fmt.Errorf("portfolio: runner result echoes another invocation or generation")
	}
	inv := Invocation{Caller: req.Caller, ID: req.ID}
	a, ok := attribute(inv)
	if !ok {
		return Segment{}, fmt.Errorf("portfolio: no attribution for invocation %s/%s", inv.Caller, inv.ID)
	}
	cost := obs.Usage.CostMicros
	if cost != nil && *cost < 0 {
		cost = nil
	}
	g := Segment{
		Invocation:  inv,
		ID:          a.Segment.ID,
		Revision:    a.Segment.Revision,
		Profile:     Profile{ID: req.Profile.ID, Revision: a.ProfileRevision, Model: req.Profile.Model},
		ActualModel: obs.ActualModel,
		Source:      a.Source,
		Subject:     a.Subject,
		Active:      a.Active,
		Elapsed:     a.Elapsed,
		Cost:        cost,
	}
	return g, nil
}
