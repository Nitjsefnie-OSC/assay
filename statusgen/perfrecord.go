package main

// perfrecord.go — statusgen's side of the opt-in, LOCAL-ONLY perf record (#3101; perf-v3
// with a selected profile since statusgen/26, see the section at the end of this file).
//
// With ASSAY_TELEMETRY exactly "1", every statusgen process that exits through exitProcess
// appends one perf-v1 line (tool "statusgen", verb = the mode that ran, wall and CPU
// milliseconds, exit and result) to the same <state dir>/perf/ day files the desk tools write,
// so `deskperf report --tool statusgen` reads them. Unset or any other value, nothing is
// created and nothing changes.
//
// The writer — schema, day-file layout, 7-day retention, install id, CPU reading — is the
// shared perfrec module, the one implementation the desk tools use too; this file owns only
// what is statusgen's: the verb roster and the exit wrapper. It is unrelated to
// telemetry.go's network payload, which keeps its own double opt-in and is untouched.
//
// NO FREE TEXT: the verb is always a member of the rosters below, never any part of argv.
// NEVER LOAD-BEARING: a record that cannot be written is dropped silently; the exit code and
// output are exactly what they were without the switch.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

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

// perfStart is the clock's zero: package initialisation, which precedes main().
var perfStart = time.Now()

// recordPerf appends this process's perf line when ASSAY_TELEMETRY=1: the perf-v3 summary when
// a profile is selected (recordV3), else the unchanged perf-v1 line. It returns nothing and
// never prints: the record is not allowed to change what statusgen does.
func recordPerf(exit int) {
	if !perfrec.Enabled() {
		return
	}
	if recordV3(exit) {
		return
	}
	rec := perfrec.New(perfToolName, perfVerb(os.Args[1:]), perfrec.RoleNone,
		time.Since(perfStart), exit, statusgenVersion, time.Now())
	if ms, ok := perfrec.CPUMillis(); ok {
		rec.CPUMs = &ms
	}
	perfrec.Append(rec)
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
//
// The shared --perf-* context flags are bound here, on the same flag set, so they are parsed by
// the real parser (a malformed value is perf coverage, never a usage error). After a clean
// parse the invocation's context is selected once and the residual "evaluate" phase opens.
func parseMainFlags() {
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	perfBindFlags(flag.CommandLine)
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			exitProcess(0)
		}
		exitProcess(2)
	}
	perfSelect()
	perfEnter("evaluate") // closed by recordPerf at exit
}

// --- perf-v3 (statusgen/26) ------------------------------------------------------------------
//
// An explicitly SELECTED profile (--perf-profile / ASSAY_PERF_PROFILE = tool | workflow, with
// ASSAY_TELEMETRY=1) replaces the perf-v1 line with one perf-v3 process-summary record. No
// profile — unset, "off", or any invalid value — keeps the perf-v1 line above byte-for-byte:
// the legacy mode mapping, its key set and its boundary do not move.
//
// Two boundaries, stated because they differ: perf-v1's wallMs runs from perfStart, statusgen's
// own package initialisation (unchanged, historical). perf-v3's runs from perfrec.ProcessEntry(),
// the mark taken while the perfrec package initialises — it precedes every statusgen package
// initialiser, so a v3 wallMs is never shorter than the v1 wallMs of the same run and the two
// are not compared. Both end at exitProcess.

// perfPhases is statusgen's fixed perf-v3 phase roster. Phases are EXCLUSIVE: entering one
// pauses the phase it nests in, so phase ms never double count (phaseOverlap stays false for
// sequential code) and n counts timed segments — a paused phase resumes as a new segment.
//
//	discover  enumerating stream, archive and findings sources (load.go)
//	parse     reading + parsing a stream README, a findings file or a brief file
//	evaluate  everything else after the command line parsed: checks, eligibility, coverage
//	render    building the board text or a JSON view
//	write     writing STATUS.md and the register views
//	deskread  one deskread child, start to exit, as seen from statusgen (n = child starts)
var perfPhases = []string{"discover", "parse", "evaluate", "render", "write", "deskread"}

// perfNames are the shared --perf-* context flags (the desk tools' adapter registers the same
// eight). statusgen binds them on its own flag set: it may not import the desk adapter.
var perfNames = []string{"perf-profile", "perf-session-id", "perf-agent-id", "perf-parent-agent-id",
	"perf-parent-invocation-id", "perf-desk-role", "perf-role", "perf-work"}

// perfInput collects the parsed --perf-* values. Only the flag package writes it (Set below),
// so values reach the resolver exactly as the real parser accepted them; it is never re-read
// from argv.
var perfInput perfrec.ContextInput

type perfFlag struct{ name string }

func (perfFlag) String() string { return "" }
func (v perfFlag) Set(s string) error {
	val := perfrec.Value{Value: s, Present: true}
	switch v.name {
	case "perf-profile":
		perfInput.Profile = val
	case "perf-session-id":
		perfInput.SessionID = val
	case "perf-agent-id":
		perfInput.AgentID = val
	case "perf-parent-agent-id":
		perfInput.ParentAgentID = val
	case "perf-parent-invocation-id":
		perfInput.ParentInvocationID = val
	case "perf-desk-role":
		perfInput.DeskRole = val
	case "perf-role":
		perfInput.Role = val
	case "perf-work":
		perfInput.Work = append(perfInput.Work, s)
		perfInput.WorkPresent = true
	}
	return nil // optional metadata never fails the parse; perfrec classifies it as coverage
}

// perfBindFlags registers the eight context flags on fs. It never rejects a value.
func perfBindFlags(fs *flag.FlagSet) {
	for _, name := range perfNames {
		if fs.Lookup(name) == nil {
			fs.Var(perfFlag{name}, name, "optional local performance context (never operational authority)")
		}
	}
}

var (
	perfOnce    sync.Once
	perfCtx     perfrec.Context
	perfSession atomic.Pointer[perfrec.Session]
)

// perfSelect resolves this invocation's context ONCE (fresh invocation id per process) from
// the parsed flags, then the documented environment keys. A selected profile starts the v3
// session at the process-init mark. Sub-commands with their own parsers, help, version and
// usage errors reach it first from recordPerf, so they resolve from the environment alone
// (plus anything parsed before the error) and record the whole interval with no phases.
func perfSelect() {
	perfOnce.Do(func() {
		perfCtx = perfrec.ResolveContext(perfInput)
		if perfCtx.Profile() == "" {
			return
		}
		perfSession.Store(perfrec.BeginSelected(perfrec.SystemProbe(), perfrec.ProcessEntry(), perfPhases, perfCtx))
	})
}

type perfFrame struct {
	name string
	stop func() // nil while paused by a nested phase
}

var (
	perfMu    sync.Mutex
	perfStack []*perfFrame
)

// perfEnter opens phase name and returns its idempotent leave. With no v3 session it is a
// no-op. Re-entering the phase already on top is a no-op too (the outer call times it).
func perfEnter(name string) (leave func()) {
	s := perfSession.Load()
	if s == nil {
		return func() {}
	}
	perfMu.Lock()
	defer perfMu.Unlock()
	if n := len(perfStack); n > 0 {
		top := perfStack[n-1]
		if top.name == name {
			return func() {}
		}
		if top.stop != nil {
			top.stop()
			top.stop = nil
		}
	}
	f := &perfFrame{name: name, stop: s.Phase(name)}
	perfStack = append(perfStack, f)
	var once sync.Once
	return func() { once.Do(func() { perfLeave(s, f) }) }
}

func perfLeave(s *perfrec.Session, f *perfFrame) {
	perfMu.Lock()
	defer perfMu.Unlock()
	i := len(perfStack) - 1
	for ; i >= 0 && perfStack[i] != f; i-- {
	}
	if i < 0 {
		return // already closed by recordPerf
	}
	if f.stop != nil {
		f.stop()
	}
	if i != len(perfStack)-1 {
		// Left out of order (a concurrent caller): drop this frame only; the top keeps timing.
		perfStack = append(perfStack[:i], perfStack[i+1:]...)
		return
	}
	perfStack = perfStack[:i]
	if i > 0 && perfStack[i-1].stop == nil {
		perfStack[i-1].stop = s.Phase(perfStack[i-1].name)
	}
}

// perfCloseAll stops every open phase, innermost first, so Finish records them.
func perfCloseAll() {
	perfMu.Lock()
	defer perfMu.Unlock()
	for i := len(perfStack) - 1; i >= 0; i-- {
		if perfStack[i].stop != nil {
			perfStack[i].stop()
		}
	}
	perfStack = nil
}

// recordV3 writes the perf-v3 process summary for a selected profile. It reports whether it
// did, so recordPerf falls back to the unchanged perf-v1 line when no profile is selected.
// Request coverage stays "missing": statusgen observes no forge request itself, and the
// deskread child reports its own (a child that does not is a coverage gap, never a zero).
func recordV3(exit int) bool {
	perfSelect()
	s := perfSession.Load()
	if s == nil {
		return false
	}
	perfCloseAll()
	role := perfrec.RoleNone
	if r := perfCtx.Data().Role; r != "" {
		role = r
	}
	perfrec.Append(s.Finish(perfToolName, perfVerb(os.Args[1:]), role, exit, statusgenVersion))
	return true
}

// perfChildFlags are the context flags handed to a deskread child. The child's desk role is
// not among them: it resolves from the operational DESK_LOOP the child inherits unchanged.
var perfChildFlags = []string{"perf-profile", "perf-session-id", "perf-agent-id", "perf-parent-agent-id",
	"perf-parent-invocation-id", "perf-role", "perf-work"}

var perfProbed sync.Map // deskread binary -> whether its help lists every perfChildFlags entry

// deskreadTakesPerf is the version gate for the context hand-off: it asks the resolved deskread
// binary, once per process, whether its help lists every flag perfChildArgs would pass. A build
// that predates them (it would refuse an unknown flag, changing statusgen's result) or that
// cannot answer gets no flags. The probe runs with an empty environment and a bound, so it
// can reach no credential, record nothing and never hang a run.
func deskreadTakesPerf(bin string) bool {
	if v, ok := perfProbed.Load(bin); ok {
		return v.(bool)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "--help")
	cmd.Env = []string{}
	out, _ := cmd.CombinedOutput()
	ok := len(out) > 0
	for _, name := range perfChildFlags {
		ok = ok && bytes.Contains(out, []byte("--"+name))
	}
	perfProbed.Store(bin, ok)
	return ok
}

// perfChildArgs returns one deskread child's argv (args[0] is the read kind) with this
// invocation's context inserted as --perf-* flags after the kind. perfrec.ContextEnv decides
// what is forwarded (fed no inherited environment, it composes only the context): the tool
// profile alone, or under the workflow profile the session, agent, role and work with this
// invocation as the child's parent. Empty values are passed as empty flags, which clear what
// the child would otherwise inherit; with no profile selected the child is told "off", so an
// overridden or cleared parent never hands its child a context it did not use. Switch off, or a
// deskread without the flags, the argv is unchanged and the child inherits as before.
func perfChildArgs(bin string, args []string) []string {
	if !perfrec.Enabled() || len(args) == 0 {
		return args
	}
	perfSelect()
	if !deskreadTakesPerf(bin) {
		return args
	}
	flags := []string{"--perf-profile=off"}
	if perfCtx.Profile() != "" {
		flags = nil
	}
	for _, kv := range perfrec.ContextEnv(nil, perfCtx, true) {
		k, v, _ := strings.Cut(kv, "=")
		name := "perf-" + strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, "ASSAY_PERF_")), "_", "-")
		if name != "perf-work" {
			flags = append(flags, "--"+name+"="+v)
			continue
		}
		var refs []json.RawMessage
		if json.Unmarshal([]byte(v), &refs) != nil || len(refs) == 0 {
			flags = append(flags, "--perf-work=")
			continue
		}
		for _, r := range refs {
			flags = append(flags, "--perf-work="+string(r))
		}
	}
	out := append([]string{args[0]}, flags...)
	return append(out, args[1:]...)
}
