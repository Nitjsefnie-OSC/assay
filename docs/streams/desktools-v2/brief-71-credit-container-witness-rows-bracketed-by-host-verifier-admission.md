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
  - "driver ruling, 2026-10-22: Linux container runs count as Linux witnesses when the host runs verifier admission immediately before and after the container run; admission inside a network-less container is the refused shape"
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
  - "cross-repository (version 2) admission receipts in container mode: out-of-scope (refused here; a later brief may admit them)"
  - "container image build, runtime selection and an adopter's runner scripts: out-of-scope (adopter-owned)"
---

# Brief 71 — Credit container witness rows bracketed by host verifier admission

## Context

files: statusgen/{verifyrun,verifieradmission,verifysandbox}.go and their tests; docs/brief-rules.md (rule 45); plugins/assay/skills/verify-desk/SKILL.md; plugins/assay/skills/pr-review-desk/SKILL.md; tools/desk/internal/deskkit/verifierattestation.go (read only).

facts:
- Rule 45 (`docs/brief-rules.md:639`) admits `--sandbox=container-netns` as an explicit opt-in for a runner already inside a `--network none` container (`:649`-`:661`). Every sandboxed row records `sandbox=unshare` or `sandbox=container-netns` in its witness Result cell. `witnessSandboxOf` reads it (`statusgen/verifysandbox.go:148`).
- `sandboxRefusal` (`verifysandbox.go:124`) refuses container-netns only in the CI lane (`ciLane`, `:118`) and with `--in-container`. It never looks at admission. `verifyrun` calls it first (`statusgen/verifyrun.go:1805`).
- Admission is conditional. `verifierAdmission` (`statusgen/verifieradmission.go:18`) requires it only when `DESK_LOOP` contains `verify` (`:19`) or `git rev-parse --git-path assay-verifier-attestation.json` names an existing file (`:20`-`:25`). Otherwise it returns no attestation and the run proceeds unattested (`:27`-`:28`). When it is required it runs `deskdispatch --check-verifier` in-process (`:30`). Inside a network-less container that is the refused shape the ruling names, because the check cannot reach the forge. When the container does not mount a linked worktree's git directory, the `git rev-parse` probe fails and an undeclared run is silently unattested.
- `verifierAdmissionReceipt` (`verifieradmission.go:40`) already parses the check's JSON output with no network. Version 1 yields `Verification-Attestation: <repo>#<issue> run=… source=… brief=… model=… tier=…`. Version 2 (cross-repository) needs absolute `BriefRoot` and `SourceHome` paths on the host (`:57`-`:61`). `verifyrun` appends the attestation line under the witness table (`verifyrun.go:2022`-`:2023`) and skips admission under `--check` (`:1828`).
- verify-desk §Pre-work admission (`SKILL.md:128`-`:164`) requires the check before any row and says verifier output files belong outside the admitted worktree, since any extra file, ignored files included, refuses admission. Lines `:373`-`:378` make verifyrun the execution witness. pr-review-desk's SKILL has no container, sandbox or Linux-witness text, so a reviewer has no rule that credits a container row.

Design-fit: proposal 11 makes independent verification "a disposable runner without
publisher credentials … and a separately authorized publisher"
(`11-target-architecture.md:281`). Proposal 13 step 3 runs the work "in a disposable
sandbox with explicit inputs". A network-less container fits both, provided admission
stays with the host, which holds the forge view. The ruling moves admission to the host
on both sides of the run. Inside the container, `verifyrun` reads the host's receipt as
an explicit input and runs no admission of its own. Proposal 29's worker/verifier
attestations row keeps "a closed authority core … and explicit required-proof
capabilities". The container bracket is therefore a required-proof capability on the
attestation: a reader that does not support it refuses to credit the row. It is not an
optional perf-v4 extension that a reader may skip. The verifier attestation format in
deskkit is unchanged. No new verb is added.

single-point-of-failure: the host's post-run admission check, which proves the home and the binding did not move while the container ran — backed by the pre-run receipt binding the run, and by the landing-phase check `deskevidence` already repeats at `--root`.

## Human decision

Recorded: driver ruling, 2026-10-22, as in sources. Still for acceptance at done: the
bracket form below and the image-digest requirement. Recommended: require the image
digest on every credited container row. A run whose digest cannot be read is
could-not-run, not a pass. Alternative: make the digest advisory. That is easier to adopt,
but then a reviewer cannot tell two container images apart after the run.

## Ground rules

Normal worker admission, one isolated worktree and one draft PR. Tests use fixtures and
fakes. No container engine, network or forge access is needed to run them. Do not relax
`sandboxRefusal`, the loopback precheck, the CI-lane refusal or any admission refusal.
Stop at implemented.

## Task

1. `verifyrun` gains `--admission-receipt <file>`, valid only with
   `--sandbox=container-netns`: the host's `deskdispatch --check-verifier` JSON output,
   placed outside the home and mounted read-only. With container-netns, `verifyrun`
   never runs `deskdispatch`. It parses the receipt with `verifierAdmissionReceipt`,
   requires version 1 (version 2 is refused, naming the limit), requires the receipt's
   brief to equal the brief being run, and emits the attestation line as today.
2. `--sandbox=container-netns` with no `--admission-receipt` is a usage refusal before
   any row runs. This applies whatever `DESK_LOOP` says and whether or not the
   linked-worktree git directory is visible. Unattested container runs end.
3. `verifyrun` gains `--container-image-digest sha256:<64 hex>`, required with
   `--sandbox=container-netns` and supplied by the host. It is emitted on the witness as
   `Container-Image: sha256:<…>`.
4. Add a host-side close step, `statusgen verifyrun --close-bracket --root <home> --brief
   <brief> --witness <file>`. It runs `deskdispatch --check-verifier` on the host, compares
   the result's run, source, brief, model and tier with the witness's attestation line, and
   on a match appends `Admission-Bracket: run=<run> source=<source> before=receipt
   after=check-verifier`. On a mismatch or refusal it appends nothing and exits non-zero.
   A row credited without this line is not a credited container row.
5. Amend rule 45 and verify-desk §Pre-work admission with the bracket sequence: host
   check, container run with receipt and digest, host close step, within one dispatch and
   with no home edits between. Amend pr-review-desk with the credit rule: a
   `sandbox=container-netns` row counts as a Linux witness only when the same Evidence
   block carries its Verification-Attestation line, a matching `Admission-Bracket` line
   and a `Container-Image` digest. Otherwise it is rejected as unattested. State the known
   limits: version 2 receipts are refused, and a linked worktree's git directory is not
   mounted in the container, so rows that need `git` history must run on the host.
   Regenerate the Cursor mirror.

## Verify

| # | Command | Expect |
|---|---|---|
| 1 | `cd statusgen && go test -run '^TestContainerNetnsRequiresAdmissionReceipt$' -count=1 -v .` | exit 0 and named PASS; with `DESK_LOOP` unset and no attestation file, container-netns without a receipt is refused before any row runs and no witness row is written |
| 2 | `cd statusgen && go test -run '^TestContainerNetnsReceipt$' -count=1 -v .` | exit 0 and named PASS; a valid version 1 receipt yields the same attestation line the in-process path renders; a version 2 receipt, a brief mismatch, an incomplete receipt and a missing image digest are each refused; no `deskdispatch` process is started (fake on PATH records zero calls) |
| 3 | `cd statusgen && go test -run '^TestCloseBracket$' -count=1 -v .` | exit 0 and named PASS; a matching host check appends `Admission-Bracket:`; a changed source, a different run or a refused host check appends nothing and exits non-zero |
| 4 | `cd statusgen && go test -run '^TestSandbox' -count=1 -v .` | exit 0 and all seven existing `TestSandbox*` tests PASS; the existing loopback precheck, CI-lane and `--in-container` refusals are unchanged |
| 5 | `grep -n 'Admission-Bracket' docs/brief-rules.md plugins/assay/skills/verify-desk/SKILL.md plugins/assay/skills/pr-review-desk/SKILL.md` | all three files match |
| 6 | `cd tools/desk && go run ./cmd/deskinstall --harness cursor --forge github --repo ../.. --bundle ../../plugins/assay --check` | exit 0; no Cursor-mirror drift reported |
| 7 | `statusgen --root . --consumers --brief desktools-v2/71 --base $(git merge-base origin/main HEAD)` | pre-merge, on the implementation branch: exit 0; declared implementation sites and follow-up references corroborate; out-of-scope judgments remain UNCHECKED and reviewer-owned; completeness is not proved |
| 8 | `statusgen --root . --lint` | exit 0; no PROBLEM line names this brief |

Pre-mortem: a container run that is still unattested when admission is not detected is
caught by row 1. Re-running admission inside the container, or accepting a receipt for a
different brief, is caught by row 2. A bracket that closes on a moved home is caught by
row 3. Loosening the existing sandbox refusals while adding the flags is caught by row 4.
Mutation check for row 1: restore the conditional `required` path for container-netns and
confirm row 1 goes red. Whether the receipt and close step together prove "immediately
before and after" is for review to judge, with the human gate as the backstop.

The implementation must create the exact named tests; an empty go-test match is not PASS.

## Evidence

<!-- Independent postmerge verifier records actual commands, source and outcomes here. -->

## Review

Gate: human (gate-why above). Independent correctness and security review of the exact head.
