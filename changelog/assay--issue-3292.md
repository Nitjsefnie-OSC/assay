### Changed
- The standing truth suite's staged twin (`ci/staged-workflows/truth-suite.yml`) adds the
  `cellctl` plan/apply mutation spec (`tools/desk/internal/cellop/mutations.json`) to its
  mutation-gate matrix, so those guards are re-mutated on every push to main and daily once a
  maintainer promotes it; until then no workflow runs that sweep.
