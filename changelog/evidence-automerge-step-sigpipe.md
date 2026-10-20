### Fixed
- The evidence auto-merge step suite no longer fails intermittently on main: its checks read captured output through here-strings instead of an early-exit reader behind a pipe under pipefail, so a match on the first line can no longer be reported as a miss.
