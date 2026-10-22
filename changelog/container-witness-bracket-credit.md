### Added
- `statusgen verifyrun` brackets a `--sandbox=container-netns` run with host verifier
  admission: `--open-bracket` writes a receipt (check JSON, home tree digest, opened time)
  outside the home; the container run needs `--admission-receipt`, `--container-image-digest`
  and `--witness`, never runs admission itself, and refuses unless its mounted root digests to
  the receipt's tree before and after the rows; `--close-bracket` re-checks on the host and
  appends one `Admission-Bracket:` line only when the run and the tree digest still match.
  The line's `witness=` value is the digest of the span from the nearest witness table header
  above it, which the Evidence landing's `(on-behalf-of human:<subject>)` stamp does not
  change. The new read-only `verifyrun --bracket-digest --witness <file>` prints one line per
  bracket with `verdict=match|mismatch|no-table|shared` and the span's line counts, and exits 0 only when all match; a
  bracket is a match only when its span holds one `Tree-Digest`, `Container-Image` and
  attestation line, its `tree=` equals that span's `Tree-Digest`, its `run=` and `source=` equal
  that span's attestation and its `opened=` is not after `closed=`. An escaped `\|` is never a
  cell delimiter, as in the Evidence landing. The close step reads the run and tree only from
  that span and refuses one holding a second `Tree-Digest`, `Container-Image` or attestation line,
  and replaces the witness by temp file and rename, never in place.
  Rule 45, verify-desk and pr-review-desk now credit a container row only inside exactly one
  closed bracket whose verdict is `match`.
- The driver's ruling of 2026-10-22 that bracketed container runs count as Linux witnesses, for ratification at merge: https://github.com/medici-finance/assay/issues/4071
