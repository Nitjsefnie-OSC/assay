package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/medici-finance/assay/tools/desk/internal/deskkit"
)

// checkAddedRowAttribution checks submissions, not historical acceptance. In particular,
// implemented/human-gated BLOCKED reports do not reach verified/done attribution lint.
// Inspect execution tables only: risk-value derivations and other explanatory tables
// are not command observations. Never manufacture missing attribution from a heading,
// commit author, landing actor or today's clock.
func checkAddedRowAttribution(content, base []byte) error {
	lines := strings.Split(evidenceCommentRe.ReplaceAllString(string(content), ""), "\n")
	start, end, ok := evidenceSpan(lines)
	if !ok {
		return nil
	}
	old := map[string]int{}
	oldLines := strings.Split(evidenceCommentRe.ReplaceAllString(string(base), ""), "\n")
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
				switch strings.ToLower(c) {
				case "date", "when":
					dateIdx = j
				case "runner":
					runnerIdx = j
				case "command", "exit", "output", "observed", "result", "observed result", "observed output", "check", "actual", "acceptance result":
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
		valid := dateIdx >= 0 && runnerIdx >= 0 && dateIdx < len(cells) && runnerIdx < len(cells)
		if valid {
			_, err := time.Parse("2006-01-02", cells[dateIdx])
			runner := strings.TrimSpace(strings.Trim(cells[runnerIdx], "`*_"))
			valid = err == nil && runner != "" && runner != "—" && runner != "-" && !strings.EqualFold(runner, "unknown-runner") && !strings.Contains(runner, "<runner>")
		}
		if !valid {
			return deskkit.Refused(fmt.Sprintf("refused: added Evidence execution row at line %d needs separate Date (YYYY-MM-DD) and Runner cells, including blocked/unrun and supplemental observations; recover attribution from this run's records or rerun, never infer it from the landing actor — nothing was written", i+1))
		}
	}
	return nil
}
