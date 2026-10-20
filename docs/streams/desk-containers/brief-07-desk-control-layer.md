---
brief: assay:assay:desk-containers:07
title: Optional Linux operator terminal layout over the shared desk attach path
wave: 11
depends: ["desk-containers/28"]
unblocks: []
effort: M
gate: model
risk: {regulatory: no, customer: no, irreversible: no, sensitive-data: no}
issues: [64]
schema: brief-v2
authored: 2026-08-22 by desk-containers scoping session
sources:
  - "medici-finance/assay#64 — the request (tmux and equivalents as a cheap way to control all the desks; win32 as well as mac; the aim is to fire up the docker/k8s pods)"
  - "docs/streams/desk-containers/spec.md — multi-desk control layer section + open question 8"
  - "containers/desk-run.sh (brief 04) — the per-desk launch path each pane invokes"
  - "freshness-checked 2026-08-22 @ b3a2067 — no control/multiplexer config exists in the tree"
why: A single Linux operator terminal should expose five independent desk sessions without requiring native Windows or adding another launcher.
version: 2
id: cbf80c98-0c5e-4309-836b-89bced9ffae5
exec-tier: strong
exec-tier-why: Terminal presentation must preserve remote session and lifecycle ownership.
consumers:
  - "containers/control/: follow-up desk-containers/07"
  - "docs/docker.md: follow-up desk-containers/07"
---

# Brief 07 — Optional Linux operator terminal layout

## Context

files: `containers/control/EVALUATION.md` (planned), `containers/control/desks.yaml` (planned), `containers/control/check.sh` (planned), `docs/docker.md`,
`changelog/linux-operator-layout.md` (planned).

facts: The 20 October 2026 direction permits cellctl inside Kubernetes and removes Windows
from this target. /28 is the single operator path. A multiplexer is optional presentation;
its absence cannot block start/attach/reconnect or require five models to stay active.
The original macOS/Windows evaluation is historical context, not a current delivery gate.

layering: A terminal layout calls /28's existing attach interface; no container flags,
credential reads, lifecycle logic or second launcher in layout scripts.

design-fit:
  owner: operator terminal presentation; execution remains loopadmin
  contract: existing cell-operation/attach boundary
  retires: [required cross-platform multiplexer evaluation for this Linux cohort]
  weight: verbs +0, flags +0 runtime, refusals +0, layout files +3
  why-add: A convenience layout can expose the five sessions without granting new authority; direct cellctl attachment remains sufficient.

## Read first

[Current target](../../proposals/non-ci-architecture/28-interactive-kubernetes-desks.md)
and [operator profile](brief-28-package-in-cluster-cellctl-and-bind-the-interactive-desk-profile.md).

## Ground rules

Branch/draft PR only; no live API, model call, deployment or workflow. No credentials in
layout or output. Independent verification owns Evidence.

## Task

1. Evaluate the smallest supported Linux terminal layout inside the operator image. Record
   actual tool/version and observed attach/detach behavior; direct one-role cellctl attachment
   remains the baseline. Do not implement native Windows, Docker orchestration or a service.
2. Expose five named role attachments through /28, using configured role aliases. Closing a
   pane/session never terminates its desk or starts a replacement; reconnect is idempotent.
   The layout composes no Pod specs and knows no Secret or credential values.
3. Add an offline fixture that replays actual layout calls against /28's fake stream backend,
   counting launches and owners. Update docs/docker.md with direct and optional layout use,
   checked versions and the distinction between disconnect and explicit lifecycle stop.

## Verify

| # | Class | Command | Expect |
|---|---|---|---|
| 1 | check +flow +dereference | `sh containers/control/check.sh --offline` | exit 0; five role panes call the single attach path; closing/reopening the layout produces zero new model invocations or duplicate role owners |
| 2 | check +mutation | `sh containers/control/check.sh --offline --self-test` | exit 0; planted start-on-attach, stop-on-pane-close and credential read all fail the fixture; no cluster contact |

## Pre-mortem

Terminal becomes launch/lifetime owner: rows 1–2. A credentials-owning parallel launcher:
row 2. Tool/version claims require a reviewer to dereference actual recorded tool output;
real stream behavior is independently qualified in /29, not inferred from layout existence.

## Evidence

<!-- Independent verifier records command, version and observed behavior. -->

## Review

Gate: model; presentation adds no privilege. Native Windows is optional separate work.
