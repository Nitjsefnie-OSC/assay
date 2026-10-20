### Changed

- statusgen's local perf record (`ASSAY_TELEMETRY=1`) is now perf-v2, like the desk tools': it times each stage of a run (parse, lint, links, claims, view, render, write, corroborate) and counts its direct forge HTTP calls, so `deskperf report` can say where statusgen's time goes. Output, exit codes and forge traffic are unchanged.
