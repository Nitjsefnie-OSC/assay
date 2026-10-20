# Land the write-policy CI and release qualification hunk

The worker App does not hold workflow-write permission. This source PR prepares
[write-policy-ci.diff](write-policy-ci.diff); it changes neither live workflow.
Use the existing human-applied workflow leg described in
[go-analysis-ci-hunk.md](go-analysis-ci-hunk.md). The proposed workflow-App decision
record is still unapproved and supplies no alternative authority.

The diff adds the deskcore suite to the unfiltered source build-test loop and to
the existing desk-tools release-test leg. The desk suite already runs actual
built-hook parity, consumer and mutation fixtures. Existing jobs, conditions,
permissions and checks are preserved. A runbook or local pass is not CI activation.

From a human-owned isolated checkout, set `REVIEWED_HEAD` to the exact full
implementation SHA reviewed on `codex/k8s-shared-write-policy`. These commands
must run under the human's own authorized Git identity, not a worker App:

```sh
: "${REVIEWED_HEAD:?Set the full implementation SHA you reviewed}"
case "$REVIEWED_HEAD" in *[!0-9a-f]*|'') exit 1 ;; esac
test "${#REVIEWED_HEAD}" -eq 40
git fetch origin codex/k8s-shared-write-policy
test "$(git rev-parse refs/remotes/origin/codex/k8s-shared-write-policy)" = "$REVIEWED_HEAD"
CI_REVIEW_TREE=$(mktemp -d "${TMPDIR:-/tmp}/write-policy-workflow.XXXXXX")
git worktree add --detach "$CI_REVIEW_TREE" "$REVIEWED_HEAD"
cd "$CI_REVIEW_TREE"
git apply --check docs/runsheets/write-policy-ci.diff
git apply docs/runsheets/write-policy-ci.diff
git diff --check
git diff -- .github/workflows/ci.yml .github/workflows/release.yml
```

Inspect the two additive test invocations, then commit and push only those files:

```sh
test "$(git diff --name-only)" = "$(printf '%s\n' .github/workflows/ci.yml .github/workflows/release.yml)"
git add .github/workflows/ci.yml .github/workflows/release.yml
git commit -m 'ci: qualify shared write policy in source and release tests'
git push origin HEAD:refs/heads/codex/k8s-shared-write-policy
```

Do not dispatch a workflow. The implementation worker must fetch and merge the
human's branch commit before another push. The independent review desk checks the
normal PR run at that exact head; merged-main verification and the normal release
lane remain separate. If the hunk stops applying or checks fail, preserve the
failure and reconcile source; do not copy a whole stale workflow over current main.

The source module replacement is build plumbing. The external consumer test uses
an explicitly synthetic offline proxy version; no compatible public deskcore
version/tag/checksum has been selected or published. The human release owner must
select and publish through the normal release policy, then the actual managed
consumer must resolve that exact pin with `GOWORK=off` and qualify its own allocation,
write-root and independent sandbox flow. Neither this workflow hunk nor importing
`managed/v1` supplies that qualification.
