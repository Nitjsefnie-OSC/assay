### Fixed
- A plain `statusgen --lint` (no `--forge`) no longer reads origin's branch heads or decays dead claims through the drive-snapshot region check: that check now uses the run's own claim read, and a region that only an unread claim can explain is reported could-not-check instead of drift (#3323).
