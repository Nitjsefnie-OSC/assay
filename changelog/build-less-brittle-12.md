### Added
- `docs/refactor-oracle-template.md`: the refactor oracle a `redesign` is coded against — intent (bound to pinned and candidate sources), failure modes mapped to `// regression:` tests, triaged characterization tests (`keep` / `drop` / `unknown`), invariants, and five runnable acceptance rows that refuse on an `unknown` or an untested failure mode.
- The test ledger reads a `// characterization: <module> <oracle-file>` tag and reports `characterization-untriaged` for a tagged test its oracle's triage table does not list.

### Changed
- The author-brief skill and the review kit's design-fit section name the `oracle:` line: a design brief from a `redesign` carries it, and its `## Oracle` triage table is a material claim at review.
