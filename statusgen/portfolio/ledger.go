package portfolio

import (
	"encoding/json"
	"sort"
)

// Quantity is one measure's known subtotal and its coverage. A record whose value is unknown
// adds to UnknownRecords and nothing to Known: the subtotal is never presented as a total.
type Quantity struct {
	Known          int64 `json:"known"`
	KnownRecords   int   `json:"known-records"`
	UnknownRecords int   `json:"unknown-records"`
}

// Complete reports whether every record behind the quantity had a known value.
func (q Quantity) Complete() bool { return q.UnknownRecords == 0 }

// Tally holds agent and human measures side by side and never mixes them: an agent
// duration is not a human count, and a human count is not a duration.
type Tally struct {
	Agent map[string]Quantity `json:"agent"`
	Human map[string]Quantity `json:"human"`
}

func newTally() Tally {
	t := Tally{Agent: map[string]Quantity{}, Human: map[string]Quantity{}}
	for _, m := range append(append([]string(nil), agentMeasures...), ElapsedProxy) {
		t.Agent[m] = Quantity{}
	}
	for _, m := range humanMeasures {
		t.Human[m] = Quantity{}
	}
	return t
}

// Entry names one report the ledger did not count as known, and why.
type Entry struct {
	Kind   string `json:"kind"`
	Key    string `json:"key"`
	Reason string `json:"reason"`
}

// Ledger is the read-only measured allocation for one policy revision over its window.
// Total is the sum of every bucket and Unattributed; unattributed work stays visible.
type Ledger struct {
	Policy         string           `json:"policy"`
	PolicyRevision string           `json:"policy-revision"`
	Scope          string           `json:"scope"`
	Window         Interval         `json:"window"`
	Weights        string           `json:"weights"`
	WeightsRev     string           `json:"weights-revision"`
	Buckets        map[string]Tally `json:"buckets"`
	Unattributed   Tally            `json:"unattributed"`
	Total          Tally            `json:"total"`
	Duplicates     []Entry          `json:"duplicates"`
	Conflicts      []Entry          `json:"conflicts"`
	Excluded       []Entry          `json:"excluded"`
}

// Measure builds the ledger for a snapshot. It validates first and reads nothing else.
// semantic: S-portfolio-measurement
func Measure(s Snapshot) (Ledger, error) {
	if err := Validate(s); err != nil {
		return Ledger{}, err
	}
	p := s.Policy
	l := Ledger{
		Policy: p.ID, PolicyRevision: p.Revision, Scope: p.Scope, Window: p.Window,
		Weights: p.Weights.ID, WeightsRev: p.Weights.Revision,
		Buckets: map[string]Tally{}, Unattributed: newTally(), Total: newTally(),
		Duplicates: []Entry{}, Conflicts: []Entry{}, Excluded: []Entry{},
	}
	owner := map[Subject]string{}
	for _, b := range p.Buckets {
		l.Buckets[b.ID] = newTally()
		for _, m := range b.Members {
			owner[m] = b.ID
		}
	}
	weights := map[string]int64{}
	for _, m := range p.Weights.Models {
		weights[m.Model] = m.Milli
	}
	bucketOf := func(s *Subject) string {
		if s == nil {
			return ""
		}
		return owner[*s]
	}
	tally := func(bucket string) Tally {
		if bucket == "" {
			return l.Unattributed
		}
		return l.Buckets[bucket]
	}
	record := func(bucket string, agent bool, measure string, v *int64) {
		for _, t := range []Tally{tally(bucket), l.Total} {
			m := t.Human
			if agent {
				m = t.Agent
			}
			q := m[measure]
			if v == nil {
				q.UnknownRecords++
			} else {
				q.Known += *v
				q.KnownRecords++
			}
			m[measure] = q
		}
	}
	measureSegments(s, &l, weights, bucketOf, record)
	measureDecisions(s, &l, bucketOf, record)
	return l, nil
}

type recordFunc func(bucket string, agent bool, measure string, v *int64)

func measureSegments(s Snapshot, l *Ledger, weights map[string]int64, bucketOf func(*Subject) string, record recordFunc) {
	reports := map[Invocation]map[string][]Segment{}
	seen := map[string]bool{}
	for _, g := range s.Segments {
		enc := encode(g)
		if seen[enc] {
			l.Duplicates = append(l.Duplicates, Entry{"segment", segKey(g), "duplicate report counted once"})
			continue // duplicate report
		}
		seen[enc] = true
		if reports[g.Invocation] == nil {
			reports[g.Invocation] = map[string][]Segment{}
		}
		reports[g.Invocation][g.ID] = append(reports[g.Invocation][g.ID], g)
	}
	reconciled := map[Invocation]Reconciliation{}
	for _, r := range s.Reconciliations {
		reconciled[r.Invocation] = r
	}
	invs := make([]Invocation, 0, len(reports))
	for inv := range reports {
		invs = append(invs, inv)
	}
	sort.Slice(invs, func(i, j int) bool {
		if invs[i].Caller != invs[j].Caller {
			return invs[i].Caller < invs[j].Caller
		}
		return invs[i].ID < invs[j].ID
	})
	for _, inv := range invs {
		groups := reports[inv]
		selected, reason := selectSegments(groups, reconciled, inv, l)
		if reason == "" {
			reason = overlap(selected)
		}
		if reason != "" {
			// A conflict is not resolved by counting one side or both: the invocation stays
			// visible as unknown in every agent measure until it is reconciled.
			l.Conflicts = append(l.Conflicts, Entry{"invocation", inv.Caller + "/" + inv.ID, reason})
			bucket := agreedBucket(groups, bucketOf)
			for _, m := range append(append([]string(nil), agentMeasures...), ElapsedProxy) {
				record(bucket, true, m, nil)
			}
			continue
		}
		for _, g := range selected {
			bucket := bucketOf(g.Subject)
			active := activeSeconds(g, l)
			if active == outside {
				l.Excluded = append(l.Excluded, Entry{"segment", segKey(g), "active interval outside the policy window"})
				continue
			}
			var secs, weighted *int64
			if active >= 0 {
				secs = &active
				if milli, ok := weights[g.ActualModel]; ok && g.ActualModel != "" {
					w := active * milli
					weighted = &w
				}
			}
			record(bucket, true, ActiveSeconds, secs)
			record(bucket, true, WeightedActive, weighted)
			record(bucket, true, CostMicros, g.Cost)
			record(bucket, true, ElapsedProxy, g.Elapsed)
		}
	}
}

// selectSegments returns the one report per segment that counts, or a conflict reason.
// Without a reconciliation every segment must have exactly one distinct report; with one,
// exactly the accepted (id, revision) reports count and the rest are listed as excluded.
func selectSegments(groups map[string][]Segment, reconciled map[Invocation]Reconciliation, inv Invocation, l *Ledger) ([]Segment, string) {
	ids := make([]string, 0, len(groups))
	for id := range groups {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	r, ok := reconciled[inv]
	if !ok {
		var out []Segment
		for _, id := range ids {
			if len(groups[id]) != 1 {
				return nil, "conflicting reports of segment " + id + " need explicit reconciliation"
			}
			out = append(out, groups[id][0])
		}
		return out, ""
	}
	accept := map[string]string{}
	for _, ref := range r.Accept {
		accept[ref.ID] = ref.Revision
	}
	var out []Segment
	for _, id := range ids {
		var match []Segment
		for _, g := range groups[id] {
			if rev, ok := accept[id]; ok && g.Revision == rev {
				match = append(match, g)
				continue
			}
			l.Excluded = append(l.Excluded, Entry{"segment", segKey(g), "not accepted by reconciliation"})
		}
		if len(match) > 1 {
			return nil, "reconciliation accepts more than one distinct report of segment " + id
		}
		out = append(out, match...)
	}
	for _, ref := range r.Accept {
		found := false
		for _, g := range out {
			found = found || (g.ID == ref.ID && g.Revision == ref.Revision)
		}
		if !found {
			return nil, "reconciliation accepts an unreported segment " + ref.ID + "@" + ref.Revision
		}
	}
	return out, ""
}

// overlap refuses two counted segments of one invocation whose active intervals overlap.
func overlap(segments []Segment) string {
	for i := range segments {
		for j := i + 1; j < len(segments); j++ {
			a, b := segments[i].Active, segments[j].Active
			if a == nil || b == nil {
				continue
			}
			as, ae, _ := bounds(*a)
			bs, be, _ := bounds(*b)
			if as.Before(be) && bs.Before(ae) {
				return "segments " + segments[i].ID + " and " + segments[j].ID + " overlap; overlapping intervals need explicit reconciliation"
			}
		}
	}
	return ""
}

const (
	unknownActive = int64(-1)
	outside       = int64(-2)
)

// activeSeconds is the measured active time inside the policy window: unknown when no
// interval was measured or the interval straddles the window, outside when disjoint from it.
func activeSeconds(g Segment, l *Ledger) int64 {
	if g.Active == nil {
		return unknownActive
	}
	ws, we, _ := bounds(l.Window)
	s, e, _ := bounds(*g.Active)
	switch {
	case !s.Before(we) || !e.After(ws):
		return outside
	case s.Before(ws) || e.After(we):
		return unknownActive
	}
	return int64(e.Sub(s).Seconds())
}

func measureDecisions(s Snapshot, l *Ledger, bucketOf func(*Subject) string, record recordFunc) {
	reports := map[string][]Decision{}
	var order []string
	seen := map[string]bool{}
	for _, d := range s.Decisions {
		enc := encode(d)
		if seen[enc] {
			l.Duplicates = append(l.Duplicates, Entry{"decision", d.ID, "duplicate report counted once"})
			continue
		}
		seen[enc] = true
		if reports[d.ID] == nil {
			order = append(order, d.ID)
		}
		reports[d.ID] = append(reports[d.ID], d)
	}
	sort.Strings(order)
	for _, id := range order {
		ds := reports[id]
		if len(ds) > 1 {
			l.Conflicts = append(l.Conflicts, Entry{"decision", id, "conflicting reports of one decision identity"})
			bucket := bucketOf(ds[0].Primary)
			units := map[string]bool{}
			for _, d := range ds[1:] {
				if bucketOf(d.Primary) != bucket {
					bucket = ""
				}
			}
			for _, d := range ds {
				if !units[d.Unit] {
					units[d.Unit] = true
					record(bucket, false, humanMeasure(d.Unit), nil)
				}
			}
			continue
		}
		d := ds[0]
		ws, we, _ := bounds(l.Window)
		start, end, _ := bounds(d.Window)
		var v *int64
		switch {
		case !start.Before(we) || !end.After(ws):
			l.Excluded = append(l.Excluded, Entry{"decision", d.ID, "declared window outside the policy window"})
			continue
		case start.Before(ws) || end.After(we):
			// A declared window straddling the policy window cannot be attributed to it.
		case d.Unit == UnitDecision:
			one := int64(1)
			v = &one
		default:
			v = d.Quantity
		}
		record(bucketOf(d.Primary), false, humanMeasure(d.Unit), v)
	}
}

func humanMeasure(unit string) string {
	if unit == UnitMinute {
		return DecisionMinutes
	}
	return Decisions
}

func agreedBucket(groups map[string][]Segment, bucketOf func(*Subject) string) string {
	bucket, first := "", true
	for _, gs := range groups {
		for _, g := range gs {
			b := bucketOf(g.Subject)
			if first {
				bucket, first = b, false
			} else if b != bucket {
				return ""
			}
		}
	}
	return bucket
}

func segKey(g Segment) string {
	return g.Invocation.Caller + "/" + g.Invocation.ID + "/" + g.ID + "@" + g.Revision
}

func encode(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// Share is one bucket's observed share of a measure's scope total, as an exact fraction.
// It is defined only when every record behind the total is known and the total is not zero.
type Share struct {
	Measure     string `json:"measure"`
	Numerator   int64  `json:"numerator"`
	Denominator int64  `json:"denominator"`
	Defined     bool   `json:"defined"`
	Reason      string `json:"reason,omitempty"`
}

// ShareOf returns bucket's share of measure (an agent measure when agent is true).
func (l Ledger) ShareOf(bucket string, agent bool, measure string) Share {
	pick := func(t Tally) Quantity {
		if agent {
			return t.Agent[measure]
		}
		return t.Human[measure]
	}
	total := pick(l.Total)
	sh := Share{Measure: measure, Numerator: pick(l.Buckets[bucket]).Known, Denominator: total.Known}
	switch {
	case !total.Complete():
		sh.Reason = "unknown coverage: the total has unknown records"
	case total.Known == 0:
		sh.Reason = "zero total: shares are undefined"
	default:
		sh.Defined = true
	}
	if !sh.Defined {
		sh.Numerator, sh.Denominator = 0, 0
	}
	return sh
}
