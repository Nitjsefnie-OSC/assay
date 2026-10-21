# Ordered attestation batching for verifyloop

The pending-Evidence reuse change removed repeated reads of the same attestation, but
one distinct successful attestation still required four provider requests. In the
qualified four-root population, 31 distinct attestations accounted for 124 of 187 total
GitHub requests. This change collects that distinct population before validating it.

## Composition and authority

`VerifyLoop.SelectQueue` supplies the existing scheduling source with its eligible brief
paths. Collection itself makes no network request. At the first pending-Evidence read
for a repository, preparation uses the same admitted Forge, complete PR/file listings,
trusted independent verifier-author/source predicates and exact-commit file cache as lazy
evaluation. Only references in relevant verifier-authored brief files enter the batch.
Historical bindings are retained; source age or a newest-only heuristic cannot establish
that an old failed/blocked receipt is irrelevant. An incomplete file listing still
contributes uncertainty in ordinary evaluation.

`deskkit.VerifierAttestationBatchReader` supports the native GitHub backend and its known
outbound decorator. It does not unwrap arbitrary wrappers or read-only credential fences.
GitLab and unsupported decorators keep the existing lazy per-target reader. The batch
reader uses `GitHubForge.doJSON`, the existing go-gh client, custody and request observer.
It introduces no credential owner, service, transport setting or global cache.

For each chunk of at most 32 distinct issue numbers, the reader performs four separate
GraphQL operations, in order:

1. Typed issue observations, exact title/body/author identity and complete current labels.
2. The canonical issue-trust selection and reducer, including explicit body-edit provenance
   and the existing bounded comment-history completeness rule.
3. New issue observations after body history, including current body, closed state and labels.
4. Label application/removal history after current labels, including actor identity.

An invocation-local adapter supplies these values to the **unchanged
`ReadVerifierAttestation` validator**. Exact target/title/body/run/source/brief/model/tier,
unedited-body provenance, closed state and dispatcher stamp checks retain their shared
owners. The adapter cannot send an unexpected extra request. Errors and missing aliases
are target-specific uncertainty; unattributable GraphQL or transport errors stop further
batch requests. Provider error messages are not echoed. An overflowing label connection
uses the existing complete native reader; its errors never become empty labels/history.
The second issue observation is never replaced by the first.

Preparation stores receipt results and errors in the existing per-pass receipt map.
Normal per-item evaluation still checks the exact Evidence binding, original/current
brief inputs, outcome records, dependency membership and wake predicate. Every proposed
hold still reconciles live PR head/base/state after file evaluation. New selection clears
all mutable observations. Dispatch, execution and Evidence publication retain their
independent admission and current-state reads.

## Instrumentation and qualification

The existing fixed 16-phase perf-v3 roster is retained. Preparation runs inside the first
per-item `pending-evidence` scope. `attestation` times an admission-reader invocation:
a native bulk invocation can cover multiple records, while the fallback remains one
invocation per record. Its `n` must not be interpreted as the number of records or requests.
Physical attempts, bytes and actual GraphQL points remain the transport observer's facts.
The three fixed batch documents join the existing closed cost-query inventory; targets
are variables and no provider text expands that inventory. Legacy consent and unknown
custom/enterprise-schema cost behavior are unchanged.

Qualification compares parent and subject on the same four frozen roots, with normal
preflight and verifier custody, native default HTTP settings, the same total-request
counter and warm credentials. The complete plan, admitted roots, physical GitHub attempts
(including identity and Git preflight), wall time, CPU and point coverage must all be
reported. The integrated measurements will be recorded here before publication.

The collection step can observe more relevant files up front than a successful early
lazy hold would. That bounded tradeoff is explicit: it excludes unrelated/non-verifier
files, keeps the original evaluation order, and qualifies positive holds separately from
the zero-hold live population. No local candidate-pruning heuristic was added.

Focused regressions cover per-target errors, missing provenance, edit/reopen/relabel
changes across stages, pagination fallbacks, scope fences, bounded chunks, candidate
scoping, missing batch results, pass refresh and fresh PR-head reconciliation. The
mutation recipe is `tools/desk/cmd/verifyloop/pending-batch-mutations.json`.
