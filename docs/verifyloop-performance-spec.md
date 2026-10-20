# Verification planning performance and diagnostics

**Status:** draft — 20 October 2026. Extends the decision-oriented workflow direction in
[desk-workflow-efficiency-spec.md](desk-workflow-efficiency-spec.md) for verification.
That source's approval does not approve this extension. The accompanying diagnostic code is
explicitly requested implementation; the read-reuse changes below remain proposed. No
runtime, cache service, collection, credentials, release or consumer cutover is activated.

**Inspected source:** public Assay main `874e509a5319ae6fb94225a3d9c88e3fc34864ed`.
Material paths: `tools/desk/cmd/verifyloop/{main,adapter,briefscan,blockerread,pending,plan,verdictrun}.go`,
`tools/desk/internal/deskkit/{verifierattestation,workerattestation,stampapply,forgeobserve,perf}.go`,
`tools/desk/cmd/deskboard/{readcost,prfacts}.go` and `perfrec/`.
Refresh source, active work and supported pins before implementation.

## 1. Outcome and preserved behavior

Reduce the work required to produce a complete verification plan, especially provider
round trips and their waiting time, while producing the same scheduling and authority
answers from equivalent observations. Measure CPU, wall time, provider attempts and
provider quota cost separately. Request counts are not GraphQL points or money.

The first useful boundary is one `verifyloop plan` invocation: startup and envelope checks,
root/brief/outcome reads, wake/blocker evaluation, pending-Evidence observation, optional
canonical base classification, and rendering. Verifier row execution, signing and Evidence
publication are separate boundaries. A faster plan does not establish faster completed
verification; later qualification must include avoided runs and publication turnaround.

Preserve the current owners and their meanings:

- `briefscan` and `queueclass` own the existing wake/hold and overlay behavior. The canonical
  queue classification remains statusgen's; `--base-queue` can only demote overlay dispatch.
- Pending Evidence must be independently authored, attested, source-current and complete.
  Retain exact plan/source bindings, declared-directory membership, blocker wake checks and
  the post-read PR-head reconciliation. Listing/file-count caps and uncertain reads stay
  visible; unavailable is not absent, unchanged, closed, green, dispatchable or idle.
- Preflight, stop flags, allowlisted coordinates, verifier custody, author/runner separation,
  claims and dispatch/landing revalidation remain. `--no-forge` is an unavailable-state
  diagnostic, not an optimization mode. No preflight bypass or changed polling cadence.
- Planning grants no permission to execute, publish, flip, merge or accept. Human and
  Evidence-only boundaries, runner isolation and consequential current-state checks remain.
- Text/JSON output, exit codes, root coverage and optional recording consent remain compatible.

## 2. Brownfield findings and quality treatment

These are source observations, not a measured savings claim:

| Existing behavior | Cost implication and preservation |
|---|---|
| `forgeIssueSource` resolves each repository once and caches resolution errors per pass. Pending discovery caches the default branch and open-change population; PR metadata and complete file lists are also cached per change. | Preserve these caches; another discovery cache would duplicate existing work. |
| `IssueState` has no result memo. Multiple held briefs can name the same qualified blocker. | Repeated observations scale with referring briefs rather than unique blockers. |
| `ReadVerifierAttestation` reads the issue, calls `attestationRecord` which reads it again plus trust history, then `StampTimelineFor` which reads it again plus label history. | A successful one-page case entails three issue-object reads and at least two history requests. Reusing the object could remove two requests, but the immutable-body and dispatcher-stamp checks remain. |
| Pending Evidence reads SHA-bound brief/outcome content and visits each Evidence binding. Its item-to-change join repeatedly walks an already cached change/file population. | Reuse immutable bytes and parsed results; index brief-to-change membership once. The index alone saves local join work, not the already cached provider file reads. Historical bindings cannot be discarded by an unproven newest-only heuristic. |
| `deskboard` has conditional GETs, cost metering and per-PR reuse in its command package. verifyloop does not inherit that behavior. | Reuse requires a reviewed extraction into an importable desk-internal package; importing another command's `main` package is not a supported seam. |
| verifyloop's existing wrapper passes no verb roster and has no phase roster or explicit scheduling request observer. | `plan` and `verdict` collapse to `other`; whole-invocation counts do not locate the expensive path. |

Recent public history includes pending-Evidence admission (#3277), implementer independence
(#3456), complete declared input scopes (#3429) and queue integration (#3376). These changes
nominate the join/admission seams for careful qualification; they do not justify a broad
rewrite. Preserve their real regression cases and populated Evidence. The quality snapshot
under `docs/quality/` is historical input, not a fresh complexity or reliability score.

Remediate repeated observation at its owning seam. Keep credential extraction, persistent
caching, parser migration and service rollout in their existing work packages. No public,
matched fleet timing/quota baseline accompanies this draft; establish a reproducible
synthetic baseline before claiming a speedup or choosing a numerical latency target.

## 3. Fixes that can ship on the supported local path

These need the current desk/perf APIs and normal review/release, not PostgreSQL, a mirror,
a service listener or new credentials. Implement and qualify them in bounded follow-ons.

### One planning-pass observation

Reuse `forgeIssueSource`, `VerifyLoop.SelectQueue`, the typed Forge seam and `planDoc`.
Collect observations under the existing allowlist/custody boundary; keep evaluation over
values. Do not build another scheduler, queue classifier, credential owner or fact ledger.
A pass is an observation set, not an atomic multi-repository snapshot.

Memoize blocker responses by qualified forge/host/repository, kind and number inside that
pass's admitted reader. Canonicalize equivalent references; never collide an issue and a
change or two repositories. Cache success and unavailable/error results for that pass;
retain caller-specific diagnostics without retaining arbitrary error text as a cache key.
A fresh pass refreshes mutable state. Dispatch and publication keep their own rechecks.

Refactor attestation validation to accept one typed issue observation plus the required
body-history and stamp observations. Reuse `attestationRecord`'s validation and
`StampTimelineFor`'s stamp meaning rather than copying predicates into verifyloop. First
prove that consolidating object reads cannot admit an edit, reopen or label movement that
current behavior rejects. If a later current-state read closes a demonstrated race, retain
it and qualify the smaller saving. Memoize admitted receipt results and failures within the
pass using qualified target and validation/policy context. An issue is mutable even when
its contract requires an unedited body; it is not a permanent immutable cache entry.

Reuse file bytes and parse results keyed by admitted repository, exact commit, path and
parser/validation version. Cover regular-file proof, missing content, directory membership
and corrupted inputs. No branch-name or current-check-result caching under a SHA key.
Initially keep this memo invocation-local; cross-process immutable reuse is the existing
source-query owner's later work. Build a brief-to-pending-change index once from complete
cached file lists, carrying unknown-file-list uncertainty conservatively. Preserve final
PR-head/state reconciliation and the existing rule that untrusted changes cannot hold work.

Do not batch provider requests until their individual duplication and coverage are known.
A later typed batch read may replace several reads only with per-target completeness,
author/stamp/history parity, bounded pagination and actual point-cost comparison on both
supported forges. A larger GraphQL operation is not automatically cheaper.

### Shared conditional transport, separately qualified

Extract the reusable mechanics from `deskboard/readcost.go` before adopting them in
verifyloop. Preserve deskboard's existing floor, atomic cache files, corrupt-entry misses
and leaf accounting. Keep command policy outside the shared transport; no global behavior
change for other tools. Scope stored validators/bodies by the current admitted provider,
repository and credential/access context so one role's cache cannot disclose another's data.
An authorized 304 may reuse the validated body; absent/invalid validators, access loss or
uncertain responses cannot authorize a stale answer. No TTL-only mutable-state hit.

Conditional GETs reduce charged REST work and payload; they still send requests and incur
round-trip latency. Test request-count and quota effects separately. GraphQL cost retains
its own unit and coverage. This extraction can ship before the future service, but follows
a cache-scope/compatibility contract and both consumers' equivalence tests.

## 4. Diagnostics implemented with this proposal

Use the shared `perfrec` perf-v3 contract ([API/wire reference](../perfrec/V3.md)),
writer/session and `deskkit.PerfRunMeasured`; add no spool, exporter,
metrics server, second schema, raw request logger or per-repository metric labels. The shared
wrapper accepts an optional fixed phase roster; callers without it remain unchanged.

verifyloop records the closed operations `plan` and `verdict`; both existing dry-run
shorthands select `verdict`. Unknown operations remain `other`. Record fixed phases:
`startup`, `guard`, `preflight`, `queue-scan`, `blocker`, `pending-discovery`,
`pending-evidence`, `attestation`, `base-queue`, `classify`, `render`, `rows` and `sign`.
The shared recorder reports invocation count and summed milliseconds per observed phase.
Nested phase time is inclusive and marked overlapping; never add it to process wall time.
Queue scanning includes wake/association work, and signing includes payload composition
and output. Absent phases mean that boundary was not observed, not zero work.

Select the existing environment-resolved diagnostic context before provider work. Compose
`RequestObserverFor` into the planner's existing sole `ForgeFor` call as the verifier role.
The shared counted transport observes actual sends, response-body lifetime, bounded outcome
and protocol/resource groups, and eligible same-operation GraphQL cost metadata. No monitoring
request, endpoint URL, query text, credential, brief or blocker becomes a diagnostic label.
The schema's actual/unknown/failed/unsupported cost states remain intact.

Existing `ASSAY_TELEMETRY=1` without a new profile keeps perf-v2 with improved verbs/phases.
Explicit tool/workflow profiles use existing perf-v3 consent and attribution. Workflow
context comes only from the documented resolver; this change adds no parser flags or new
context propagation. The separate Cobra migration keeps CLI override integration.

Coverage is deliberately partial: scheduling-provider reads are composed; preflight's
separate transports, credential subprocesses, row commands and native/verifier children do
not become fully observed by inheritance. Keep invocation request coverage missing unless
the entire own-request inventory is established. Children need their own context and
instrumentation. Cache and waiter counters remain unavailable until their owning cache
seams supply measurements. Per-phase provider-request counts are not in the existing
record schema; timed phases and aggregate own-request histograms are distinct observations.

Recording is off by default and local only; retain permissions, seven-UTC-day retention,
legacy reader separation and best-effort sink failures. Enabled/disabled runs must preserve
stdout/stderr, exit codes, provider business selections and send counts. The existing
provider adapter may add qualified cost metadata in the same request; unsupported schemas
remain unchanged. The instrumentation is not a performance optimization or release proof.

## 5. Fit with existing delivery

| Work/semantic owner | Treatment of this extension | Dependency and retirement |
|---|---|---|
| desk-tools/23–25; desktools-v2/61, /63–65 | Reuse recorder, audit/token improvements, contexts, provider observer and mixed readers. The accompanying implementation is verifyloop consumer integration. | No reimplementation or reopening their completed Evidence. Installed pins must include the required APIs. |
| desk-tools/37–44 workflow-efficiency program | Reuse matched-cost/oracle method and observation/evaluation boundary. Allocate verifier-specific local follow-ons after approval. | Review-only Task/Verify scopes are not silently expanded; no dependency on their unshipped review CLI. |
| desktools-v2/62 | Reuse the matched hot-path qualification owner and report grouping; include the verifier cohort when a citing follow-on is authored. | Local read fixes can be tested independently; full combined startup/credential claims require their actual prerequisites. |
| desktools-v2/58–60 | Credential-provider extraction, concurrent lookup sharing/expiry and activation-discovery reuse remain separate startup improvements. | Not prerequisites to blocker/attestation/file reuse or current diagnostics. No token custody moves into perfrec/statusgen. |
| desktools-v2/40 | Existing verifyloop/reviewloop Cobra/Viper migration owns parser compatibility and explicit performance overrides. | This change preserves the current parser and does not consume its unfinished CLI. |
| desk-supervision/16, /24 | Preserve wake conditions and per-file outcome records. | No change to their accepted semantics or historical Evidence. |
| desk-supervision/26 | Reuse proposed Evidence-only batches for publication savings. | Independent of planner reuse; qualify per-entry gates, batch freeze and added publication delay. Status-flip landings retain their own authority. |
| forge-neutral/37–38, /39–45 | Shared observations, refresh coalescing/events, authenticated serving, write-through receipts, SHA queries, indexes and qualification retain these owners. | Future integration needs their real R1/R2/R3/R4 prerequisites; no new mirror/cache-service project. |
| performance-collection/01–02 | Reuse optional diagnostic ingestion/export only when separately qualified. | Local diagnosis never waits on remote collection or its database. |

This is scope routing, not an executable wave plan or approval of changed briefs. Author
citing briefs with the existing author-brief process after the design is approved; preserve
active checkpoints and completed Task/Verify/Evidence. The immediate diagnostics are already
requested; larger behavioral changes are not implemented by this proposal.

## 6. Future-state dependencies

The [non-CI delivery sequence](proposals/non-ci-architecture/16-delivery-sequence.md) gives
R5 the complete verifier use case and R6 shared forge observations. The
[forge-service extension](proposals/non-ci-architecture/22-forge-service-and-source-queries.md)
provides the existing destination for cross-process reuse. Migrate the local observation
adapter to that qualified contract; preserve pure evaluation and role-bound effects.

R6 must supply admitted scope, freshness, pagination/coverage and access revocation,
coalesced refresh, polling repair and authenticated event hints. Events request reevaluation;
they are not authoritative current-state or acceptance records. A service outage cannot fan
out into one full upstream scan per consumer. Missing coverage holds affected work.

Persistent SHA-query reuse and dependency-aware invalidation remain forge-neutral/42–43.
Shared Git objects/fetches remain /44's go/no-go trial, not a new daemon prerequisite.
R4 publication and confirmed write receipts retain the effect-ledger/custody owners; cached
observations grant no write authority. Cross-process mutable reuse and whole-repository
verifier cutover require the accepted runtime, revocation/recovery proof, supported versions
and explicit operator activation. None is needed to ship the local fixes in section 3.

## 7. Qualification and acceptance

Freeze equivalent observation inputs, policy/time, source SHAs, module versions, root set
and output schema. Compare dispositions, reasons, authority, prompts, ordering, gaps,
completeness and exit codes against current production logic. Reuse existing wake, pending,
cross-repository, risk and canonical-queue fixtures; add skewed/failed cases at the changed
boundary. Provider reads are observations at different times: explicitly classify races and
intended observation reuse instead of pretending to preserve an atomic snapshot.

| Scenario | Acceptance and evidence owner |
|---|---|
| Many briefs share one blocker; one differs by repo/kind; one read fails | Per-pass request count follows unique qualified targets; repeated failures do not resend; a fresh pass sees closure/reopen. Local memo follow-on owns the call-count oracle. |
| Reused attestation object; edited/reopened record; moved labels; incomplete history | Same admission/refusal meanings, all required history checked; retain any proved necessary consequential reread. Attestation owner supplies production-path race/fault fixtures. |
| Pending batch with overlapping files, historical bindings, moved head, unsupported author, partial file list or new declared-directory member | Same held/dispatch/unavailable populations and exact-source checks; trusted complete listings alone may establish absence. Pending-join follow-on owns index equivalence. |
| Cold/warm immutable reads, corrupt bytes, two admissions and missing path | Correct typed bytes/results or explicit uncertainty; no cross-admission hit or fabricated absence. Source reader/cache owner proves keys and failures. |
| Conditional 200/304, denied access, changed validator, corrupt cache and floor breach | Actual sends, quota units, source-current results and refusal behavior preserved separately for both consumers. Shared transport extraction owns qualification. |
| Recording off/on; tool/workflow profile; unknown command; both dry-run aliases; failed sink | Same operational output/exit and sends; closed verbs/phases and bounded own-request observations only. Accompanying diagnostics tests exercise production paths. |
| Local read optimization versus full future cutover | Local synthetic parity needs no live service. Served/runtime/custody/recovery acceptance remains with R5/R6 and their qualification owners. |

Use matched repeated fake-provider runs with identical response bodies and controlled delay;
report cold/warm cohorts, root/brief/PR/binding/unique-blocker counts and uncertainty. Alternate
baseline and subject, including telemetry off/on overhead. Assert request-count scaling,
not a brittle wall-clock test. Publish wall/CPU/request/cost distributions separately with
source/configuration/coverage and uncertainty. No universal latency budget or percentage
saving is selected here; the citing qualification owner proposes targets from the matched
baseline for maintainer acceptance. Existing passive history cannot prove causation.

Qualification must observe the real adapter/reader/record/report flow, including JSON parseability,
legacy histories and sink failure. Prove at least a relevant planted regression is detected
(e.g. removed provider observer, missing phase roster or alias attribution, memo-key collision,
failed-read retry storm, lost final-head reconciliation). A helper-only cache benchmark is
insufficient. No live production access is needed for development qualification.

## 8. Decisions, rollout and recovery

Recommended first slice: accept the bounded diagnostics, then approve local observation reuse
with no persistent mutable cache, no added upstream concurrency and unchanged preflight.
Review the separate shared conditional-transport extraction when its scope contract and
matched results are concrete. These are review recommendations, not recorded approval.

Release the diagnostic producer with the existing desk/perf dependency closure; adopt normal
consumer pins and independently verify the installed binary. New phase/verb labels do not
rewrite old records or make the old `other` cohort separable. Reverting caller instrumentation
restores the old diagnostics; disabling the opt-in restores unrecorded operation. No data or
authority migration is needed. Read-reuse changes require their own semantic/cost acceptance;
service adoption later requires R5/R6's separate release, cutover and rollback contract.

Maintainer decisions still needed: approval of the behavioral extension and any persistent
cache scope/retention contract. Numerical production targets and an operational observation
window are qualification decisions after a matched baseline; no additional operator action
is required to review this draft or run its offline diagnostic tests.
