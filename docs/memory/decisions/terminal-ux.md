D:terminal-ux 2026-10-06 cli — Terminal UX decisions for docs/blueprint/design-system.md

# Terminal UX decisions for `cli` (2026-10-06)

Elicited ahead of `/vwf:design-system`, which halted at the import
(`docs/design/site/design-system.md` not authored yet — run
`/design-session site`). The next `/vwf:design-system` run writes these into
the doc's **Terminal UX** section verbatim in meaning.

- **Output formatting** — human-readable text by default; every command takes
  `--json`, which prints exactly one JSON document to stdout and nothing else.
  Results go to stdout; progress, warnings and errors go to stderr. `--quiet`
  shows errors only; `--verbose` shows each per-file decision.
- **Color semantics** — color carries meaning only: `success`
  (created / unchanged), `warning` (drift found, kept files), `error`, and
  `emphasis` (paths, command names). Mapped to the design system's semantic
  role names, never fixed hex — the terminal theme decides the shade. Color is
  off when `NO_COLOR` is set, when stdout/stderr is not a TTY, with
  `--no-color`, and always under `--json`.
- **Progress** — an animated spinner per step when interactive (TTY); when
  piped, in CI or non-TTY, one stable line per finished step on stderr instead.
- **Errors & exit codes** — `0` success, no drift; `1` drift found (`check`)
  or changes declined; `2` usage error (bad flag/argument — prints short
  usage); `3` failure (unreadable config, template error, write failed). Every
  error states what happened, why, and the exact next command; no stack trace
  unless `--verbose`.
- **Help & naming** — commands are single verbs: `bootstrap init`, `check`,
  `update`, `add <group>`. Every flag has a long form (`--dry-run`, `--json`);
  short forms only for the most common (`-q`, `-v`, `-y`). Booleans negate with
  `--no-<flag>`. Help: one-line purpose, usage, flags, then 1–2 real examples.
  `bootstrap` with no command prints help and exits `2`.
