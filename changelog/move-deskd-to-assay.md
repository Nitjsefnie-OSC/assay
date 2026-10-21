# Move deskd into Assay

Add a standalone Apache-2.0 `deskd` module with the existing command, adapters,
index, contract generator, authorization controls, embedded web view and tests.
The source checkout and existing deployments remain unchanged; no runtime
activation, release, image or consumer retirement is included.

Human operations now require explicit `--human-login` at startup. Missing or
empty selection disables those operations; invalid selection refuses startup.
The non-secret identity policy freezes once, while credentials remain scoped to
individual requests. Existing deployments must explicitly select their prior
principal. Per-cell, origin, capability, namespace and MCP custody restrictions
remain in place. Synthetic fixtures replace adopter identifiers.

Two adopter-configuration guards remain enforced at their original home and
are excluded from this reusable module; every generic control stays included.
Consumer migration and the cell-api refactor/qualification remain separate.
