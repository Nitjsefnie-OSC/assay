### Fixed
- `deskdispatch --kit verifier` now allocates a per-claim outcome-record directory outside the
  verifier home and the dispatcher root (`<prompt dir>/outcomes/<claim key>/` beside
  `--prompt-file`, else `<user cache>/assay/outcomes/<claim key>/`, owner-only, cache form
  pruned after a week) and names it in the verifier
  prompt with the exact `deskevidence --outcome-record … --root <home>` landing command; an
  allocation that resolves inside either tree is refused, never moved. A claim key that is
  not a single directory name (a separator, a `..`) is refused on both forms, never rewritten.
  An outcomes directory (or, in the cache form, `<user cache>/assay`) that is a symlink or not
  a directory, or that other users can write to, is refused before anything is pruned or
  created, so the week-prune can never reach another tree.
- `deskevidence --outcome-record` with `--root` refuses a relative path, or an absolute one
  inside the verifier home (symlink aliases included), before any forge read, mirroring the
  Evidence-fragment guard. Without `--root`, behaviour is unchanged.
