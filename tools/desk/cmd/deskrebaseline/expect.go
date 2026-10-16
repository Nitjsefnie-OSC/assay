package main

import (
	"regexp"
	"strconv"
	"strings"
)

// expect.go — judge a probed command against the row's own Expect cell, the way the
// statusgen witness does (statusgen/verifyrun.go parseExpect / verdict). statusgen is a
// separate module, so the reader is mirrored here and held to the same vectors in
// expect_test.go. Before #1953 the probe read any non-zero exit as a changed result, so a row
// whose Expect IS `exit 1` was refused as behaviour-changed while it still passed.

// judgement is the three-state result of one probed row.
type judgement string

const (
	judgePass        judgement = "pass"
	judgeFail        judgement = "fail"
	judgeCouldNotRun judgement = "could-not-run"
)

// expectSpec is the machine-decidable part of an Expect cell: the process exit (default 0),
// an "output is" line, and a "≥ N" floor on the last output line's trailing integer.
type expectSpec struct {
	exitCode int
	wantLine string
	minCount int
	minSet   bool
}

var (
	expectOutputIsRe = regexp.MustCompile("(?i)output\\s+is\\s+`([^`]*)`")
	expectExitRe     = regexp.MustCompile(`(?i)\bexit(?:\s+code)?\s*=?\s*(\d+)`)
	expectMinRe      = regexp.MustCompile("(?i)(?:≥|>=|at least)\\s*`?(\\d+)`?")
	expectCodeSpanRe = regexp.MustCompile("`[^`]*`")
	trailingIntRe    = regexp.MustCompile(`(\d+)\D*$`)
)

// parseExpect reads the RAW Expect cell. The output convention is read and removed first, and
// the exit is read only from text outside code spans: a backticked `exit=1` is a printed
// token, not the row's required exit.
func parseExpect(cell string) expectSpec {
	spec, rest := expectSpec{}, cell
	if m := expectOutputIsRe.FindStringSubmatch(cell); m != nil {
		spec.wantLine = strings.TrimSpace(m[1])
		rest = expectOutputIsRe.ReplaceAllString(cell, " ")
	}
	if m := expectExitRe.FindStringSubmatch(expectCodeSpanRe.ReplaceAllString(rest, " ")); m != nil {
		spec.exitCode, _ = strconv.Atoi(m[1])
	}
	if m := expectMinRe.FindStringSubmatch(rest); m != nil {
		spec.minCount, _ = strconv.Atoi(m[1])
		spec.minSet = true
	}
	return spec
}

// judge decides one probed run. A run that could not execute is never a pass or a fail.
func (e expectSpec) judge(ran bool, rc int, out string) judgement {
	if !ran {
		return judgeCouldNotRun
	}
	if rc != e.exitCode {
		return judgeFail
	}
	if e.wantLine != "" && !hasLine(out, e.wantLine) {
		return judgeFail
	}
	if e.minSet {
		n, ok := lastInt(out)
		if !ok {
			return judgeCouldNotRun
		}
		if n < e.minCount {
			return judgeFail
		}
	}
	return judgePass
}

func hasLine(out, want string) bool {
	for _, l := range strings.Split(out, "\n") {
		if strings.TrimSpace(l) == want {
			return true
		}
	}
	return false
}

// lastInt reads the trailing integer of the last non-empty output line.
func lastInt(out string) (int, bool) {
	lines := strings.Split(out, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		l := strings.TrimSpace(lines[i])
		if l == "" {
			continue
		}
		m := trailingIntRe.FindStringSubmatch(l)
		if m == nil {
			return 0, false
		}
		n, err := strconv.Atoi(m[1])
		return n, err == nil
	}
	return 0, false
}
