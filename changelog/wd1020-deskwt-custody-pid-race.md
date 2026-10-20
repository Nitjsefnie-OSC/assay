### Fixed
- The `deskwt` legacy-coordinator crash test no longer fails intermittently on unrelated builds: its child now publishes its pid with an atomic rename instead of a truncate-then-write that a polling reader could catch half done.
