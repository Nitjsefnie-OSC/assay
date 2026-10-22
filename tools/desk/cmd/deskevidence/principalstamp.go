package main

// On-behalf-of principal stamp for App-run Evidence rows.
//
// statusgen's attribution lint requires every App-authored Evidence row dated at or after
// its cutover to carry the `on-behalf-of human:<subject>` annotation in its Runner cell
// (statusgen/attribution.go, principalAttributionProblems). verifyrun renders that
// annotation when it writes a witness row, but a row that reaches this tool any other way
// (a transcribed table, a re-run assembled from several witnesses) can arrive without it —
// and the only remedy was to hand-edit the Runner cell, which is exactly the kind of
// model-written attribution the annotation exists to replace.
//
// So the tool stamps it at ingest. The principal comes from ONE place: deskkit's
// ResolvePrincipal, the same config-home roster source the commit trailer
// (OnBehalfOfCommitSuffix) and verifyrun's own annotation read. It is never taken from a
// flag, an environment variable, or the content being landed. When it cannot be resolved
// the landing refuses (exit 5) and nothing is written — a placeholder is never stamped.
//
// What is touched, and what never is:
//   - Only data rows of a `## Evidence` table that THIS landing adds (a line not already on
//     the branch). A row already on the branch stays byte-identical, stamped or not.
//   - Only a Runner cell that names an App identity (a leading `<slug>[bot]` identifier). A
//     human runner (`human:<name>`) needs no principal and is never read.
//   - An added App cell is classified with the LINT'S OWN reading of the annotation
//     (onBehalfOfPrincipalOf: the LAST `on-behalf-of human:<p>`, case-insensitive), against
//     the subject the roster resolves for this target and the same roster's human set:
//       - no `on-behalf-of` text at all: stamp;
//       - exactly one annotation, naming the resolved subject: left byte-identical;
//       - an annotation not in the `human:` form (a desk role, a bare name), one naming no
//         recognised roster human, more than one annotation, or the SAME human in the other
//         form (the login on a public target, the neutral name on a private one): every
//         on-behalf-of fragment is stripped and the resolved one stamped — the strip-then-
//         append rule deskkit.AppendOnBehalfOf applies to the commit trailer;
//       - any `human:` annotation naming a DIFFERENT recognised roster human: the landing
//         refuses (exit 5), names the row, and writes nothing. One human's attribution is
//         never overwritten with another's automatically.
//     So the cell content can never choose the principal: it is the roster's, or the
//     landing does not happen.
//   - The annotation goes immediately before a trailing runner-source tag
//     (`(forge-identity)`, `(ci-env)`, `(git-config)`), which is the order verifyrun renders,
//     so the tag itself is never rewritten; with no such tag it is appended to the cell.
//   - The stamp is always ONE parenthetical, ` (on-behalf-of human:<who>)`, in the Runner
//     cell, and a strip removes each old fragment with exactly one whitespace byte before it;
//     every other byte of the row is left as it was. The written stamp is always the exact
//     form `(on-behalf-of human:<who>)` and nothing broader, which is the only form a container
//     witness bracket's digest drops from a Runner cell (it then collapses the cell's
//     whitespace). So landing the stamp, or restamping one exact-form annotation to another,
//     never moves that digest; restamping a non-exact annotation (a desk role, an empty
//     principal) does, because the old annotation was inside the hash, and the bracket then
//     fails closed.
//
// Table reading mirrors statusgen's parseEvidenceRows (header = the row followed by a
// separator; Runner column by its header name, else the last cell; `\|` is cell content;
// closed HTML comments are not rows), so the row this stamps is the row the lint reads.

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/medici-finance/assay/tools/desk/internal/deskkit"
)

// evidenceSeparatorRe matches a Markdown table separator row (statusgen's separatorRowRe).
var evidenceSeparatorRe = regexp.MustCompile(`^[\s:|-]+$`)

// evidenceCommentRe matches a CLOSED HTML comment (statusgen's htmlCommentClosedRe).
var evidenceCommentRe = regexp.MustCompile(`(?s)<!--.*?-->`)

// runnerSourceTags are the trailing runner-source parentheticals verifyrun renders after
// the on-behalf-of annotation (statusgen/verifyrun.go runnerSource* constants).
var runnerSourceTags = []string{" (forge-identity)", " (ci-env)", " (git-config)"}

const onBehalfOfCellMarker = "on-behalf-of"

// principalSubjectFn resolves the subject the annotation names for a landing into repo.
// A seam only so a test can count resolutions; production is resolveStampSubject.
var principalSubjectFn = resolveStampSubject

// resolveStampSubject resolves the annotation's subject from the roster — the neutral name
// on a public target, the login on a target the roster states private (deskkit.Principal.Subject) — or
// refuses, wrapping ResolvePrincipal's own refusal so the operator sees which step needed it.
func resolveStampSubject(repo string) (string, error) {
	p, err := deskkit.ResolvePrincipal("", repo)
	if err != nil {
		return "", deskkit.Refused("refused: an App-run Evidence row in this landing needs its on-behalf-of " +
			"principal stamped or checked and the tool cannot stamp one — nothing was written: " + err.Error())
	}
	subject := p.Subject()
	if subject == "" {
		return "", deskkit.Refused("refused: an App-run Evidence row in this landing needs its on-behalf-of " +
			"principal stamped or checked and the roster resolved an empty one, so the tool cannot stamp one — nothing was written")
	}
	return subject, nil
}

// stampEvidencePrincipal returns content with the on-behalf-of annotation stamped into every
// App-run Evidence row content adds relative to base (see the file comment), plus a map from
// each stamped or restamped row's trimmed original text to its trimmed stamped text. resolve
// is called at most once, and only when the landing adds some App-run row (every such row is
// stamped or checked against it); its error is returned unchanged. A row naming a different
// recognised human refuses the whole landing (needsPrincipalStamp).
func stampEvidencePrincipal(content, base []byte, resolve func() (string, error)) ([]byte, map[string]string, error) {
	lines := strings.Split(string(content), "\n")
	start, end, ok := evidenceSpan(lines)
	if !ok {
		return content, nil, nil
	}
	onBranch := map[string]bool{}
	for _, l := range strings.Split(string(base), "\n") {
		onBranch[l] = true
	}
	// Blank closed comments out byte-for-byte (newlines kept) so offsets and line numbers in
	// the blanked view are the original's.
	section := strings.Join(lines[start:end], "\n")
	blanked := evidenceCommentRe.ReplaceAllStringFunc(section, func(m string) string {
		b := []byte(m)
		for i := range b {
			if b[i] != '\n' {
				b[i] = ' '
			}
		}
		return string(b)
	})
	view := strings.Split(blanked, "\n")

	subject := ""
	resolved := false
	var stamped map[string]string
	runnerIdx := -1
	for i := 0; i < len(view); i++ {
		line := strings.TrimSpace(view[i])
		if !strings.HasPrefix(line, "|") {
			runnerIdx = -1
			continue
		}
		if evidenceSeparatorRe.MatchString(strings.Trim(line, "|")) {
			continue
		}
		spans := cellSpans(view[i])
		if i+1 < len(view) && evidenceSeparatorRe.MatchString(strings.Trim(strings.TrimSpace(view[i+1]), "|")) {
			runnerIdx = -1
			for j, sp := range spans {
				if strings.ToLower(strings.TrimSpace(view[i][sp[0]:sp[1]])) == "runner" {
					runnerIdx = j
				}
			}
			i++
			continue
		}
		if len(spans) == 0 {
			continue
		}
		orig := lines[start+i]
		if onBranch[orig] {
			continue
		}
		sp := spans[len(spans)-1]
		if runnerIdx >= 0 && runnerIdx < len(spans) {
			sp = spans[runnerIdx]
		}
		// Decide on the comment-blanked view (what the lint reads); insert into the original
		// bytes at the same offset, so everything else in the line is left exactly as it was.
		cell := view[i][sp[0]:sp[1]]
		if !isAppRunnerCell(strings.TrimSpace(cell)) {
			continue
		}
		if !resolved {
			s, err := resolve()
			if err != nil {
				return nil, nil, err
			}
			subject, resolved = s, true
		}
		verdict, other := needsPrincipalStamp(cell, subject)
		switch verdict {
		case stampLeave:
			continue
		case stampRefuse:
			return nil, nil, deskkit.Refused(fmt.Sprintf("refused: the App-run Evidence row at line %d names a different "+
				"recognised roster human (%s) than the principal the roster resolves for this target; "+
				"this tool never overwrites one human's attribution with another's — nothing was written. Leave the Runner "+
				"cell unannotated and land again, or have that human land it: %s",
				start+i+1, other, strings.TrimSpace(orig)))
		}
		// stampAdd / stampRestamp: strip every on-behalf-of fragment (none, for stampAdd) from
		// both the blanked view and the original at the same offsets, then insert the resolved
		// annotation before any trailing source tag.
		cellView, cellOrig := cell, orig[sp[0]:sp[1]]
		for _, cut := range onBehalfOfFragments(cellView) {
			cellView = cellView[:cut[0]] + cellView[cut[1]:]
			cellOrig = cellOrig[:cut[0]] + cellOrig[cut[1]:]
		}
		runner := strings.TrimSpace(cellView)
		lead := len(cellView) - len(strings.TrimLeft(cellView, " \t"))
		pos := lead + len(runner)
		if tag, ok := trailingSourceTag(runner); ok {
			pos -= len(tag)
		}
		annotation := " (" + onBehalfOfCellMarker + " human:" + subject + ")"
		stampedView := cellView[:pos] + annotation + cellView[pos:]
		// Defensive: the stamped cell must read, to the lint, as exactly one annotation naming
		// the resolved subject. Anything else means the strip mis-parsed the cell.
		if p, ok := onBehalfOfPrincipalOf(stampedView); !ok || p != strings.ToLower(subject) ||
			strings.Count(strings.ToLower(stampedView), onBehalfOfCellMarker) != 1 {
			return nil, nil, deskkit.Unverifiable(fmt.Sprintf("on-behalf-of stamp: the restamped Runner cell at line %d "+
				"does not read back as one annotation naming the resolved principal — refusing rather than land it: %s",
				start+i+1, strings.TrimSpace(orig)), nil)
		}
		newLine := orig[:sp[0]] + cellOrig[:pos] + annotation + cellOrig[pos:] + orig[sp[1]:]
		if stamped == nil {
			stamped = map[string]string{}
		}
		stamped[strings.TrimSpace(orig)] = strings.TrimSpace(newLine)
		lines[start+i] = newLine
	}
	if stamped == nil {
		return content, nil, nil
	}
	return []byte(strings.Join(lines, "\n")), stamped, nil
}

// isAppRunnerCell reports whether a trimmed Runner cell names an App identity — a leading
// `<slug>[bot]` identifier — rather than a human runner (statusgen's isAppRunnerToken /
// isHumanRunnerToken reading).
func isAppRunnerCell(runner string) bool {
	if runner == "" || strings.HasPrefix(strings.ToLower(runner), "human:") {
		return false
	}
	id := runner
	if i := strings.IndexAny(id, " ("); i >= 0 {
		id = id[:i]
	}
	return strings.HasSuffix(id, "[bot]")
}

// stampVerdict is what the stamp does with one added App-run Runner cell.
type stampVerdict int

const (
	stampAdd     stampVerdict = iota // no on-behalf-of text: append the resolved annotation
	stampLeave                       // one annotation naming the resolved subject: byte-identical
	stampRestamp                     // a non-principal annotation: strip every fragment, then stamp
	stampRefuse                      // names a different recognised human: refuse, write nothing
)

// needsPrincipalStamp classifies an App-run Runner cell (comment-blanked) against the subject
// the roster resolved for this target — see the file comment for the rule. On stampRefuse it
// also returns the other human's principal as the cell names it.
//
// The annotation is read with the lint's own parse (onBehalfOfPrincipalOf: the LAST
// occurrence, case-insensitive). Recognition reads the same roster as the subject's
// resolution (deskkit.EffectiveConfig), so it never depends on the target's lint config.
func needsPrincipalStamp(cell, subject string) (stampVerdict, string) {
	lower := strings.ToLower(cell)
	n := strings.Count(lower, onBehalfOfCellMarker)
	if n == 0 {
		return stampAdd, ""
	}
	subjectLogin := rosterHumanLogin(subject)
	// Every `human:` fragment in the cell, not only the one the lint reads: an earlier
	// fragment naming another recognised human is still that human's name, and stripping it
	// would be the overwrite this rule forbids.
	for _, p := range onBehalfOfPrincipals(cell) {
		if login := rosterHumanLogin(p); login != "" && login != subjectLogin {
			return stampRefuse, p
		}
	}
	if p, ok := onBehalfOfPrincipalOf(cell); ok && n == 1 && p == strings.ToLower(strings.TrimSpace(subject)) {
		return stampLeave, ""
	}
	return stampRestamp, ""
}

// rosterHumanLogin returns the login a principal names when the roster recognises it as a
// human — a login in the human set, or a neutral name the human-login map carries for one —
// and "" otherwise. KEEP IN SYNC with statusgen's onBehalfOfPrincipalRecognised
// (statusgen/principal.go), which accepts exactly these two forms.
func rosterHumanLogin(principal string) string {
	p := strings.ToLower(strings.TrimSpace(principal))
	if p == "" {
		return ""
	}
	c := deskkit.EffectiveConfig()
	if _, ok := c.Humans[p]; ok {
		return p
	}
	if login, ok := c.HumanLogins[p]; ok {
		if _, human := c.Humans[login]; human {
			return login
		}
	}
	return ""
}

// onBehalfOfPrincipalOf is statusgen's parse of a cell's annotation (statusgen/principal.go,
// onBehalfOfPrincipalOf), copied so the stamp classifies a cell exactly as the lint reads it:
// the LAST case-insensitive `on-behalf-of human:` occurrence; the principal runs to `)`,
// whitespace or `|`, lowercased. KEEP IN SYNC — TestPrincipalStampLintParity runs the real
// lint against this reading.
func onBehalfOfPrincipalOf(cell string) (string, bool) {
	const marker = onBehalfOfCellMarker + " human:"
	idx := strings.LastIndex(strings.ToLower(cell), marker)
	if idx < 0 {
		return "", false
	}
	p := principalAt(cell[idx+len(marker):])
	return p, p != ""
}

// onBehalfOfPrincipals returns the principal of EVERY `on-behalf-of human:` occurrence in a
// cell, in order, with onBehalfOfPrincipalOf's per-occurrence parse.
func onBehalfOfPrincipals(cell string) []string {
	const marker = onBehalfOfCellMarker + " human:"
	var out []string
	lower := strings.ToLower(cell)
	for off := 0; ; {
		idx := strings.Index(lower[off:], marker)
		if idx < 0 {
			return out
		}
		at := off + idx + len(marker)
		if p := principalAt(cell[at:]); p != "" {
			out = append(out, p)
		}
		off = at
	}
}

func principalAt(rest string) string {
	end := 0
	for end < len(rest) {
		c := rest[end]
		if c == ')' || c == ' ' || c == '\t' || c == '\n' || c == '|' {
			break
		}
		end++
	}
	return strings.ToLower(strings.TrimSpace(rest[:end]))
}

// onBehalfOfFragments returns the [start,end) byte spans, in DESCENDING order (so they can
// be cut one after another without shifting each other), of every on-behalf-of fragment in a
// Runner cell, each with the one whitespace byte before it (if any):
//   - a fragment opened by `(` (whitespace allowed between) runs to the next `)` inclusive,
//     or to the end of the cell's text when unclosed;
//   - a bare fragment runs to the next `(` or `)` (trailing whitespace left in place), or to
//     the end of the cell's text.
//
// The cell's own trailing padding is never part of a fragment.
func onBehalfOfFragments(cell string) [][2]int {
	lower := strings.ToLower(cell)
	textEnd := len(strings.TrimRight(cell, " \t"))
	var spans [][2]int
	for off := 0; off < textEnd; {
		idx := strings.Index(lower[off:], onBehalfOfCellMarker)
		if idx < 0 {
			break
		}
		idx += off
		start, end := idx, textEnd
		j := idx - 1
		for j >= off && (cell[j] == ' ' || cell[j] == '\t') {
			j--
		}
		if j >= off && cell[j] == '(' {
			start = j
			if k := strings.IndexByte(cell[idx:textEnd], ')'); k >= 0 {
				end = idx + k + 1
			}
		} else {
			after := idx + len(onBehalfOfCellMarker)
			if k := strings.IndexAny(cell[after:textEnd], "()"); k >= 0 {
				end = after + k
			}
			for end > start && (cell[end-1] == ' ' || cell[end-1] == '\t') {
				end--
			}
		}
		// Exactly ONE whitespace byte before a fragment goes with it — the shape the stamp
		// inserts (` (on-behalf-of human:<who>)`). So stripping an old exact-form stamp and
		// inserting the new one leaves every other byte of the row as it was.
		if start > off && (cell[start-1] == ' ' || cell[start-1] == '\t') {
			start--
		}
		spans = append(spans, [2]int{start, end})
		if end <= idx {
			end = idx + len(onBehalfOfCellMarker)
		}
		off = end
	}
	for a, b := 0, len(spans)-1; a < b; a, b = a+1, b-1 {
		spans[a], spans[b] = spans[b], spans[a]
	}
	return spans
}

// trailingSourceTag reports the runner-source tag a trimmed Runner cell ends with, if any. The
// annotation goes before it — verifyrun's rendering order — so the tag is never rewritten.
func trailingSourceTag(runner string) (string, bool) {
	for _, tag := range runnerSourceTags {
		if strings.HasSuffix(runner, tag) {
			return tag, true
		}
	}
	return "", false
}

// evidenceSpan locates the `## Evidence` section's lines: [start,end), end being the next
// `## ` heading or the end of content (statusgen's evidenceSectionSpan).
func evidenceSpan(lines []string) (start, end int, ok bool) {
	start = -1
	for i, l := range lines {
		if strings.TrimSpace(l) == "## Evidence" {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return 0, 0, false
	}
	end = len(lines)
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "## ") {
			end = i
			break
		}
	}
	return start, end, true
}

// cellSpans returns each table cell's [start,end) byte offsets in line, splitting on `|` but
// not on an escaped `\|`, and dropping at most one blank cell at each end — the cell
// boundaries statusgen's splitRowEscaped reads.
func cellSpans(line string) [][2]int {
	var spans [][2]int
	cur := 0
	for i := 0; i < len(line); i++ {
		switch {
		case line[i] == '\\' && i+1 < len(line) && line[i+1] == '|':
			i++
		case line[i] == '|':
			spans = append(spans, [2]int{cur, i})
			cur = i + 1
		}
	}
	spans = append(spans, [2]int{cur, len(line)})
	if len(spans) > 0 && strings.TrimSpace(line[spans[0][0]:spans[0][1]]) == "" {
		spans = spans[1:]
	}
	if n := len(spans); n > 0 && strings.TrimSpace(line[spans[n-1][0]:spans[n-1][1]]) == "" {
		spans = spans[:n-1]
	}
	return spans
}

// stampFragment applies stampEvidencePrincipal's line map to a --brief-path fragment, so the
// fragment merged into the brief is the stamped one and the block-level idempotency check
// compares like with like.
func stampFragment(fragment []byte, stamped map[string]string) []byte {
	lines := strings.Split(string(fragment), "\n")
	for i, l := range lines {
		key := strings.TrimSpace(l)
		if to, ok := stamped[key]; ok {
			lines[i] = strings.Replace(l, key, to, 1)
		}
	}
	return []byte(strings.Join(lines, "\n"))
}
