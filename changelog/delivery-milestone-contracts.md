### Added
- Draft delivery milestone and stream contracts define supported outcomes, versioned capability dependencies, criterion coverage and evidence-bound acceptance. Canonical IDs use `cell:repo:stream:m-NN` for milestones and `cell:repo:stream:b-NN` for briefs; shorter colon forms are context-based reference elisions.

### Changed
- `author-spec` captures milestone promises and acceptance criteria; `author-brief` maps contributing work and separates start requirements from final acceptance prerequisites. Shared guidance keeps unsupported declarations as scoping proposals until coordinated reader support exists.
- Define `stream:b-NN` as the preferred new brief label after negotiated reader support, deprecating unprefixed labels for new authoring while preserving canonical identities, legacy reads and history. Authoring skills select the supported identifier/schema pair.
- Use colon-only prefixed references (`cell:repo:stream:m-NN@version` and `cell:repo:stream:b-NN[@version]`), with context-based elisions, exact optional brief revision pins, and legacy alias compatibility. Existing reader support remains unchanged.
