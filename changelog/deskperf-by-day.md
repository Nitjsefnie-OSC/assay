### Added
- `deskperf report --by-day`: the perf history reads as a timeseries — one row per tool × UTC day with invocation count, wall p50/p90/max per schema group (legacy, perf-v2 and perf-v3 walls stay separate), the forge count pooled over every schema, and the day's perf-v3 request rollup; text and `--json`.
- `deskperf report` gains a perf-v3 requests-by-tool section: request attempts by protocol/resource class, request-latency p50/p90 estimated from the merged duration histogram, and rate-limit points summed where measured.
- `perfrec` exports `EstimateDurationPercentile` and `DurationBinBounds`, so histogram percentile estimates reuse the writer's own bucket-edge definition instead of a second copy.
