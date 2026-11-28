# Plan — Audit docs/setup/tools.md against macos/tools/ (2026-11-28)

## Goal

After this plan lands, each `### <tool>` section of `docs/setup/tools.md`, and
its sections "How a tool works" and "The packs", match the current
`macos/tools/` and `macos/packs.toml`. Each "why" rule of the old Notes of
macos-setup that is still true is in the doc. B18 is `Done`.

B18 asked to move the Notes of `docs/tool-reference.md` of macos-setup into a
bootstrap doc before a cleanup deletes that file. The user ruled that a full
audit of the doc against the current tools is the better frame: the Notes are a
checklist, not a source (D1).

There is no reversal. Ruling G3 of `2026-11-04-docs-setup` said that "a later
cleanup must move them first". That order broke: commit `a4ec825` of
macos-setup (2026-11-19) deleted the file before B18 ran. This plan does the
move late, with the git history as the source of the checklist.

## Facts the survey established

- **The old text.** `git -C ../macos-setup show
  a4ec825^:docs/tool-reference.md`,
  lines 71-201: the six parts Packages, Shells and the init stage, Secrets
  (`fnox`), Identity, Claude Code (`claude-cli`) and Maintenance
  (`macos-setup`). About a third of the claims are false now. Unit U5 holds
  the claims that are true and not documented (C1 to C32) and the false ones.
- **The doc.** `docs/setup/tools.md` has 1256 lines on 2026-11-28: "How a tool
  works" (line 6, with the item legend at lines 41-66), "The packs" (line 68,
  11 packs), and "The tools" (line 90). The numbers move when
  `2026-11-27-tool-skills` lands, with 110 `### <tool>` sections from
  `1password` to `zsh`. The doc is written by hand: no generator, test or task
  reads it. Only CLAUDE.md states the rule of one section for each folder of
  `macos/tools/`.
- **The tools.** `macos/tools/` has 110 folders on 2026-11-28, one for each
  section. The one order difference: the doc has `fzf` before `fonts`.
- **Renames since the old Notes.** No `macos-setup` or `devtools` tool;
  `updateall` is a task of `mise`. `claude-cli` was split into `claude` and
  `claude-cli`. zoxide and dasel are tools of their own. `[bootstrap.packages]`
  is `[mise.bootstrap.packages]`, and `[tools]` is `[mise.tools]`.
  `remove_items` is gone: a merge spec uses `"$mode" = "remove"`.
- **Other plans.** `2026-11-27-mac-backup-target` (complete and archived on
  2026-11-28) and `2026-11-27-tool-skills` (approved) change `macos/tools/` and
  `docs/setup/tools.md`. This plan requires `2026-11-27-tool-skills` (D4).
- **Stale comments.** The survey found five stale comments in the tool files.
  They are parked in
  `docs/memory/gaps/2026-11-28-mac-tool-stale-comments-parked.md`.
- **The gates.** `mise tasks ls --hidden` lists `code:format`, `code:lint`,
  `code:precommit` and `test`. No test covers the docs, so the plan gate is
  `code:precommit --all` (without `--all`, it runs only on the uncommitted
  files, and at the end of the plan every unit has committed).
- **The commit convention.** `.config/git-conventional-commits.yaml`: ten
  types, no scopes. Each commit of this plan is `docs`.
- **The backlog.** B18 is a draft issue of the GitHub project `bootstrap` of
  the owner of the repo (the backlog of `/vwf:backlog`), not a file of the
  tree. `/vwf:execute` sets it `Done` at the landing, from `backlog` of
  `index.json`. No unit edits it.
- **The hand-off.** The planner commits this folder and the parked record
  `docs/memory/gaps/2026-11-28-mac-tool-stale-comments-parked.md`, and lands
  them on `main`, before any run.
- **The project.** The one project is `bootstrap` (folder `cli/`). The plan
  changes no file of `cli/`, `repo/` or `macos/`.

## Unit contract

Every unit prompt carries, in order: its ruling quoted from this plan, its
owned paths plus "touch nothing outside this list", the facts section, the
plan's shared files, and the return block below. A unit never bumps a version,
never runs a generator, never edits a doc outside its owned paths, never adds
a dependency, never commits. A unit deletes with plain `rm`, never `git rm` —
it stages nothing.

A unit returns exactly this block and nothing else — no file contents, no diff:

    CHANGED: <path> — <one line>            (one per file)
    DECIDED: <what> — <why>                 (choices made inside scope, or none)
    DOCS FALSIFIED: <path> — <passage>      (reported, never edited; or none)
    QUESTION Q<n>: <the question, readable alone>
    OPTIONS Q<n>: <option 1, the recommendation> — <consequence> | <option 2> — <consequence> [| <option 3> | <option 4>]
    WHY Q<n>: <path:line or plan section> — <why the plan does not decide it>

The three QUESTION lines repeat per question, `n` counting from 1 in the
block, with two to four options; a block with no question carries the one
line `QUESTION: none`. A question is a point the plan does not decide and the
unit cannot settle inside its scope; the unit never proceeds on an
assumption.

A unit completes every edit that does not depend on an answer, never builds on
a part an answer can change, then returns all its open questions in one block.
Its edits stay uncommitted in the worktree.

## Out of scope

- A test that examines the sections of `docs/setup/tools.md`: a separate gate
  change. Today only CLAUDE.md states the rule.
- Any change in macos-setup. Its cleanup is B20.
- Any change to the engine (`cli/`), to `repo/` or to `macos/tools/`.
- The stale comments in the tool files (parked, below).

## Parked

- The stale comments in the shipped files of the Mac tools:
  `docs/memory/gaps/2026-11-28-mac-tool-stale-comments-parked.md`. The audit
  files of this plan add the comments that the run finds.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-28-mac-tools-doc-audit

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
