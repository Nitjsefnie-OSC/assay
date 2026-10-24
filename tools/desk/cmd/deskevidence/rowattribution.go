package main

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/medici-finance/assay/tools/desk/internal/deskkit"
)

// checkAddedRowAttribution checks submissions, not historical acceptance. In particular,
// implemented/human-gated BLOCKED reports do not reach verified/done attribution lint.
// Inspect execution tables only: risk-value derivations and other explanatory tables
// are not command observations. Never manufacture missing attribution from a heading,
// commit author, landing actor or today's clock.
func checkAddedRowAttribution(content, base []byte) error {
	lines, origin := stripClosedComments(string(content))
	start, end, ok := evidenceSpan(lines)
	if !ok {
		return nil
	}
	old := map[string]int{}
	oldLines, _ := stripClosedComments(string(base))
	if s, e, ok := evidenceSpan(oldLines); ok {
		header := ""
		for i := s; i < e; i++ {
			l := oldLines[i]
			if !strings.HasPrefix(strings.TrimSpace(l), "|") {
				header = ""
				continue
			}
			if i+1 < e && evidenceSeparatorRe.MatchString(strings.Trim(strings.TrimSpace(oldLines[i+1]), "|")) {
				header = strings.TrimSpace(l)
				i++
				continue
			}
			old[header+"\x00"+l]++
		}
	}
	dateIdx, runnerIdx := -1, -1
	execution := false
	header := ""
	fenced := false
	for i := start; i < end; i++ {
		raw := lines[i]
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~") {
			fenced = !fenced
			execution = false
			continue
		}
		if fenced {
			continue
		}
		if !strings.HasPrefix(line, "|") {
			execution = false
			continue
		}
		if evidenceSeparatorRe.MatchString(strings.Trim(line, "|")) {
			continue
		}
		spans := cellSpans(raw)
		cells := make([]string, len(spans))
		for j, s := range spans {
			cells[j] = strings.TrimSpace(raw[s[0]:s[1]])
		}
		if i+1 < end && evidenceSeparatorRe.MatchString(strings.Trim(strings.TrimSpace(lines[i+1]), "|")) {
			dateIdx, runnerIdx, execution = -1, -1, false
			header = strings.TrimSpace(raw)
			for j, c := range cells {
				// Date and Runner match exactly, as the lint does: a spelling the lint does not
				// find would pass here and fail there.
				switch strings.ToLower(c) {
				case "date":
					dateIdx = j
				case "runner":
					runnerIdx = j
				}
				if executionHeader(c) {
					execution = true
				}
			}
			i++
			continue
		}
		if old[header+"\x00"+raw] > 0 {
			old[header+"\x00"+raw]--
			continue
		}
		if !execution {
			continue
		}
		// The lint reads Date as the cell immediately left of Runner (flipDateRunnerRows,
		// evidenceVerifierInfo), so any other Date position passes here and fails the flip.
		valid := dateIdx >= 0 && runnerIdx == dateIdx+1 && runnerIdx < len(cells)
		if valid {
			_, err := time.Parse("2006-01-02", cells[dateIdx])
			valid = err == nil && !emptyRunner(cells[runnerIdx]) && !placeholderRunner(strings.TrimSpace(strings.Trim(cells[runnerIdx], "`*_")))
		}
		if !valid {
			return deskkit.Refused(fmt.Sprintf("refused: added Evidence execution row at line %d of the landed brief (a fragment is read merged into it) needs separate Date (YYYY-MM-DD) and Runner cells, including blocked/unrun and supplemental observations; recover attribution from this run's records or rerun, never infer it from the landing actor — nothing was written", origin[i]+1))
		}
	}
	return nil
}

// stripClosedComments removes every closed HTML comment exactly as the lint does
// (evidenceCommentRe.ReplaceAllString, which JOINS the lines on either side of a comment
// that crosses a line break) and returns the resulting lines with, for each, the 0-based
// line of the submitted file on which it starts, so a refusal names a line the author can find.
func stripClosedComments(s string) (lines []string, origin []int) {
	var out strings.Builder
	at, line := 0, 0
	origin = []int{0}
	keep := func(seg string) {
		for i := 0; i < len(seg); i++ {
			if seg[i] == '\n' {
				line++
				origin = append(origin, line)
			}
		}
		out.WriteString(seg)
	}
	for _, m := range evidenceCommentRe.FindAllStringIndex(s, -1) {
		keep(s[at:m[0]])
		line += strings.Count(s[m[0]:m[1]], "\n")
		at = m[1]
	}
	keep(s[at:])
	return strings.Split(out.String(), "\n"), origin
}

// executionHeader reports whether a header cell names a command or its observation.
// Emphasis, backticks and invisible characters are ignored and a keyword may carry a
// suffix ("Commands", "Exit code", "Observed (exit + output)"), so a spelling the author
// reads as the same column is not a way out of the check.
func executionHeader(cell string) bool {
	c := strings.ToLower(strings.TrimSpace(strings.Trim(stripInvisible(cell), "`*_ ")))
	switch c {
	case "cmd", "check", "checks", "actual":
		return true
	}
	for _, p := range []string{"command", "exit", "output", "observed", "result", "outcome", "acceptance result"} {
		if strings.HasPrefix(c, p) {
			return true
		}
	}
	return false
}

// stripInvisible drops whitespace, format characters (zero-width space and joiners, BOM,
// word joiner) and the Hangul fillers, which render as nothing.
func stripInvisible(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || unicode.Is(unicode.Cf, r) {
			return -1
		}
		switch r {
		case '\u3164', '\u115F', '\u1160', '\uFFA0', '\u2800':
			return -1
		}
		return r
	}, s)
}

// emptyRunner reports whether a Runner cell names nobody: empty, or only what the lint's
// normalizeMark empties (an em dash, hyphen or en dash), or only invisible characters.
func emptyRunner(cell string) bool {
	switch strings.TrimSpace(strings.Trim(stripInvisible(cell), "`*_")) {
	case "", "—", "-", "–":
		return true
	}
	return false
}

// placeholderRunner reports whether a Runner cell holds a stand-in rather than an identity.
func placeholderRunner(runner string) bool {
	switch strings.ToLower(runner) {
	case "unknown-runner", "unknown", "n/a", "na", "tbd", "todo", "none", "?", "??":
		return true
	}
	return strings.Contains(runner, "<runner>")
}
