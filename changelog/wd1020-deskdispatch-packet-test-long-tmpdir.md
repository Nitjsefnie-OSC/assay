### Fixed
- The `deskdispatch` packet tests no longer depend on the length of the host's temp-directory path: three tests that compared a temp path against a value the packet cuts at 240 characters now pass under a long `TMPDIR` too. The cap itself is unchanged.
