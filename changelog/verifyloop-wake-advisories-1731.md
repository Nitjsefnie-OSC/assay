### Fixed
- Preserve legacy numeric verdict execution while retaining unsupported scheduling rows as uncertainty; surface malformed latest wake receipts, reject observable held/PASS overlap and clarify partial-run execution.
- Retain all nonempty scheduling row labels as uncertainty, preserve earlier diagnostic reasons, exclude offline-deferred rows from partial dispatch and refuse row-key casing/count overlap at the writer. Re-anchor the command-reader class guard without removing any assertion.
