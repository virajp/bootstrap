# U5 — Init options and prompts

- **Wave:** 3
- **Depends on:** U2
- **Owns:** `cli/src/commands/init/options.ts`,
  `cli/src/commands/init/prompts.ts`, `cli/test/commands/init/options.test.ts`,
  `cli/test/commands/init/prompts.test.ts`
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/commands/init/options.test.ts` and `prompts.test.ts`
  — every row of flow 110's mode table (interactive, interactive `-y`,
  non-interactive, non-interactive `-y`, `--dry-run`, `--json`) resolves values
  and tools as the table says; each invalid flag (`--merge-*` not `direct`|`pr`,
  `--scope` not kebab-case, `--repo` not two or more segments, an unknown tool,
  `core` or a core tool name in `--tool`) is a usage error naming the flag and
  the allowed form; a missing required value names the flag; `-y` with no repo
  default names `--repo`; duplicate `--tool` names are used once; interactive
  prompts repeat on an invalid answer, the scopes prompt splits one
  comma-separated answer (empty = zero scopes), the tool picker is a
  multi-select of the non-core tools, the confirm's "no" is a decline; Ctrl-C at
  a prompt is an interrupt. Fails because the module does not exist.
- **Read first:** flow 110 steps 3–5 and the mode table, `cli/src/repo/**`,
  `cli/src/tool/catalog.ts`, `cli/src/setup-config/schema.ts`.
- **Lazy-load:** `effect/cli` `Flag` and `Prompt` docs via Context7.

## Ruling

| # | Decision | Ruling                                                              | Rejected                                               | Unit |
| - | -------- | ------------------------------------------------------------------- | ------------------------------------------------------ | ---- |
| 3 | Prompts  | the `effect/cli` `Prompt` module (`Text`, `MultiSelect`, `Confirm`) | `@clack/prompts` (superseded by the Effect CLI choice) | U5   |

## Edits

1. **`cli/src/commands/init/options.ts`** — the `init` flags (`--repo`,
   `--scope` repeatable, `--merge-develop`, `--merge-main`, `--tool` repeatable,
   `-y/--yes`, `--dry-run`, `--json`, plus the global `--quiet`, `--verbose`,
   `--no-color`), their validation, and the mode-table resolution into setup
   values and the selected tools, with typed usage errors (exit 2) naming the
   flag.
2. **`cli/src/commands/init/prompts.ts`** — the value prompts, the scopes
   prompt, the tool picker and the confirm, using `Prompt`; used only when the
   mode table says so.
3. **`cli/test/commands/init/**`** — the test-first cases, prompts driven by a
   test terminal layer.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.

## Guardrails

- Do not touch `cli/src/cli.ts` (U6 wires the command).
- Do not re-implement git or origin logic; use `cli/src/repo/**`.

## Commit

`feat: resolve init values from flags and prompts`
