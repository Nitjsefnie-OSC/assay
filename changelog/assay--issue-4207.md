### Fixed

- The verify-runner image now ships git 2.56.0 at `/usr/bin/git`, built from the same pinned tarball and sha256 as the base image, instead of the Debian bookworm git 2.39.5. The admitted runner calls `git --attr-source` (git 2.41 or later) through its fixed `PATH`, so briefs with an eol-converted file were refused on the old git. The build now runs the git-floor check as its last step, and a unit test holds the two images to one pinned git.
