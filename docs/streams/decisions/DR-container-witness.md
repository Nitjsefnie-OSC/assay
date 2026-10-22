---
id: DR-container-witness
date: "2026-10-22"
title: "Credit a Linux container witness row only inside a host verifier-admission bracket, with the image digest recorded"
consequence: major
decided-by: "human:<name>"
alternatives:
  - "Run verifier admission inside the network-less container. Ruled out: this is the refused shape. The admission check must reach the forge, and a container started with no network cannot, so the row either fails admission or runs unattested."
  - "Leave container-sandboxed rows unattested, or credit them without an admission bracket. Ruled out: a Linux container is the only Linux runner many verifiers have, and an unattested row is not an admitted independent-verification witness, so reviewers reject it."
  - "Make the container image digest advisory. Ruled out: it is easier to adopt, but a reviewer cannot tell two container images apart after the run, and the in-container tree digest is computed by the verifyrun binary inside that image, so an unpinned image could report a digest it did not compute."
accepted:
  - "A Linux container run counts as a Linux witness only when the host runs verifier admission immediately before the container run (the receipt) and immediately after it (the close-step check). Inside the container, verifyrun reads the host's receipt as an explicit input and runs no admission of its own."
  - "The bracket binds the admission, the bytes the container ran (a tree digest taken on the host at the open, by the container before any row, and on the host at the close) and the witness it closes. A reader that does not support the bracket refuses to credit the row."
  - "Every credited container row records its image digest. A run whose digest cannot be read is could-not-run, never a pass."
  - "Cross-repository (version 2) admission receipts are refused in container mode. A later brief may admit them."
  - "Building the container image, choosing the runtime and writing runner scripts stay with the adopter."
---

**Status:** draft — not approved. This record is not an approval. Approving a design record on a
`gate: human` brief is the driver's act. A model or a worker cannot set it. When the driver
merges the pull request that adds this record, that merge ratifies the record.

## Basis

The decision text above comes from two places: the `sources:` line of
[desktools-v2/71](../desktools-v2/brief-71-credit-container-witness-rows-bracketed-by-host-verifier-admission.md),
and the body of that brief's decision issue, #4071. The public record of the ruling it
transcribes is a citation, not a quote. It is the driver's merge of #4069, the pull request
that authored the brief, merged 2026-10-22. No ruling comment exists yet on #4071, so this record has
no `ruling:` link, and `decided-by` keeps the placeholder a public repository uses.

The implementation is #4078. That merge records what was built. It is not an approval of
this record.

## What this record does not decide

It does not stand in for the `gate: human` acceptance at done. That acceptance still decides
whether the bracket rule credits container rows (brief 71, `## Human decision`). It does not
loosen `sandboxRefusal`, the loopback precheck, the CI-lane refusal or any admission
refusal. It is a separate record from
[DR-verify-in-container](DR-verify-in-container.md), which chooses the Windows runner and is
still a draft of its own.

`consequence: major`: if this design is wrong, a container row that was never admitted
could be credited as an independent-verification witness.
