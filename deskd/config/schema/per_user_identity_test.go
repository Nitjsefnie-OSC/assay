package schema

import "testing"

// ── the service contract Task 9 — the per-user-identity ROLLOUT GATE ───────────
//
// These tests enforce the configuration half of the per-cell rollout gate:
//
//	1. the ENABLE is a NEW explicit per-cell field, `per_user_identity:
//	   enabled`; ABSENT means OFF. A dev `environments:` declaration is a
//	   PRECONDITION for setting it and never an enable by itself.
//	4. cellsvalidate refuses EVERY private-cell opt-in unconditionally,
//	   except a cell named in the proving-cell allowlist.
//	5. "private" = any cell without `allow_anonymous_read: true` (the only
//	   static signal an offline validator has).
//
// Every case below is written so the FAIL-CLOSED direction is the one that
// breaks it: a rule that stopped refusing would fail here, not merely stop
// being exercised.

// puiDoc builds a one-cell document with the given cell name, the given extra
// cell-level lines, and (optionally) a dev environments block.
func puiDoc(name, extra string, withDevEnv bool) string {
	doc := `
cells:
  - name: ` + name + `
    repos: [example-org/toolkit]
    streams_root: docs/streams
    loops:
      - role: review
        schedule: "*/10 * * * *"
        models: [deepseek-chat]
        app: assay-reviewer-app
    silence_threshold: 48h
    notification_tiers: {review: push, question: badge, notify: none}
` + extra
	if withDevEnv {
		doc += `    environments:
      - name: ` + name + `-dev
        kind: k8s
        tier: dev
        cluster: fixture-cluster
        namespaces: [example-dev-app]
        access: [read]
        runner_pool: pool-` + name + `
`
	}
	return doc
}

// (b) The opt-in WITHOUT a dev environment is refused: the precondition is
// real, not decorative.
func TestPerUserIdentity_OptInWithoutDevEnvironmentIsRefused(t *testing.T) {
	err := parseErr(t, puiDoc("assay", "    per_user_identity: enabled\n", false))
	assertMentions(t, err, "per_user_identity")
	assertMentions(t, err, "environments")
	assertMentions(t, err, "assay")
}

// (c) The opt-in on a PRIVATE, NON-proving cell is refused unconditionally —
// the ruling's fail-closed answer to forks 4 and 5. A dev environment does not
// rescue it: the allowlist is the only exception.
func TestPerUserIdentity_OptInOnPrivateNonProvingCellIsRefused(t *testing.T) {
	err := parseErr(t, puiDoc("tracker", "    per_user_identity: enabled\n", true))
	assertMentions(t, err, "per_user_identity")
	assertMentions(t, err, "tracker")
	// The refusal must say WHY, so an operator reading cellsvalidate's stderr
	// learns the rule rather than only that they were refused.
	assertMentions(t, err, "allow_anonymous_read")
}

// (d) The opt-in on the PROVING cell, with a dev environment, is accepted —
// and resolves to enabled.
func TestPerUserIdentity_OptInOnProvingCellWithDevEnvironmentPasses(t *testing.T) {
	f, err := Parse([]byte(puiDoc("assay", "    per_user_identity: enabled\n", true)))
	if err != nil {
		t.Fatalf("the proving cell with a dev environment must validate: %v", err)
	}
	if !f.Cells[0].PerUserIdentityEnabled() {
		t.Fatal("the proving cell opted in with a dev environment must resolve ENABLED")
	}
}

// A PUBLIC cell (verified anonymous-read opt-in) is not private, so the
// allowlist does not apply to it: the refusal is scoped to private cells
// exactly as ruled, and does not quietly become "only ever the assay cell".
func TestPerUserIdentity_OptInOnPublicCellPasses(t *testing.T) {
	f, err := Parse([]byte(puiDoc("public-cell",
		"    allow_anonymous_read: true\n    per_user_identity: enabled\n", true)))
	if err != nil {
		t.Fatalf("a public cell with a dev environment may opt in: %v", err)
	}
	if !f.Cells[0].PerUserIdentityEnabled() {
		t.Fatal("a public opted-in cell with a dev environment must resolve ENABLED")
	}
}

// ABSENT is OFF, and absence stays valid: this is the state every committed
// cell is in today and the state the whole fleet defaults to.
func TestPerUserIdentity_AbsentIsOffAndValid(t *testing.T) {
	f, err := Parse([]byte(puiDoc("assay", "", true)))
	if err != nil {
		t.Fatalf("a cell with no per_user_identity key must validate: %v", err)
	}
	if f.Cells[0].PerUserIdentity != "" {
		t.Fatalf("absent per_user_identity must be the zero value, got %q", f.Cells[0].PerUserIdentity)
	}
	if f.Cells[0].PerUserIdentityEnabled() {
		t.Fatal("ABSENT must mean OFF — this is the default the whole rollout gate rests on")
	}
}

// `enabled` is the ONLY spellable value. There is deliberately no `disabled`,
// `true`, `on` or `yes`: a closed enum whose only member ENABLES means a
// typo can never accidentally turn the path on, and OFF is expressed by
// deleting the key rather than by a second value that must be spelled right.
func TestPerUserIdentity_OnlyEnabledIsSpellable(t *testing.T) {
	for _, v := range []string{"disabled", "true", "on", "yes", "Enabled", "1", "off"} {
		err := parseErr(t, puiDoc("assay", "    per_user_identity: "+v+"\n", true))
		assertMentions(t, err, "per_user_identity")
	}
}

// The proving-cell allowlist is a COMPILE-TIME constant, not a config field.
// If it ever becomes readable from cells.yaml, the same edit that sets the
// opt-in could also widen the allowlist that is meant to refuse it — a
// self-authorizing config. This asserts the shape as well as the contents.
func TestPerUserIdentity_ProvingCellAllowlistIsSmallAndClosed(t *testing.T) {
	if len(ProvingCells) != 1 || !ProvingCells["assay"] {
		t.Fatalf("ProvingCells = %v, want exactly {assay} — widening it is a reviewed CODE change (relayed direction, toolkit 1963 fork 3)", ProvingCells)
	}
	if !IsProvingCell("assay") || IsProvingCell("tracker") || IsProvingCell("") {
		t.Fatal("IsProvingCell must name exactly the proving cell and fail closed on everything else")
	}
}

// The adopter-specific real-config assertion remains enforced in its original
// source repository. It is outside this reusable module: no adopter cells.yaml
// is transferred. The private handoff pins the retained function and its hash.

// The committed fixtures are the ones Verify 12/12b run cellsvalidate over.
// Asserting them here keeps them red (and the green one green) on every CI
// run, so a row that says "cellsvalidate exits non-zero on this file" is
// backed by a file that is still refused rather than one that quietly started
// passing.
func TestPerUserIdentity_CommittedFixturesKeepTheirVerdicts(t *testing.T) {
	for _, name := range []string{"private-cell-opt-in", "opt-in-without-dev-environment"} {
		if _, err := Load("../testdata/peruseridentity/" + name + ".yaml"); err == nil {
			t.Errorf("%s.yaml: expected the fixture to be REFUSED, got nil — the fixture has stopped being red", name)
		}
	}
	f, err := Load("../testdata/peruseridentity/proving-cell-opt-in.yaml")
	if err != nil {
		t.Fatalf("proving-cell-opt-in.yaml must validate: %v", err)
	}
	if len(f.Cells) != 2 {
		t.Fatalf("the row-12 fixture must carry TWO cells so one process can serve an ON and an OFF cell, got %d", len(f.Cells))
	}
	if !f.Cells[0].PerUserIdentityEnabled() {
		t.Error("the proving cell in the fixture must resolve ENABLED")
	}
	if f.Cells[1].PerUserIdentityEnabled() {
		t.Error("the second cell in the fixture must resolve OFF — that difference is what row 12 proves")
	}
}

// PerUserIdentityEnabled is the RUNTIME predicate deskd fails closed on. It
// must re-derive all three conditions itself rather than trusting that the
// document was validated: deskd is a second layer, on its own signal.
func TestPerUserIdentity_RuntimePredicateReDerivesEveryCondition(t *testing.T) {
	dev := []Environment{{Name: "e", Kind: EnvKindK8s, Tier: EnvTierDev,
		Namespaces: []string{"example-dev-app"}, Access: []AccessLevel{AccessRead}, RunnerPool: "p"}}
	cases := []struct {
		name string
		cell Cell
		want bool
	}{
		{"proving cell, opted in, dev env", Cell{Name: "assay", PerUserIdentity: PerUserIdentityEnabled, Environments: dev}, true},
		{"no opt-in", Cell{Name: "assay", Environments: dev}, false},
		{"opt-in, no environments at all", Cell{Name: "assay", PerUserIdentity: PerUserIdentityEnabled}, false},
		{"opt-in, private, not the proving cell", Cell{Name: "tracker", PerUserIdentity: PerUserIdentityEnabled, Environments: dev}, false},
		{"opt-in, public, not the proving cell", Cell{Name: "tracker", PerUserIdentity: PerUserIdentityEnabled, AllowAnonymousRead: true, Environments: dev}, true},
		{"unspellable value that slipped past validation", Cell{Name: "assay", PerUserIdentity: "true", Environments: dev}, false},
	}
	for _, c := range cases {
		if got := c.cell.PerUserIdentityEnabled(); got != c.want {
			t.Errorf("%s: PerUserIdentityEnabled()=%v want %v", c.name, got, c.want)
		}
	}
}
