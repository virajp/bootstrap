# U2 — Command surface of init, usage errors and the error model

- **Wave:** 2
- **Depends on:** U1
- **Owns:** `cli/src/cli.ts`, `cli/src/bin.ts`, `cli/src/init/command.ts` (new),
  `cli/src/init/usage.ts` (new), `cli/src/report/errors.ts` (new),
  `cli/test/cli.test.ts`, `cli/test/init/usage.test.ts` (new),
  `cli/test/report/errors.test.ts` (new), `cli/test/catalog-driven.test.ts`
- **Model:** opus
- **Kind:** code
- **Test first:**
  - `cli/test/init/usage.test.ts` (in-process, through `run(args)`): each usage
    error of flow 110 (the list before step 1) exits 2, writes nothing and
    prints a message that names the flag and the allowed form — an unknown flag;
    `--merge-develop squash`; `--merge-main x`; `--add-scope Api`;
    `--remove-scope a--b`; `--repo onlyone`; `--reset-scope --add-scope api`;
    `--reset-scope --remove-scope api`; `--add-scope api --remove-scope api`;
    `--tool nosuch`; two `--tool` names of one `max: one` category (taken from
    the catalog data, not a literal pair). All usage errors run before any git
    or disk access: the test runs them in a directory that is not a git
    repository.
  - `cli/test/cli.test.ts`: `init --help` exits 0 and prints the init help;
    `--help` wins over other input; `init --version` and `init -v --version`
    exit 2; `init -v` parses as `--verbose`; `bootstrap --version` exits 0 and
    prints the version; bare `bootstrap` still exits 2 with the top-level help.
  - `cli/test/report/errors.test.ts`: each error kind carries its exit code
    (declined 1, usage 2, failure 3, interrupted 130), a what, a why and a next
    command; under `--json` the error is one JSON document on stdout with `exit`
    and `error` (with `next_command` when the error has one), and
    `dry_run: true` when `--dry-run` is given; without `--json`, the error goes
    to stderr as `✗ error:` text.
  - `cli/test/catalog-driven.test.ts`: the scan covers every `.ts` file under
    `cli/src` except `cli/src/tool/catalog.ts`,
    `cli/src/tool-category/catalog.ts` and `cli/src/git/**`.
  - They fail because `init` is not registered and no error model exists.
- **Read first:** `docs/blueprint/flows/cli/110-setup-repository/index.md`
  (usage errors, step 3 flags), `docs/blueprint/conventions.md` `#errors`,
  `docs/blueprint/design-system.md` `#terminal-ux` (flags, help, errors), every
  owned file.
- **Lazy-load:** the `effect/cli` docs via Context7 (flags, subcommands, parse
  errors, help); `cli/src/tool/catalog.ts`, `cli/src/tool-category/catalog.ts`.

## Ruling

| #  | Decision                                                | Ruling                                                                                                                                                                                                          | Rejected                   | Unit   |
| -- | ------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------- | ------ |
| 3  | G2: effect/cli exits 1 on an unknown flag               | Map the framework's parse errors to a usage error, exit 2 with the short usage, at the command boundary                                                                                                         | An own argument pre-parser | U2     |
| 4  | G3: effect/cli gives `-v` to `--version`                | `-v` is `--verbose` on each command; `--version` only at the top level, with no short form                                                                                                                      | Keep the framework default | U2     |
| 7  | Plan constraint: catalog-driven code (user, 2026-10-08) | The literal scan of plan 1 (`cli/test/catalog-driven.test.ts`) extends to every `.ts` file under `cli/src` except the two catalogs and `cli/src/git/`. `git` and `mise` there are process names, not tool logic | An allow-list of literals  | U2, U3 |
| 10 | Command surface before the pipeline                     | U2 registers `init`, its flags and usage checks; its handler ends in a placeholder that U8 replaces with the pipeline before the review                                                                         | One unit for both          | U2, U8 |

## Edits

1. **`cli/src/report/errors.ts`** — the error model: one tagged error type per
   kind with its exit code (`Runtime.errorExitCode`, as `NoCommand` does), the
   what, the why and an optional next command; one function that prints an error
   as human text on stderr or as the JSON error document on stdout per
   `conventions.md` `#errors`. Later units raise these errors and never edit
   this file.
2. **`cli/src/init/command.ts`** — the `init` subcommand with every flag of the
   facts list, long forms, `-y`, `-q`, `-v`; repeatable `--tool`, `--add-scope`,
   `--remove-scope`. The handler runs `usage.ts`, then a placeholder that exits
   3 "not implemented" (row 10).
3. **`cli/src/init/usage.ts`** — the usage checks as a pure function over the
   parsed flags and the catalog data. All errors found are reported together.
4. **`cli/src/cli.ts`** — register `init`; map parse errors to the usage error
   (row 3); `--version` only at the root, `-v` free for `--verbose` (row 4).
5. **`cli/src/bin.ts`** — only the edits that rows 3 and 4 need at the process
   boundary.
6. **Tests** — the test-first cases; adapt `cli/test/cli.test.ts`.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:build` exits 0.
- `mise x -- mise run p:cli:e2e` exits 0.

## Guardrails

- Do not touch `cli/src/tool/**`, `cli/src/tool-category/**`,
  `cli/src/setup-config/**` or `cli/templates/**` (plan 1).
- No git and no disk access before the usage checks pass.
- Copy message texts from the flow and the conventions byte for byte.

## Commit

`feat: add the init command surface with usage errors and exit codes`
