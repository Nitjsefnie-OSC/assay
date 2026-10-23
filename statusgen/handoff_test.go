package main

import (
	"os"
	"path/filepath"
	"testing"
)

// regression: class verification-handoff
func TestHandoffScanner(t *testing.T) {
	streams, root := scanFixtureRepo(t, map[string]string{
		"issue-805.md": closedPlaceholderFile("alpha", 805, "bug"),
	})
	archive := filepath.Join(root, "docs/streams/alpha/done")
	if err := os.MkdirAll(archive, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(archive, "issue-806.md"), []byte(closedPlaceholderFile("alpha", 806, "bug")), 0600); err != nil {
		t.Fatal(err)
	}
	issues := []ghIssue{
		{Number: 801, Labels: lbl("bug", "raised-by:verifier")},
		{Number: 802, Labels: lbl("bug", "to:worker")},
		{Number: 803, Labels: lbl("bug", "to:worker", "to:intake")},
		{Number: 804, Labels: lbl("bug", "to:")},
		{Number: 805, Labels: lbl("bug", "to:worker")},
		{Number: 806, Labels: lbl("bug", "to:worker")},
		{Number: 807, Labels: lbl("bug", "needs-decision")},
	}
	plans, closes, notices, skipped := planScan(root, streams, fixtureLister(map[string][]ghIssue{scanHomeRepo(): issues}, ""), blessAll, labelHoldClearAll)
	if len(plans) != 1 || plans[0].Issue != 801 || len(closes) != 0 || len(skipped) != 0 {
		t.Fatalf("addressed scanner policy: plans=%+v closes=%+v skipped=%+v", plans, closes, skipped)
	}
	if len(notices) != 5 {
		t.Fatalf("missing addressed/unknown holds: %v", notices)
	}
	_, closes, _, skipped = planScan(root, streams, fixtureLister(nil, scanHomeRepo()), blessAll, labelHoldClearAll)
	if len(skipped) != 1 || len(closes) != 0 {
		t.Fatalf("unreadable observation became empty: %v %v", skipped, closes)
	}
}
