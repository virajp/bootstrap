# Plan — [files] tables and merge mode for repo tools (2026-11-04)

## Goal

After this plan lands, a repo tool states the condition and the mode of each
of its files in a `[files."<target>"]` table of its `tools.toml`: `mode`
(`generated`, `seed` or `merge`), `requires` (exact paths), `requires_found`
(file names searched down to 4 folders), `format`, and for merge `remove`
and `remove_items`. The old keys `seed`, `[seed_requires]` and the `files`
and `requires` of a mise entry stop the sync with code 2 and name the new
key; a mise entry uses `requires_found`. Merge mode moves into
`sync-engine/merge`; no repo tool uses it yet. Every `tools.toml` of
bootstrap moves to the new keys, and the empty `seed` keys go.

Plan 3 of 16 — requires `docs/plans/2026-11-04-sync-engine-core`; required by
`docs/plans/2026-11-04-packs-merged`.

## Facts the survey established

- **The design rulings** come from the planning session of 2026-11-04
  (artifact https://claude.ai/artifact/7LKisXPLbSGtrFfVD36ez7). Each unit
  quotes the rulings it implements; no unit reads another file for them.

`A` is `sync-engine/repo-sync` (line numbers from before plan 1; they can
differ by a few lines). `B` is
`MS/tools/macos-setup/.config/mise/tasks/_scripts/setup-sync` and
`MS` = `/Users/virajpatel/Projects/github.com/virajp/macos-setup` (read only).

- **The manifest reader.** `_RS_MANIFEST_TEMPLATE` (A:435-444) prints
  `default_env`, `depends`, `extension`, `seed`, `seed_requires`, `mise`,
  `mise_requires`, `mise_files` and `root_only` lines. `_rs_mise_index`
  (A:474-533) reads every manifest of the clone in one gomplate call, and
  checks the `files` of a mise entry (needs `requires`, each path a file of
  the tool folder).
- **Seeds today:** `_rs_resolve` (A:671-677) fills `_RS_SEEDS` and
  `_RS_SEED_REQUIRES`; `_rs_file` (A:1539-1551) keeps an existing seed and
  skips a seed whose `seed_requires` file (an exact path, `-e`) is absent.
- **mise `requires`/`files` today:** `_rs_resolve` (A:681-690) fills
  `_RS_MISE_REQUIRES` and `_RS_FILE_REQUIRES`; `_rs_file_skips` (A:887-899)
  and `_rs_skip_required` (A:901-924) skip a file and remove a copy with the
  header; `_rs_has_file` (A:963-971) is the search (git ls-files, depth 4);
  `_rs_mise_env` (A:1021-1028) drops a mise tool whose `requires` file is
  absent.
- **The tools.toml files** (29 tools: 11 with a seed, 18 with `seed = []`). With
  a seed: bun (`bunfig.toml`),
  docker (`.tools/docker-compose.yaml`), fnox (`.config/fnox.toml.tmpl`), git
  (`.config/wt.toml`), gitleaks (`.config/gitleaks.toml.tmpl`), grype
  (`.config/grype.yaml`), kotlin (`.config/detekt.yml`,
  `.config/ktlint.editorconfig`), mise
  (`.config/mise/conf.d/zz-repo/mise.ci.toml`,
  `.config/mise/conf.d/zz-repo/mise.dev.toml`), node (`pnpm-workspace.yaml`,
  with `[seed_requires] "pnpm-workspace.yaml" = "package.json"`),
  osv-scanner (`.config/osv-scanner.toml`), swift
  (`.config/swift-format.json`, `.config/swiftlint.yml`). Every other tool
  has `seed = []`. mise `requires`/`files`: python only — `ruff` (files
  `.config/mise/tasks/code/format/ruff`, `.config/mise/tasks/code/lint/ruff`)
  and `pipx:pyrefly` (files `.config/mise/tasks/code/lint/pyrefly`), each
  `requires = "pyproject.toml"`. The `files` keys of `github` and `gitlab`
  (`tools.toml:13-14`) are pre-commit hook regexes in `[[precommit.hooks]]`,
  not mise entries: they stay.
- **Merge mode in B:** `_SS_MERGE_JQ` (B:872-~928), `_ss_merge`
  (B:1049-1163), the `[files]` schema and the merge rules in the header of B
  (B:64-84; the merge rules B:178-214), the config errors and their messages in
  `MS/.config/mise/tasks/test/setup-sync` (:1246-1347). B runs jq and yq
  through its pinned wrappers; the engine now uses `_se_jq` and `_se_yq` of
  `sync-engine/common`.
- **After plan 2:** `sync-engine/write` has `_se_format`, `_se_parse_check`,
  `_se_tmp`, `_se_seed_kept`, `_se_unblock` and `_se_finish`;
  `sync-engine/common` has `_se_has_header` and the tool wrappers.
- **Tests:** the python `files` cases of `test:repo-sync` (about :406-470)
  and the node `seed_requires` cases use the old keys.
- **Docs that describe the keys:** `CLAUDE.md` ("Layout rule": the five keys,
  `seed`, `[seed_requires]`; "mise config": `requires` and `files`),
  `docs/design/tools.md`, `docs/design/mise.md`, `docs/sync/tools/*.md` (the
  Seeds row and the mise tool notes), the header comment of A.

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
    GAP: <what the plan left unspecified and the assumption taken>   (or none)
    UNRESOLVED: <the ruling needed>         (or none)

A `GAP:` is a hole in the plan the unit could proceed past on a stated
assumption; it is recorded and the run continues. An `UNRESOLVED:` is a ruling
the unit could not proceed without; it blocks the unit and its dependents.

## Out of scope

- A repo tool that uses merge mode: none until a real need (the design
  rulings of 2026-11-04).
- The manifest and the stale rule: plan 5. This plan keeps the removal of
  today for a generated file whose condition fails (D7).
- The Mac side of `requires_found` (exit 2 there): plan 9.
- A `docs/sync/migrate.md` item (D6).
- `mise run setup:configs` in this repository and in each consumer (D4).

## Parked

See P16 Parked (`docs/plans/2026-11-04-docs-setup`).

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-04-repo-files-table

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
