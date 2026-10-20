# Review-scope case corpus — first-pass inventory and the blocking boundary

Synthetic packets for the review-scope contract (brief 20). An independent reviewer scores
each case's outcome against the contract in `review-prompt` clause 13 and the executable
specification in `tools/desk/internal/deskkit/reviewscope.go`, and records the model and
version, the packet hashes and any disagreement in the brief's Evidence.

Every packet below is SYNTHETIC. No private incident identifier, review transcript or
production detail appears here; the "PRs", files and findings are invented to exercise the
three scope outcomes offline. The three named unit tests
(`TestReviewScopeFirstPassInventory`, `TestReviewScopeRequiredAndUnrelated`,
`TestReviewScopeNoSilentPromotion`) drive the same inputs through the specification; this
corpus is the human-readable half a reviewer scores independently.

The contract in one line: **the first pass inventories and declares; a blocker names a
concrete failure and its scope basis; a late sibling keeps its class and round count.** The
five scope bases a blocker may name are `changed-behaviour`, `acceptance-obligation`,
`material-claim`, `safety-consequence`, `design-fit`.

---

## Packet pairs — one first-pass review packet and one follow-up packet per scenario

Each scenario below carries an explicitly labelled **first-pass review packet** (the first
review of the change, before any fix) paired with an explicitly labelled **follow-up review
packet** (the re-review of the worker's fix at a new head). Case ids are stable: the eight
originally scored cases keep their ids and their expected outcomes, and the packets added to
complete each pair take the next free id in their scenario.

| scenario | Verify row | first-pass review packet | follow-up review packet |
|---|---|---|---|
| A — first-pass inventory | 1 | A.1 (with A.2, the non-conforming first pass it is contrasted with) | A.3 |
| B — required, unrelated and safety findings | 2 | the B first-pass packet, whose four findings are B.1 to B.4 | B.5 |
| C — no silent promotion, class continuity | 3 | C.3 | the C follow-up packet, whose two findings are C.1 and C.2 |

---

## Case A — first-pass inventory over a multi-file class (Verify row 1)

A change to a payment source file asserts, in a code comment and a doc, that a payment
"settles in one block". The claim is false: settlement takes up to three blocks. This is one
false-claim class with occurrences in more than one file. (All file names below are
synthetic illustration, shown inside code blocks; they name no real path in this repository.)

### Case A — FIRST-PASS REVIEW PACKET

#### A.1 — first-pass review packet (complete)

```
packet: first-pass review (round 1 of class settles-in-one-block)
finding-class: settles-in-one-block
first-pass search:
  command:   grep -rn 'settles in one block' src/ docs/
  scope:     changed surface (src/pay.go), the item's required deliverables
             (docs/overview.md is a listed deliverable), and references to the
             settlement claim across the repository
  exclusions: vendored third-party trees under vendor/ (read, not this item's deliverable)
  revision:  <synthetic head>
  complete:  yes
in-scope occurrences (reported TOGETHER):
  - src/pay.go:42        the code comment
  - docs/overview.md:9   the same assertion in prose
excluded (read, not in scope):
  - vendor/x/readme.md:3 an incidental substring match in a third-party readme
verdict: CHANGES_REQUESTED — the false settlement claim (class settles-in-one-block)
         appears at src/pay.go:42 and docs/overview.md:9; fix both.
```

**Expected score:** the packet records the search, its scope, its exclusions and the
revision, and names ALL in-scope occurrences in one verdict. It does not leave the second
occurrence for a later round.

#### A.2 — the "incomplete search" negative (a non-conforming first pass, not the follow-up)

```
packet: first-pass review (non-conforming variant of A.1)
finding-class: settles-in-one-block
first-pass search:
  command:   grep -n 'settles in one block' src/pay.go   # one file only
  scope:     src/pay.go
  complete:  no (docs/ and the rest of the repository were not searched)
verdict: <would-be> APPROVED — no other occurrence found
```

**Expected score:** REJECT the verdict. An incomplete search is reported incomplete and
never certified clean: this packet searched one file and cannot assert the class is clear.

### Case A — FOLLOW-UP REVIEW PACKET

#### A.3 — follow-up review packet (re-review of the fix to A.1)

```
packet: follow-up review (re-review of A.1's verdict at the worker's fix)
finding-class: settles-in-one-block
prior packet: A.1 (CHANGES_REQUESTED on src/pay.go:42 and docs/overview.md:9)
worker fix: both occurrences now read "settles within three blocks"
re-review search (A.1's recorded search, re-run at the new head):
  command:   grep -rn 'settles in one block' src/ docs/
  scope:     same as A.1 — changed surface, required deliverables, and references to
             the settlement claim across the repository
  exclusions: vendored third-party trees under vendor/ (read, not this item's deliverable)
  revision:  <synthetic fix head>
  complete:  yes
in-scope occurrences: none
excluded (read, not in scope):
  - vendor/x/readme.md:3 the same incidental third-party substring match as in A.1
verdict: APPROVED for class settles-in-one-block — both occurrences A.1 named are fixed,
         and a complete, recorded re-run finds no in-scope occurrence left.
```

**Expected score:** ACCEPT the verdict. A complete, recorded search with no in-scope
occurrence may certify the class clean; the excluded third-party match stays excluded and is
not turned into a new blocker. Because A.1 reported both occurrences together, the class
closes on this one re-review with no sibling left over for another round. (Contrast A.2: the
same "clean" claim from an incomplete search is rejected.)

---

## Case B — required, unrelated, and safety findings (Verify row 2)

A change adds a new settlement mode to the same synthetic payment source file.

### Case B — FIRST-PASS REVIEW PACKET

The first review of the change. Its record and verdict come first; its four findings, each
scored on its own, are B.1 to B.4.

```
packet: first-pass review (round 1)
first-pass search:
  command:   read the brief's required deliverables; grep -rn 'settlement mode' src/ docs/;
             trace the callers of the new mode's code paths
  scope:     changed surface (src/pay.go), the item's required deliverables, and code
             paths the new mode reaches
  exclusions: none
  revision:  <synthetic head>
  complete:  yes
findings: B.1, B.2, B.3, B.4 (below)
verdict: CHANGES_REQUESTED — blockers B.1 (acceptance-obligation) and B.3
         (safety-consequence); B.2 and B.4 filed as linked follow-up issues, holding nothing.
```

#### B.1 — omitted required operator-state table → BLOCKING

```
finding: the brief requires an operator-state table documenting the new mode's states in
         docs/ops/runbook.md; the diff does not add it.
location: docs/ops/runbook.md (untouched by the diff)
scope basis: acceptance-obligation
```

**Expected score:** BLOCKER. Untouched is not automatically exempt — a required deliverable
is in scope even when the worker omitted it from the diff.

#### B.2 — unrelated pre-existing prose → FOLLOW-UP

```
finding: NOTICE:1 carries a stale copyright year.
location: NOTICE:1
scope basis: none — pre-existing, unrelated to the settlement change
```

**Expected score:** FOLLOW-UP, filed as a linked issue. It names no concrete failure of this
change and holds the PR of nothing.

#### B.3 — safety consequence outside the edited lines → BLOCKING

```
finding: the new mode reaches a delete path in src/gc.go whose guard the change disarms;
         the guarded delete is now fail-open.
location: src/gc.go:88 (outside the diff's edited lines)
scope basis: safety-consequence
```

**Expected score:** BLOCKER. A demonstrated safety consequence of the change blocks even
outside the edited lines.

#### B.4 — co-location only → FOLLOW-UP

```
finding: src/pay_notes.md:2 has a typo; it shares the directory with the change.
location: src/pay_notes.md:2
scope basis: none — sharing a directory/substring is not a basis
```

**Expected score:** FOLLOW-UP. Co-location is not scope.

### Case B — FOLLOW-UP REVIEW PACKET

#### B.5 — follow-up review packet (re-review of the fix to B.1 and B.3)

```
packet: follow-up review (re-review of the B first-pass packet at the worker's fix)
prior packet: the B first-pass packet (blockers B.1, B.3; follow-ups B.2, B.4)
worker fix: docs/ops/runbook.md now carries the operator-state table for the new mode;
            the src/gc.go:88 guard is re-armed, so the guarded delete fails closed again
revision: <synthetic fix head>
re-check of each prior finding:
  - B.1  the required table is present and lists every state of the new mode -> resolved
  - B.3  the delete path is guarded again; the change no longer disarms it  -> resolved
  - B.2  NOTICE:1 unchanged; still a linked follow-up issue                   -> not re-raised
  - B.4  src/pay_notes.md:2 unchanged; still a linked follow-up issue         -> not re-raised
new findings: none
verdict: APPROVED — both blockers are resolved at this head; the two follow-ups stay linked
         issues and hold nothing.
```

**Expected score:** ACCEPT the verdict. Each blocker is cleared only by evidence read at the
new head, and the re-review keeps the first pass's boundary: the unrelated prose (B.2) and the
co-located typo (B.4) are still follow-ups, not blockers, because nothing about them changed —
no scope basis, no changed impact and no new evidence.

---

## Case C — no silent promotion, and class continuity (Verify row 3)

### Case C — FIRST-PASS REVIEW PACKET

#### C.3 — first-pass review packet (round 1)

C.3 is numbered after C.1 and C.2 because those ids are stable; it is the round-1 review
that C.1 and C.2 follow up.

```
packet: first-pass review (round 1)
finding-class: settles-in-one-block
first-pass search:
  command:   grep -rn 'settles in one block' src/ docs/
  scope:     changed surface (src/pay.go), the item's required deliverables
             (docs/overview.md), and references to the settlement claim across the
             repository
  exclusions: vendored third-party trees under vendor/ (read, not this item's deliverable)
  revision:  <synthetic head>
  complete:  yes
in-scope occurrences (reported TOGETHER):
  - src/pay.go:42        the false settlement claim in a code comment
  - docs/overview.md:9   the same assertion in prose
other findings:
  - docs/notes.md:12     an advisory note; no scope basis in this change -> NON-BLOCKING,
                         filed as a linked follow-up
verdict: CHANGES_REQUESTED — class settles-in-one-block at src/pay.go:42 and
         docs/overview.md:9 (round 1); docs/notes.md:12 holds nothing.
```

**Expected score:** a conforming first pass: it records its search, scope, exclusions and
revision, names both in-scope occurrences together, opens class `settles-in-one-block` at
round 1, and judges the advisory note non-blocking for want of a basis. Its search is
literal text, so it does not match a sibling that states the same proposition in other words
(docs/api.md:30, "settles instantly") — a review coverage gap that C.2 must not charge to the
worker as a fresh class.

### Case C — FOLLOW-UP REVIEW PACKET

The round-2 re-review of the worker's fix to C.3. Its record and verdict come first; its two
findings, each scored on its own, are C.1 and C.2.

```
packet: follow-up review (round 2 re-review of C.3)
prior packet: C.3 (class settles-in-one-block at src/pay.go:42 and docs/overview.md:9;
              docs/notes.md:12 non-blocking)
worker fix: both occurrences corrected; the worker also edited an unrelated file
revision: <synthetic fix head>
findings: C.1, C.2 (below)
verdict: CHANGES_REQUESTED — class settles-in-one-block only, at round 2: the late sibling
         C.2 surfaced is in scope and unfixed; docs/notes.md:12 (C.1) stays a follow-up.
```

#### C.1 — a previously non-blocking occurrence is not promoted by an unrelated edit

```
round 1: docs/notes.md:12 is an advisory note; judged NON-BLOCKING (no basis in the change).
round 2: the worker edits an unrelated file. A reviewer now wants to block on docs/notes.md:12.
         Nothing about the note itself changed: no changed impact, no new evidence.
```

**Expected score:** it stays a FOLLOW-UP. A previously non-blocking occurrence cannot become
blocking merely because another file was edited — promotion requires changed impact or new
evidence, explicitly recorded. (If the reviewer records genuine changed impact or new
evidence, promotion IS permitted — that is the line between real new evidence and a bypass of
the standing rejection.)

#### C.2 — a late sibling retains its original class and round count

```
class: settles-in-one-block  (round count = 2: the C.3 first pass, then this re-review)
late discovery: a third occurrence surfaces at docs/api.md:30, re-noticed under the
                re-discovered label "settles-instantly".
```

**Expected score:** the late sibling is folded into the EXISTING class
`settles-in-one-block`. Its round count stays 2 — it is not reset to zero, not incremented as
a fresh class, and the worker is not charged a new class for review's own missed coverage.

---

## Scoring record (independent reviewer fills at verify time)

| case | packet | expected outcome | scored outcome | model / version | agree? |
|---|---|---|---|---|---|
| A.1 | A first-pass | complete packet names all in-scope occurrences together | | | |
| A.2 | A first-pass (non-conforming variant) | incomplete search rejected — never certified clean | | | |
| A.3 | A follow-up | complete recorded re-run with no in-scope occurrence certifies the class clean | | | |
| B.1 | B first-pass | omitted required operator table — BLOCKER | | | |
| B.2 | B first-pass | unrelated pre-existing prose — FOLLOW-UP | | | |
| B.3 | B first-pass | safety consequence outside edited lines — BLOCKER | | | |
| B.4 | B first-pass | co-location only — FOLLOW-UP | | | |
| B.5 | B follow-up | blockers cleared only by evidence at the new head; follow-ups not promoted | | | |
| C.3 | C first-pass | recorded first pass opens the class at round 1; advisory note non-blocking | | | |
| C.1 | C follow-up | no promotion without changed impact/new evidence | | | |
| C.2 | C follow-up | late sibling keeps class + round count | | | |

Any disagreement between the scored and expected outcome prevents acceptance.
