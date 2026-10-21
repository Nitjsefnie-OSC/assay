# Deskboard decision observations — issue #3373

`dispatch` and `throughput` retain their canonical queue builders, caps and admission
rules. This change removes repeated decision discovery and issue reads within one
admission pass. It does not replace the board classifier, introduce a scheduler or
deploy a shared service.

## Observation algorithm and authority

1. Read each local brief and evaluate its execution tier before remote work.
2. Resolve the existing role-bound forge once per repository that needs a decision read.
3. Lazily discover decision markers once per repository, through a fixed typed mode of
   `SearchIssues`. Walk complete pages or report unknown. Include closed and decided
   issues: the lowest matching issue must retain its existing meaning.
4. Index discovered marker bodies by exact spelling. A brief selects its literal id
   plus the same optional legacy spelling that `deskdispatch` searches; another
   project's colon marker is not an alias. Reconfirm membership
   using fresh `GetIssue` observations and the same lowest-numbered-candidate helper
   as `deskdispatch`. A search body's marker alone cannot hold a row.
5. Share each issue observation and canonical `BriefDecisionWait` result within this
   pass. Read remaining issues under the existing global six-worker pool. Fold results
   in input order, including held and could-not-check explanations.

A failed discovery never becomes an empty result. An explicit `decision-issue` does
not depend on discovery. State, labels, label re-application, immutable ratifier identity,
comment history completeness and missing-ratifier handling retain their existing
evaluator. Unknown rows remain visible and named, as before. Every new admission pass
starts with fresh observations; no ruling is persisted or reused across sweeps.

The discovery mode retains the existing admitted, read-only forge seam and provider
clients. It accepts no caller qualifiers. GitHub checks `total_count`,
`incomplete_results`, bodies, duplicate identities and next-page signals, bounded by
the provider's ten-page/1,000-result search window. GitLab walks its existing bounded
issue-page window, checks continuity and rejects duplicate or missing pages. Neither
adapter retries through a different identity or permission scope.

The observation owner is in `deskkit`; `deskboard` owns only bounded scheduling and
presentation. This is an in-process step toward a shared read view, compatible with
desktools-v2's typed query boundary and the proposed shared-observation architecture.
It does not claim the future mirror, service, custody transition or event projection
has shipped. Downstream actors retain independent authority checks before effects.

## Matched perf-v3 experiment

The repeatable `TestAdmissionPerfV3` fixture runs the actual admission path, native
GitHub adapter and perf-v3 process wrapper/writer/reader. Only the remote forge is
replaced. It supplies twelve distinct, human-gated brief rows in one repository,
complete unruled decision histories, one marker-discovery page and 8 ms of injected
latency per HTTP request. All twelve rows must remain held with zero unknowns.

The baseline is production source `cfe5c927ec7049728f4b97306beaedf43a444b9f` with only
the identical test harness added. The repaired production source and harness are
`b2d0b9355b7b6d07d86c4baa396ff3ea009107df`. Subsequent publication changes to the
mutation specification and this report do not change either measured path.
Both local test binaries carry their respective source stamp. Five independent
processes per variant were run in alternating before/after order; `-test.count=5`
in one process is unsuitable because process-init time would accumulate.

| Quantity | Before | After |
|---|---:|---:|
| HTTP attempts, every sample | 60 | 37 |
| Search attempts, every sample | 12 | 1 |
| Other REST attempts, every sample | 36 | 24 |
| GraphQL attempts, every sample | 12 | 12 |
| Process wall time, ms | 573, 557, 545, 581, 586 | 79, 75, 71, 79, 80 |
| Median process wall time, ms | 573 | 79 |
| Own process CPU, ms | 30, 23, 20, 51, 62 | 23, 19, 21, 18, 30 |
| Median own process CPU, ms | 30 | 21 |

That is **38.3% fewer HTTP attempts** and **86.2% less median elapsed time in this
latency fixture**. It is not a production speedup claim. Small CPU samples and their
variance do not establish a general CPU improvement. Unique explicit issue references
save no discovery calls; repositories needing more pages or having many shared
decisions have different savings. The full CLI's root/queue/review work is outside
this bounded experiment and is not subtracted from its production telemetry.

Native request observations exactly match the fixture's independent HTTP witness.
The per-process records have measured request coverage and no delegated work. REST
points are unsupported; GraphQL points remain unknown on this local endpoint. The
unchanged twelve GraphQL requests are not reported as zero-cost or measured savings.
No wall/CPU totals from overlapping parent/child records or legacy schemas are added.
Telemetry consent is confined to test processes and their disposable fixture homes;
no operator consent, workflow grouping, audit ledger or live pin changes.

Binary SHA-256 values:

- Before: `ffebc60bf0fa6f942a0c490d5249903e091169ac8dbc4bcdfae6518341c741a8`
- After: `ce77f10e89b2c4dc2594a61200228e1bbe9cf2f97a1e0380c296c19036b20f28`

To reproduce from each respective source tree with the same test file:

```sh
cd tools/desk
go test -c ./cmd/deskboard -o /absolute/scratch/board.test \
  -ldflags '-X github.com/medici-finance/assay/tools/desk/internal/deskkit.SourceSHA=<full-source-commit>'
ASSAY_TELEMETRY=1 /absolute/scratch/board.test \
  -test.run '^TestAdmissionPerfV3$' -test.v
```

Each invocation prints one schema-qualified `ADMISSION_PERF_V3` record. Repeat as
fresh processes, alternating variants. These are local builds, not installed releases.

## Regression qualification

Tests compare shared observations with the single-brief evaluator and independently
assert the required ruling outcomes. They cover overlapping concurrent readers,
fresh rulings on a second sweep, both brief-id spellings, current marker membership,
lowest issue identity, explicit references, repository isolation, incomplete discovery,
both native adapters' pagination, execution tier before remote work, and stable output
under deliberately different completion order.

The dedicated classic mutation harness is:

```sh
cd tools/desk
go run ./cmd/muhar -j 2 -spec cmd/deskboard/decision-admission-mutations.json
```

Existing review snapshots, conditional transports, per-PR reuse, quota floors,
root resolution, queue semantics, trust controls and perf-v3 observation machinery
are retained. No existing test is retired.
