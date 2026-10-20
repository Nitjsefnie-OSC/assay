---
id: DR-desktools-v2-63
date: "2026-10-20"
title: "Define local perf-v3 attribution: explicit `tool` and `workflow` profiles over one existing writer; only `workflow` admits bounded, typed, self-asserted local work and agent linkage; readers keep v1/v2 groups separate, dedupe by install and invocation, and never merge measured, estimated and unknown points; no sender, service, spool or credential read"
consequence: major
decided-by: "human:<name>"
alternatives:
  - "Option 2 on the brief's decision issue, \"Hold for correction — identify the field, lifecycle or failure boundary that needs correction\" — not chosen: the driver answered 1 (https://github.com/medici-finance/assay/issues/3536#issuecomment-6278971374). No rationale was recorded, and none is supplied here."
  - "Larger work lists, or request-detail records with unbounded sample storage — not delivered: the delivering pull request's Desk-decided entry bounds the wire contract to eight work references, 4096 input bytes, 64 request groups and 32 log2-millisecond duration bins instead. This was the desk's decision on the pull request, not the driver's; the driver's ruling does not mention it."
  - "Resolving identity by searching branches, files, Git, process ids or a native harness — ruled out by the brief's task 2 (credential-free resolution with a random per-invocation identity). The delivered `ResolveContext` reads only the named environment keys and the adapter's parsed flag values."
  - "Merging measured, estimated and unknown cost points, or treating missing cost as zero — ruled out by the brief's task 3. The delivered reducer keeps units and coverage states apart and rejects a measured point with no value."
  - "A sender, spool, alias mapper, service or remote activation in this brief — out of scope: option 1 itself retains separate remote activation controls, and the brief's design-fit admits \"no new sender, service or credential holder\"."
accepted:
  - "Under the `workflow` profile, records hold identifiable local data: qualified brief and issue references and opaque session, agent and parent ids, kept for seven UTC days in 0700 directories and 0600 files. `ASSAY_PERF_PROFILE=workflow` also grants reader permission, so a shell or agent that inherits it can print those references by choosing a workflow grouping or filter."
  - "All attribution is self-asserted and unauthenticated. A role or work label in a performance record is never an identity proof and never drives a gate, budget, custody, claim or completion decision."
  - "The wire bounds (8 work references, 4096 input bytes, 64 request groups, 32 duration bins) become a persisted contract: changing them after adoption needs versioned compatibility handling."
  - "The pre-append symlink and regular-file check is a best-effort `Lstat`-then-open sequence, not atomic protection against a concurrent process of the same operating-system user replacing a file between the two."
---

**This record transcribes a recorded human ruling and states the delivered design. It is
NOT approved.** The ruling below answers option 1 on the brief's decision issue, but its text
names neither this record nor the brief, so neither the design-approval gate nor the
decision-gate hold can use it. The design approval is pending on its own decision issue:
https://github.com/medici-finance/assay/issues/3597
Until the driver's comment approving this record exists there, nothing here claims the design
was approved.

## The ruling recorded

The brief's decision issue is medici-finance/assay#3536, "Decision: assay:assay:desktools-v2:63 —
Define perf-v3 attribution, profiles and compatible local readers". It carries the marker
`<!-- decision-gate: assay:assay:desktools-v2:63 -->` and offers option 1, "Accept the bounded
implementation — accept the demonstrated local disclosure and isolation; retain separate
remote activation controls", and option 2, "Hold for correction". Its default is "none —
blocks until answered".

The driver's own comment (type User, 2026-10-19T22:42:30Z, not edited), verbatim: https://github.com/medici-finance/assay/issues/3536#issuecomment-6278971374

> 1

The issue was closed three seconds later and still carries `needs-decision`. The comment is
its only comment (read 2026-10-20). The ruling's text names no record and no brief.

**Timing.** The ruling was posted before the reviewer's CHANGES_REQUESTED review of the
delivering pull request (submitted 2026-10-19T22:45:03Z at `d36f0674`). It was also posted
before the current head `5ed262fe` was committed (22:57:01Z). The ruling does not say which
head it accepts.

**Option labels.** The ruling's "1" matches option 1 on the decision issue, and the brief's
own Human decision uses the same numbering and wording. This record states that design as
option 1. It does not extend the ruling to anything the option does not name.

**Why this record carries no `ruling:` key yet.** `decided-by` is the public placeholder
`human:<name>`. Under `spec/registers-v1.md` §7.5's placeholder rule, a placeholder names
nobody: it is never looked up in a login map and never corroborated by a pull-request
approval. Its only route is a `ruling:` link to an unedited human comment whose own text names
this record by its id, posted on a decision issue that carries this record's decision-gate
marker. The comment above fails that test: it names no record (condition 5), so a `ruling:`
pointing at it would fail closed with `record-not-named`. For the same reason the brief's
decision-gate hold (`spec/lifecycle-v1.md` §4.5, layer two) refuses it: that check needs the
comment's text to name the brief by its board id, `desktools-v2/63`.
The decision issue for this record's approval carries the marker and asks the driver for a new
comment naming the record. The `ruling:` key is added to this record, by a later push, only
after that unedited comment exists. It links only the driver's own comment choosing option 1
and naming this record. A comment from any other login, or one that holds or rejects the
design, never qualifies. Until then this record is not approved, and the corroboration check
correctly reports `decided-by` as `placeholder-unratified`.

## The design stated for approval

The brief is `docs/streams/desktools-v2/brief-63-define-perf-v3-attribution-profiles-and-compatible-local-readers.md`.
Its delivering pull request is medici-finance/assay#3550 (draft, head `5ed262fe`). The frozen
contract is the V3.md file that pull request adds under perfrec. What it delivers, read
from that diff:

1. **Explicit profiles, opt-in only.** `ASSAY_TELEMETRY=1` with no selected profile keeps each
   producer's existing schema. `ASSAY_PERF_PROFILE=tool|workflow`, or its parsed flag, selects
   `perf-v3` only together with that exact telemetry switch. Empty or `off` falls back to the
   legacy schema, and an invalid profile disables v3. No profile sends anything or changes an
   operation's authority or result.
2. **Bounded, typed attribution.** A v3 line keeps the v2 interval fields unchanged and adds one
   `v3` object: `profile`, `attribution` (a fresh 128-bit random `invocationId`; optional
   session, agent, parent agent and parent invocation ids, `deskRole`, assigned `role` and
   `work`; and coverage for every optional field), one `observationKind`, at most 64 request
   groups, request coverage and drops, and an optional query observation. The `tool` profile
   requires every workflow field absent and its coverage `unsupported`, with top-level role
   `none`. Only `workflow` admits identity and work. `work` holds at most eight typed brief or
   issue references (at most one primary, 4096 input bytes in all), qualified by repository
   and, for an issue, by forge and instance. Raw URLs, free strings and unknown keys are
   invalid.
3. **Resolution owned by perfrec, never by the operation.** `ResolveContext` reads only the
   named environment keys and the adapter's parsed values. A flag replaces the environment,
   a present-but-empty value clears it, and an invalid value becomes `invalid` coverage with
   no copied input. The writer and the decoder each validate a hand-built record
   independently of the resolver. A failed write is swallowed and the operation's exit code
   is returned unchanged. `deskkit.PerfRunMeasured` passes the session to the parser callback
   explicitly; no process-global attribution or request sink exists.
4. **Request and query observations carry metadata only.** A request observation records
   class, origin, protocol, resource, outcome, duration coverage and optional bytes and cost
   points from fixed vocabularies. No URLs, bodies, GraphQL text, keys or credential material
   are representable. Point units stay distinct, estimates never merge with measured points,
   and missing cost is never zero. Only a client counts logical queries.
5. **Readers keep meanings apart.** `ReduceV3` deduplicates by install and invocation before
   filtering. An exact replay counts once, and a conflicting replay is exposed and excluded.
   Grouping or filtering by work, session, agent, desk role or assigned role needs workflow
   permission. v1 and v2 remain separate report groups, and the legacy `--role` filter keeps
   its meaning and excludes v3. `deskperf report --json` adds `perfV3` only when v3 data or
   conflicts exist, and makes no provider calls.
6. **Local files only.** Seven UTC days of retention, 0700 directories and 0600 regular files,
   with the best-effort symlink check stated under `accepted`. There is no sender, alias
   mapper, spool, service, new credential read or operational accounting. A future importer
   must obtain workflow consent and alias raw local references independently.

The brief and its delivering diff are the authority. This summary does not amend either.

## Consequence level — transcriber's classification

`consequence: major` is not part of the ruling, which states no level. The transcriber took
it from the brief's risk answers (`sensitive-data: yes`, the other three `no`, including
`irreversible: no`) and its `gate-why`: "Workflow attribution exposes local work linkage". A
wrong design could keep work or agent linkage on disk that opt-out or tool-only use should
not keep. The records stay local, and a route back exists, though changing the persisted
bounds after adoption needs versioned compatibility handling. The driver can correct the level
before the record's approval is linked.

## Open / not decided

- **The design approval.** It is the driver's, and it is pending on https://github.com/medici-finance/assay/issues/3597
  It counts only as the driver's own unedited comment there naming `DR-desktools-v2-63`,
  linked from this record's `ruling:` key. A pull-request approval, a desk relay or a reviewer
  verdict does not stand in for it.
- **The review of the delivering pull request.** The last review of medici-finance/assay#3550
  requested changes at `d36f0674`, with two blocking findings. F-perf-mutation-target: two
  mutation targets (the review suite's perf test, and desktools-v2/21's mutation map) name
  `perf.go` text the pull request removed. F-lifecycle-lint: the brief moved to `implemented`
  with no `design:` record and no `decision-issue:`. At head `5ed262fe` the pull request's
  diff changes both mutation files and holds the row at `blocked` with `decision-issue: 3536`,
  but no review exists at that head. Whether the findings are resolved is not decided here.
- **Independent verification.** The brief's Verify table (TestPerfV3ContextAndConsent,
  TestPerfV3RecordToReportFlow, TestPerfV3IsolationAndOverhead and the consumer
  corroboration) runs after merge by a non-implementer. This record attests none of it,
  including the overhead figures in the pull request's body.
- **Integration and rollout.** CLI and launcher propagation (desktools-v2/64), own forge
  request timing (desktools-v2/65), desk hot paths (desktools-v2/62) and the perfdata
  importer are later briefs' work. Any remote activation stays under its own controls, per
  option 1.
- **The brief's human sign-off.** desktools-v2/63 is `gate: human`. Its `verified` and `done`
  flips need the human sign-off on a Verify run of the merged implementation. This record is
  a design record only.

## What happens after merge

**Merging this record alone authorizes no flip.** A change that moves desktools-v2/63 to
`implemented` and adds `design: DR-desktools-v2-63` to the brief's frontmatter waits for ALL of:

1. the driver's own unedited comment on https://github.com/medici-finance/assay/issues/3597
   approving `DR-desktools-v2-63` (option 1, naming the record);
2. a push adding this record's `ruling:` key, linking that comment;
3. a clean `statusgen --corroborate` read of this record.

The brief's decision-gate hold needs its own ruling as well. The brief's `ruling:` must link
an unedited comment from the driver, on its recorded `decision-issue:`, whose text names
`desktools-v2/63`. The decision issue for this record carries both markers. One comment
there naming both ids can therefore serve both checks, if the brief's `decision-issue:`
records that issue.

No tool enforces these conditions. statusgen's ruling check (`statusgen/decisionruling.go`)
accepts a comment from any login the human-login map resolves to a human when `decided-by` is
the placeholder, and it does not read whether the comment approves or rejects. A passing
`lint` on the flip change is not evidence of approval either. The design gate checks only
that the record exists (https://github.com/medici-finance/assay/issues/2704), and the
corroboration job skips a change that touches only a decision record
(https://github.com/medici-finance/assay/issues/2669).
