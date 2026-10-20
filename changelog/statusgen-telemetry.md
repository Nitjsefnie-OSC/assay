### Changed

- statusgen's local perf record (`ASSAY_TELEMETRY=1`) is now perf-v2, like the desk tools': it times each stage of a run (board parse, source lint and its heaviest checks, link lint, the claim read, view build, render, write, and `--corroborate`) and counts its direct forge HTTP calls, so `deskperf report` can say where most of statusgen's time goes (process start and exit fall outside every stage). Output, exit codes and forge traffic are unchanged.
