### Fixed
- The review cost fixture now checks every cohort's result against a literal expectation and reads each board observation's degraded-row and truncation signals, so a partial cohort can no longer be reported as `ok` while the test passes. Each issue-enumeration invocation must also record its own HTTP attempts.
