### Fixed
- `cellctl` now carries a configured `ASSAY_TELEMETRY=1` (from `cell.env`, `defaults.env` or the launching shell) into house, Codex and scrubbed desk launches, so the desk tools' local perf history records for launched desks. Only the exact value `1` opts in; an explicit `0` stays off, an unset key stays absent, and no other environment is forwarded.
