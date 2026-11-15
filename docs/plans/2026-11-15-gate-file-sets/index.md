# Plan — The file sets of the code:format and code:lint gates (2026-11-15)

## Goal

After this plan lands, `code:format` and `code:lint` examine only the
uncommitted files of the current repository by default, or every file of the
repository with `--all`. They never examine a file of a submodule or of another
repository, and they always obey `.gitignore`. Each gate gets its list from one
Python helper, `gate_files`, which applies a global include and exclude list
(the key `code_files` of the mise tool) and the `#FILES` header of the gate. In
a gate run the helper is the only place that removes a file, and `--debug`
names each removed file.

The user asked for this to make the gates fast: "We need mechanism where by
default only uncommitted (or staged) files are formatted and lint'ed. This is
to optimize performance. There can be `--all` flag to include all files from
the repo." The design was settled in the artifact "Gate File Sets"
(https://claude.ai/artifact/AcPtfD2jdD91999G7TecMv), Q1 to Q8, then in the plan
interview.

**Reversal (D20):** the CLAUDE.md gate rule of "Task scripts" (a gate without
files lists them with `git ls-files`; a self-scanning tool uses the list only to
find whether a file exists) is reversed. The name `--all` comes back with a new
meaning after D100 of `2026-11-14-config-folder` removed the member flag.

## Facts the survey established

- **Git (examined with git 2.56.1 in a scratch repo):** `git ls-files` has no
  flag for the staged changes (`--stage` lists every tracked file with its
  mode), so the default set uses `git diff --cached`. `--modified` also lists a
  deleted file. `--others` lists a nested repository as one path that ends with
  `/`. `--cached` lists a submodule as one gitlink (mode `160000`), never its
  files, and a change inside a submodule is not listed.
- **Tools (Context7):** pyrefly ignores `project-excludes` and the ignore files
  when it gets `FILES...`. dprint applies its config `excludes` to given paths
  and to `--stdin-files` (dprint 0.55.0 and later), which also avoids the
  argument limit. ESLint warns for an ignored or unmatched given file unless
  `warnIgnored` is false. `dart format` has no exclude config. SwiftLint and
  ruff use `--force-exclude` today. `ARG_MAX` is 1,048,576 bytes on macOS.
- **The linter (`@virajp.dev/linter` 2.1.0):** its built-in ignores are in
  `dist/config/ignores.js`; its presets cover astro, css, html, json, jsonc,
  md, toml, yaml/yml, and the JavaScript and TypeScript types. A consumer adds
  ignores in `.config/linter.yaml`.
- **Python:** every consumer has the python tool (`pre-commit` depends on it,
  and `pre-commit` is required in the base pack), pinned to 3.14 in the `dev`
  and `ci` environments (`repo/tools/python/tools.toml:29-31`). No shipped file
  is Python yet.
- **Exclude lists in `repo/tools/`:** dprint `[defaults] excludes`
  (`repo/tools/dprint/tools.toml:15`), rendered by
  `dprint.jsonc.merge.toml:42-46`; pre-commit `exclude` (hook scope, not a
  gate list); the SwiftLint seed `excluded`
  (`repo/tools/swift/files/.config/swiftlint.yml:5-10`). No tool contributes
  excludes to dprint today.
- **The gates:** 18 entries (`code/format/*`, `code/lint/*`) in the
  `tools.toml` of dprint, flutter, kotlin, python, swift, linter, go, android
  and mise. Three gates skip generated files in code (ktlint, detekt:
  `*/generated/*`; swift-format: `*.generated.swift`). `shell_files_in_scope`
  of `helpers` has one caller, `code-lint-linter`.
- **The hooks:** the `format` hook runs `code:format --fix` and the `lint` hook
  `code:lint --fix`, each with `pass_filenames = true`. `code:precommit` runs
  `pre-commit run --all-files`.
- **Tests:** no test runs the two orchestrators; `test:repo` runs
  `code-lint-bootstrap` (`run_gate`) and checks the placement of some gate
  entries.
- **The docs that describe the gates today:** CLAUDE.md (Task scripts, the gate
  rules), README.md (the hooks, `60-bootstrap`), `docs/design/mise.md` (the
  orchestrator table, the tool-to-gate map), `docs/design/repo.md` (the member
  flags, the `lint` hook), `docs/design/tools.md` (numbered leaves),
  `docs/design/cli.md` (the hook and CI), `docs/sync/index.md`,
  `docs/sync/migrate.md`, and
  `docs/sync/tools/{mise,linter,dprint,flutter,kotlin,swift,python,go,android}.md`.
- **CI:** no shipped CI workflow runs the gates; the PR templates only name
  `mise run code` in a checklist.
- **The queued plans:** six edit the same files (`repo/tools/mise/`, the
  android and swift gates, CLAUDE.md, `docs/`); task-ports makes `own_tasks`
  take only `NN-` names. This plan requires the last one (D15).
- **The values folder:** the engine reads only named files of
  `.config/bootstrap/` (`repo.toml`, `tools.toml`, `releases.toml`), so a
  rendered `config-files.toml` there is safe.

## Unit contract

Every unit prompt carries, in order: its ruling quoted from this plan, its
owned paths plus "touch nothing outside this list", the facts section, the
plan's shared files, and the return block below. A unit never bumps a version,
never runs a generator, never edits a doc, never adds a dependency the plan
does not list, never commits. A unit deletes with plain `rm`, never `git rm` —
it stages nothing. A unit never writes a lock file — the file a package
manager or a toolchain manager writes to pin what it resolved — with Write,
Edit, `sed` or a script: a lock file changes only as the output of the CLI
that owns it, the command the stack's package-manager conventions name, and
the unit keeps what that CLI wrote. A unit whose manifest change moves a lock
file lists that lock file in its `owns` beside the manifest.

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
Its edits stay uncommitted in the worktree. The orchestrator asks the user,
records each answer in `decisions`, and resumes the unit with the question ids
and answers; the resumed unit applies them and returns the whole block again
(the execute skill's `references/questions.md`).

## Out of scope

- `code:sec` and its scanners: they keep their own file lists (D21).
- The shipped PR templates and CI: no shipped workflow runs the gates (D21).
- The queued plans: their `code:lint` gates stay; this session edits no other
  plan folder.
- The SwiftLint seed: a seed is the consumer's after the first sync (D8).
- A change of the linter package: parked (D10).

## Parked

- The built-in ignore list of `@virajp.dev/linter` duplicates the `#FILES`
  header of `code-lint-linter`. An option in the linter repository that turns
  off the list, or `warnIgnored: false`, lets the helper hold the only list.
  Record: `docs/memory/gaps/2026-11-15-linter-builtin-ignores-parked.md`.
- The vwf gate lines of a consumer (the wave and plan gates that run after the
  units commit) must use `--all` for `code:lint` and `code:format`; a change
  in claude-plugins (D55). Record:
  `docs/memory/gaps/2026-11-15-vwf-gates-all-parked.md`.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-15-gate-file-sets

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
