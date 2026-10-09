# U3 — Git access, origin parser and preflight

- **Wave:** 3
- **Depends on:** U2
- **Owns:** `cli/src/git/**` (new), `cli/test/git/**` (new)
- **Model:** opus
- **Kind:** code
- **Test first:**
  - `cli/test/git/origin.test.ts` (pure): the step 3 table of flow 110 —
    `git@<host>:o/r.git` and `https://<host>/o/r.git` give repo `o/r` and the
    host; `https://<host>/a/b/c.git` gives `a/b/c`;
    `ssh://git@github.com/o/r.git` and `http://…` give the host and no repo; no
    `origin` or an unreadable URL gives neither. `.git` is optional.
  - `cli/test/git/preflight.test.ts` (in a temporary git repository): outside a
    git repository the preflight fails with exit 3 "not a git repository — run
    `git init` first"; from a subdirectory the root is the repository root; with
    no `mise` on `PATH` it fails with exit 3 "mise not installed" and the
    install command; with `mise` at a path other than `~/.local/bin/mise` it
    returns one warning naming that path.
  - `cli/test/git/status.test.ts` (in a temporary git repository): for a list of
    paths it reports each that is modified, staged, untracked, ignored, or
    tracked and deleted in the working tree without a committed deletion; a
    clean tracked path and an absent untracked path report nothing.
  - They fail because `cli/src/git/` does not exist.
- **Read first:** `docs/blueprint/flows/cli/110-setup-repository/index.md` (step
  1, step 3 table), `docs/blueprint/conventions.md` `#errors`, `#safety`,
  `cli/src/report/errors.ts`.
- **Lazy-load:** the Effect platform child-process docs via Context7.

## Ruling

| # | Decision                                                | Ruling                                                                                                                                                                                                          | Rejected                         | Unit   |
| - | ------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------- | ------ |
| 5 | Git access                                              | `git` runs as a child process through the Effect platform                                                                                                                                                       | A git library (a new dependency) | U3     |
| 7 | Plan constraint: catalog-driven code (user, 2026-10-08) | The literal scan of plan 1 (`cli/test/catalog-driven.test.ts`) extends to every `.ts` file under `cli/src` except the two catalogs and `cli/src/git/`. `git` and `mise` there are process names, not tool logic | An allow-list of literals        | U2, U3 |

## Edits

1. **`cli/src/git/origin.ts`** — the pure parser of the `origin` URL.
2. **`cli/src/git/git.ts`** — run `git` as a child process: the repository root,
   the `origin` URL, the status of a list of paths (porcelain status plus
   `check-ignore`).
3. **`cli/src/git/preflight.ts`** — the preflight per `conventions.md`
   `#errors`: git repository, `mise` on `PATH`, the warning for another `mise`
   path. It returns the root and the warnings.
4. **Tests** — the test-first cases.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:e2e` exits 0.

## Guardrails

- Do not touch `cli/src/report/errors.ts`; raise its errors.
- No `node:fs` and no `node:child_process`; use the Effect platform services.
- The names `git` and `mise` appear only under `cli/src/git/` (row 7).

## Commit

`feat: read git state, parse origin and run the preflight`
