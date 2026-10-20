package main

import (
	"strings"
	"testing"

	"github.com/medici-finance/assay/tools/desk/internal/deskkit"
)

// A refused acquire whose `show` says the key is FREE but another spelling of the same PR is
// ALIAS-HELD is a live-holder collision (medici-finance/assay#3700), not "no live holder, a
// claim-acquire error", and it is never classified STALE from the alias line's own state/age.
func TestAliasHeldIsReportedAsALiveHolder(t *testing.T) {
	s := &stub{}
	_, root := s.install(t)
	plantScripts(t, root)
	s.replies = []reply{
		{match: "remote get-url origin", stdout: "git@github.com:medici-finance/assay.git"},
		{match: "dispatch-claim.sh acquire", code: deskkit.ExitRefused},
		{match: "dispatch-claim.sh show", stdout: "dispatch-claim: FREE assay--item-1 (no refs/dispatch/assay--item-1 in the repo)\n" +
			"dispatch-claim: ALIAS-HELD other--item-1 — dispatch-claim other--item-1 owner=x state=claimed branch=feat/y " +
			"at=2026-08-24T00:00:00Z age=500m (same PR as assay--item-1)"},
	}
	err := cmdDispatch([]string{"item-1", "--root", root})
	if err == nil || deskkit.ExitCodeOf(err) != deskkit.ExitRefused {
		t.Fatalf("want a refusal, got %v", err)
	}
	for _, want := range []string{"already claimed by a LIVE holder", "another spelling", "other--item-1"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal does not say %q: %s", want, err.Error())
		}
	}
	if strings.Contains(err.Error(), "STALE") || strings.Contains(err.Error(), "NOT a collision") {
		t.Errorf("an alias holder was mislabelled: %s", err.Error())
	}
	if s.ran("steal") {
		t.Fatal("deskdispatch stole a claim held under another spelling")
	}
}
