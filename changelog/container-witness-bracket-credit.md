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
  bracket with `verdict=match|mismatch|no-table|shared` and the span's line counts, and exits 0 only when all match and no
  `sandbox=container-netns` row lies outside every bracket's span; a
  bracket is a match only when its span holds one `Tree-Digest`, `Container-Image` and
  attestation line, its `tree=` equals that span's `Tree-Digest`, its `run=` and `source=` equal
  that span's attestation, its `opened=` is strictly before `closed=` and the span's digests are
  `sha256:<64 hex>`. An escaped `\|` is never a
  cell delimiter, as in the Evidence landing. The close step reads the run and tree only from
  that span and refuses one holding a second `Tree-Digest`, `Container-Image` or attestation line,
  and replaces the witness by temp file and rename, never in place.
  Rule 45, verify-desk and pr-review-desk now credit a container row only inside exactly one
  closed bracket whose verdict is `match`, and state that `verdict=match` shows only that the
  bracket is consistent with its span, not who wrote it: a row is credited only when the
  bracket line and every span line were landed by the verifier's landing step (per-line
  `author-mail` with the pinned id, never the display name), the span's whole attestation line
  (run, source, brief, model and tier) matches the verifier's attestation record for that run
  (the closed, never-edited `[verification-attestation] <run>` issue authored by the dispatcher
  identity), and that attestation names the brief that contains it and the source under
  verification. Per-line provenance is author-asserted attribution, not authorization; the
  forge-attributed record is the anchor. No tool checks those conditions yet.
- The driver's ruling of 2026-10-22 that bracketed container runs count as Linux witnesses, for ratification at merge: #4069
