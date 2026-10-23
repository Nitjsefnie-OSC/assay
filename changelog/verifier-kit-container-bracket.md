### Changed
- The verifier dispatch kit (and verifyloop's template) now tells a dispatched verifier to run a `check:ci` row the host cannot sandbox through the `--sandbox=container-netns` admission bracket (`--open-bracket` / `--admission-receipt` / `--close-bracket`) instead of recording it could-not-run.
