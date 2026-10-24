---
brief: assay:assay:desk-containers:30
title: Role instructions for Codex windows and dispatched agents
why: >-
  A project whose rules differ by desk role can already give each Claude Code role window its own
  instruction file through `CELL_ROLE_CONTEXT`. Two gaps stop it from moving role rules out of the
  one shared instruction file: a role that runs on Codex is refused outright, and the text never
  reaches the agents a desk dispatches, so a worker or reviewer that needs a rule would lose it.
  This brief binds the same file to Codex windows and lets a role name, per dispatch kit, a file
  that deskdispatch quotes verbatim into every prompt of that kit. A missing or changed file
  refuses rather than launching or dispatching without the rules.
wave: 0
depends: []
unblocks: []
effort: L
gate: model
risk: {regulatory: no, customer: no, irreversible: no, sensitive-data: no}
issues: []
schema: brief-v2
version: 1
authored: 2026-10-24 by the-desk (assay:author-brief, opus-5.5; requested by human:<name>)
exec-tier: strong
exec-tier-why: "(a) the Codex binding is chosen at pickup from a probe; (b) one value crosses cellctl, the launch environment and deskdispatch; (c) it carries the rules dispatched agents follow, and a silent miss passes every happy-path test"
sources:
  - "Driver request 2026-10-24: split role-specific rules out of a project's shared instruction file into per-role files that reach each desk window and the agents it dispatches, on every harness the cell runs"
  - "docs/cellctl-role-context.md — the `instructions` key (\"Not passed to agents it dispatches\"), the custody rule (no declaration or named file under `<cell-dir>/worktrees/`, `<cell-dir>/memory/` or a tree a worktrees/ link leads to), the Refusals list (\"The role runs on a harness other than Claude Code and has a non-empty entry\")"
  - "tools/desk/cmd/cellctl/rolecontext.go — loadRoleContext / roleContextFor; the Claude binding (`--append-system-prompt`, inline `--settings`); the non-Claude refusal; roleContextVerified"
  - "tools/desk/cmd/cellctl/launch.go deskLaunch — the Codex argv (`codex <policy -c ...> ... -C <wt> -m <model> <role prompt>`), codexStandingPrompt for a standing window, and the launch environment, where the repair-admission key (deskkit.EnvRepairAdmission, `ASSAY_REPAIR_ADMISSION`) is the precedent for a value cellctl composes and deskdispatch reads"
  - "tools/desk/cmd/deskdispatch/prompt.go assemblePrompt, kits.go (commonKitPath, the closed kitFile map: worker, worker-objective, review, verifier), dispatchrecord.go buildDispatchRecord"
  - "docs/research/codex-harness-capabilities.md §3.1 — Codex reads AGENTS.md from ~/.codex, the repo root and the working directory, 32 KiB combined; no inline instruction override is recorded there"
  - "tools/cellctl/tests/harness.test.sh — the Codex arm appends the resident-rules fragment to the role worktree's AGENTS.md; that file is session-writable, which is why this brief does not use it for role rules"
  - freshness-checked 2026-10-24 @ 032dfc79d (origin/main) — `instructions` binds only on Claude Code; a Codex role with any non-empty entry is refused; deskdispatch reads no role-context value and its prompt carries only the shipped kits
domain: complicated
outcome: none
consumers:
  - "tools/desk/cmd/cellctl/rolecontext.go (the declaration gains `dispatch_instructions`; the Codex refusal narrows to the keys with no Codex binding): follow-up desk-containers/30 (this brief; flips to fixed-here when the implementation edits the path)"
  - "tools/desk/cmd/cellctl/launch.go (composes ASSAY_DISPATCH_INSTRUCTIONS into the window environment; Codex argv carries `instructions`): follow-up desk-containers/30 (this brief; flips to fixed-here when the implementation edits the path)"
  - "tools/desk/internal/deskkit/ (the env-key name and the shared kit-name list both binaries read): follow-up desk-containers/30 (this brief; flips to fixed-here when the implementation edits the path)"
  - "tools/desk/cmd/deskdispatch/prompt.go (reads the key, verifies the digest, quotes the file after the class kit): follow-up desk-containers/30 (this brief; flips to fixed-here when the implementation edits the path)"
  - "tools/desk/cmd/deskdispatch/dispatchrecord.go (the dispatched record names the file and digest): follow-up desk-containers/30 (this brief; flips to fixed-here when the implementation edits the path)"
  - "docs/cellctl-role-context.md (key table, Codex binding, dispatch delivery, refusals): follow-up desk-containers/30 (this brief; flips to fixed-here when the implementation edits the path)"
  - "tools/cellctl/examples/role-context.json (the shipped example names a review-kit dispatch file beside builtin:lean-reviewer): follow-up desk-containers/30 (this brief; flips to fixed-here when the implementation edits the path)"
  - "container and scrubbed cells: out-of-scope (CELL_ROLE_CONTEXT is refused there today and stays refused; this brief adds no binding for a launch composed inside a container)"
---

# Brief 30 — Role instructions for Codex windows and dispatched agents

## Context
files:
- `tools/desk/cmd/cellctl/rolecontext.go` — parse and validate `dispatch_instructions`; Codex binding for `instructions`; narrow the non-Claude refusal
- `tools/desk/cmd/cellctl/launch.go` — compose `ASSAY_DISPATCH_INSTRUCTIONS` for the role; add the Codex `instructions` binding to the Codex argv
- `tools/desk/cmd/cellctl/check.go` — per-role `dispatch_instructions` rows; MISS on a missing or unreadable file or an unknown kit
- `tools/desk/cmd/cellctl/desk.go` — the dry-run lines for the new key and the Codex binding
- `tools/desk/cmd/cellctl/rolecontext_test.go` — new and extended tests named in Verify
- `tools/desk/internal/deskkit/dispatchinstructions.go` (planned) — `EnvDispatchInstructions`, the value's grammar, its parser, and `DispatchKits` (the kit names)
- `tools/desk/internal/deskkit/dispatchinstructions_test.go` (planned)
- `tools/desk/cmd/deskdispatch/prompt.go` — quote the file after the class kit
- `tools/desk/cmd/deskdispatch/kits.go` — the kit map's keys checked against `deskkit.DispatchKits`
- `tools/desk/cmd/deskdispatch/dispatchrecord.go` — record source and digest
- `tools/desk/cmd/deskdispatch/dispatchinstructions_test.go` (planned)
- `docs/cellctl-role-context.md`
- `tools/cellctl/examples/role-context.json`
- `changelog/desk-containers-30-role-instructions.md` (planned)

facts:
- **What exists.** `CELL_ROLE_CONTEXT` names one operator-owned JSON declaration (`{"version":1,"roles":{<role>:{...}}}`). Its `instructions` key is a file whose text is appended to the role window's system prompt (`--append-system-prompt`). It is not passed to agents the window dispatches. An agent definition with `inherit_instructions: false` (the shipped `builtin:lean-reviewer`) starts without the project's instruction files too, so today a rule that lives only in a role file reaches no dispatched agent at all.
- **Codex today.** A role whose resolved harness is not Claude Code and whose entry is non-empty is refused ("only the Claude Code binding exists"). The Codex argv already carries `-c` overrides from the model policy and ends with the role prompt as its last argument.
- **Custody.** The whole Claude binding travels inline on the command line, which the harness ranks above every settings file a session can write. The declaration and every file it names are refused when they lie under `<cell-dir>/worktrees/`, `<cell-dir>/memory/`, or a tree a `worktrees/` entry leads to. Writing role rules into a role worktree's `AGENTS.md` (the resident-rules precedent) would put them where the session can edit them, so this brief does not do that.
- **The dispatch side.** deskdispatch assembles each prompt in `assemblePrompt` from the common clauses and one class kit, quoted verbatim, chosen by `--kit` from a closed set (`worker`, `worker-objective`, `review`, `verifier`). A shepherding run (`--pr`) uses the worker kit. deskdispatch reads no `CELL_*` value today. cellctl already composes one key for it, `ASSAY_REPAIR_ADMISSION` (`deskkit.EnvRepairAdmission`).
- **Codex inline instructions are unverified.** The capability survey records no Codex flag or `-c` key that adds developer instructions inline. That has to be checked against the installed Codex at pickup (Task 2); it is not assumed here.
- **Risk answers and `deskkit`.** The new `deskkit` file holds an env-key name, a kit-name list and a value parser. It handles no credential, token or customer data and makes no external write, so the four risk answers stay `no`. The control this brief adds is fail-closed (a refusal when the rules cannot be delivered), and no existing refusal is loosened except the Codex one, which is narrowed to keys that now have a binding.
- **Role names.** Declaration role keys are cellctl's role names. desk-tools/53 plans a rename of one of them and keeps role-keyed cell state readable; the two briefs share no ordering edge. <!-- graph: no edge desk-tools/53 — a role-name rename is read through the same keys either way -->

layering: No new component. The declaration's owner (`cellctl` role context) gains one key and one harness binding. deskdispatch gains one reader. The value both sides share (the env-key name, its grammar and the kit names) moves into `deskkit`, so neither binary restates it (Task 1, Verify 9).

design-fit:
  owner: cellctl role context (`tools/desk/cmd/cellctl/rolecontext.go`) owns the declaration; deskdispatch owns prompt assembly
  contract: none — the declaration schema is documented in docs/cellctl-role-context.md and has no semantic-owner row; the new key is additive under `version: 1`
  retires: [the blanket "non-Claude harness with a non-empty entry" refusal, narrowed to keys that have no Codex binding]
  weight: verbs +0, flags +0, refusals +4 (deskdispatch - file unreadable, digest mismatch, malformed value; cellctl - unknown kit name), rule-text lines +0 (the kits are unchanged; the quoted text is operator-owned and not shipped)
  why-add: The dispatch-time refusals cannot live in cellctl, which has finished by the time a window dispatches; deskdispatch is the only reader at that moment. Two removals were considered. Having cellctl write the text into each dispatched home was rejected, because homes are created after launch and a worktree is session-writable. Leaving the window to paste its own instructions into a dispatch was rejected, because the text would no longer be quoted verbatim or carry a digest.

## Ground rules
- NEVER git push / trigger workflows / run mutating infra commands. Commit only per the task instructions.
- Stop at `implemented` — you do not set verified/done (a different, non-implementing identity does).
- If anything is unclear or contradicts repo state: report NEEDS_CONTEXT, don't guess.

## Task
1. **The shared value (`deskkit`).** Add `EnvDispatchInstructions = "ASSAY_DISPATCH_INSTRUCTIONS"`, `DispatchKits` (the four kit names) and a parser for the value. Grammar: comma-separated `<kit>=<absolute path>@sha256:<hex>`. Each kit appears at most once, the kit must be in `DispatchKits`, and the path must be absolute. Anything else is a parse error. A test in deskdispatch fails when its `kitFile` keys and `DispatchKits` differ.
2. **Codex binding for `instructions`.** Probe the installed Codex for a command-line way to add developer or system instructions inline (a flag, or a `-c` key). Record the version probed and the result in `docs/cellctl-role-context.md`, beside the Claude Code version floor.
   - If an inline key exists, bind `instructions` to it.
   - If none does, bind it by putting the text ahead of the role prompt in the last argument, under a fixed heading that says it is the role's instructions. Do this for both the standing and the cadence form.
   - Either way, the binding is argv-only: nothing is written to disk, and the same declaration composes the same argv on every launch.
   - Narrow the non-Claude refusal: on Codex, `instructions` and `dispatch_instructions` are accepted. Every other non-empty key is still refused and named in the message. A `cursor` role keeps today's refusal for every key.
   - The `[context]` and `[dry-run] context` lines print the source and sha256 for Codex roles too.
3. **`dispatch_instructions` key.** The value is an object mapping a kit name to a file, for example `{"worker": "rules/worker.md"}`.
   - Validate each file like `instructions`: the custody rule, the 64 KiB limit, readable, a regular file. Reject a kit outside `DispatchKits` and a repeated kit. One file may serve several kits, and may also be the role's `instructions` file: a role whose window and dispatched agents follow the same rules names one file once per use.
   - When the role's entry declares it, `cellctl desk` and `cellctl up` set `ASSAY_DISPATCH_INSTRUCTIONS` in the window environment. The value uses absolute resolved paths and the digest read at launch. Set it on every harness binding, Claude and Codex. A dry run prints it on one line, `[dry-run] env ASSAY_DISPATCH_INSTRUCTIONS=<value> (role-context dispatch_instructions)`, so an adopter can feed the exact launch value to `deskdispatch --dry-run`.
   - When the role declares no entry, set nothing. Remove any inherited value from the window environment, so a launching shell's export cannot apply unnoticed.
   - `cellctl check` prints one row per role and kit (kit, path, size, sha256). A problem is a `MISS`, and like every role-context MISS it refuses every role of the cell.
4. **deskdispatch reads it.** When `ASSAY_DISPATCH_INSTRUCTIONS` is set and names the run's kit (shepherding runs count as `worker`), read the file before any claim, ref, worktree or other side effect.
   - Recompute its sha256 and compare it with the recorded one.
   - Refuse with a non-zero exit when the value does not parse, the file cannot be read, or the digest differs. The message names the kit and the file, and says to relaunch the window after a deliberate edit.
   - On success, append the file's bytes verbatim after the class kit, under `## Role dispatch instructions (<kit>) — source <path> sha256:<hex>`.
   - Add `dispatch_instructions` (path and digest) to the dispatched record.
   - When the key is unset, empty, or names other kits only, the prompt and the record are byte-identical to today's.
5. **Docs and example.**
   - In `docs/cellctl-role-context.md`: add the key to the table; change the `instructions` row to say how dispatched agents get role rules; add the Codex binding to "How it is applied"; add the new refusals to "Refusals".
   - State the limit. The env value is visible to the session, which could unset it for its own deskdispatch call. That is the same posture as "a session with a shell can still start another harness process". The digest stops an edited or swapped file, not a session that decides to drop the key.
   - Extend `tools/cellctl/examples/role-context.json` so the review role names a `review` dispatch file next to `builtin:lean-reviewer`. That is the pairing the doc says needs a self-contained prompt.
6. Add the changelog fragment.

## Verify (executable — no prose-only DoD items)
| # | Class | Command | Expect |
|---|-------|---------|--------|
| 1 | check | `cd tools/desk && go build ./... && go vet ./cmd/cellctl/ ./cmd/deskdispatch/ ./internal/deskkit/ && echo BUILD-OK` | exit 0; output contains `BUILD-OK` |
| 2 | check +flow | `cd tools/desk && go test ./cmd/cellctl/ -run '^TestCodexRoleInstructions$' -count=1 -v > "${TMPDIR:-/tmp}/r30.out" 2>&1 && grep -F -e '--- PASS: TestCodexRoleInstructions' "${TMPDIR:-/tmp}/r30.out"` | exit 0; output holds `--- PASS: TestCodexRoleInstructions`. A Codex role whose entry names `instructions` launches with the file's text in its argv through the Task 2 binding, for both a standing and a cadence launch. The `[context]` line names the source and sha256. A file listing of the role worktree is identical before and after. Two dry runs print byte-identical argv. |
| 3 | check +mutation | `cd tools/desk && go test ./cmd/cellctl/ -run '^TestCodexUnboundKeysRefused$' -count=1 -v > "${TMPDIR:-/tmp}/r30.out" 2>&1 && grep -F -e '--- PASS: TestCodexUnboundKeysRefused' "${TMPDIR:-/tmp}/r30.out"` | exit 0; output holds `--- PASS: TestCodexUnboundKeysRefused`. A Codex role that declares `memory_dir`, `tools_off` or `agents` is refused, and the message names the key. A `cursor` role with `instructions` is refused. Evidence records a run with the narrowed refusal removed, in which this test fails. |
| 4 | check +flow | `cd tools/desk && go test ./cmd/deskdispatch/ -run '^TestDispatchRulesInPrompt$' -count=1 -v > "${TMPDIR:-/tmp}/r30.out" 2>&1 && grep -F -e '--- PASS: TestDispatchRulesInPrompt' "${TMPDIR:-/tmp}/r30.out"` | exit 0; output holds `--- PASS: TestDispatchRulesInPrompt`. This is the dispatched-agent row. With the key naming a `worker` file, a fresh `--kit worker` prompt and a shepherding `--pr` prompt each contain the file's bytes verbatim, after the class kit, under the fixed heading with the digest. The dispatched record carries the path and digest. A `--kit review` run in the same environment, with no review entry, is byte-identical to a run with the key unset. |
| 5 | check +flow +dereference | `cd tools/desk && go test ./cmd/cellctl/ -run '^TestBinaryDispatchRulesFlow$' -count=1 -v > "${TMPDIR:-/tmp}/r30.out" 2>&1 && grep -F -e '--- PASS: TestBinaryDispatchRulesFlow' "${TMPDIR:-/tmp}/r30.out"` | exit 0; output holds `--- PASS: TestBinaryDispatchRulesFlow`. The test builds both binaries. It declares `worker` and `review` files for two roles of a scratch cell and takes `ASSAY_DISPATCH_INSTRUCTIONS` from each role's dry-run `[dry-run] env ASSAY_DISPATCH_INSTRUCTIONS=` line. It then runs `deskdispatch --dry-run` under that environment for the role's own kit and compares the quoted block with the file on disk byte for byte. |
| 6 | check +mutation | `cd tools/desk && go test ./cmd/deskdispatch/ -run '^TestDispatchRulesRefusals$' -count=1 -v > "${TMPDIR:-/tmp}/r30.out" 2>&1 && grep -F -e '--- PASS: TestDispatchRulesRefusals' "${TMPDIR:-/tmp}/r30.out"` | exit 0; output holds `--- PASS: TestDispatchRulesRefusals`. Each of these exits non-zero and leaves no claim ref, no worktree and no record line: a missing file, an unreadable file, a file edited after launch (digest mismatch), a malformed value, an unknown kit and a relative path. Each message names the kit and the file. Evidence records a run with the digest comparison removed, in which the mismatch case fails. |
| 7 | check | `cd tools/desk && go test ./cmd/cellctl/ -run '^TestBinaryCheckDispatchRules$' -count=1 -v > "${TMPDIR:-/tmp}/r30.out" 2>&1 && grep -F -e '--- PASS: TestBinaryCheckDispatchRules' "${TMPDIR:-/tmp}/r30.out"` | exit 0; output holds `--- PASS: TestBinaryCheckDispatchRules`. `cellctl check` exits non-zero with a `MISS` row naming the role and kit when a dispatch file is missing, when a kit name is unknown, or when the file lies under `<cell-dir>/worktrees/` or `<cell-dir>/memory/`. `cellctl desk` refuses every role of that cell before a worktree is created. With the file present, the row prints the kit, path, size and sha256. One file named for `worker`, for `worker-objective` and as the role's `instructions` passes, with one row per kit. |
| 8 | check | `cd tools/desk && go test ./cmd/cellctl/ -run '^TestDispatchKeyAbsent$' -count=1 -v > "${TMPDIR:-/tmp}/r30.out" 2>&1 && grep -F -e '--- PASS: TestDispatchKeyAbsent' "${TMPDIR:-/tmp}/r30.out"` | exit 0; output holds `--- PASS: TestDispatchKeyAbsent`. A role with no `dispatch_instructions` launches with no `ASSAY_DISPATCH_INSTRUCTIONS` in its environment, even when the launching shell exported one. An undeclared cell's argv and environment are unchanged from today. |
| 9 | check | `cd tools/desk && go test ./cmd/deskdispatch/ -run '^TestDispatchKitNamesShared$' -count=1 -v > "${TMPDIR:-/tmp}/r30.out" 2>&1 && grep -F -e '--- PASS: TestDispatchKitNamesShared' "${TMPDIR:-/tmp}/r30.out"` | exit 0; output holds `--- PASS: TestDispatchKitNamesShared` (the kit map's keys equal `deskkit.DispatchKits`) |
| 10 | check +dereference | `cd tools/desk && go test ./cmd/cellctl/ -run '^TestRoleContextShippedExample$' -count=1 -v > "${TMPDIR:-/tmp}/r30.out" 2>&1 && grep -F -e '--- PASS: TestRoleContextShippedExample' "${TMPDIR:-/tmp}/r30.out"` | exit 0; output holds `--- PASS: TestRoleContextShippedExample`. The shipped example, now carrying a `review` dispatch file, loads through the real parser, and the doc's quoted dry-run lines match what the binary prints. |
| 11 | check | `cd tools/desk && go test ./cmd/cellctl/ ./cmd/deskdispatch/ ./internal/deskkit/ -count=1 && echo PKG-OK` | exit 0; output contains `PKG-OK`. The whole packages pass, including every existing role-context, prompt and kit-text test. |
| 12 | gate:model +dereference | `cmd: cellctl desk example-cell worker-desk` — on a machine with Codex installed, in a scratch house-kind cell, run this role on Codex and a second role (`pr-review-desk`) on Claude Code. Give each an `instructions` file holding a unique sentinel sentence, launch each, and ask the window to quote its role instructions. | Each window quotes its sentinel sentence. The Codex version and the binding used (inline key or prompt prefix) match what `docs/cellctl-role-context.md` records. |
| 13 | check | `statusgen --root . --lint > "${TMPDIR:-/tmp}/r30-lint.out" 2>&1 && grep -q '^LINT: PASS' "${TMPDIR:-/tmp}/r30-lint.out"` | exit 0; the lint output ends `LINT: PASS` |
| 14 | check | `statusgen --root . --consumers --brief desk-containers/30 --base $(git merge-base origin/main HEAD)` | exit 0, run pre-merge on the implementation branch; every declared implementation site and follow-up reference corroborates against the branch diff; the out-of-scope entry stays UNCHECKED and reviewer-owned |

## Evidence
<!-- appended at implementation time by a NON-implementer: one row per Verify item
     (command, exit code, output line(s) or hash, date, runner). -->

## Review
Gate: model (from frontmatter — all four risk answers no). Reviewer records verdict + date in the stream README table.
