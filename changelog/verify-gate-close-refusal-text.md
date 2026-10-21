### Fixed
- `verify-gate-close`: a refused human close now posts statusgen's own refusal text on the gate card, sanitized and fenced, instead of a fixed "not currently verified" line that hid the real reason (a HELD or could-not-check row, for example).
