### Fixed
- `deskdispatch --kit verifier` now allocates a per-run outcome-record directory outside the
  verifier home and the dispatcher root (`<prompt dir>/outcomes/<claim key>/` beside
  `--prompt-file`, else `<user cache>/assay/outcomes/<claim key>/`, owner-only, cache form
  pruned after a week) and names it in the verifier
  prompt with the exact `deskevidence --outcome-record … --root <home>` landing command; an
  allocation that resolves inside either tree is refused, never moved.
- `deskevidence --outcome-record` with `--root` refuses a relative path, or an absolute one
  inside the verifier home (symlink aliases included), before any forge read, mirroring the
  Evidence-fragment guard. Without `--root`, behaviour is unchanged.
