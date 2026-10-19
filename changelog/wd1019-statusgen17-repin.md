### Fixed
- `statusgen/17` Verify row 10 pins the current text of the seven test functions it guards. A later change to the scan planner's signature edited two of them, so the pinned hash no longer matched main; the row now re-pins to those functions as they stand, with the same command shape and the same Expect. The brief's `version:` moves from 1 to 2.
