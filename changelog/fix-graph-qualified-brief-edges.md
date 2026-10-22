### Fixed
- `statusgen --graph dot|jsonl` now emits the `depends`, `unblocks`, `sources` and `issues`
  edges of brief files whose `brief:` id is the qualified `<cell>:<alias>:<stream>:<NN>` form.
  The export keyed each brief file on that raw id, which never matched a `<stream>/<NN>` node,
  so every brief-v2 brief's edges were dropped without an error; it now keys on the
  filename-derived `<stream>/<NN>` id, the same one `--eligibility` uses. The brief-id class
  guard now also follows a local copy of the raw id, the shape this defect took.
