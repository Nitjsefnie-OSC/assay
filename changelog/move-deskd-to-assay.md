### Added
- A standalone Apache-2.0 `deskd` Go module: the command, adapters, index, contract generator, offline schema validator, authorization controls, embedded web view and tests. Existing deployments are unchanged; no runtime activation, release, image or consumer retirement is included.

### Changed
- Human operations require an explicit `--human-login` at startup. A missing or empty value disables them, an invalid one refuses startup, and the login is matched case-insensitively. Existing deployments must select their prior principal explicitly.
- Every MCP tool call is authorized against the cell it names: the caller must be a configured loop of that cell, and a filing repository must belong to it.
- State-changing HTTP routes refuse cross-site and foreign-Host requests, and JSON endpoints require a JSON content type.
- `/v1` reads are filtered to the cells the viewer may see, and the local terminal child no longer inherits configured chat key variables.
- Three adopter-configuration guards stay at their original home; every generic control is included. Consumer migration and the cell-api refactor remain separate.
