# U4 — Docs

- **Read first:** every owned file that the list below names, at the named
  passages, before editing.
- **Lazy-load:** the diffs of U1 and U2 (`git log -p` of the branch), the
  `DOCS FALSIFIED:` lines that the earlier units returned.

## Ruling

D4: "The user, by hand, after the landing. It is out of scope, as in the
earlier plans."

D5: "Only in dev, like [tasks], when the tool is in dev. CI is not a machine
to provision."

D6: "When the tools.toml of a tool contains <%, the engine renders that file
as a gomplate template into a temp copy, and the generator reads the copy, so
the generated file holds the rendered value. The render context has .config
(the values of the tool), .all (the values of every selected tool) and .env
(HOME, USER), on both sides. The overrides of the values file are not
rendered. A <% %> expression sits inside a TOML basic string ("..."), and
every string literal inside the expression is a Go raw string in backticks. A
failed render (a required or test.Fail check) stops the mise config of that
tool with code 2 on the current failure path: the existing conf.d files of the
tool stay. A rendered copy that is not valid TOML gives the message <tools
root>/<tool>/tools.toml: not valid TOML after render: <message>, code 2, for
that tool."

D7: "One file conf.d/<tool>/mise.toml for each tool with content. It writes
every valid section, the dev-only sections included, because the Mac uses no
environments. A [mise_env.*] table, the envs key of a mise tool and the
requires_found key of a mise tool are tool errors: the files of that tool
stay, the other tools continue, and the run returns 2. An
[overrides.mise_env.*] table is a values-file error: the whole run stops with
code 2 before any write. [overrides.mise.*] works as on the repo side, an
override cannot add a key, and the comment is # override: macos.toml (the file
name only)."

## Edits

1. **The explicit list.** Find each passage by its text.
   - **`CLAUDE.md`** "mise config": the dev-only list adds `[bootstrap.*]`
     and `[dotfiles]` (D5); a `[mise.*]` string can hold `<% %>`, rendered
     with `.config`, `.all` and `.env`, each string literal inside the
     expression in backticks (D6); the generator is `sync-engine/mise`, with
     the sides `repo` and `macos` (D7, in one sentence; P15 moves the full
     rule to the design docs).
   - **`docs/design/mise.md`**: the same three points, with the D6 rule for
     quotes and its two messages, the macos side rules of D7, and the
     own-config entry for `zz-macos`.
   - **`docs/design/tools.md`**: the valid `[mise.*]` sections.
   - **`docs/sync/index.md`** and **`README.md`**: each list of the valid
     sections.
2. **`vwf:docs-sync`.** Run it over the full branch delta
   (`/Users/virajpatel/Projects/github.com/virajp/claude-plugins/.dev-marketplace/plugins/vwf/skills/docs-sync/SKILL.md`),
   and apply each finding and each `DOCS FALSIFIED:` line of the earlier
   units that the list does not cover, inside the owned files.
3. No `docs/sync/migrate.md` item: the generated config of every current
   tool does not change.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -l 'dotfiles' CLAUDE.md docs/design/mise.md` finds both files.
- `grep -l 'sync-engine/mise' CLAUDE.md docs/design/mise.md` finds both
  files.
- `mise x -- dprint check README.md CLAUDE.md docs` passes.

## Guardrails

- Touch only the owned files. Report a doc outside the list as a `GAP:`.
- Write identifiers in code spans.
- Delete with `rm`, never `git rm`.

## Commit

`docs: describe the shared mise generator` — written by the orchestrator
after the unit gate and the wave gate, not by the unit.
