### Fixed
- Reconcile excludes Evidence-only PRs through the shared diff classifier and records the remaining derived-board rollout and Go-analysis CI slices as partial deliveries. Missing required patches or head content derive unknown and hold the whole write.
- Trailer-only apply limits admission reads to the todo/in-progress rows it can move, explicitly reporting other rows as unknown without provenance. Full audits, backfill and online regen drift still classify every derived row; shared per-run caches avoid duplicate file/shape reads.
