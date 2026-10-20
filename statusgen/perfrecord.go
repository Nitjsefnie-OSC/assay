package main

// perfrecord.go — statusgen's side of the opt-in, LOCAL-ONLY perf record (#3101), written in
// perf-v2: the same session format the desk tools write (perfrec.Begin/Phase/Finish).
//
// With ASSAY_TELEMETRY exactly "1", every statusgen process that exits through exitProcess
// appends one perf-v2 line to the same <state dir>/perf/ day files the desk tools write, so
// `deskperf report --tool statusgen` reads them:
//
//   - wall and CPU from perfrec's process-init mark to the exit, one boundary;
//   - per-stage phase timings (perfPhases, written in roster order), so most of the wall time
//     has a breakdown: board parse, source lint, link lint, the forge claim read, the view
//     build, render, the STATUS.md write or compare, and --corroborate's forge corroboration.
//     The phases do not sum to the wall time: process start before the first stage and the
//     exit after the last one fall inside the wall time but outside every phase;
//   - the `http` counter: every direct HTTP attempt at statusgen's forge seams (the GitHub
//     REST and GraphQL clients, the GitLab client, and the telemetry sender), counted at the
//     call, so it is MEASURED even when it is zero. The other counters are listed
//     unavailable: statusgen's many `gh`/`git` child processes have no shared seam to count.
//
// Unset or any other value, nothing is created and nothing changes: the session is nil and
// every phase call is a no-op. The writer — schema, day-file layout, 7-day retention, install
// id, CPU reading — is the shared perfrec module; this file owns only what is statusgen's: the
// verb and phase rosters, the forge counter and the exit wrapper. It is unrelated to
// telemetry.go's network payload, which keeps its own double opt-in.
//
// NO FREE TEXT: the verb and every phase name are members of the rosters below, never any part
// of argv. NEVER LOAD-BEARING: measuring adds no forge call, no output and no exit path; a
// record that cannot be written is dropped silently; the exit code and output are exactly what
// they were without the switch.

import (
	"errors"
	"flag"
	"os"
	"strings"
	"sync/atomic"

	"github.com/medici-finance/assay/perfrec"
)

// perfToolName is the record's tool field. It is a constant, not derived from argv[0].
const perfToolName = "statusgen"

// perfDefaultVerb is the mode of a run that names none: the plain board regeneration.
const perfDefaultVerb = "regen"

// perfSubcommands are the positional sub-commands main() dispatches before flag parsing
// ("version" also answers to --version). perfModeFlags are the flags that select a mode of
// the default command. Both are compiled-in rosters; TestPerfModeRosterIsRegistered pins the
// flags against the flag set main() really registers.
var (
	perfSubcommands = []string{
		"version", "init", "verifyrun", "verifyclosure", "verifyflip", "mergecheck",
		"shardcheck", "phantoms", "enforcement-status", "backfill", "reconcile", "outcomes",
		"regen", "conform", "patterns", "instance", "runtime-resource", "migrate", "brief",
		"newbrief", "lint", "verify-gate-close",
	}
	perfModeFlags = []string{
		"check", "lint", "lint-audit", "instrument-audit", "record", "verify-issues",
		"decision-issues", "owed-issues", "drive-issues", "watchdog", "drive-snapshot",
		"signoff-digest", "scan-issues", "transcribe-scan", "transcribe-verdict",
		"transcribe-scan-delta", "close-verify", "auto-flip-model", "register-links", "alarms",
		"verif-backlog", "autonomy", "ladder", "issues", "self-improvement", "cynefin", "dora",
		"dora-timing", "dora-json", "trend", "roadmap", "consumers", "corroborate",
		"bottleneck", "intake-exits", "intake-debt", "throughput", "leadtime",
		"flow-efficiency", "flow", "first-pass-yield", "review-rework", "decision-latency",
		"net-flow", "assayscore", "launch", "requirements-rollup", "export-evidence",
		"export-audit-pack", "graph", "gate-scores", "next-up", "eligibility", "coverage",
		"cluster-pending-queue", "gate-telemetry", "telemetry", "telemetry-dry-run",
	}
)

// perfVerb maps argv (without the program name) onto a roster member. A positional
// sub-command wins; else the first argument that is a mode flag; else the default regen.
// It returns the ROSTER's spelling, never the input.
func perfVerb(args []string) string {
	if len(args) == 0 {
		return perfDefaultVerb
	}
	if args[0] == "--version" {
		return "version"
	}
	if v := perfrec.Verb(perfSubcommands, args); v != perfrec.Other {
		return v
	}
	for _, a := range args {
		if !strings.HasPrefix(a, "-") {
			continue
		}
		name := strings.TrimLeft(a, "-")
		if i := strings.IndexByte(name, '='); i >= 0 {
			name = name[:i]
		}
		if v := perfrec.Verb(perfModeFlags, []string{name}); v != perfrec.Other {
			return v
		}
	}
	if !strings.HasPrefix(args[0], "-") {
		return perfrec.Other // a positional token that is no sub-command: unknown, not regen
	}
	return perfDefaultVerb
}

// perfPhases is statusgen's fixed phase roster. perfrec writes phases in roster order, which
// is the order a run first enters them except for one pair: the view build is entered before
// the claim read and resumed after it, yet "claims" is listed before "view". A record's phase
// order is therefore the roster's, never a trace of the run. A stage entered more than once
// (one run() per root, the source lint resumed around its named sub-stages, the view build
// either side of the claim read) accumulates into one entry whose n counts the entries. The
// sub-stages split out of "lint" are the checks that each took more than ~5% of a full
// --lint/--check of this repository; every other check stays inside "lint".
var perfPhases = []string{
	"parse",       // budget specs, stream and finding load, placeholders, archived streams
	"lint",        // the per-stream/per-brief source checks not named below
	"briefs",      // checkBriefFiles: every brief file's frontmatter and sections
	"ordering",    // orderingGateNotices: README/brief prose that reads as an ordering gate
	"evidence",    // evidenceActorGate: Evidence-row actors against git history
	"links",       // the doc-tree block: fixture corpora, doc walk and the lints that read it
	"claims",      // the claim read (a forge round-trip with --forge or on the write path)
	"view",        // findings, next-up, history ages, drives: the board's model
	"render",      // emit: the STATUS.md bytes
	"write",       // record / check-compare / write of STATUS.md and the register views
	"corroborate", // --corroborate: forge corroboration of human stamps
}

// perfSession measures this process: started at package initialisation from perfrec's
// process-init mark when the switch is on, nil (every call a no-op) when it is off.
var perfSession = perfBegin()

func perfBegin() *perfrec.Session {
	if !perfrec.Enabled() {
		return nil
	}
	return perfrec.Begin(perfrec.SystemProbe(), perfrec.ProcessEntry(), perfPhases)
}

// perfPhase starts a timed stage and returns its stop (idempotent; a no-op when off).
func perfPhase(name string) func() { return perfSession.Phase(name) }

// perfStage times a sequence of non-overlapping stages: enter closes the open stage and opens
// the next, done closes whichever is open. A function with many returns defers done once.
type perfStage struct{ stop func() }

func (p *perfStage) enter(name string) {
	p.done()
	p.stop = perfPhase(name)
}

func (p *perfStage) done() {
	if p.stop != nil {
		p.stop()
		p.stop = nil
	}
}

// perfHTTPCalls counts direct HTTP attempts at the forge seams, process-wide. It is
// incremented unconditionally (an atomic add is cheaper than the branch that would skip it)
// and read only when a record is built.
var perfHTTPCalls atomic.Int64

// perfCountHTTP records one direct HTTP attempt. Each forge seam calls it once, immediately
// before its client's Do, so a retry, a 304 and a transport failure each count; a redirect
// the client follows inside that Do is part of the same attempt and does not count again.
func perfCountHTTP() { perfHTTPCalls.Add(1) }

// perfSourceStamp is the ONE place the record's build stamp is read: the commit the binary was
// built from (perfrec.Revision: the toolchain's vcs.revision). statusgen has no link-time commit
// stamp, hence the empty one. It is never statusgenVersion: that is a release TAG, which
// perfrec.SourceSHA's hex-only gate records as "unpinned".
func perfSourceStamp() string { return perfrec.Revision("") }

// perfBuildRecord ends the session and builds this process's record; ok is false when the
// switch is off.
func perfBuildRecord(exit int) (perfrec.Record, bool) {
	if perfSession == nil {
		return perfrec.Record{}, false
	}
	perfSession.Add(perfrec.CounterHTTP, perfHTTPCalls.Load())
	return perfSession.Finish(perfToolName, perfVerb(os.Args[1:]), perfrec.RoleNone, exit, perfSourceStamp()), true
}

// recordPerf appends this process's perf-v2 line when ASSAY_TELEMETRY=1. It returns nothing
// and never prints: the record is not allowed to change what statusgen does.
func recordPerf(exit int) {
	if !perfrec.Enabled() {
		return
	}
	if rec, ok := perfBuildRecord(exit); ok {
		perfrec.Append(rec)
	}
}

// exitProcess is the ONE way main() ends the process: it records the perf line (a no-op when
// the switch is off) and then exits with code. main.go calls no os.Exit directly, so no exit
// path escapes the record (TestMainExitsOnlyThroughRecorder).
func exitProcess(code int) {
	recordPerf(code)
	os.Exit(code)
}

// parseMainFlags is flag.Parse() with the exit routed through exitProcess: the default
// command line's ExitOnError would call os.Exit inside the flag package on a usage error,
// an exit path no record could see. Behaviour is the flag package's: -h/--help exits 0, any
// other parse error exits 2, and the usage text is printed by the flag set as before.
func parseMainFlags() {
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			exitProcess(0)
		}
		exitProcess(2)
	}
}
