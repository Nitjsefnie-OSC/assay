### Fixed
- The desktools-v2/64 CLI context-precedence test now derives the binaries it exercises from the performance entrypoint inventory and fails when an exercised row's parser owner or state disagrees with the CLI migration registry, so a changed parser exemption on a claim or token-custody entrypoint is caught by that row's own test filter.
