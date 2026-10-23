package main

import (
	"github.com/medici-finance/assay/tools/desk/internal/deskkit"
	"strings"
	"testing"
)

const unattributed4114 = "| # | Class | Command | Acceptance result | Supplemental host result (no credit) |\n|---|---|---|---|---|\n| 9 | check | statusgen --lint | pass | exit 0 |\n| 10 | check:ci | check | could-not-run: closure refused | no credit |\n"
const attributedBlocked = "| # | Command | Observed | Date | Runner |\n|---|---|---|---|---|\n| 9 | statusgen --lint | exit 0 | 2026-10-23 | human:operator |\n| 10 | check | could-not-run: closure refused | 2026-10-23 | human:operator |\n"

func TestEvidenceSubmissionAttribution(t *testing.T) {
	base := "# Brief\n\n## Evidence\n"
	for name, tc := range map[string]struct {
		content, old string
		refuse       bool
	}{
		"4114":                    {base + unattributed4114, base, true},
		"blocked":                 {base + attributedBlocked, base, false},
		"historical":              {base + unattributed4114 + "\n" + attributedBlocked, base + unattributed4114, false},
		"duplicate":               {base + unattributed4114 + unattributed4114, base + unattributed4114, true},
		"missing_date":            {base + strings.ReplaceAll(attributedBlocked, "2026-10-23", ""), base, true},
		"invalid_date":            {base + strings.ReplaceAll(attributedBlocked, "2026-10-23", "2026-02-30"), base, true},
		"missing_runner":          {base + strings.ReplaceAll(attributedBlocked, "human:operator", ""), base, true},
		"placeholder":             {base + strings.ReplaceAll(attributedBlocked, "human:operator", "Unknown-runner"), base, true},
		"header_reinterpretation": {base + strings.ReplaceAll(attributedBlocked, "| Runner |", "| Unknown |"), base + attributedBlocked, true},
		"legacy_when":             {base + strings.ReplaceAll(attributedBlocked, "Date", "When"), base, false},
		"commented_header":        {base + strings.ReplaceAll(unattributed4114, "Command", "Com<!-- note -->mand"), base, true},
		"actual_header":           {base + "| # | Check | Expected | Actual |\n|---|---|---|---|\n| 1 | check | pass | blocked |\n", base, true},
		"comments":                {base + strings.ReplaceAll(unattributed4114, "| 9 |", "| <!--comment-->9 |"), base, true},
		"quoted":                  {base + "```\n" + unattributed4114 + "```\n", base, false},
		"derivation":              {base + "| Rank | Value | Derivation |\n|---|---|---|\n| 1 | 5 | derived |\n", base, false},
	} {
		t.Run(name, func(t *testing.T) {
			err := checkAddedRowAttribution([]byte(tc.content), []byte(tc.old))
			if (err != nil) != tc.refuse {
				t.Fatalf("error=%v want refuse=%v", err, tc.refuse)
			}
		})
	}
}
func TestEvidenceAttributionWriteGate(t *testing.T) {
	for _, fragment := range []bool{false, true} {
		for _, dry := range []bool{false, true} {
			for _, valid := range []bool{false, true} {
				name := "direct"
				if fragment {
					name = "fragment"
				}
				if dry {
					name += "_dry"
				}
				if valid {
					name += "_attributed"
				}
				t.Run(name, func(t *testing.T) {
					f, _ := setupFake(t)
					path := "docs/streams/x/brief.md"
					base := "# Brief\n\n## Evidence\n"
					table := unattributed4114
					if valid {
						table = attributedBlocked
					}
					f.setFile(path, base)
					root := rootWithFile(t, path, base+table)
					args := []string{"example-org/tracker", "evidence", "--root", root, "--evidence-file", path}
					if fragment {
						args = []string{"example-org/tracker", "evidence", "--root", root, "--evidence-file", writeRepoFile(t, "rows.md", table), "--brief-path", path}
					}
					if dry {
						args = append(args, "--dry-run")
					}
					code := run(args)
					want := deskkit.ExitRefused
					if valid {
						want = deskkit.ExitOK
					}
					if code != want {
						t.Fatalf("exit=%d want=%d", code, want)
					}
					writes := 0
					if valid && !dry {
						writes = 1
					}
					if f.putCalls != writes {
						t.Fatalf("writes=%d want=%d", f.putCalls, writes)
					}
					if valid && !dry && !strings.Contains(f.putContent, "could-not-run: closure refused") {
						t.Fatal("blocked disposition lost")
					}
				})
			}
		}
	}
}
