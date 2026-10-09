# U6 — Change set and target checks

- **Wave:** 4
- **Depends on:** U3, U4
- **Owns:** `cli/src/init/changes.ts` (new), `cli/test/init/changes.test.ts`
  (new)
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/init/changes.test.ts` (in a temporary git
  repository) — from the rendered files, the resolution of U4, the recorded
  config and the disk:
  - A path that does not exist is `created`; a path whose content or mode
    differs is `changed`; an equal path is `unchanged`; an existing
    `create_only` file is `kept` and untouched; a recorded path no longer
    rendered and present on disk is `orphaned`; a path of a dropped or replaced
    tool is `deleted`, except its `create_only` files, which are `kept`; a
    replacement is listed as `{from, to}`.
  - A target that is modified, staged, untracked or ignored, and whose render
    differs, refuses the run with exit 1 listing the path and "commit or stash
    these files, then run again"; the same for a tracked target deleted without
    a committed deletion. A dirty target whose content and mode equal the render
    is `unchanged` and does not refuse. `.config/bootstrap.yaml` follows the
    same rule (row 9).
  - A target that is a directory or a symlink, or whose parent is a symlink or a
    regular file, fails with exit 3 naming the path; with both kinds present,
    the error has exit 3 and lists every path of both kinds.
  - A first run reports no `orphaned` path. Every path list is sorted.
  - The setup config's next content comes from plan 1's refresh of `files` (full
    mode) and its rewrite rule (init mode).
  - It fails because `changes.ts` does not exist.
- **Read first:** `docs/blueprint/flows/cli/110-setup-repository/index.md`
  (steps 6–8), `docs/blueprint/conventions.md` `#safety`,
  `docs/blueprint/entities/setup-config/index.md` (invariants 1–7),
  `cli/src/git/git.ts`, `cli/src/init/resolve.ts`, `cli/src/report/errors.ts`.
- **Lazy-load:** `cli/src/tool/render.ts`, `cli/src/setup-config/**`.

## Ruling

| # | Decision                                | Ruling                                                                                                                                                              | Rejected | Unit   |
| - | --------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------ |
| 9 | Plan 1 row 9 (setup-config invariant 1) | Lands here: a dirty target, `.config/bootstrap.yaml` included, refuses the run only when its render differs from the working copy; the setup config is written last | —        | U5, U6 |

## Edits

1. **`cli/src/init/changes.ts`** — compute the change set in memory: compare
   each render with the disk (content and mode), apply the git status of U3,
   classify every path, and return the change set for U5 and the result lists
   for U7, or the declined (exit 1) or failure (exit 3) error.
2. **`cli/test/init/changes.test.ts`** — the test-first cases.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:e2e` exits 0.

## Guardrails

- Writes nothing; reads the disk and git only.
- Reads `create_only` from the catalog data; no tool name literal (row 7 of
  `index.md`).

## Commit

`feat: compute the init change set and refuse unsafe targets`
