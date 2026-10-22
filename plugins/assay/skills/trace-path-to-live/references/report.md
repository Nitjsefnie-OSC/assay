# Report presentation

### Plain-language titles and counts

Use these visible headings: **Progress over the last month**, **Work still to finish**,
**What to do next**, **Longest sequence of dependent work**, **Work dependencies**, and
**What was checked and what is unknown**. Technical terms may appear in explanatory details,
not unexplained titles, chart legends or headline cards. In particular, do not label cards
"combined closure", "full-service" or "path points". Name each confirmed outcome in ordinary
language; use brief ids as secondary references. Define a brief once as a tracked piece of work.

Distinguish **Total work** (all distinct briefs in scope, including completed work) from
**All work remaining** (every brief whose status is not `done`). A `verified` brief is
**Awaiting final approval** and still counts as remaining. Do not equate dependency satisfaction
or a zero path weight with full completion. Show weighted-path scores only in an optional
method explanation as relative work estimates, explicitly not days or brief counts.

### Progress chart — first substantive content

Immediately below the objective and its plain-language scope, before summary cards or graphs,
show one chart with **six lines**, a labelled legend and an accessible table of the same values:

| Line | Count |
|------|-------|
| Total work | all briefs present in the selected scope at that reading, including `done` |
| All work remaining | all present briefs not marked `done` |
| Not started | `todo` |
| Being implemented | `in-progress` |
| Awaiting verification | `implemented` |
| Awaiting final approval | `verified` |

Use the report's stated timezone and a 28-day window: three preceding seven-day blocks,
oldest first, then the latest seven calendar days individually. This gives ten readings.
For each older block, show the count at its last day's end, label the full date range, and
say "week-end count", never imply it is a sum or average. Daily readings use day-end counts;
today uses the captured source revision and is labelled "so far". Separate the weekly and
daily portions visually. Use real date spacing or clearly labelled period slots, not an
unlabelled change of time scale. All six values must be inspectable at every reading.

For history, follow the **same current confirmed scope**: the deduplicated prerequisite set
computed above. Do not retroactively re-resolve a different objective for each day. Read each
root's default-branch state at the end of the period from recorded historian data or an immutable
Git snapshot selected from that branch's history. Record each source revision, cutoff and
timezone in coverage. Match stable brief identity across moves/renames when available.
Distinguish a brief not yet present in a fully readable historical tree from an unreadable or
unmatched record. The former contributes to neither total nor remaining; the latter is unknown,
never zero or an invented flat line. Mark incomplete readings and leave affected series as gaps.
Explain that these are historical counts for today's selected work; newly authored scope can
make both total and remaining rise. This is not a historical reconstruction of the entire plan.
Disclose other lifecycle states; include them in remaining when not `done`, and do not silently
assign them to one of the four named stages. No projection or ideal burn line is required.

### Remaining-work graph and coverage

Write ONE self-contained HTML page: inline CSS and JS, no external fetch, no CDN. In addition
to the progress chart and plain-language counts above, it contains:

- the objective and the confirmed targets;
- the critical path, overall and per target;
- the ranked moves;
- the graph, rendered from only nodes whose status is **not `done`**:
  - each node shows its status, gate, eligibility, PR and claim badges;
  - the critical path is highlighted;
  - exclude `done` nodes and their edges from rendering, search, click-isolation and layout;
    keep them in the underlying analysis, historical counts and provenance;
  - keep `verified` nodes visible as awaiting final approval, including any sign-off action;
  - calculate dependency readiness and weighted paths on the complete graph before filtering;
    never invent a direct dependency across a hidden completed node;
  - clicking a node isolates its remaining prerequisites and dependents; its detail panel may
    separately list completed prerequisites as text;
- a coverage block. It lists each root with its SHA, every `could-not-check` root, hold or reader,
  every fallback used, and the cross-root dependents.

The page names private boards when a root is private, so write it **outside any repository** in
the operator's scratch location and never commit it. Hand the operator its absolute path.
