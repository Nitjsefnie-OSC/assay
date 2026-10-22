### Added
- `statusgen verifyrun` brackets a `--sandbox=container-netns` run with host verifier
  admission: `--open-bracket` writes a receipt (check JSON, home tree digest, opened time)
  outside the home; the container run needs `--admission-receipt`, `--container-image-digest`
  and `--witness`, never runs admission itself, and refuses unless its mounted root digests to
  the receipt's tree before and after the rows; `--close-bracket` re-checks on the host and
  appends one `Admission-Bracket:` line only when the run and the tree digest still match.
  Rule 45, verify-desk and pr-review-desk now credit a container row only inside a closed
  bracket.
