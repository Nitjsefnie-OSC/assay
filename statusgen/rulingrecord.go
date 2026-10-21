package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// `statusgen --corroborate --ruling-record <path>` corroborates ONE decision
// record's human ruling outside a pull request (#2079).
//
// The PR mode of --corroborate gates the records a PR touches, and a real
// decided-by name there can corroborate against the PR's own reviews. A caller
// that holds no PR — a dispatcher checking that a gate: human brief's recorded
// ruling stands before it starts work — has no such anchor. The only anchor left
// is the record's ruling: link (registers-v1 §7.5), so this mode demands one and
// resolves it through the same resolveRuling every PR-mode record goes through:
// same link grammar, same repository, unedited non-bot comment by a mapped human
// that names the record, on a decision issue for it. Nothing is relaxed.
//
// Exit codes: 0 = the ruling corroborates; 1 = it does not (or the record is
// unreadable, or carries no ruling: link); 2 = could-not-check (usage, an
// undeterminable repository, or a forge read that failed).

// rulingRecordClientFn builds the forge client this mode reads through — a seam
// so a test can serve the forge from an httptest server.
var rulingRecordClientFn = rulingForgeClient

func runCorroborateRecord(root, rel string) int {
	rel = filepath.ToSlash(filepath.Clean(strings.TrimSpace(rel)))
	if !isDecisionRecordPath(rel) {
		fmt.Fprintf(os.Stderr, "statusgen: --ruling-record %q is not a decision record (a repo-relative docs/streams/%s/<name>.md, README.md excluded)\n", rel, decisionsDirName)
		return 2
	}
	repo := repoFromOriginAtFn(root)
	if repo == "" {
		fmt.Fprintf(os.Stderr, "statusgen: --ruling-record %s: COULD-NOT-CHECK — cannot determine the GitHub repository from the origin remote of %s\n", rel, root)
		return 2
	}
	rec := loadDecisionRecordStamp(root, rel)
	o := resolveDecisionRulings(rulingRecordClientFn(), root, repo, []decisionRecordStamp{rec}, scanEffectiveConfig().HumanLogins)[rel]
	switch {
	case o.Present && o.Reason == rulingOK:
		fmt.Printf("%s: CORROBORATED — ruling comment by %s (mapped human %q) on this record's decision issue: %s\n", rel, o.Author, o.Name, o.Detail)
		return 0
	case o.Reason == rulingUnresolvable:
		fmt.Printf("%s: COULD-NOT-CHECK — ruling %s: %s\n", rel, o.Reason, o.Detail)
		return 2
	case !o.Present:
		// No ruling: link. A PR's own reviews can anchor a real decided-by name; outside a PR
		// nothing can, so the record does not corroborate whatever its decided-by says.
		fmt.Printf("%s: MISSING-CORROBORATION — ruling %s: the record carries no ruling: link; outside a pull request only a ruling: link to the human's ruling comment can corroborate it (registers-v1 §7.5)\n",
			rel, rulingPlaceholderUnratified)
		return 1
	default:
		fmt.Printf("%s: MISSING-CORROBORATION — ruling %s: %s\n", rel, o.Reason, o.Detail)
		return 1
	}
}
