# Issue discovery projections

The issue reader selects a population from the complete typed open-issue listing
before paying trust, comment and label-history reads. Collection stays in the
existing role-bound forge adapter; pure selection lives in
`tools/desk/internal/issueflow`. No cache, service, new credential or pool is added.

| Caller | Producer | Population and required enrichment |
|---|---|---|
| Worker skill and dispatch runbook | `issueboard issues --projection work-ready` | Unrepresented candidates without excluded/decision/addressed/help-wanted labels. Selected external authors still need current blessing and numeric identity; selected work still needs complete label history. Missing active placeholders retain positive state reads to detect a partial listing. A history hold is AWAIT, never dispatchable. |
| Intake skill | `issueboard issues --projection intake` | All open issues and active closed placeholders, with original trust/quarantine, escalation, addressed-response, label-history and positive closure behavior. `issueboard intake` separately reads local raw intake. |
| Addressed desk inbox | `issueboard issues --projection inbox --to ROLE` | Only the named addressee, selected before trust and comment enrichment; retains complete escalation histories and conservative overflow. Role aliases remain the existing addressee owner's responsibility. |
| Human decision inbox script | `deskinbox discovery REPOS...` | Case-insensitive, space-trimmed union of urgent, needs-decision, question and help wanted, ordered urgency then creation age. This is distinct from the addressed-role population. Selection precedes detail; the script retains its existing detail, positive ruling screen, cards, flow and partial-result exit semantics. |

Projection output declares population, current invocation cutoff, enumeration
completeness and row-level history degradation. It is an observation set, not an
atomic provider snapshot, and has no validity beyond the collecting invocation.
An empty worker result cannot establish intake/inbox idle. Excluded external issues
are absent from the worker projection; full/intake still show their quarantine. Failed complete-list
reads fail the whole issueboard projection; existing row-level history overflow
still degrades conservatively. The human inbox preserves partial output plus exit 2.
Neither observation grants write, claim, admission or executor authority.

## Compatibility and rollback

| Plugin/script and producer | Supported path |
|---|---|
| Older plugin with new producer | Legacy full/default/`issues`/`intake`/`--to` forms remain available, with unchanged tables and exit behavior. |
| New plugin with matching producer | Check `issueboard --version` for `issue-projection-v1` before selecting projected forms. New source is qualified together in this PR; a released pin must contain this capability. |
| New plugin with older producer | Use the documented legacy spellings before discovery, preserving manual readiness checks. Do not send an unsupported selector or retry after a partial projected read. |
| Inbox script with compatible producer | With a nonempty desk-loop identity, auto mode checks `deskinbox --version` for `issue-discovery-v1` before any discovery. It runs exactly one typed listing per requested repository; no legacy label queries run in that arm. `ASSAY_DESKINBOX` selects a pinned executable. |
| Human terminal without a desk-loop identity | Auto refuses with could-not-check before either typed or ambient discovery. Choose explicit `ASSAY_INBOX_READER=legacy` for an ambient `gh` read. Explicit typed mode still requires the canonical App role. |
| Inbox script with absent/older producer | With a nonempty desk-loop identity, auto selects legacy before discovery. `ASSAY_INBOX_READER=typed` instead refuses before any discovery; use it to qualify the migrated arm without broad reads. |
| Explicit rollback | `ASSAY_INBOX_READER=legacy` restores shell label discovery. Restore the previous plugin and producer pins together for worker/intake. No live pin, installation or executor switch is performed here. |

Typed discovery reads with the session’s minted App identity; the shell’s existing
per-item detail and trust-screen reads use the operator’s ambient `gh` identity.
Those identities can see different repositories: failures stay partial, and the
operator can select `ASSAY_INBOX_READER=legacy` before discovery. Typed stderr
retains effective-configuration and coverage diagnostics as well as read failures.

Typed rows must name a requested repository and carry the documented field types
before they can drive detail reads. The existing trust screen still judges every
row; discovery supplies no author or ruling authority.

Never fall back after a compatible producer starts: permission errors, malformed
output and page-cap overflow remain incomplete observations. The existing typed
reader's bounded serial fallback is used; no unsupported batching API is invented.
Its per-process caps do not coordinate independent processes or guarantee fleet quota.

## Remaining broad readers and acceptance holds

The default full board, `tools/desk/README.md` operator recipe and independently
selected legacy compatibility paths remain supported broad readers. Deskboard's
intake-depth note correctly names the separate local `issueboard intake` read;
dispatch/stalled/health remain their own board populations and are not switched to
a narrow worker result. Independent watcher health, all trust/edit checks, audit,
budget/breaker/admission controls and consequential rechecks remain required.

Both actual mains use the shared `PerfRunMeasured` recorder. The real parser
selects immutable context before effects; compiled operation names keep private
argv out of measurement labels. Explicit invocation-owned observers reach the
existing native forge constructors. No ambient measurement setter is added.

With local telemetry consent and a selected profile, these callers emit the
canonical perf-v3 schema from desktools-v2/63–65. An unselected profile keeps the
legacy recorder path; consent off creates no record or optional cost selection.
Qualified GraphQL cost uses the same provider operation and its actual scalar
metadata, never a shared quota gauge or an extra request. Provider/child intervals
are not added to process wall time.

Issueboard's own native request inventory is covered. Inbox discovery and table
are native list-only modes with complete own-request coverage. Walk and HTML
observe native list/body reads but retain the legacy raw comment reader: its
pagination differs from the typed comment API, so changing it here would change
the existing bound. Those modes explicitly retain missing request coverage rather
than reporting raw comment traffic as measured zero. Credential-minter and other
subprocess work remains separately owned; parent measurements cannot prove full
child coverage or fleet cost.

Local caller flows exercise the actual native transport, writer, mixed history
reader and reducer, including consent off, per-invocation context isolation,
metadata failure, response parity and the raw-comment gap. Shared dependencies
remain subject to their own review and human acceptance. The separate
desk-tools/38 verification prerequisite is retained: source tests here do not
certify it or authorize runtime activation, releases, installations or pin changes.

Flow-only shell modes (`--flow`, including `--html`) read existing flow emitters and perform no issue discovery, producer capability probe or issue query. They therefore need no loop identity or reader selection. Shell discovery modes return 6 before any read when auto has no loop identity; select `ASSAY_INBOX_READER=legacy` explicitly for ambient discovery.
