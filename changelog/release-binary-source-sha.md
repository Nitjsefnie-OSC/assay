### Fixed
- Released `statusgen` binaries now record the commit they were built from as `sourceSHA` in the local perf record instead of `unpinned`; any tool whose link-time commit stamp is empty (statusgen, or a desk tool built without the stamp) falls back to the commit Go embeds in the binary, and a build with no VCS information still records `unpinned`.
