# verifyrun: cancel ordinary descendants at the row deadline

A timed-out Verify row now uses statusgen's existing Unix process-group cancellation
helper. Killing only its shell left ordinary descendants running with the output pipe open,
so verifyrun could wait past the unchanged row deadline and the child could keep working
before the next row. Timeout witnesses remain could-not-run. Exact row IDs, commands,
order, output hashes and all admission, source, fail-first, sandbox and Evidence checks
remain unchanged.

Normal background output is still read completely within the original deadline. No short
pipe cutoff, cached sandbox probe, command cache or skipped verification work was added.
Windows retains its existing direct-child cancellation limit; a descendant that deliberately
leaves its Unix process group can also outlive cancellation. This is process cleanup, not a
new sandbox or proof of total process-tree containment.

Five alternating fresh-process fixture pairs kept the same 150 ms row deadline and command.
For the planted ordinary-child timeout, median executor time fell from 2,019 ms to 153.3 ms;
a successful background-output control stayed about 1.2 seconds before and after. These controlled
measurements show the deadline repair, not faster real verification checks or a historical
30-minute-run speedup. CPU savings are not established. Original logs and exact fixture,
source and executable hashes are retained.

The existing cancellation helper remains the implementation owner. The class sweep found
related statusgen Git-reader and desk hook/callout adapters; this change creates no new
execution library and moves no credential or admission authority.

Issue: #4360
