# Verifier regression floor

`inventory.yaml` pins the selected test sources at an immutable baseline. The
inventory check resolves each test and compares its formatted function body with
an AST-body SHA256 frozen from that baseline. `freeze.go` reads the pinned Git
objects with replacement objects disabled to reproduce those provenance hashes
at authoring time; it does not generate semantic expectations. Routine tests
need no history, credentials or network. Missing/invalid/mismatched body hashes
fail the check. The optional checkout-head log is metadata, not a historical
binary execution or authenticity claim.

`cases/` contains hand-written raw Verify/Evidence facts and expected semantic
results extracted from the original assertions, plus an explicit passing control.
No witness renderer or evaluator generated the expected results. Fixture object
IDs and dates are synthetic. `record.json` freezes the candidate/acceptance
binding from the domain's candidate-a example, independently of future changes
to that example. The old and new adapters execute the current checked-out source:
old uses the existing package-main observation and eligibility adapters, new uses
`domain.Assess`; both reuse the single `evaluation` owner. The log names each
adapter, evaluator version/policy digest and source baseline. These are adapter
comparisons, not independent implementations of the coverage policy.

The graph and offline-reader tests retain their existing frozen inputs and are
reused unchanged, including through `TestFloorReused`. The domain does not ship
replacement reader/serializer APIs. Runner conformance stays in its own module;
all selected kit tests are inventoried and run unchanged. Claim-holder intent is
explicitly transferred to graph-execution/19; this floor does not qualify its
PostgreSQL replacement. Contract links on reuse rows identify the relevant
cross-cutting semantic boundary, not a claim those domain APIs replace the kit.

`differences.yaml` is empty because no semantic change is authorized here. A
future reviewed allowance must name its case, reason and exact old/new results;
stale, unknown, duplicate and mismatched allowances fail. Expectations continue
to pin the old result. Mutation controls use synthetic allowances only.

Run from the repository root:

```
(cd statusgen && GOWORK=off go test -timeout 90s -count=1 -v -run '^(TestRegressionFloor|TestFloorReused)' .)
python3 statusgen/testdata/regressionfloor/mutate.py
(cd loopadmin && GOWORK=off go test -timeout 90s -count=1 ./runner/conformance/)
```

The mutation script restores each modified owned file in `finally`, requires a
named behavioral failure, and never treats a compile error as a mutation kill.
Do not run another test or source consumer concurrently with these mutations.
No baseline test is removed or weakened. The existing testledger remains the
only departure report. Local test results confer no independent verification,
publication, isolation or recovery qualification.
