---
brief: assay:assay:desktools-v2:71
title: Credit container witness rows bracketed by host verifier admission
why: "A Linux container is the only Linux runner many verifiers have, but container-sandboxed rows either run unattested or attempt admission where the forge is unreachable, so reviewers reject them."
wave: 1
depends: []
unblocks: []
effort: M
gate: human
risk: {regulatory: no, customer: no, irreversible: no, sensitive-data: no}
gate-why: "Changes which runs count as admitted independent-verification witnesses (proposal 11 independent verification; proposal 13 verifier transaction step 3). Human acceptance of the bracket rule is required at done."
issues: []
schema: brief-v2
version: 1
authored: 2026-10-22 by statusgen newbrief
sources:
  - 'desktools-v2: Credit container witness rows bracketed by host verifier admission'
  - "driver ruling, 2026-10-22, for ratification at merge (its public record is the driver's merge of this gate: human brief PR): Linux container runs count as Linux witnesses when the host runs verifier admission immediately before and after the container run; admission inside a network-less container is the refused shape"
  - docs/brief-rules.md (rule 45)
  - docs/proposals/non-ci-architecture/13-migration-and-decisions.md (§The verifier transaction, end to end, step 3)
  - docs/proposals/non-ci-architecture/29-extensible-json-contracts.md (worker/verifier attestations row)
  - freshness-checked 2026-10-22 @ 9a2c6ba2bc85558d77c4227984d93a728a7f7e94 (origin/main)
exec-tier: strong
exec-tier-why: "Admission and witness credit are an independent-verification boundary; the fail-closed path and the bracket binding need adversarial review."
domain: complicated
outcome: none
consumers:
  - "statusgen/verifyrun.go, statusgen/verifieradmission.go, statusgen/verifysandbox.go: follow-up desktools-v2/71"
  - "docs/brief-rules.md rule 45: follow-up desktools-v2/71"
  - "plugins/assay/skills/verify-desk/SKILL.md and plugins/assay/skills/pr-review-desk/SKILL.md, plus the generated Cursor mirror: follow-up desktools-v2/71"
  - "tools/desk/cmd/deskevidence/ landing test of a bracketed witness: follow-up desktools-v2/71"
  - "cross-repository (version 2) admission receipts in container mode: out-of-scope (refused here; a later brief may admit them)"
  - "container image build, runtime selection and an adopter's runner scripts: out-of-scope (adopter-owned)"
---

# Brief 71 — Credit container witness rows bracketed by host verifier admission

## Context

files: statusgen/verifyrun.go, statusgen/verifieradmission.go, statusgen/verifysandbox.go and their tests; docs/brief-rules.md (rule 45); plugins/assay/skills/verify-desk/SKILL.md; plugins/assay/skills/pr-review-desk/SKILL.md; tools/desk/cmd/deskevidence/ (test only); tools/desk/internal/deskkit/verifierattestation.go (read only)

facts:
- Rule 45 (`docs/brief-rules.md:639`) admits `--sandbox=container-netns` as an explicit opt-in for a runner already inside a `--network none` container (`:649`-`:661`). Every sandboxed row records `sandbox=unshare` or `sandbox=container-netns` in its witness Result cell. `witnessSandboxOf` reads it (`statusgen/verifysandbox.go:148`).
- `sandboxRefusal` (`verifysandbox.go:124`) refuses container-netns only in the CI lane (`ciLane`, `:118`) and with `--in-container`. It never looks at admission. `verifyrun` calls it first (`statusgen/verifyrun.go:1805`).
- Admission is conditional. `verifierAdmission` (`statusgen/verifieradmission.go:18`) requires it only when `DESK_LOOP` contains `verify` (`:19`) or `git rev-parse --git-path assay-verifier-attestation.json` names an existing file (`:20`-`:25`). Otherwise it returns no attestation and the run proceeds unattested (`:27`-`:28`). When it is required it runs `deskdispatch --check-verifier` in-process (`:30`). Inside a network-less container that is the refused shape the ruling names, because the check cannot reach the forge. When the container does not mount a linked worktree's git directory, the `git rev-parse` probe fails and an undeclared run is silently unattested.
- `verifierAdmissionReceipt` (`verifieradmission.go:40`) already parses the check's JSON output with no network. Version 1 yields `Verification-Attestation: <repo>#<issue> run=… source=… brief=… model=… tier=…`. Version 2 (cross-repository) needs absolute `BriefRoot` and `SourceHome` paths on the host (`:57`-`:61`). `verifyrun` appends the attestation line under the witness table (`verifyrun.go:2022`-`:2023`) and skips admission under `--check` (`:1828`).
- The landing path rewrites witness bytes. `deskevidence` runs `stampEvidencePrincipal` on every landing (`tools/desk/cmd/deskevidence/deskevidence.go:445`, `:459`), which inserts `(on-behalf-of human:<subject>)` into the Runner cell of each added App row that has none, placed immediately before any trailing runner-source tag such as `(git-config)` (`principalstamp.go:78`, `:152`-`:156`; `needsPrincipalStamp` at `:171`). `verifyrun` renders that annotation itself only when a principal resolves (`statusgen/verifyrun.go:236`), which a network-less container without the host roster usually cannot do. The annotation is always a parenthetical, so `runnerKey` (`statusgen/verifiedrunneragree.go:74`) already ignores it for runner comparison. Brief 73 widens the rewrite to restamping planted annotations.
- verify-desk §Pre-work admission (`SKILL.md:128`-`:164`) requires the check before any row and says verifier output files belong outside the admitted worktree, since any extra file, ignored files included, refuses admission. Lines `:373`-`:378` make verifyrun the execution witness. pr-review-desk's SKILL has no container, sandbox or Linux-witness text, so a reviewer has no rule that credits a container row.

Design-fit: proposal 11 makes independent verification "a disposable runner without
publisher credentials … and a separately authorized publisher"
(`11-target-architecture.md:281`). Proposal 13 step 3 runs the work "in a disposable
sandbox with explicit inputs". A network-less container fits both, provided admission
stays with the host, which holds the forge view. The ruling moves admission to the host
on both sides of the run. Inside the container, `verifyrun` reads the host's receipt as
an explicit input and runs no admission of its own. Proposal 29's worker/verifier
attestations row keeps "a closed authority core … and explicit required-proof
capabilities". The bracket binds three things the host can check: the admission
(receipt before, check-verifier after), the bytes the container ran (a tree digest
computed on the host at the open, by the container before any row, and on the host
again at the close) and the witness it closes (a hash of the witness on the bracket
line). The container bracket is therefore a required-proof capability on the
attestation: a reader that does not support it refuses to credit the row. It is not an
optional perf-v4 extension that a reader may skip. The verifier attestation format in
deskkit is unchanged. No new verb is added.

single-point-of-failure: the host close step's three-way tree-digest match (host at open, container before its rows, host at close), which binds the bytes the container ran to the admitted home — backed by the host's post-run admission check (the home and the binding did not move), the image digest that pins the `verifyrun` binary computing the in-container digest, and the landing-phase check `deskevidence` already repeats at `--root`.

## Human decision

For ratification at merge: driver ruling, 2026-10-22, as in sources; its public record is the driver's merge of this gate: human brief PR. Still for acceptance at done: the
bracket form below and the image-digest requirement. Recommended: require the image
digest on every credited container row. A run whose digest cannot be read is
could-not-run, not a pass. Alternative: make the digest advisory. That is easier to adopt,
but then a reviewer cannot tell two container images apart after the run. The image digest
also matters to the tree binding: the in-container tree digest is computed by the
`verifyrun` binary in that image, so an unpinned image could report a digest it did not
compute.

Default if no answer: the recommended shape (host receipt plus host close-step bracket; image digest required on every credited container row) after 2026-10-24.

The default above fixes only the design the implementer builds to at dispatch, including
the tree binding of Tasks 1, 3 and 6 (over the receipt `tree` that Task 2 records). It never substitutes for the `gate: human` acceptance at done,
which still decides whether the bracket rule credits container rows.

## Ground rules

Normal worker admission, one isolated worktree and one draft PR. Tests use fixtures and
fakes. No container engine, network or forge access is needed to run them. Do not relax
`sandboxRefusal`, the loopback precheck, the CI-lane refusal or any admission refusal.
Stop at implemented.

## Task

1. Tree digest. Define one function, used by every step below: sha256 over the sorted
   list of `<path> NUL <mode> NUL <sha256 of content or link target> NUL` for every regular
   file and symlink under the root, excluding the top-level `.git` entry. Admission already
   refuses any untracked or ignored file in the home, so on an admitted home this set is
   exactly the tracked tree at `source`. Rendered as `sha256:<64 hex>`.
   Witness digest. Define a second function, exposed read-only as `statusgen verifyrun
   --bracket-digest --witness <file>`, used by the close step and by the credit rule.
   A witness table header is a line equal, after trimming, to the first line of
   `witnessHeader` (`statusgen/verifyrun.go:151`) or of `witnessHeaderFailFirst`
   (`statusgen/failfirst.go:74`). A bracket's span is fixed by the bracket's position
   alone: it starts at the nearest witness table header ABOVE that position and ends at
   the line immediately before it. For an existing `Admission-Bracket:` line the position
   is that line; for the close step, which has no line yet, it is the line it is about to
   append. The span covers that one table and the `Tree-Digest:`, `Container-Image:` and
   `Verification-Attestation:` lines under it. Other witness tables in the same file or
   Evidence block, above or below, are outside it, so a brief with several witness tables
   needs no extra marker. A position with no witness table header above it has no span.
   The header and every table row are split into cells exactly as `cellSpans` splits them
   (`tools/desk/cmd/deskevidence/principalstamp.go:219`-`:242`, the boundaries statusgen's
   `splitRowEscaped` at `statusgen/verifyrows.go:60` reads): on `|`, but never on an
   escaped `\|`, dropping at most one blank cell at each end. A Command cell that contains
   `\|` therefore stays one cell and never shifts which cell is the Runner cell. Use the
   same splitter on both sides, never a plain split on `|`.
   The span is normalised before hashing: trailing whitespace stripped, blank lines
   dropped, lines joined with LF, and each table row's Runner cell (the column the span's
   header names `Runner`) rewritten in three steps. First, every occurrence of exactly
   `(on-behalf-of human:<subject>)` is removed, where `<subject>` is one or more characters
   other than whitespace, `(` and `)`, and the match is case-sensitive. No looser form is
   removed: a parenthetical such as `(on-behalf-of <desk role>)`, `(On-Behalf-Of human:x)`
   or `(on-behalf-of human: x)` stays in the hash. Second, every run of spaces and tabs
   inside the cell becomes one space. Third, the cell is trimmed. The bytes between the
   Runner cell's two `|` delimiters are then replaced by one space, the normalised value
   and one space, and every other byte of the row is kept. The whitespace step is what
   makes the digest identical before and after landing. `verifyrun` writes a trailing
   source tag such as `(git-config)` after the runner (`statusgen/verifyrun.go:239`-`:241`),
   and `deskevidence` inserts ` (on-behalf-of human:<subject>)` immediately before that tag
   (`tools/desk/cmd/deskevidence/principalstamp.go:151`-`:156`). For example the container
   cell `example-verifier[bot] @ <tree> (git-config)` lands as
   `example-verifier[bot] @ <tree>  (on-behalf-of human:<subject>)(git-config)`.
   Removing the stamp alone would leave a double space before `(git-config)`; collapsing
   whitespace restores the container's bytes. The exact-form stamp is therefore outside the
   hash: landing it, or restamping one exact-form annotation to another under brief 73,
   never changes the digest, while any other byte in the span does.
   Output: `--bracket-digest` prints one line per `Admission-Bracket:` line in the file, in
   file order, naming the bracket it checked:
   `bracket=L<bracket line> header=L<header line> witness=sha256:<64 hex>
   recorded=<the bracket's witness= value> verdict=<v>`. The verdict is `match` when the
   computed and recorded digests are equal, `mismatch` when they differ, `no-table` when no
   witness table header lies above the bracket, and `shared` when another bracket in the
   file has the same nearest header. It exits 0 only when every bracket's verdict is
   `match`. A container row is credited only when it lies inside exactly one bracket's
   span. Rows under a table that two brackets both cover lie in two spans and are not
   credited; that is what `shared` reports.
2. Host open step: `statusgen verifyrun --open-bracket --root <home> --brief <brief> --out
   <file>`. It runs `deskdispatch --check-verifier` on the host, refuses if `--out` lies
   inside the home, computes the home's tree digest, and writes the receipt file: the check's
   JSON verbatim, `tree` (the digest) and `opened` (UTC timestamp). A refused check writes
   nothing and exits non-zero.
3. `verifyrun` gains `--admission-receipt <file>`, valid only with
   `--sandbox=container-netns`: the open step's file, mounted read-only. With container-netns,
   `verifyrun` never runs `deskdispatch`. It parses the embedded check output with
   `verifierAdmissionReceipt`, requires version 1 (version 2 is refused, naming the limit),
   requires the receipt's brief to equal the brief being run, and emits the attestation line
   as today. Before any row runs it computes the tree digest of its `--root` and refuses
   unless it equals the receipt's `tree`. After the last row it computes the digest again;
   if the two differ the run writes no attestation, `Tree-Digest` or bracket-ready witness
   and exits non-zero as could-not-run. This covers a mount that is a copy rather than the
   home, which the host close step cannot see. On a match it records
   `Tree-Digest: sha256:<…>` on the witness, between the table and the attestation line.
4. `--sandbox=container-netns` with no `--admission-receipt` is a usage refusal before
   any row runs. This applies whatever `DESK_LOOP` says and whether or not the
   linked-worktree git directory is visible. Unattested container runs end.
5. `verifyrun` gains `--container-image-digest sha256:<64 hex>`, required with
   `--sandbox=container-netns` and supplied by the host. It is emitted on the witness as
   `Container-Image: sha256:<…>`.
6. Host close step: `statusgen verifyrun --close-bracket --root <home> --brief <brief>
   --admission-receipt <file> --witness <file>`. It runs `deskdispatch --check-verifier` on
   the host and compares the result's run, source, brief, model and tier with the
   `Verification-Attestation:` line inside the span that ends at the append position (Task
   1). It recomputes the home's tree digest and requires it to equal both the receipt's
   `tree` and that span's `Tree-Digest`. Lines of these kinds outside the span (another
   table's) are ignored, and a span carrying more than one `Verification-Attestation:`,
   `Tree-Digest:` or `Container-Image:` line is refused. It refuses a witness that already
   carries an `Admission-Bracket` line. On a full match it appends one line:
   `Admission-Bracket: run=<run> source=<source> tree=sha256:<…> witness=sha256:<…>
   opened=<UTC> closed=<UTC> before=receipt after=check-verifier`, where `witness` is the
   witness digest of Task 1 over the span ending just above the line and `opened` comes
   from the receipt. It refuses when no witness table header lies above the append
   position. On any
   mismatch or refusal it appends nothing and exits non-zero. A row credited without this
   line is not a credited container row.
7. Amend rule 45 and verify-desk §Pre-work admission with the bracket sequence: host open
   step, container run with receipt and image digest, host close step, within one dispatch
   and with no home edits between. Amend pr-review-desk with the credit rule: a
   `sandbox=container-netns` row counts as a Linux witness only when the row lies inside
   exactly one `Admission-Bracket` line's span (Task 1), and that same span, not merely
   the same Evidence block, carries exactly one `Verification-Attestation:` line, exactly
   one `Tree-Digest:` line and exactly one `Container-Image:` digest line. The bracket's
   `tree` must equal that span's `Tree-Digest`, never a `Tree-Digest` belonging to another
   table. `statusgen verifyrun --bracket-digest` run on the landed Evidence block must
   report `verdict=match` on that bracket's own `bracket=` line (never a raw `sha256sum`,
   which the landing stamp would break), and the bracket's `opened` must precede `closed`. Otherwise it is rejected as unattested. State the known
   limits: version 2 receipts are refused; a linked worktree's git directory is not mounted
   in the container, so rows that need `git` history must run on the host; and the
   in-container digest is only as trustworthy as the pinned image that computes it.
   Regenerate the Cursor mirror.

## Verify

| # | Class | Command | Expect |
|---|---|---|---|
| 1 | check:ci +mutation | `cd statusgen && go test -run '^TestContainerNetnsRequiresAdmissionReceipt$' -count=1 -v .` | exit 0 and named PASS; with `DESK_LOOP` unset and no attestation file, container-netns without a receipt is refused before any row runs and no witness row is written; mutation: restore the conditional `required` path for container-netns — the named test must fail |
| 2 | check:ci | `cd statusgen && go test -run '^TestContainerNetnsReceipt$' -count=1 -v .` | exit 0 and named PASS; a valid version 1 receipt yields the same attestation line the in-process path renders and a `Tree-Digest` line; a version 2 receipt, a brief mismatch, an incomplete receipt, a missing image digest and a mounted tree whose digest differs from the receipt's `tree` (one file's content changed, one file added) are each refused before any row runs; a row that changes a file under the mounted root makes the run exit non-zero with no attestation or `Tree-Digest` line; no `deskdispatch` process is started (fake on PATH records zero calls) |
| 3 | check:ci | `cd statusgen && go test -run '^TestCloseBracket$' -count=1 -v .` | exit 0 and named PASS; a matching host check appends one `Admission-Bracket:` line carrying `tree=`, `witness=` (equal to `--bracket-digest` of the span above it), `opened=` and `closed=`; the witness digest is unchanged when an `(on-behalf-of human:<p>)` parenthetical is added to or replaced in a Runner cell, changes when a Result, Command or Date cell changes, and changes when a non-exact form such as `(on-behalf-of <desk role>)` or `(On-Behalf-Of human:x)` is added; a Runner cell carrying a `(git-config)` source tag digests the same unstamped and with the stamp inserted before the tag; in a file with two witness tables, each bracketed, `--bracket-digest` prints one line per bracket naming its own `bracket=` and nearest `header=` line, and a change in the first table changes only the first bracket's digest; a second bracket under the same table reports `verdict=shared` for both and exits non-zero; a bracket with no witness table header above it reports `verdict=no-table`; a witness `Tree-Digest` that differs from the home's recomputed digest, a home whose digest changed between open and close, a changed source, a different run, a witness that already carries a bracket line, a span whose only `Tree-Digest:` or `Verification-Attestation:` line belongs to another table, a span with two `Tree-Digest:` lines, or a refused host check each appends nothing and exits non-zero |
| 4 | check:ci +flow +mutation | `cd tools/desk && go test ./cmd/deskevidence/ -run '^TestBracketedWitnessSurvivesLanding$' -count=1 -v` | exit 0 and named PASS; a bracketed container witness fixture with unannotated App Runner cells, at least one of which ends in a runner-source tag such as `(git-config)`, and at least one row whose Command cell contains an escaped pipe (`grep -e a \| wc -l`), is passed through `deskevidence`'s landing stamp (`stampEvidencePrincipal` with a fixture roster through the existing `principalSubjectFn` seam, no forge), which adds the annotation before the tag; `statusgen verifyrun --bracket-digest`, built from this branch by the test, on the stamped block reports `verdict=match` and its `witness=` equals the bracket's; the same fixture with one Result cell edited does not; mutation: make the digest split on every `|`, escaped or not; the escaped-pipe row's stamp is then not normalised away, the stamped block reports `verdict=mismatch`, and the named test must fail |
| 5 | check:ci | `cd statusgen && go test -run '^TestSandbox' -count=1 -v .` | exit 0 and all seven existing `TestSandbox*` tests PASS; the existing loopback precheck, CI-lane and `--in-container` refusals are unchanged |
| 6 | check:ci | `grep -n 'Admission-Bracket' docs/brief-rules.md plugins/assay/skills/verify-desk/SKILL.md plugins/assay/skills/pr-review-desk/SKILL.md` | all three files match |
| 7 | check:ci | `cd tools/desk && go run ./cmd/deskinstall --harness cursor --forge github --repo ../.. --bundle ../../plugins/assay --check` | exit 0; no Cursor-mirror drift reported |
| 8 | check | `statusgen --root . --consumers --brief desktools-v2/71 --base $(git merge-base origin/main HEAD)` | pre-merge, on the implementation branch: exit 0; declared implementation sites and follow-up references corroborate; out-of-scope judgments remain UNCHECKED and reviewer-owned; completeness is not proved |
| 9 | check | `statusgen --root . --lint` | exit 0; no PROBLEM line names this brief |

Pre-mortem: a container run that is still unattested when admission is not detected is
caught by row 1. Re-running admission inside the container, or accepting a receipt for a
different brief, or running a mounted tree that is not the admitted one, is caught by
row 2. A bracket that closes on a moved home, on a witness from a different tree, or on a
witness that was already bracketed is caught by row 3; a bracket line moved to another
witness no longer matches its `witness` hash. An honest row rejected because the landing
stamp changed the hashed bytes, including the double space a stamp leaves before a
runner-source tag, is caught by rows 3 and 4. A bracket checked against the wrong
witness table, or two brackets crediting one table, is caught by row 3. A copied mount changed during
the run is caught by row 2's after-rows digest. Loosening the existing sandbox refusals
while adding the flags is caught by row 5.
Mutation check for row 1: restore the conditional `required` path for container-netns and
confirm row 1 goes red. Whether the receipt and close step together prove "immediately
before and after" is for review to judge, with the human gate as the backstop.

The implementation must create the exact named tests; an empty go-test match is not PASS.

## Evidence

<!-- Independent postmerge verifier records actual commands, source and outcomes here. -->

## Review

Gate: human (gate-why above). Independent correctness and security review of the exact head.
