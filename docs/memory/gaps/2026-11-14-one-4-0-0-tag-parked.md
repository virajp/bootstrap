GAP 2026-11-14 ★4 | plan:2026-11-14-config-folder,
plan:2026-11-14-contract-commands,
decision:2026-11-14-bootstrap-plugins-split | one bootstrap@4.0.0 tag after R1a,
R1b, R1c and R2 | parked: the R1a folder and the claude-plugins copy of the
design record still name the old tag points

**Source** The R1b interview of `docs/plans/2026-11-14-contract-commands/`
(decision D16). The user ruled: "Keep the release to 4.0.0 since we will be
doing all these changes and only then releasing", and for the scope: "all, we
will be building a local binary and testing it".

What is out of scope for R1b, and where each point belongs:

- **R1a, before it runs.** `docs/plans/2026-11-14-config-folder/` is APPROVED
  and did not run. Its D69 says "bootstrap@4.0.0 is tagged at once after R1a
  lands", and its U8 (`08-docs.md`), U9 (`09-gates.md`) and its Parked hand
  step quote D69. If R1a runs as written, its docs unit writes the old tag
  points into the docs and the design record. Amend R1a's D69, U8, U9 and
  Parked to D16 before `/vwf:execute` runs it. R1b's U7 also fixes the design
  record, after R1a.
- **The claude-plugins copy of the design record.**
  `docs/memory/decisions/2026-11-14-bootstrap-plugins-split.md` in
  claude-plugins needs the R1a amendments (D1, D2, D3, D5, D10) and D16 of R1b.
  R1b's U7 edits only the bootstrap copy.
- **R2 planning.** The design says R2 "keeps the old task names as aliases for
  one release". Under one 4.0.0 tag no release comes between, so the R2
  interview decides whether R2 keeps the aliases at all.
